package entity

import (
	"fmt"
	"slices"
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

// IsPassRequiredForUser checks if a pass is required for the given user
func (e *Event) IsPassRequiredForUser(user *User, excludedRoles []string) bool {
	if !e.PassRequired {
		return false
	}

	if slices.Contains(excludedRoles, user.Role.String()) {
		return false
	}

	return true
}

// CalculateScheduledAt calculates the scheduled time for pass sending based on cron schedule
func (e *Event) CalculateScheduledAt() time.Time {
	loc := location.Location()
	st := e.StartTime.In(loc)

	// Determine timeBeforeEvent based on event day
	var timeBeforeEvent time.Duration
	dow := st.Weekday()
	if dow >= time.Monday && dow <= time.Friday {
		timeBeforeEvent = 24 * time.Hour
	} else {
		timeBeforeEvent = 48 * time.Hour
	}

	// Calculate send day
	sendDay := st.Add(-timeBeforeEvent)

	// Determine send time based on send day
	sendDow := sendDay.Weekday()
	var sendHour int
	if sendDow >= time.Monday && sendDow <= time.Friday {
		sendHour = 16 // Weekdays 16:00
	} else {
		sendHour = 12 // Weekends 12:00
	}

	// Return the send time
	return time.Date(sendDay.Year(), sendDay.Month(), sendDay.Day(), sendHour, 0, 0, 0, loc)
}
