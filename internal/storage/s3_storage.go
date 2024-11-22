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

type Uploader interface {
	Upload(input *s3manager.UploadInput) (*s3manager.UploadOutput, error)
}

type S3Storage struct {
	Uploader Uploader
	Bucket   string
}

type S3Uploader struct {
	uploader *s3manager.Uploader
}

func NewS3Uploader(uploader *s3manager.Uploader) *S3Uploader {
	return &S3Uploader{uploader: uploader}
}

func (a *S3Uploader) Upload(input *s3manager.UploadInput) (*s3manager.UploadOutput, error) {
	return a.uploader.Upload(input)
}

func NewS3Storage(region, bucket, accessKey, secretKey string) (*S3Storage, error) {
	if region == "" {
		return nil, fmt.Errorf("region cannot be empty")
	}
	if accessKey == "" {
		return nil, fmt.Errorf("accessKey cannot be empty")
	}
	if secretKey == "" {
		return nil, fmt.Errorf("secretKey cannot be empty")
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}
	realUploader := s3manager.NewUploader(sess)

	return &S3Storage{
		Uploader: NewS3Uploader(realUploader),
		Bucket:   bucket,
	}, nil
}

// SaveContentBlocks uploads content blocks to S3
func (s *S3Storage) SaveContentBlocks(ctx context.Context, blocks []model.ContentBlock, folder string) error {
	for _, block := range blocks {
		data, err := json.Marshal(block)
		if err != nil {
			return fmt.Errorf("failed to marshal block %d: %v", block.ID, err)
		}

		key := fmt.Sprintf("%s/%d.json", folder, block.ID)
		_, err = s.Uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(s.Bucket),
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
