package entity

import (
	"fmt"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/location"
)

type Event struct {
	ID                    string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             gorm.DeletedAt
	ClubID                string `gorm:"not null;type:uuid"`
	Name                  string `gorm:"not null"`
	Description           string `gorm:"not null"`
	AfterRegistrationText string
	Location              string    `gorm:"not null"`
	StartTime             time.Time `gorm:"not null"`
	EndTime               time.Time
	RegistrationEnd       time.Time `gorm:"not null"`
	MaxParticipants       int
	ExpectedParticipants  int
	QRCodeID              string
	QRFileID              string
	AllowedRoles          pq.StringArray `gorm:"type:text[]"`
	PassRequired          bool           `gorm:"default:false"`
}

// IsOver checks if the event is over, considering the additional time
// if additionalTime is 0, the function just checks if the event has started
// if additionalTime is positive, the function checks if the event has started
// and if the event has ended before the current time minus additionalTime
// if additionalTime is negative, the function checks if the event has started
// and if the event has ended after the current time plus additionalTime
func (e *Event) IsOver(additionalTime time.Duration) bool {
	return e.StartTime.Before(time.Now().In(location.Location()).Add(-additionalTime))
}

// Link generates a link to the event in the bot
//
// The link is in the format https://t.me/<botName>?start=event_<eventID>
//
// The link can be used to open the event in the bot
func (e *Event) Link(botName string) string {
	return fmt.Sprintf("https://t.me/%s?start=event_%s", botName, e.ID)
}
