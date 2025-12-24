package logger

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected slog.Level
	}{
		{"debug", "debug", slog.LevelDebug},
		{"DEBUG", "DEBUG", slog.LevelDebug},
		{"info", "info", slog.LevelInfo},
		{"warn", "warn", slog.LevelWarn},
		{"warning", "warning", slog.LevelWarn},
		{"error", "error", slog.LevelError},
		{"empty", "", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getLogLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGet(t *testing.T) {
	defaultLogger = nil

	t.Run("Get logger instance", func(t *testing.T) {
		log1 := Get("debug")
		assert.NotNil(t, log1)
		assert.NotNil(t, log1.Logger)

		log2 := Get("info")
		assert.Equal(t, log1, log2, "Get should return the same instance")
	})
}
