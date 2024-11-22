package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Feride3d/backup-creator/internal/model"
)

type LocalStorage struct {
	storagePath string
}

func NewLocalStorage(storagePath string) *LocalStorage {
	return &LocalStorage{storagePath: storagePath}
}

// SaveContentBlocks saves content blocks to the local file system
func (s *LocalStorage) SaveContentBlocks(ctx context.Context, blocks []model.ContentBlock, folder string) error {
	backupPath := filepath.Join(s.storagePath, folder)

	if err := os.MkdirAll(backupPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	for _, block := range blocks {
		filePath := filepath.Join(backupPath, fmt.Sprintf("%d.json", block.ID))
		data, err := json.MarshalIndent(block, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal block %d: %v", block.ID, err)
		}
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write block %d to file: %v", block.ID, err)
		}
	}
	fmt.Printf("Saved content blocks to local directory: %s\n", backupPath)
	return nil
}
