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

// EventParticipant holds the schema definition for the EventParticipant entity.
type EventParticipant struct {
	ent.Schema
}

// Mixin of the EventParticipant.
func (EventParticipant) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.SoftDeleteMixin{},
		mixin.TimeMixin{},
	}
}

// Fields of the EventParticipant.
func (EventParticipant) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.UUID("event_id", uuid.UUID{}),

		field.UUID("user_id", uuid.UUID{}),

		field.Enum("attendance_method").
			Values("not_visited", "user_qr", "event_qr").
			Default("not_visited").
			Comment("Посещено ли мероприятие и с помощью чего оно было посещено"),

		field.Time("attended_at").
			Optional().
			Comment("Время посещения мероприятия"),
	}
}

// Edges of the EventParticipant.
func (EventParticipant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("event", Event.Type).
			Required().
			Unique().
			Field("event_id"),
		// Удаление управляется через soft delete hook

		edge.To("user", User.Type).
			Required().
			Unique().
			Field("user_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the EventParticipant.
func (EventParticipant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "event_id").
			Unique(),
	}
}
