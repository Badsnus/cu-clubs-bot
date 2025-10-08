package types

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.SugaredLogger
	LogsPath   string
	Name       string
	FileWriter *os.File
}

// Log represents a log entry
type Log struct {
	Timestamp  time.Time
	Caller     string
	LoggerName string
	Level      zapcore.Level
	Message    string
}

// LogHook is a function that will be called for each log entry
type LogHook func(log Log)
