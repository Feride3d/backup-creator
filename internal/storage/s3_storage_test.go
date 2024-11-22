package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Feride3d/backup-creator/internal/model"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUploader struct {
	mock.Mock
}

func (m *MockUploader) Upload(input *s3manager.UploadInput) (*s3manager.UploadOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*s3manager.UploadOutput), args.Error(1)
}

func NewTestS3Storage(mockUploader Uploader, bucket string) *S3Storage {
	return &S3Storage{
		Uploader: mockUploader,
		Bucket:   bucket,
	}
}

func TestS3Storage_SaveContentBlocks(t *testing.T) {
	mockUploader := new(MockUploader)
	bucket := "test-bucket"
	storage := NewTestS3Storage(mockUploader, bucket)

	blocks := []model.ContentBlock{
		{ID: 1, Name: "Block1", Content: "Content1"},
		{ID: 2, Name: "Block2", Content: "Content2"},
	}
	folder := "backup_20241121"

	for _, block := range blocks {
		data, _ := json.Marshal(block)
		mockUploader.On("Upload", &s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(fmt.Sprintf("%s/%d.json", folder, block.ID)),
			Body:   bytes.NewReader(data),
		}).Return(&s3manager.UploadOutput{}, nil)
	}

	ctx := context.Background()
	err := storage.SaveContentBlocks(ctx, blocks, folder)

	assert.NoError(t, err)
	mockUploader.AssertExpectations(t)
}

func TestS3Storage_SaveContentBlocks_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name          string
		blocks        []model.ContentBlock
		mockSetup     func(*MockUploader)
		expectedError string
	}{
		{
			name: "Uploader error",
			blocks: []model.ContentBlock{
				{ID: 2, Name: "ValidBlock", Content: "Some content"},
			},
			mockSetup: func(mockUploader *MockUploader) {
				mockUploader.On("Upload", mock.Anything).Return(&s3manager.UploadOutput{}, fmt.Errorf("mocked upload error"))
			},
			expectedError: "failed to upload block 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUploader := new(MockUploader)
			bucket := "test-bucket"
			storage := NewTestS3Storage(mockUploader, bucket)

			tt.mockSetup(mockUploader)

			ctx := context.Background()
			err := storage.SaveContentBlocks(ctx, tt.blocks, "backup_20241121")

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)

			mockUploader.AssertExpectations(t)
		})
	}
}

func TestS3Storage_SaveContentBlocks_MarshalError(t *testing.T) {
	mockUploader := new(MockUploader)
	bucket := "test-bucket"
	storage := NewTestS3Storage(mockUploader, bucket)

	blocks := []model.ContentBlock{
		{
			ID:      1,
			Name:    "InvalidBlock",
			Content: make(chan int),
		},
	}

	ctx := context.Background()
	err := storage.SaveContentBlocks(ctx, blocks, "backup_20241121")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal block 1")

	mockUploader.AssertNotCalled(t, "Upload", mock.Anything)
}

func TestNewS3Storage(t *testing.T) {
	tests := []struct {
		name        string
		region      string
		bucket      string
		accessKey   string
		secretKey   string
		expectError bool
	}{
		{
			name:        "Valid configuration",
			region:      "us-east-1",
			bucket:      "test-bucket",
			accessKey:   "valid-access-key",
			secretKey:   "valid-secret-key",
			expectError: false,
		},
		{
			name:        "Invalid region",
			region:      "",
			bucket:      "test-bucket",
			accessKey:   "valid-access-key",
			secretKey:   "valid-secret-key",
			expectError: true,
		},
		{
			name:        "Missing access key",
			region:      "us-east-1",
			bucket:      "test-bucket",
			accessKey:   "",
			secretKey:   "valid-secret-key",
			expectError: true,
		},
		{
			name:        "Missing secret key",
			region:      "us-east-1",
			bucket:      "test-bucket",
			accessKey:   "valid-access-key",
			secretKey:   "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewS3Storage(tt.region, tt.bucket, tt.accessKey, tt.secretKey)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, storage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, storage)
				assert.Equal(t, tt.bucket, storage.Bucket)
				assert.NotNil(t, storage.Uploader)
			}
		})
	}
}
