package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/Feride3d/backup-creator/internal/model"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Storage struct {
	uploader *s3manager.Uploader
	bucket   string
}

func NewS3Storage(region, bucket, accessKey, secretKey string) *S3Storage {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	}))

	return &S3Storage{
		uploader: s3manager.NewUploader(sess),
		bucket:   bucket,
	}
}

// SaveContentBlocks uploads content blocks to S3
func (s *S3Storage) SaveContentBlocks(ctx context.Context, blocks []model.ContentBlock, folder string) error {
	for _, block := range blocks {
		data, err := json.Marshal(block)
		if err != nil {
			return fmt.Errorf("failed to marshal block %d: %v", block.ID, err)
		}

		key := fmt.Sprintf("%s/%d.json", folder, block.ID)
		_, err = s.uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
			Body:   bytes.NewReader(data),
		})
		if err != nil {
			return fmt.Errorf("failed to upload block %d: %v", block.ID, err)
		}

		fmt.Printf("Uploaded content block %d to S3 as %s\n", block.ID, key)
	}
	return nil
}
