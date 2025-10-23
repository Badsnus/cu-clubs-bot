package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/secondary/postgres/ent/schema/mixin"
)

// Notification holds the schema definition for the Notification entity.
type Notification struct {
	ent.Schema
}

// Mixin of the Notification.
func (Notification) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.SoftDeleteMixin{},
		mixin.TimeMixin{},
	}
}

// Fields of the Notification.
func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.UUID("event_id", uuid.UUID{}),

		field.UUID("user_id", uuid.UUID{}),

		field.Enum("type").
			Values("day", "hour"),
	}
}

// Edges of the Notification.
func (Notification) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("event", Event.Type).
			Field("event_id").
			Required().
			Unique(),
		// Удаление управляется через soft delete hook

		edge.To("user", User.Type).
			Field("user_id").
			Required().
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the Notification.
func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "event_id").
			Unique(),
	}
}
