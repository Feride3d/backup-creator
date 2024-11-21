package service

import (
	"context"
	"time"

	"github.com/Feride3d/backup-creator/internal/model"
)

type ContentProvider interface {
	GetUpdatedContentBlocks(ctx context.Context, lastRun time.Time) ([]model.ContentBlock, error)
}

type FetchService struct {
	Provider ContentProvider
}

func NewFetchService(Provider ContentProvider) *FetchService {
	return &FetchService{Provider: Provider}
}

func (s *FetchService) FetchUpdatedContentBlocks(ctx context.Context, lastRun time.Time) ([]model.ContentBlock, error) {
	return s.Provider.GetUpdatedContentBlocks(ctx, lastRun)
}
