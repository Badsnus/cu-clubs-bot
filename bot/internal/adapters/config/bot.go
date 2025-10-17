package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type BotConfig interface {
	Token() string
	AdminIDs() []int64
	GrantChatID() int64
	MailingChannelID() int64
	AvatarChannelID() int64
	IntroChannelID() int64
	PassChannelID() int64
	QRChannelID() int64
	ValidEmailDomains() []string
}

type botConfig struct {
	token             string
	adminIDs          []int64
	grantChatID       int64
	mailingChannelID  int64
	avatarChannelID   int64
	introChannelID    int64
	passChannelID     int64
	qrChannelID       int64
	validEmailDomains []string
}

func NewBotConfig() BotConfig {
	admins := viper.GetIntSlice("bot.admin-ids")
	adminIDs := make([]int64, len(admins))
	for i, v := range admins {
		adminIDs[i] = int64(v)
	}
	return &botConfig{
		token:             viper.GetString("bot.token"),
		adminIDs:          adminIDs,
		grantChatID:       viper.GetInt64("bot.auth.grant-chat-id"),
		mailingChannelID:  viper.GetInt64("bot.mailing.channel-id"),
		avatarChannelID:   viper.GetInt64("bot.avatar.channel-id"),
		introChannelID:    viper.GetInt64("bot.intro.channel-id"),
		passChannelID:     viper.GetInt64("settings.pass.channel-id"),
		qrChannelID:       viper.GetInt64("bot.qr.channel-id"),
		validEmailDomains: viper.GetStringSlice("bot.auth.valid-email-domains"),
	}
}

func (cfg *botConfig) Token() string {
	return cfg.token
}

func (cfg *botConfig) AdminIDs() []int64 {
	return cfg.adminIDs
}

func (cfg *botConfig) GrantChatID() int64 {
	return cfg.grantChatID
}

func (cfg *botConfig) MailingChannelID() int64 {
	return cfg.mailingChannelID
}

func (cfg *botConfig) AvatarChannelID() int64 {
	return cfg.avatarChannelID
}

func (cfg *botConfig) IntroChannelID() int64 {
	return cfg.introChannelID
}

func (cfg *botConfig) PassChannelID() int64 {
	return cfg.passChannelID
}

func (cfg *botConfig) QRChannelID() int64 {
	return cfg.qrChannelID
}

func (cfg *botConfig) ValidEmailDomains() []string {
	return cfg.validEmailDomains
}

func (cfg *botConfig) Validate() {
	if len(cfg.adminIDs) == 0 {
		fmt.Printf("Warning: Bot.AdminIDs is empty, admin functionality may not work properly\n")
	}
	if cfg.mailingChannelID == 0 {
		fmt.Printf("Warning: Bot.MailingChannelID is empty, mailing functionality may not work\n")
	}
	if cfg.avatarChannelID == 0 {
		fmt.Printf("Warning: Bot.AvatarChannelID is empty, avatar uploads may not work\n")
	}
	if cfg.introChannelID == 0 {
		fmt.Printf("Warning: Bot.IntroChannelID is empty, intro uploads may not work\n")
	}
}
