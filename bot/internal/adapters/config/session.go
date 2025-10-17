package config

import (
	"time"

	"github.com/spf13/viper"
)

type SessionConfig interface {
	TTL() time.Duration
	AuthTTL() time.Duration
	ResendTTL() time.Duration
	EmailTTL() time.Duration
	EventIDTTL() time.Duration
}

type sessionConfig struct {
	ttl        time.Duration
	authTTL    time.Duration
	resendTTL  time.Duration
	emailTTL   time.Duration
	eventIDTTL time.Duration
}

func NewSessionConfig() SessionConfig {
	return &sessionConfig{
		ttl:        viper.GetDuration("bot.session.ttl"),
		authTTL:    viper.GetDuration("bot.session.auth-ttl"),
		resendTTL:  viper.GetDuration("bot.session.resend-ttl"),
		emailTTL:   viper.GetDuration("bot.session.email-ttl"),
		eventIDTTL: viper.GetDuration("bot.session.event-id-ttl"),
	}
}

func (cfg *sessionConfig) TTL() time.Duration {
	return cfg.ttl
}

func (cfg *sessionConfig) AuthTTL() time.Duration {
	return cfg.authTTL
}

func (cfg *sessionConfig) ResendTTL() time.Duration {
	return cfg.resendTTL
}

func (cfg *sessionConfig) EmailTTL() time.Duration {
	return cfg.emailTTL
}

func (cfg *sessionConfig) EventIDTTL() time.Duration {
	return cfg.eventIDTTL
}
