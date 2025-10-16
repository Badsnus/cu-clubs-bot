package config

import (
	"github.com/spf13/viper"
)

type SMTPConfig interface {
	Host() string
	Port() int
	Login() string
	Password() string
	Email() string
	Domain() string
}

type smtpConfig struct {
	host     string
	port     int
	login    string
	password string
	email    string
	domain   string
}

func NewSMTPConfig() SMTPConfig {
	return &smtpConfig{
		host:     viper.GetString("infrastructure.smtp.host"),
		port:     viper.GetInt("infrastructure.smtp.port"),
		login:    viper.GetString("infrastructure.smtp.login"),
		password: viper.GetString("infrastructure.smtp.pass"),
		email:    viper.GetString("infrastructure.smtp.email"),
		domain:   viper.GetString("infrastructure.smtp.domain"),
	}
}

func (cfg *smtpConfig) Host() string {
	return cfg.host
}

func (cfg *smtpConfig) Port() int {
	return cfg.port
}

func (cfg *smtpConfig) Login() string {
	return cfg.login
}

func (cfg *smtpConfig) Password() string {
	return cfg.password
}

func (cfg *smtpConfig) Email() string {
	return cfg.email
}

func (cfg *smtpConfig) Domain() string {
	return cfg.domain
}
