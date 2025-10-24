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
//
// Правила отправки:
// - События в Воскресенье или Понедельник → отправка в предыдущую Субботу 12:00
// - События во Вторник-Пятницу → отправка в предыдущий день 16:00
// - События в Субботу → отправка в Пятницу 16:00
func (e *Event) CalculateScheduledAt() time.Time {
	loc := location.Location()
	eventStart := e.StartTime.In(loc)
	eventWeekday := eventStart.Weekday()

	var scheduledAt time.Time

	switch eventWeekday {
	case time.Sunday:
		// Воскресенье → отправка в Субботу 12:00
		daysToSubtract := 1
		sendDay := eventStart.AddDate(0, 0, -daysToSubtract)
		scheduledAt = time.Date(sendDay.Year(), sendDay.Month(), sendDay.Day(), 12, 0, 0, 0, loc)

	case time.Monday:
		// Понедельник → отправка в Субботу 12:00
		daysToSubtract := 2
		sendDay := eventStart.AddDate(0, 0, -daysToSubtract)
		scheduledAt = time.Date(sendDay.Year(), sendDay.Month(), sendDay.Day(), 12, 0, 0, 0, loc)

	case time.Saturday:
		// Суббота → отправка в Пятницу 16:00
		daysToSubtract := 1
		sendDay := eventStart.AddDate(0, 0, -daysToSubtract)
		scheduledAt = time.Date(sendDay.Year(), sendDay.Month(), sendDay.Day(), 16, 0, 0, 0, loc)

	default:
		// Вторник-Пятница → отправка в предыдущий день 16:00
		daysToSubtract := 1
		sendDay := eventStart.AddDate(0, 0, -daysToSubtract)
		scheduledAt = time.Date(sendDay.Year(), sendDay.Month(), sendDay.Day(), 16, 0, 0, 0, loc)
	}

	return scheduledAt
}
