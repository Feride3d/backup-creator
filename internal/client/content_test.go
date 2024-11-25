package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Feride3d/backup-creator/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestContentClient_GetUpdatedContentBlocks_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse map[int]string
		serverStatus   int
		expectError    bool
		errorMessage   string
		expectedBlocks []model.ContentBlock
	}{
		{
			name: "Empty result from API",
			serverResponse: map[int]string{
				1: `{
					"items": [],
					"page": 1,
					"pageSize": 50,
					"count": 0
				}`,
			},
			serverStatus:   http.StatusOK,
			expectError:    false,
			expectedBlocks: []model.ContentBlock{},
		},
		{
			name:           "API returns 500 Internal Server Error",
			serverResponse: nil,
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
			errorMessage:   "API error",
		},
		{
			name: "Malformed JSON response",
			serverResponse: map[int]string{
				1: `{
					"items": [
						{"id": 1, "name": "Block1", "modifiedDate": "invalid-date", "content": "Content1"}
					]
				}`,
			},
			serverStatus: http.StatusOK,
			expectError:  true,
			errorMessage: "failed to decode response",
		},
		{
			name:           "Request timeout",
			serverResponse: nil,
			serverStatus:   http.StatusOK,
			expectError:    true,
			errorMessage:   "failed to send request",
		},
		{
			name: "Last page contains fewer items than pageSize",
			serverResponse: map[int]string{
				1: `{
					"items": [
						{"id": 1, "name": "Block1", "modifiedDate": "2023-11-21T10:00:00Z", "content": "Content1"}
					],
					"page": 1,
					"pageSize": 50,
					"count": 1
				}`,
			},
			serverStatus: http.StatusOK,
			expectError:  false,
			expectedBlocks: []model.ContentBlock{
				{ID: 1, Name: "Block1", ModifiedDate: time.Date(2023, 11, 21, 10, 0, 0, 0, time.UTC), Content: "Content1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var query struct {
					Page struct {
						Page     int `json:"page"`
						PageSize int `json:"pageSize"`
					} `json:"page"`
				}

				_ = json.NewDecoder(r.Body).Decode(&query)
				page := query.Page.Page

				if tt.name == "Request timeout" {
					time.Sleep(2 * time.Second)
					return
				}

				response, exists := tt.serverResponse[page]
				if !exists {
					w.WriteHeader(tt.serverStatus)
					return
				}

				w.WriteHeader(tt.serverStatus)
				_, _ = w.Write([]byte(response))
			}))
			defer server.Close()

			client := NewContentClient(server.URL, "test_token")
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			lastRun := time.Date(2023, 11, 21, 0, 0, 0, 0, time.UTC)

			blocks, err := client.GetUpdatedContentBlocks(ctx, lastRun)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBlocks, blocks)
			}
		})
	}
}
