package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/secondary/postgres/ent/schema/mixin"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/valueobject"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Mixin of the User.
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.Int64("telegram_id").
			Positive().
			Unique(),

		field.Enum("localization").
			Values("ru").
			Default("ru").
			Comment("Локализация пользователя"),

		field.String("username").
			Optional().
			Comment("Telegram username пользователя"),

		field.Enum("role").
			GoType(valueobject.Role("")),

		field.String("email").
			GoType(valueobject.Email("")).
			Optional().
			Unique(),

		field.String("fio").
			GoType(valueobject.FIO{}),

		field.UUID("qr_payload", uuid.UUID{}).
			Optional().
			Unique().
			Comment("Payload QR-кода"),

		field.String("qr_file_id").
			Optional().
			Comment("File ID сгенерированного qr-кода"),

		field.Bool("is_banned").
			Default(false),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("owned_clubs", Club.Type).
			Through("club_owners", ClubOwner.Type),

		edge.To("events", Event.Type).
			Through("event_participants", EventParticipant.Type),

		edge.To("ignore_mailings", IgnoreMailing.Type),

		edge.From("notifications", Notification.Type).
			Ref("user"),

		edge.From("passes", Pass.Type).
			Ref("user"),
	}
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_banned"),

		index.Fields("role"),
	}
}
