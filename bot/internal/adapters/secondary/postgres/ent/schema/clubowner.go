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

// ClubOwner holds the schema definition for the ClubOwner entity.
type ClubOwner struct {
	ent.Schema
}

// Mixin of the ClubOwner.
func (ClubOwner) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.SoftDeleteMixin{},
		mixin.TimeMixin{},
	}
}

// Fields of the ClubOwner.
func (ClubOwner) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.UUID("user_id", uuid.UUID{}),

		field.UUID("club_id", uuid.UUID{}),

		field.Bool("warnings").
			Default(false).
			Comment("Отправлять ClubOwner-у варнинги"),
	}
}

// Edges of the ClubOwner.
func (ClubOwner) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Required().
			Unique().
			Field("user_id").
			Annotations(entsql.OnDelete(entsql.Cascade)), // Hard delete при удалении User

		edge.To("club", Club.Type).
			Required().
			Unique().
			Field("club_id"),
	}
}

// Indexes of the ClubOwner.
func (ClubOwner) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "club_id").
			Unique(),
	}
}
