package service

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/Feride3d/backup-creator/internal/model"
)

// Storage defines an interface for saving content blocks to a storage system
type Storage interface {
	SaveContentBlocks(ctx context.Context, blocks []model.ContentBlock, folder string) error
}

type BackupService struct {
	storage Storage
}

func NewBackupService(storage Storage) *BackupService {
	return &BackupService{storage: storage}
}

// Saving each block to storage is an independent operation using goroutines
func (s *BackupService) SaveContent(ctx context.Context, blocks []model.ContentBlock, folder string) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(blocks))

	for _, block := range blocks {
		wg.Add(1)
		go func(b model.ContentBlock) {
			defer wg.Done()
			if err := s.storage.SaveContentBlocks(ctx, []model.ContentBlock{b}, folder); err != nil {
				errCh <- fmt.Errorf("block ID %d: %v", b.ID, err)
			}
		}(block)
	}

	wg.Wait()
	close(errCh)

	var finalErr error
	for err := range errCh {
		log.Printf("Error saving block: %v", err)
		if finalErr == nil {
			finalErr = err
		}
	}

	return finalErr
}
