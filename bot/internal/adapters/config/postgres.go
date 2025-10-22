package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type PGConfig interface {
	DSN() string
	EntDSN() string
}

type pgConfig struct {
	host     string
	user     string
	password string
	port     int
	dbName   string
	sslMode  string
	timeZone string
}

func NewPGConfig() PGConfig {
	return &pgConfig{
		host:     viper.GetString("infrastructure.database.host"),
		user:     viper.GetString("infrastructure.database.user"),
		password: viper.GetString("infrastructure.database.password"),
		port:     viper.GetInt("infrastructure.database.port"),
		dbName:   viper.GetString("infrastructure.database.name"),
		sslMode:  viper.GetString("infrastructure.database.ssl-mode"),
		timeZone: viper.GetString("settings.timezone"),
	}
}

func (cfg *pgConfig) DSN() string {
	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=%s TimeZone=%s",
		cfg.user,
		cfg.password,
		cfg.dbName,
		cfg.host,
		cfg.port,
		cfg.sslMode,
		cfg.timeZone,
	)
}

func (cfg *pgConfig) EntDSN() string {
	// Temporarily hardcode a different database name for Entgo to avoid conflicts with GORM
	// TODO: Configure this properly in config.yaml
	return fmt.Sprintf("user=%s password=%s dbname=%s_ent host=%s port=%d sslmode=%s TimeZone=%s",
		cfg.user,
		cfg.password,
		cfg.dbName,
		cfg.host,
		cfg.port,
		cfg.sslMode,
		cfg.timeZone,
	)
}
