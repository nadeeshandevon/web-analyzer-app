package repositorysql

import (
	"testing"
	"web-analyzer-api/app/internal/model"
	"web-analyzer-api/app/internal/util/logger"

	"github.com/stretchr/testify/assert"
)

func TestWebAnalyzerRepo(t *testing.T) {
	log := logger.Get("info")
	repo := NewWebAnalyzerRepo(log)

	t.Run("Save and GetById", func(t *testing.T) {
		analysis := model.WebAnalyzer{
			URL: "http://test.test",
		}

		id, err := repo.Save(analysis)
		assert.NoError(t, err)
		assert.NotEmpty(t, id)

		// Get by valid ID
		found, err := repo.GetById(id)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, id, found.ID)
		assert.Equal(t, "http://test.test", found.URL)

		// Get by invalid ID
		notFound, err := repo.GetById("123")
		assert.NoError(t, err)
		assert.Nil(t, notFound)
	})

	t.Run("Update", func(t *testing.T) {
		analysis := model.WebAnalyzer{
			URL: "http://test.test",
		}

		id, _ := repo.Save(analysis)

		updatedAnalysis := model.WebAnalyzer{
			ID:     id,
			URL:    "http://updated.test",
			Status: "success",
		}

		updatedID, err := repo.Update(updatedAnalysis)
		assert.NoError(t, err)
		assert.Equal(t, id, updatedID)

		// Verify update
		found, _ := repo.GetById(id)
		assert.Equal(t, "http://updated.test", found.URL)
		assert.Equal(t, "success", found.Status)

		// Update unavailable record
		invalidUpdate := model.WebAnalyzer{ID: "123"}
		_, err = repo.Update(invalidUpdate)
		assert.Error(t, err)
		assert.Equal(t, "record not found", err.Error())
	})
}
