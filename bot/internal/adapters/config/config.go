package config

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/location"
)

// Config represents the unified configuration for the entire application
type Config struct {
	Logger    Logger
	PG        PGConfig
	RedisConf RedisConfig
	SMTP      SMTPConfig
	Bot       BotConfig
	App       AppConfig
	Banner    BannerConfig
	Session   SessionConfig
}

func NewConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Search config in multiple locations
	viper.AddConfigPath("/opt/config") // docker mounted config
	viper.AddConfigPath(".")           // current directory
	viper.AddConfigPath("../")         // parent directory
	viper.AddConfigPath("../../")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	loggerCfg, err := NewLoggerConfig()
	if err != nil {
		return nil, err
	}

	bannerCfg := NewBannerConfig()
	if err := bannerCfg.Validate(); err != nil {
		return nil, fmt.Errorf("banner configuration validation failed: %w", err)
	}

	cfg := &Config{
		Logger:    loggerCfg,
		PG:        NewPGConfig(),
		RedisConf: NewRedisConfig(),
		SMTP:      NewSMTPConfig(),
		Bot:       NewBotConfig(),
		App:       NewAppConfig(),
		Banner:    bannerCfg,
		Session:   NewSessionConfig(),
	}

	location.Init(cfg.App.Timezone())

	// Log warnings for empty config values
	if cfg.Logger.LogToChannel() && cfg.Logger.ChannelID() == 0 {
		fmt.Printf("Warning: Logger.ChannelID is empty, but LogToChannel is enabled\n")
	}
	if cfg.App.VersionNotifyOnStartup() && cfg.App.VersionChannelID() == 0 {
		fmt.Printf("Warning: App.VersionChannelID is empty, but VersionNotifyOnStartup is enabled\n")
	}
	if len(cfg.Bot.AdminIDs()) == 0 {
		fmt.Printf("Warning: Bot.AdminIDs is empty, admin functionality may not work properly\n")
	}
	if cfg.Bot.MailingChannelID() == 0 {
		fmt.Printf("Warning: Bot.MailingChannelID is empty, mailing functionality may not work\n")
	}
	if cfg.Bot.AvatarChannelID() == 0 {
		fmt.Printf("Warning: Bot.AvatarChannelID is empty, avatar uploads may not work\n")
	}
	if cfg.Bot.IntroChannelID() == 0 {
		fmt.Printf("Warning: Bot.IntroChannelID is empty, intro uploads may not work\n")
	}
	if len(cfg.App.PassLocationSubstrings()) == 0 {
		fmt.Printf("Warning: App.PassLocationSubstrings is empty, pass location validation may not work\n")
	}

	return cfg, nil
}
