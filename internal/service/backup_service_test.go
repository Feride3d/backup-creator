package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Feride3d/backup-creator/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorage is a mock implementation of the Storage interface
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) SaveContentBlocks(ctx context.Context, blocks []model.ContentBlock, folder string) error {
	args := m.Called(ctx, blocks, folder)
	return args.Error(0)
}

func TestBackupService_SaveContentBlocks_Success(t *testing.T) {

	mockStorage := new(MockStorage)
	backupService := NewBackupService(mockStorage)

	ctx := context.Background()
	blocks := []model.ContentBlock{
		{ID: 1, Name: "Block 1", Content: "Content 1"},
		{ID: 2, Name: "Block 2", Content: "Content 2"},
	}
	folder := "backup_20230101"

	// Mock behavior: no errors
	for _, block := range blocks {
		mockStorage.On("SaveContentBlocks", ctx, []model.ContentBlock{block}, folder).Return(nil).Once()
	}

	err := backupService.SaveContent(ctx, blocks, folder)

	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

func TestBackupService_SaveContentBlocks_PartialFailure(t *testing.T) {

	mockStorage := new(MockStorage)
	backupService := NewBackupService(mockStorage)

	ctx := context.Background()
	blocks := []model.ContentBlock{
		{ID: 1, Name: "Block 1", Content: "Content 1"},
		{ID: 2, Name: "Block 2", Content: "Content 2"},
		{ID: 3, Name: "Block 3", Content: "Content 3"},
	}
	folder := "backup_20230101"

	// Mock behavior: one block fails
	mockStorage.On("SaveContentBlocks", ctx, []model.ContentBlock{blocks[0]}, folder).Return(nil).Once()
	mockStorage.On("SaveContentBlocks", ctx, []model.ContentBlock{blocks[1]}, folder).Return(errors.New("disk full")).Once()
	mockStorage.On("SaveContentBlocks", ctx, []model.ContentBlock{blocks[2]}, folder).Return(nil).Once()

	err := backupService.SaveContent(ctx, blocks, folder)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disk full")
	mockStorage.AssertExpectations(t)
}

func TestBackupService_SaveContentBlocks_AllFailure(t *testing.T) {

	mockStorage := new(MockStorage)
	backupService := NewBackupService(mockStorage)

	ctx := context.Background()
	blocks := []model.ContentBlock{
		{ID: 1, Name: "Block 1", Content: "Content 1"},
		{ID: 2, Name: "Block 2", Content: "Content 2"},
	}
	folder := "backup_20230101"

	// Mock behavior: all blocks fail
	for _, block := range blocks {
		mockStorage.On("SaveContentBlocks", ctx, []model.ContentBlock{block}, folder).Return(errors.New("network error")).Once()
	}

	err := backupService.SaveContent(ctx, blocks, folder)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network error")
	mockStorage.AssertExpectations(t)
}

func TestBackupService_SaveContentBlocks_Concurrency(t *testing.T) {

	mockStorage := new(MockStorage)
	backupService := NewBackupService(mockStorage)

	ctx := context.Background()
	blocks := []model.ContentBlock{
		{ID: 1, Name: "Block 1", Content: "Content 1"},
		{ID: 2, Name: "Block 2", Content: "Content 2"},
		{ID: 3, Name: "Block 3", Content: "Content 3"},
	}
	folder := "backup_20230101"

	var wg sync.WaitGroup
	wg.Add(len(blocks))

	// Mock behavior with concurrency handling
	for _, block := range blocks {
		blockCopy := block
		mockStorage.On("SaveContentBlocks", ctx, []model.ContentBlock{blockCopy}, folder).
			Run(func(args mock.Arguments) {
				defer wg.Done()
			}).Return(nil).Once()
	}

	go func() {
		err := backupService.SaveContent(ctx, blocks, folder)
		assert.NoError(t, err)
	}()

	wg.Wait()
	mockStorage.AssertExpectations(t)
}
