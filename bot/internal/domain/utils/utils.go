package utils

import (
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/location"
	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v3"
	"slices"
	"time"
)

func IsAdmin(userID int64) bool {
	return slices.Contains(viper.GetIntSlice("bot.admin-ids"), int(userID))
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

func GetMaxRegisteredEndTime(startTimeStr string) string {
	const layout = "02.01.2006 15:04"

	startTime, _ := time.ParseInLocation(layout, startTimeStr, location.Location())
	return time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location()).Add(-24 * time.Hour).Add(16 * time.Hour).Format(layout)
}
