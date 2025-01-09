package minioctrl

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	TranslatedResourcesBucket = "translated-resources"
	TranslatedChunksBucket    = "translated-chunks"
)

type MinioService struct {
	client *minio.Client
}

func NewMinioService(endpoint, accessKeyID, secretAccessKey string, useSSL bool) (*MinioService, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %v", err)
	}

	return &MinioService{
		client: client,
	}, nil
}

func (s *MinioService) EnsureBucketExists(ctx context.Context, bucketName string) error {
	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %v", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %v", err)
		}
	}

	return nil
}

func (s *MinioService) GetObject(ctx context.Context, bucketName, objectName string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %v", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %v", err)
	}

	return data, nil
}

func (s *MinioService) PutObject(ctx context.Context, bucketName, objectName string, data []byte) error {
	reader := bytes.NewReader(data)
	_, err := s.client.PutObject(ctx, bucketName, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return fmt.Errorf("failed to put object: %v", err)
	}

	return nil
}

func (s *MinioService) DeleteObject(ctx context.Context, bucketName, objectName string) error {
	err := s.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %v", err)
	}

	return nil
}

func (s *MinioService) DeleteObjects(ctx context.Context, bucketName string, objectNames []string) error {
	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)
		for _, name := range objectNames {
			objectsCh <- minio.ObjectInfo{
				Key: name,
			}
		}
	}()

	for err := range s.client.RemoveObjects(ctx, bucketName, objectsCh, minio.RemoveObjectsOptions{}) {
		if err.Err != nil {
			return fmt.Errorf("failed to delete object %s: %v", err.ObjectName, err.Err)
		}
	}

	return nil
}

func (s *MinioService) GetBucketAndObjectFromURL(minioURL string) (string, string) {
	// MinioURL format: bucket-name/object-name
	parts := strings.SplitN(minioURL, "/", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}
