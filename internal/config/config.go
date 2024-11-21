package config

import (
	"fmt"
	"os"
)

type Config struct {
	AuthURL      string
	APIURL       string
	ClientID     string
	ClientSecret string
	StoragePath  string
	S3Bucket     string
	S3Region     string
	S3AccessKey  string
	S3SecretKey  string
}

func Load() Config {
	authBaseURL := os.Getenv("AUTH_URL")
	authURL := fmt.Sprintf("%s/v2/token", authBaseURL)
	apiBaseURL := os.Getenv("API_URL")
	apiURL := fmt.Sprintf("%s/asset/v1/content/assets", apiBaseURL)
	return Config{
		AuthURL:      authURL,
		APIURL:       apiURL,
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		StoragePath:  os.Getenv("STORAGE_PATH"),
		S3Bucket:     os.Getenv("S3_BUCKET"),
		S3Region:     os.Getenv("S3_REGION"),
		S3AccessKey:  os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:  os.Getenv("S3_SECRET_KEY"),
	}
}
