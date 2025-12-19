package logger

import (
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

var defaultLogger *Logger

func Get(level string) *Logger {
	if defaultLogger == nil {
		var programLevel = new(slog.LevelVar)
		h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
		defaultLogger = &Logger{Logger: slog.New(h)}
		slog.SetDefault(defaultLogger.Logger)
	}
	return defaultLogger
}
