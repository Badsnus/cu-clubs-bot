package mixin

import (
	"context"
	"fmt"
	"log"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"

	gen "github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/secondary/postgres/ent"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/secondary/postgres/ent/club"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/secondary/postgres/ent/event"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/secondary/postgres/ent/hook"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/secondary/postgres/ent/intercept"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/secondary/postgres/ent/pass"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/valueobject"
)

// SoftDeleteMixin implements the soft delete pattern for schemas.
type SoftDeleteMixin struct {
	mixin.Schema
}

// Fields of the SoftDeleteMixin.
func (SoftDeleteMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("deleted_at").
			Optional(),
	}
}

// Indexes of the SoftDeleteMixin.
func (SoftDeleteMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("deleted_at").
			Annotations(entsql.IndexAnnotation{
				Where: "deleted_at IS NULL",
			}),
	}
}

type softDeleteKey struct{}

// SkipSoftDelete returns a new context that skips the soft-delete interceptor/mutators.
func SkipSoftDelete(parent context.Context) context.Context {
	return context.WithValue(parent, softDeleteKey{}, true)
}

// Interceptors of the SoftDeleteMixin.
func (d SoftDeleteMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			// Skip soft-delete, means include soft-deleted entities.
			if skip, _ := ctx.Value(softDeleteKey{}).(bool); skip {
				return nil
			}
			d.P(q)
			return nil
		}),
	}
}

// Hooks of the SoftDeleteMixin.
func (d SoftDeleteMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(
			func(next ent.Mutator) ent.Mutator {
				return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
					// Skip soft-delete, means delete the entity permanently.
					if skip, _ := ctx.Value(softDeleteKey{}).(bool); skip {
						return next.Mutate(ctx, m)
					}

					mx, ok := m.(interface {
						SetOp(ent.Op)
						Client() *gen.Client
						SetDeletedAt(time.Time)
						WhereP(...func(*sql.Selector))
					})
					if !ok {
						return nil, fmt.Errorf("unexpected mutation type %T", m)
					}

					// ========================================
					// ФАЗА 1: Каскадное удаление зависимых записей
					// ========================================

					// Обработка удаления Club
					if clubMx, ok := m.(*gen.ClubMutation); ok {
						if err := cascadeDeleteClubDependencies(ctx, clubMx); err != nil {
							return nil, fmt.Errorf("failed to cascade delete club dependencies: %w", err)
						}
					}

					// Обработка удаления Event
					if eventMx, ok := m.(*gen.EventMutation); ok {
						if err := cascadeDeleteEventDependencies(ctx, eventMx); err != nil {
							return nil, fmt.Errorf("failed to cascade delete event dependencies: %w", err)
						}
					}

					// ========================================
					// ФАЗА 2: Soft Delete родительской записи
					// ========================================
					d.P(mx)
					mx.SetOp(ent.OpUpdate)
					mx.SetDeletedAt(time.Now())
					return mx.Client().Mutate(ctx, m)
				})
			},
			ent.OpDeleteOne|ent.OpDelete,
		),
	}
}

// cascadeDeleteClubDependencies удаляет все зависимые записи при удалении Club
func cascadeDeleteClubDependencies(ctx context.Context, m *gen.ClubMutation) error {
	clubID, exists := m.ID()
	if !exists {
		return nil
	}

	client := m.Client()

	log.Printf("[SoftDelete] Cascading delete for Club ID: %s", clubID)

	// 1. Soft delete всех событий клуба
	affectedEvents, err := client.Event.Update().
		Where(
			event.HasClubWith(club.ID(clubID)),
			event.DeletedAtIsNil(),
		).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to soft delete events: %w", err)
	}
	log.Printf("[SoftDelete] Soft deleted %d events for club %s", affectedEvents, clubID)

	// 2. Soft delete ClubOwner записи
	affectedOwners, err := client.ClubOwner.Update().
		Where(func(s *sql.Selector) {
			s.Where(sql.EQ("club_id", clubID))
			s.Where(sql.IsNull("deleted_at"))
		}).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to soft delete club owners: %w", err)
	}
	log.Printf("[SoftDelete] Soft deleted %d club owners for club %s", affectedOwners, clubID)

	// 3. Soft delete IgnoreMailing записи
	affectedIgnore, err := client.IgnoreMailing.Update().
		Where(func(s *sql.Selector) {
			s.Where(sql.EQ("club_id", clubID))
			s.Where(sql.IsNull("deleted_at"))
		}).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to soft delete ignore mailings: %w", err)
	}
	log.Printf("[SoftDelete] Soft deleted %d ignore mailings for club %s", affectedIgnore, clubID)

	return nil
}

// cascadeDeleteEventDependencies удаляет все зависимые записи при удалении Event
func cascadeDeleteEventDependencies(ctx context.Context, m *gen.EventMutation) error {
	eventID, exists := m.ID()
	if !exists {
		return nil
	}

	client := m.Client()

	log.Printf("[SoftDelete] Cascading delete for Event ID: %s", eventID)

	// 1. Soft delete EventParticipant записи
	affectedParticipants, err := client.EventParticipant.Update().
		Where(func(s *sql.Selector) {
			s.Where(sql.EQ("event_id", eventID))
			s.Where(sql.IsNull("deleted_at"))
		}).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to soft delete event participants: %w", err)
	}
	log.Printf("[SoftDelete] Soft deleted %d participants for event %s", affectedParticipants, eventID)

	// 2. Отменить активные пропуска (не удаляем, а меняем статус)
	affectedPasses, err := client.Pass.Update().
		Where(
			pass.HasEventWith(event.ID(eventID)),
			pass.StatusEQ(valueobject.PassStatusPending),
		).
		SetStatus(valueobject.PassStatusCancelled).
		SetNotes("Event deleted").
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to cancel passes: %w", err)
	}
	log.Printf("[SoftDelete] Cancelled %d passes for event %s", affectedPasses, eventID)

	// 3. Soft delete Notification записи
	affectedNotifications, err := client.Notification.Update().
		Where(func(s *sql.Selector) {
			s.Where(sql.EQ("event_id", eventID))
			s.Where(sql.IsNull("deleted_at"))
		}).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to soft delete notifications: %w", err)
	}
	log.Printf("[SoftDelete] Soft deleted %d notifications for event %s", affectedNotifications, eventID)

	return nil
}

// P adds a storage-level predicate to the queries and mutations.
func (d SoftDeleteMixin) P(w interface{ WhereP(...func(*sql.Selector)) }) {
	w.WhereP(
		sql.FieldIsNull(d.Fields()[0].Descriptor().Name),
	)
}
