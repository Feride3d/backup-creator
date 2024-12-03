package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Feride3d/backup-creator/internal/model"
	"github.com/stretchr/testify/assert"
)

type mockAuthClient struct {
	GetAccessTokenFunc func() (model.Token, error)
}

func (m *mockAuthClient) GetAccessToken() (model.Token, error) {
	return m.GetAccessTokenFunc()
}

func TestEnsureTokenValid(t *testing.T) {
	mockAuth := &mockAuthClient{
		GetAccessTokenFunc: func() (model.Token, error) {
			return model.Token{AccessToken: "new_token", ExpiresIn: 3600, ExpiryTime: time.Now().Add(time.Hour)}, nil
		},
	}

	token := &model.Token{AccessToken: "expired_token", ExpiryTime: time.Now().Add(-time.Minute)}
	client := NewContentClient("https://api.example.com", token, mockAuth)

	err := client.EnsureTokenValid()
	assert.NoError(t, err)
	assert.Equal(t, "new_token", client.token.AccessToken)
}

func TestFetchPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test_token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		response := map[string]interface{}{
			"items": []model.ContentBlock{
				{ID: 1, Name: "Block1"},
				{ID: 2, Name: "Block2"},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	token := &model.Token{AccessToken: "test_token"}
	client := NewContentClient(server.URL, token, nil)

	query := make(map[string]interface{})
	items, err := client.FetchPage(context.Background(), query, 1, 50)
	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "Block1", items[0].Name)
}

func TestFetchPage_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	token := &model.Token{AccessToken: "test_token"}
	client := NewContentClient(server.URL, token, nil)

	query := make(map[string]interface{})
	_, err := client.FetchPage(context.Background(), query, 1, 50)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
}

func TestGetUpdatedContentBlocksConcurrent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var query map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		page := int(query["page"].(map[string]interface{})["page"].(float64))

		response := map[string]interface{}{
			"items": []model.ContentBlock{
				{ID: page, Name: fmt.Sprintf("Block%d", page)},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	token := &model.Token{AccessToken: "test_token"}
	client := NewContentClient(server.URL, token, nil)

	query := make(map[string]interface{})
	items, err := client.GetUpdatedContentBlocksConcurrent(context.Background(), time.Now().Add(-time.Hour), 2, query)

	assert.NoError(t, err)
	assert.Len(t, items, 5)
	for i, item := range items {
		assert.Equal(t, fmt.Sprintf("Block%d", i+1), item.Name)
	}
}

func TestGetUpdatedContentBlocksConcurrent_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	token := &model.Token{AccessToken: "test_token"}
	client := NewContentClient(server.URL, token, nil)

	query := make(map[string]interface{})
	_, err := client.GetUpdatedContentBlocksConcurrent(context.Background(), time.Now().Add(-time.Hour), 2, query)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
}
