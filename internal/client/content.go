package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Feride3d/backup-creator/internal/model"
)

type ContentClient struct {
	apiURL string
	token  string
}

func NewContentClient(apiURL, token string) *ContentClient {
	return &ContentClient{apiURL, token}
}

func (c *ContentClient) GetUpdatedContentBlocks(ctx context.Context, lastRun time.Time) ([]model.ContentBlock, error) {
	var allItems []model.ContentBlock
	page := 1
	pageSize := 50

	for {
		url := fmt.Sprintf("%s/query", c.apiURL)
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"leftOperand": map[string]interface{}{
					"property":       "modifiedDate",
					"simpleOperator": "greaterThan",
					"value":          lastRun.Format(time.RFC3339),
				},
			},
			"page": map[string]interface{}{
				"page":     page,
				"pageSize": pageSize,
			},
			"fields": []string{"id", "name", "modifiedDate", "content"},
		}

		queryJSON, err := json.Marshal(query)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal query: %v", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(queryJSON))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API error: %s (status: %d, response: %s)", url, resp.StatusCode, string(body))
		}

		var result struct {
			Items    []model.ContentBlock `json:"items"`
			Page     int                  `json:"page"`
			PageSize int                  `json:"pageSize"`
			Count    int                  `json:"count"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %v", err)
		}

		allItems = append(allItems, result.Items...)

		if len(result.Items) < pageSize {
			break
		}

		page++
	}

	if len(allItems) == 0 {
		return []model.ContentBlock{}, nil
	}

	return allItems, nil
}
