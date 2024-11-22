package scheduler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Feride3d/backup-creator/internal/model"
	mock_service "github.com/Feride3d/backup-creator/internal/scheduler/mocks"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetLastRunTime(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		fileContent  string
		expectedTime time.Time
		expectError  bool
	}{
		{
			name:         "Valid RFC3339 timestamp",
			fileContent:  "2023-11-20T15:04:05Z",
			expectedTime: time.Date(2023, 11, 20, 15, 4, 5, 0, time.UTC),
			expectError:  false,
		},
		{
			name:         "Empty file",
			fileContent:  "",
			expectedTime: time.Time{},
			expectError:  true,
		},
		{
			name:         "Invalid timestamp format",
			fileContent:  "not-a-timestamp",
			expectedTime: time.Time{},
			expectError:  true,
		},
		{
			name:         "Future timestamp",
			fileContent:  "2099-01-01T00:00:00Z",
			expectedTime: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC),
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := tmpDir + "/lastrun.txt"
			err := os.WriteFile(tmpFile, []byte(tt.fileContent), 0644)
			assert.NoError(t, err)

			s := NewScheduler(nil, nil, tmpFile)

			result, err := s.GetLastRunTime()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTime, result)
			}
		})
	}
}

func TestExecuteBackup(t *testing.T) {
	tests := []struct {
		name            string
		mockFetchBlocks func(ctx context.Context, lastRun time.Time) ([]model.ContentBlock, error)
		mockSaveBlocks  func(ctx context.Context, blocks []model.ContentBlock, folder string) error
		expectedError   string
	}{
		{
			name: "Successful backup",
			mockFetchBlocks: func(ctx context.Context, lastRun time.Time) ([]model.ContentBlock, error) {
				return []model.ContentBlock{{ID: 1, Name: "Block1", Content: "Content1"}}, nil
			},
			mockSaveBlocks: func(ctx context.Context, blocks []model.ContentBlock, folder string) error {
				return nil
			},
			expectedError: "",
		},
		{
			name: "Fetch error",
			mockFetchBlocks: func(ctx context.Context, lastRun time.Time) ([]model.ContentBlock, error) {
				return nil, fmt.Errorf("fetch error")
			},
			mockSaveBlocks: nil,
			expectedError:  "failed to fetch content blocks: fetch error",
		},
		{
			name: "Save error",
			mockFetchBlocks: func(ctx context.Context, lastRun time.Time) ([]model.ContentBlock, error) {
				return []model.ContentBlock{{ID: 1, Name: "Block1", Content: "Content1"}}, nil
			},
			mockSaveBlocks: func(ctx context.Context, blocks []model.ContentBlock, folder string) error {
				return fmt.Errorf("save error")
			},
			expectedError: "failed to save content blocks: save error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			lastRunFile := filepath.Join(tmpDir, "lastrun.txt")
			err := os.WriteFile(lastRunFile, []byte("2023-11-22T09:00:00Z"), 0644)
			assert.NoError(t, err)

			mockFetchService := new(mock_service.ContentProvider)
			mockBackupService := new(mock_service.Backuper)

			mockFetchService.On("GetUpdatedContentBlocks", mock.Anything, mock.Anything).
				Return(tt.mockFetchBlocks(context.Background(), time.Date(2024, 11, 22, 9, 0, 0, 0, time.UTC)))

			if tt.mockSaveBlocks != nil {
				mockBackupService.On(
					"SaveContent",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(tt.mockSaveBlocks(context.Background(), []model.ContentBlock{}, ""))
			}

			s := &Scheduler{
				fetchService:  mockFetchService,
				backupService: mockBackupService,
				lastRunFile:   lastRunFile,
			}

			err = s.ExecuteBackup(context.Background())

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockFetchService.AssertExpectations(t)
			if tt.mockSaveBlocks != nil {
				mockBackupService.AssertExpectations(t)
			}
		})
	}
}

type MockBackupExecutor struct {
	mock.Mock
}

func (m *MockBackupExecutor) ExecuteBackup(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestScheduler_Run(t *testing.T) {
	mockFetchService := new(mock_service.ContentProvider)
	mockBackupService := new(mock_service.Backuper)

	mockFetchService.On("GetUpdatedContentBlocks", mock.Anything, mock.Anything).Return([]model.ContentBlock{
		{ID: 1, Name: "Block1", Content: "Content1"},
	}, nil)

	mockBackupService.On("SaveContent", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	scheduler := &Scheduler{
		cronScheduler: cron.New(),
		fetchService:  mockFetchService,
		backupService: mockBackupService,
	}

	go scheduler.Run("@every 1s")
	time.Sleep(2 * time.Second)
	scheduler.cronScheduler.Stop()

	mockFetchService.AssertExpectations(t)
	mockBackupService.AssertExpectations(t)
}
