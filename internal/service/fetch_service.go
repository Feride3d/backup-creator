package service

import (
	"context"
	"time"

	"github.com/Feride3d/backup-creator/internal/model"
)

type ContentProvider interface {
	GetUpdatedContentBlocksConcurrent(ctx context.Context, lastRun time.Time, workerCount int, query map[string]interface{}) ([]model.ContentBlock, error)
	FetchPage(ctx context.Context, query map[string]interface{}, page, pageSize int) ([]model.ContentBlock, error)
}

type FetchService struct {
	Provider ContentProvider
}

func NewFetchService(Provider ContentProvider) *FetchService {
	return &FetchService{Provider: Provider}
}

func (s *FetchService) GetUpdatedContentBlocks(ctx context.Context, lastRun time.Time) ([]model.ContentBlock, error) {
	query := make(map[string]interface{})
	workerCount := 5
	return s.Provider.GetUpdatedContentBlocksConcurrent(ctx, lastRun, workerCount, query)
}
