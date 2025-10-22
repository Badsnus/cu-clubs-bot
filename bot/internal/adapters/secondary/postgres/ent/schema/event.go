package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/secondary/postgres/ent/schema/mixin"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/valueobject"
)

// Event holds the schema definition for the Event entity.
type Event struct {
	ent.Schema
}

// Mixin of the Event.
func (Event) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.SoftDeleteMixin{},
		mixin.TimeMixin{},
	}
}

// Fields of the Event.
func (Event) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.UUID("club_id", uuid.UUID{}),

		field.String("name").
			Comment("Название мероприятия"),

		field.String("description").
			Optional().
			Comment("Описание мероприятия"),

		field.String("after_registration_text").
			Optional().
			Comment("Текст который отображается после регистрации пользователя на мероприятие"),

		field.String("location").
			Comment("Локация мероприятия"),

		field.Time("start_time").
			Comment("Время начала мероприятия"),

		field.Time("end_time").
			Optional().
			Comment("Время завершения мероприятия"),

		field.Time("registration_end_time").
			Comment("Время завершения регистрации на мероприятие"),

		field.Int("max_participants").
			Optional().
			NonNegative().
			Comment("Максимальное количество участников на мероприятии"),

		field.Int("expected_participants").
			Optional().
			NonNegative().
			Comment("Ожидаемое количество участников мероприятия (по достижении отпраялется варнинг)"),

		field.UUID("qr_payload", uuid.UUID{}).
			Optional().
			Unique().
			Comment("Payload QR-кода"),

		field.String("qr_file_id").
			Optional().
			Comment("File ID сгенерированного qr-кода"),

		field.JSON("allowed_roles", valueobject.Roles{}).
			Optional().
			Comment("Роли которым доступно мероприятие"),

		field.Bool("pass_required").
			Default(false).
			Comment("Нужен ли пропуск для посещения мероприятия"),
	}
}

// Edges of the Event.
func (Event) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("club", Club.Type).
			Ref("events").
			Field("club_id").
			Required().
			Unique(),

		edge.From("participants", User.Type).
			Ref("events").
			Through("event_participants", EventParticipant.Type),

		edge.From("notifications", Notification.Type).
			Ref("event"),

		edge.From("passes", Pass.Type).
			Ref("event"),
	}
}

// Indexes of the Event.
func (Event) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("qr_payload").
			Unique(),

		index.Fields("club_id", "start_time"),

		index.Fields("allowed_roles").
			Annotations(entsql.IndexAnnotation{
				Type: "GIN",
			}),

		index.Fields("deleted_at").
			Annotations(entsql.IndexAnnotation{
				Where: "deleted_at IS NULL",
			}),

		index.Fields("registration_end_time"),
	}
}

// Annotations of the Event.
func (Event) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"end_time_after_start": "end_time IS NULL OR end_time > start_time",
			"reg_before_start":     "registration_end_time <= start_time",
		}),
	}
}
