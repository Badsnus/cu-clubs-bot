package dto

import (
	"time"

	"github.com/lib/pq"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/location"
)

type UserEvent struct {
	ID                    string
	ClubID                string
	Name                  string
	Description           string
	AfterRegistrationText string
	Location              string
	StartTime             time.Time
	EndTime               time.Time
	RegistrationEnd       time.Time
	MaxParticipants       int
	ExpectedParticipants  int
	AllowedRoles          pq.StringArray
	IsVisited             bool
}

func NewUserEventFromEntity(event entity.Event, isVisited bool) UserEvent {
	return UserEvent{
		ID:                    event.ID,
		ClubID:                event.ClubID,
		Name:                  event.Name,
		Description:           event.Description,
		AfterRegistrationText: event.AfterRegistrationText,
		Location:              event.Location,
		StartTime:             event.StartTime,
		EndTime:               event.EndTime,
		RegistrationEnd:       event.RegistrationEnd,
		MaxParticipants:       event.MaxParticipants,
		ExpectedParticipants:  event.ExpectedParticipants,
		AllowedRoles:          event.AllowedRoles,
		IsVisited:             isVisited,
	}
}

func (e *UserEvent) IsOver(additionalTime time.Duration) bool {
	return e.StartTime.Before(time.Now().In(location.Location()).Add(-additionalTime))
}
