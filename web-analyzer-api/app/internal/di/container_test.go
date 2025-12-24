package di

import (
	"testing"
	"web-analyzer-api/app/internal/util/logger"

	"github.com/stretchr/testify/assert"
)

func TestNewContainer(t *testing.T) {
	log := logger.Get("debug")

	t.Run("Initialize Container", func(t *testing.T) {
		container := NewContainer(log)

		assert.NotNil(t, container)
		assert.NotNil(t, container.HTTPHandlers.WebAnalyzerHandler)
	})
}
