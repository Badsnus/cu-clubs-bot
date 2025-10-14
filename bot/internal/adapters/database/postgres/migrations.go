package postgres

import "github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"

// Migrations is a list of all gorm migrations for the database.
var Migrations = []interface{}{
	&entity.User{},
	&entity.Club{},
	&entity.ClubOwner{},
	&entity.IgnoreMailing{},
	&entity.Event{},
	&entity.EventParticipant{},
	&entity.EventNotification{},
	&entity.Pass{},
}
