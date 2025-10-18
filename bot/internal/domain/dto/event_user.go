package dto

import (
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

type EventUser struct {
	User      entity.User
	UserVisit bool
}

func NewEventUserFromEntity(user entity.User, userVisit bool) EventUser {
	return EventUser{
		User:      user,
		UserVisit: userVisit,
	}
}
