package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type RedisConfig interface {
	Host() string
	Port() string
	Password() string
	Address() string
}

type redisConfig struct {
	host     string
	port     string
	password string
}

func NewRedisConfig() RedisConfig {
	return &redisConfig{
		host:     viper.GetString("service.redis.host"),
		port:     viper.GetString("service.redis.port"),
		password: viper.GetString("service.redis.password"),
	}
}

func (cfg *redisConfig) Host() string {
	return cfg.host
}

func (cfg *redisConfig) Port() string {
	return cfg.port
}

func (cfg *redisConfig) Password() string {
	return cfg.password
}

func (cfg *redisConfig) Address() string {
	return fmt.Sprintf("%s:%s", cfg.host, cfg.port)
}
