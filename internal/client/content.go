package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/Feride3d/backup-creator/internal/model"
)

type AuthProvider interface {
	GetAccessToken() (model.Token, error)
}

type ContentClient struct {
	apiURL     string
	token      *model.Token
	authClient AuthProvider
}

func NewContentClient(apiURL string, token *model.Token, authClient AuthProvider) *ContentClient {
	return &ContentClient{apiURL: apiURL, token: token, authClient: authClient}
}

func (c *ContentClient) EnsureTokenValid() error {
	if c.token.IsExpired() {
		newToken, err := c.authClient.GetAccessToken()
		if err != nil {
			return fmt.Errorf("failed to refresh token: %w", err)
		}
		c.token = &newToken
	}
	return nil
}

func (c *ContentClient) GetUpdatedContentBlocksConcurrent(ctx context.Context, lastRun time.Time, workerCount int, query map[string]interface{}) ([]model.ContentBlock, error) {
	var allItems []model.ContentBlock
	jobs := make(chan int)
	results := make(chan []model.ContentBlock)
	errors := make(chan error, workerCount)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case page, ok := <-jobs:
					if !ok {
						return
					}
					items, err := c.FetchPage(ctx, query, page, 50)
					if err != nil {
						errors <- err
						cancel()
						return
					}
					results <- items
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		page := 1
		for {
			select {
			case jobs <- page:
				page++
				if page > 5 {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	var firstErr error
	for {
		select {
		case res, ok := <-results:
			if !ok {
				sort.Slice(allItems, func(i, j int) bool {
					return allItems[i].ID < allItems[j].ID
				})
				if firstErr != nil {
					return nil, firstErr
				}
				return allItems, nil
			}
			allItems = append(allItems, res...)
		case err := <-errors:
			if firstErr == nil {
				firstErr = err
			}
		case <-ctx.Done():
			if firstErr != nil {
				return nil, firstErr
			}
			return nil, ctx.Err()
		}
	}
}

func (c *ContentClient) FetchPage(ctx context.Context, query map[string]interface{}, page, pageSize int) ([]model.ContentBlock, error) {
	localQuery := make(map[string]interface{})
	for k, v := range query {
		localQuery[k] = v
	}
	localQuery["page"] = map[string]interface{}{
		"page":     page,
		"pageSize": pageSize,
	}

	queryJSON, err := json.Marshal(localQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/query", c.apiURL), bytes.NewBuffer(queryJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s (status: %d, response: %s)", c.apiURL, resp.StatusCode, string(body))
	}

	var result struct {
		Items []model.ContentBlock `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return result.Items, nil
}
