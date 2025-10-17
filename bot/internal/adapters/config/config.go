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

	// Validate configuration and print warnings
	warningsManager := NewWarningsManager()
	warningsManager.ValidateConfig(cfg)
	warningsManager.PrintWarnings()

	return cfg, nil
}
