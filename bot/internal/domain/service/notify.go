package service

import (
	"strings"

	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
	"go.uber.org/zap/zapcore"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type NotifyService struct {
	bot    *tele.Bot
	layout *layout.Layout
	logger *types.Logger
}

func NewNotifyService(bot *tele.Bot, layout *layout.Layout, logger *types.Logger) *NotifyService {
	return &NotifyService{
		bot:    bot,
		layout: layout,
		logger: logger,
	}
}

// LogHook returns a log hook for the specified channel
//
// Parameters:
//   - channelID is the channel to send the log to
//   - locale is the locale to use for the layout
//   - level is the minimum log level to send
func (s *NotifyService) LogHook(channelID int64, locale string, level zapcore.Level) (types.LogHook, error) {
	chat, err := s.bot.ChatByID(channelID)
	if err != nil {
		return nil, err
	}
	return func(log types.Log) {
		if log.Level >= level {
			_, err = s.bot.Send(chat, s.layout.TextLocale(locale, "log", log))
			if err != nil && !strings.Contains(log.Message, "failed to send log to channel") {
				s.logger.Errorf("failed to send log to channel %d: %v\n", channelID, err)
			}
		}
	}, nil
}
