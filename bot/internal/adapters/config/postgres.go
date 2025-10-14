package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type PGConfig interface {
	DSN() string
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
		host:     viper.GetString("service.database.host"),
		user:     viper.GetString("service.database.user"),
		password: viper.GetString("service.database.password"),
		port:     viper.GetInt("service.database.port"),
		dbName:   viper.GetString("service.database.name"),
		sslMode:  viper.GetString("service.database.ssl-mode"),
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
