package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Feride3d/backup-creator/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestAuthClient_GetAccessToken(t *testing.T) {
	tests := []struct {
		name           string
		authURL        string
		serverResponse string
		serverStatus   int
		expectedToken  model.Token
		expectError    bool
		errorMessage   string
	}{
		{
			name:         "Invalid auth URL",
			authURL:      "://invalid-url",
			expectError:  true,
			errorMessage: "invalid auth URL",
		},
		{
			name:    "Successful token retrieval",
			authURL: "http://valid-url",
			serverResponse: `{
				"access_token": "test_token",
				"expires_in": 1080,
				"token_type": "Bearer"
			}`,
			serverStatus: http.StatusOK,
			expectedToken: model.Token{
				AccessToken: "test_token",
				ExpiresIn:   1080,
				ExpiryTime:  time.Now().Add(18 * time.Minute),
			},
			expectError: false,
		},
		{
			name:    "Server error",
			authURL: "http://valid-url",
			serverResponse: `{
				"error": "invalid_client",
				"error_description": "Invalid client ID."
			}`,
			serverStatus: http.StatusUnauthorized,
			expectError:  true,
			errorMessage: "failed to get access token: invalid_client - Invalid client ID.",
		},
		{
			name:           "Invalid JSON in response",
			authURL:        "http://valid-url",
			serverResponse: `{invalid_json}`,
			serverStatus:   http.StatusOK,
			expectError:    true,
			errorMessage:   "failed to decode token response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Invalid auth URL" {
				client := NewAuthClient(tt.authURL, "client_id", "client_secret")
				token, err := client.GetAccessToken()
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Empty(t, token.AccessToken)
				return
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v2/token", r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				w.WriteHeader(tt.serverStatus)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			client := NewAuthClient(server.URL+"/v2/token", "client_id", "client_secret")

			token, err := client.GetAccessToken()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Empty(t, token.AccessToken)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken.AccessToken, token.AccessToken)
				assert.WithinDuration(t, time.Now().Add(time.Duration(tt.expectedToken.ExpiresIn)*time.Second), token.ExpiryTime, time.Second)
			}
		})
	}
}

func TestAuthClient_GetAccessTokenMarshal(t *testing.T) {
	tests := []struct {
		name           string
		authURL        string
		clientSecret   interface{}
		serverResponse string
		serverStatus   int
		expectError    bool
		errorMessage   string
	}{
		{
			name:         "Payload marshal error",
			authURL:      "http://valid-url",
			clientSecret: make(chan int),
			expectError:  true,
			errorMessage: "json: unsupported type: chan int",
		},
		{
			name:         "HTTP request error",
			authURL:      "http://invalid-url",
			clientSecret: "valid_secret",
			expectError:  true,
			errorMessage: "failed to send token request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Payload marshal error" {
				payload := map[string]interface{}{
					"client_secret": tt.clientSecret,
				}
				_, err := json.Marshal(payload)

				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				return
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			client := NewAuthClient(server.URL, "client_id", tt.clientSecret.(string))

			token, err := client.GetAccessToken()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Empty(t, token.AccessToken)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
