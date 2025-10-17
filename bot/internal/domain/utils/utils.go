package utils

import (
	"slices"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/location"
)

func IsAdmin(userID int64, adminIDs []int64) bool {
	return slices.Contains(adminIDs, userID)
}

func ChangeMessageText(msg *tele.Message, text string) interface{} {
	if msg.Photo != nil {
		msg.Photo.Caption = text
		return msg.Photo
	}
	if msg.Video != nil {
		msg.Video.Caption = text
		return msg.Video
	}
	if msg.Audio != nil {
		msg.Audio.Caption = text
		return msg.Audio
	}
	if msg.Document != nil {
		msg.Document.Caption = text
		return msg.Document
	}
	return text
}

func GetMessageText(msg *tele.Message) string {
	switch {
	case msg.Text != "":
		return msg.Text
	case msg.Caption != "":
		return msg.Caption
	default:
		return ""
	}
}

func GetMaxRegisteredEndTime(startTime time.Time) time.Time {
	if startTime.Weekday() == time.Sunday {
		return time.Date(startTime.Year(), startTime.Month(), startTime.Day()-1, 12, 0, 0, 0, location.Location())
	} else if startTime.Weekday() == time.Monday {
		return time.Date(startTime.Year(), startTime.Month(), startTime.Day()-2, 12, 0, 0, 0, location.Location())
	}

	return time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location()).Add(-24 * time.Hour).Add(16 * time.Hour)
}
