package scheduler

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Feride3d/backup-creator/internal/service"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	CronScheduler *cron.Cron
	FetchService  *service.FetchService
	BackupService *service.BackupService
	LastRunFile   string
	mu            sync.Mutex
}

func NewScheduler(fetch *service.FetchService, backup *service.BackupService, lastRunFile string) *Scheduler {
	return &Scheduler{
		CronScheduler: cron.New(),
		FetchService:  fetch,
		BackupService: backup,
		LastRunFile:   lastRunFile,
		mu:            sync.Mutex{},
	}
}

func (s *Scheduler) Run(cronExpr string) {
	_, err := s.CronScheduler.AddFunc(cronExpr, func() {
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

	s.CronScheduler.Start()
	log.Printf("Scheduler started with cron expression: %s", cronExpr)
}

func (s *Scheduler) ExecuteBackup(ctx context.Context) error {
	lastRun, err := s.GetLastRunTime()
	if err != nil {
		log.Printf("Warning: unable to determine last run time: %v. Using default.", err)
		lastRun = time.Now().Add(-24 * time.Hour)
	}

	log.Println("Fetching updated content blocks...")
	blocks, err := s.FetchService.FetchUpdatedContentBlocks(ctx, lastRun)
	if err != nil {
		return fmt.Errorf("failed to fetch content blocks: %w", err)
	}

	folder := time.Now().Format("backup_20060102")
	log.Println("Saving content blocks...")
	if err := s.BackupService.SaveContentBlocks(ctx, blocks, folder); err != nil {
		return fmt.Errorf("failed to save content blocks: %w", err)
	}

	return s.UpdateLastRunTime()
}

func (s *Scheduler) GetLastRunTime() (time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.LastRunFile)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, string(data))
}

func (s *Scheduler) UpdateLastRunTime() error {
	s.mu.Lock() // instead of mutex as option github.com/gofrs/flock
	defer s.mu.Unlock()
	data := []byte(time.Now().Format(time.RFC3339))
	return os.WriteFile(s.LastRunFile, data, 0644)
}
