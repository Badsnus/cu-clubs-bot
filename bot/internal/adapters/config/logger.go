package config

import (
	"time"

	"github.com/spf13/viper"
)

type Logger interface {
	Debug() bool
	LogToFile() bool
	LogsDir() string
	TimeLocation() *time.Location
	LogToChannel() bool
	ChannelID() int64
	ChannelLocale() string
	ChannelLogLevel() int
}

type loggerConfig struct {
	debug           bool
	logToFile       bool
	logsDir         string
	timeLocation    *time.Location
	logToChannel    bool
	channelID       int64
	channelLocale   string
	channelLogLevel int
}

func NewLoggerConfig() (Logger, error) {
	location, err := time.LoadLocation(viper.GetString("settings.timezone"))
	if err != nil {
		return nil, err
	}

	return &loggerConfig{
		debug:           viper.GetBool("settings.logging.debug"),
		logToFile:       viper.GetBool("settings.logging.log-to-file"),
		logsDir:         viper.GetString("settings.logging.logs-dir"),
		timeLocation:    location,
		logToChannel:    viper.GetBool("settings.logging.log-to-channel"),
		channelID:       viper.GetInt64("settings.logging.channel-id"),
		channelLocale:   viper.GetString("settings.logging.locale"),
		channelLogLevel: viper.GetInt("settings.logging.channel-log-level"),
	}, nil
}

func (cfg *loggerConfig) Debug() bool {
	return cfg.debug
}

func (cfg *loggerConfig) LogToFile() bool {
	return cfg.logToFile
}

func (cfg *loggerConfig) LogsDir() string {
	return cfg.logsDir
}

func (cfg *loggerConfig) TimeLocation() *time.Location {
	return cfg.timeLocation
}

func (cfg *loggerConfig) LogToChannel() bool {
	return cfg.logToChannel
}

func (cfg *loggerConfig) ChannelID() int64 {
	return cfg.channelID
}

func (cfg *loggerConfig) ChannelLocale() string {
	return cfg.channelLocale
}

func (cfg *loggerConfig) ChannelLogLevel() int {
	return cfg.channelLogLevel
}
