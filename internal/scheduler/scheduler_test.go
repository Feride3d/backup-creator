package scheduler_test

import (
	"os"
	"testing"
	"time"

	"github.com/Feride3d/backup-creator/internal/scheduler"
	"github.com/stretchr/testify/assert"
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

			s := scheduler.NewScheduler(nil, nil, tmpFile)

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
