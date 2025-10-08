package config

import "fmt"

// Config represents the unified configuration for the entire application
type Config struct {
	Logger    Logger
	PG        PGConfig
	RedisConf RedisConfig
	SMTP      SMTPConfig
	Bot       BotConfig
	App       AppConfig
	Banner    BannerConfig
}

func NewConfig() (*Config, error) {
	loggerCfg, err := NewLoggerConfig()
	if err != nil {
		return nil, err
	}

	bannerCfg := NewBannerConfig()
	if err := bannerCfg.Validate(); err != nil {
		return nil, fmt.Errorf("banner configuration validation failed: %w", err)
	}

	return &Config{
		Logger:    loggerCfg,
		PG:        NewPGConfig(),
		RedisConf: NewRedisConfig(),
		SMTP:      NewSMTPConfig(),
		Bot:       NewBotConfig(),
		App:       NewAppConfig(),
		Banner:    bannerCfg,
	}, nil
}
