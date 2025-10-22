package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/secondary/postgres/ent/schema/mixin"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/valueobject"
)

// Pass holds the schema definition for the Pass entity.
type Pass struct {
	ent.Schema
}

// Mixin of the Pass.
func (Pass) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
	}
}

// Fields of the Pass.
func (Pass) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.UUID("event_id", uuid.UUID{}),

		field.UUID("user_id", uuid.UUID{}),

		field.Enum("type").
			GoType(valueobject.PassType("")).
			Default(string(valueobject.PassTypeEvent)).
			Comment("Тип пропуска (event, manual, api)"),

		field.Enum("status").
			GoType(valueobject.PassStatus("")).
			Default(string(valueobject.PassStatusPending)).
			Comment("Статус пропуска (pending, sent, cancel)"),

		field.Enum("requester_type").
			GoType(valueobject.RequesterType("")).
			Default(string(valueobject.RequesterTypeUser)).
			Comment("Тип запросившего пропуск"),

		field.UUID("requester_id", uuid.UUID{}).
			Optional().
			Comment("ID запросившего пропуск"),

		field.Time("scheduled_at").
			Comment("Когда запланирована отправка пропуска"),

		field.Time("sent_at").
			Optional().
			Comment("Когда был отправлен пропуск"),

		field.String("notes").
			Optional().
			Comment("Дополнительная информация по пропуску"),

		field.Bool("email_sent").
			Default(false).
			Comment("Пропуск отправлен на почту"),

		field.Bool("telegram_sent").
			Default(false).
			Comment("Пропуск отправлен в Telegram"),
	}
}

// Edges of the Pass.
func (Pass) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("event", Event.Type).
			Field("event_id").
			Required().
			Unique(),

		edge.To("user", User.Type).
			Field("user_id").
			Required().
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the Pass.
func (Pass) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("event_id", "user_id", "status").
			Unique().
			Annotations(entsql.IndexAnnotation{
				Where: "status != 'cancelled'",
			}),

		index.Fields("status", "scheduled_at"),
	}
}
