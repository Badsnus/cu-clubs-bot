package schema

import (
	"fmt"
	"net/url"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/secondary/postgres/ent/schema/mixin"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/valueobject"
)

// Club holds the schema definition for the Club entity.
type Club struct {
	ent.Schema
}

// Mixin of the Club.
func (Club) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.SoftDeleteMixin{},
		mixin.TimeMixin{},
	}
}

// Fields of the Club.
func (Club) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.String("name").
			Unique().
			Comment("Название клуба"),

		field.String("description").
			Optional().
			Comment("Описание клуба"),

		field.String("link").
			Optional().
			Validate(func(s string) error {
				if s == "" {
					return nil
				}
				u, err := url.Parse(s)
				if err != nil || !u.IsAbs() {
					return fmt.Errorf("invalid URL format")
				}
				return nil
			}).
			Comment("Ссылка в профиле клуба (Например ссылка на канал/чат)"),

		field.String("avatar_id").
			Optional().
			Comment("id изображения которое отображается в профиле клуба"),

		field.String("intro_id").
			Optional().
			Comment("id кружка клуба который отправлятся перед профилем"),

		field.Bool("visible_in_tour").
			Default(false).
			Comment("Виден ли клуб в туре"),

		field.JSON("allowed_roles", valueobject.Roles{}).
			Optional().
			Comment("Роли для которых клуб может создавать Event"),

		field.Bool("qr_allowed").
			Default(false).
			Comment("Разрешено ли использование qr-кода мероприятия"),
	}
}

// Edges of the Club.
func (Club) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("events", Event.Type),

		edge.From("owners", User.Type).
			Ref("owned_clubs").
			Through("club_owners", ClubOwner.Type),

		edge.To("ignore_mailings", IgnoreMailing.Type),
	}
}

// Indexes of the Club.
func (Club) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("visible_in_tour"),

		index.Fields("allowed_roles").
			Annotations(entsql.IndexAnnotation{
				Type: "GIN",
			}),
	}
}
