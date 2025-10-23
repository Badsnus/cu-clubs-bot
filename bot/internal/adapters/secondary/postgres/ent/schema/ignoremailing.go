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

// IgnoreMailing holds the schema definition for the IgnoreMailing entity.
type IgnoreMailing struct {
	ent.Schema
}

// Mixin of the IgnoreMailing.
func (IgnoreMailing) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.SoftDeleteMixin{},
		mixin.TimeMixin{},
	}
}

// Fields of the IgnoreMailing.
func (IgnoreMailing) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.UUID("user_id", uuid.UUID{}),

		field.UUID("club_id", uuid.UUID{}),
	}
}

// Edges of the IgnoreMailing.
func (IgnoreMailing) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("ignore_mailings").
			Unique().
			Required().
			Field("user_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),

		edge.From("club", Club.Type).
			Ref("ignore_mailings").
			Unique().
			Required().
			Field("club_id"),
		// Удаление управляется через soft delete hook
	}
}

// Indexes of the IgnoreMailing.
func (IgnoreMailing) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "club_id").
			Unique(),
	}
}
