package logger

import (
	"log/slog"
	"os"
	"strings"
)

type Logger struct {
	*slog.Logger
}

var defaultLogger *Logger

func Get(level string) *Logger {
	if defaultLogger == nil {
		logLevel := parseLogLevel(level)
		h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})
		defaultLogger = &Logger{Logger: slog.New(h)}
		slog.SetDefault(defaultLogger.Logger)
	}
	return defaultLogger
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
