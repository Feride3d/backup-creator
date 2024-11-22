package scheduler

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Feride3d/backup-creator/internal/model"
	"github.com/Feride3d/backup-creator/internal/service"
	"github.com/robfig/cron/v3"
)

type Backuper interface {
	SaveContent(ctx context.Context, blocks []model.ContentBlock, folder string) error
}

type ContentProvider interface {
	GetUpdatedContentBlocks(ctx context.Context, lastRun time.Time) ([]model.ContentBlock, error)
}

type BackupExecutor interface {
	ExecuteBackup(ctx context.Context) error
}

type Scheduler struct {
	cronScheduler *cron.Cron
	executor      BackupExecutor
	fetchService  ContentProvider
	backupService Backuper
	lastRunFile   string
	mu            sync.Mutex
}

func NewScheduler(fetch *service.FetchService, backup *service.BackupService, lastRunFile string) *Scheduler {
	return &Scheduler{
		cronScheduler: cron.New(),
		fetchService:  fetch,
		backupService: backup,
		lastRunFile:   lastRunFile,
		mu:            sync.Mutex{},
	}
}

func (s *Scheduler) Run(cronExpr string) {
	_, err := s.cronScheduler.AddFunc(cronExpr, func() {
		log.Println("Starting scheduled backup...")
		ctx := context.Background()
		if err := s.ExecuteBackup(ctx); err != nil {
			log.Printf("Backup failed: %v", err)
		} else {
			log.Println("Backup completed successfully!")
		}
	})
	if err != nil {
		log.Fatalf("Failed to add cron job: %v", err)
	}

	s.cronScheduler.Start()
	log.Printf("Scheduler started with cron expression: %s", cronExpr)
}

func (s *Scheduler) ExecuteBackup(ctx context.Context) error {
	lastRun, err := s.GetLastRunTime()
	if err != nil {
		log.Printf("Warning: unable to determine last run time: %v. Using default.", err)
		lastRun = time.Now().Add(-24 * time.Hour)
	}
	if s.fetchService == nil {
		return fmt.Errorf("fetchService is not initialized")
	}
	log.Println("Fetching updated content blocks...")
	blocks, err := s.fetchService.GetUpdatedContentBlocks(ctx, lastRun)
	if err != nil {
		return fmt.Errorf("failed to fetch content blocks: %w", err)
	}

	if s.backupService == nil {
		return fmt.Errorf("backupService is not initialized")
	}
	folder := time.Now().Format("backup_20241121")
	log.Println("Saving content blocks...")
	if err := s.backupService.SaveContent(ctx, blocks, folder); err != nil {
		return fmt.Errorf("failed to save content blocks: %w", err)
	}

	return s.UpdateLastRunTime()
}

func (s *Scheduler) GetLastRunTime() (time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.lastRunFile)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, string(data))
}

func (s *Scheduler) UpdateLastRunTime() error {
	s.mu.Lock() // instead of mutex as option github.com/gofrs/flock
	defer s.mu.Unlock()
	data := []byte(time.Now().Format(time.RFC3339))
	return os.WriteFile(s.lastRunFile, data, 0644)
}
