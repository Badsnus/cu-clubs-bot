package config

import (
	"github.com/spf13/viper"
)

type BotConfig interface {
	Token() string
	PassChannelID() int64
	QRChannelID() int64
}

type botConfig struct {
	token         string
	passChannelID int64
	qrChannelID   int64
}

func NewBotConfig() BotConfig {
	return &botConfig{
		token:         viper.GetString("bot.token"),
		passChannelID: viper.GetInt64("settings.pass.channel-id"),
		qrChannelID:   viper.GetInt64("bot.qr.channel-id"),
	}
}

func (cfg *botConfig) Token() string {
	return cfg.token
}

func (cfg *botConfig) PassChannelID() int64 {
	return cfg.passChannelID
}

func (cfg *botConfig) QRChannelID() int64 {
	return cfg.qrChannelID
}
