package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Feride3d/backup-creator/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNewLocalStorage(t *testing.T) {
	storagePath := "/tmp/test_storage"
	localStorage := NewLocalStorage(storagePath)

	assert.NotNil(t, localStorage)
	assert.Equal(t, storagePath, localStorage.storagePath)
}

func TestLocalStorage_SaveContentBlocks(t *testing.T) {
	t.Run("Save content blocks successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		localStorage := NewLocalStorage(tmpDir)

		blocks := []model.ContentBlock{
			{ID: 1, Name: "Block1", Content: "Content1"},
			{ID: 2, Name: "Block2", Content: "Content2"},
		}
		folder := "backup_20241121"

		ctx := context.Background()
		err := localStorage.SaveContentBlocks(ctx, blocks, folder)

		assert.NoError(t, err)

		for _, block := range blocks {
			filePath := filepath.Join(tmpDir, folder, fmt.Sprintf("%d.json", block.ID))
			assert.FileExists(t, filePath)

			data, err := os.ReadFile(filePath)
			assert.NoError(t, err)

			var savedBlock model.ContentBlock
			err = json.Unmarshal(data, &savedBlock)
			assert.NoError(t, err)
			assert.Equal(t, block, savedBlock)
		}
	})

	t.Run("Error creating directory", func(t *testing.T) {
		invalidPath := "/invalid_path/test_storage"
		localStorage := NewLocalStorage(invalidPath)

		blocks := []model.ContentBlock{
			{ID: 1, Name: "Block1", Content: "Content1"},
		}
		folder := "backup_20241121"

		ctx := context.Background()
		err := localStorage.SaveContentBlocks(ctx, blocks, folder)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create backup directory")
	})

	t.Run("Error writing to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		localStorage := NewLocalStorage(tmpDir)

		blocks := []model.ContentBlock{
			{ID: 1, Name: "Block1", Content: make(chan int)},
		}
		folder := "backup_20241121"

		ctx := context.Background()
		err := localStorage.SaveContentBlocks(ctx, blocks, folder)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal block")
	})
}
