package main

import (
	"log"

	"github.com/Feride3d/backup-creator/internal/client"
	"github.com/Feride3d/backup-creator/internal/config"
	"github.com/Feride3d/backup-creator/internal/scheduler"
	"github.com/Feride3d/backup-creator/internal/service"
	"github.com/Feride3d/backup-creator/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	cfg := config.Load()

	authClient := client.NewAuthClient(cfg.AuthURL, cfg.ClientID, cfg.ClientSecret)
	token, err := authClient.GetAccessToken()
	if err != nil {
		log.Fatalf("Failed to get access token: %v", err)
	}

	if token.IsExpired() {
		log.Println("Token expired, requesting a new one")
		token, err = authClient.GetAccessToken()
		if err != nil {
			log.Fatalf("Failed to refresh token: %v", err)
		}
	}

	contentClient := client.NewContentClient(cfg.APIURL, token.AccessToken)

	var selectedStorage service.Storage
	if cfg.S3Bucket != "" {
		selectedStorage = storage.NewS3Storage(cfg.S3Region, cfg.S3Bucket, cfg.S3AccessKey, cfg.S3SecretKey)
	} else {
		selectedStorage = storage.NewLocalStorage(cfg.StoragePath)
	}

	fetchService := service.NewFetchService(contentClient)
	backupService := service.NewBackupService(selectedStorage)

	scheduler := scheduler.NewScheduler(fetchService, backupService, "lastrun.txt")
	cronExpr := "0 0 * * *" // cron job every day at midnight
	scheduler.Run(cronExpr)

	log.Println("Backup completed successfully!")
}
