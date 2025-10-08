package config

import (
	"github.com/spf13/viper"
)

type AppConfig interface {
	Timezone() string
	EmailConfirmationTemplate() string
	PassEmails() []string
	QRLogoPath() string
	VersionNotifyOnStartup() bool
	VersionChannelID() int64
}

type appConfig struct {
	timezone                  string
	emailConfirmationTemplate string
	passEmails                []string
	qrLogoPath                string
	versionNotifyOnStartup    bool
	versionChannelID          int64
}

func NewAppConfig() AppConfig {
	return &appConfig{
		timezone:                  viper.GetString("settings.timezone"),
		emailConfirmationTemplate: viper.GetString("settings.html.email-confirmation"),
		passEmails:                viper.GetStringSlice("settings.pass-emails"),
		qrLogoPath:                viper.GetString("settings.qr.logo-path"),
		versionNotifyOnStartup:    viper.GetBool("settings.version.notify-on-startup"),
		versionChannelID:          viper.GetInt64("settings.version.channel-id"),
	}
}

func (cfg *appConfig) Timezone() string {
	return cfg.timezone
}

func (cfg *appConfig) EmailConfirmationTemplate() string {
	return cfg.emailConfirmationTemplate
}

func (cfg *appConfig) PassEmails() []string {
	return cfg.passEmails
}

func (cfg *appConfig) QRLogoPath() string {
	return cfg.qrLogoPath
}

func (cfg *appConfig) VersionNotifyOnStartup() bool {
	return cfg.versionNotifyOnStartup
}

func (cfg *appConfig) VersionChannelID() int64 {
	return cfg.versionChannelID
}
