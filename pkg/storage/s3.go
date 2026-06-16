package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// newMinioClient is overridable in tests.
var newMinioClient = func(endpoint string, opts *minio.Options) (minioClient, error) {
	return minio.New(endpoint, opts)
}

// minioClient abstracts the minio.Client methods used by this package.
type minioClient interface {
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
}

type FileStorage interface {
	Upload(r io.Reader, objectName string, size int64, contentType string) (string, error)
}

var _ FileStorage = (*S3Client)(nil)

type S3Client struct {
	client     minioClient
	bucketName string
	endpoint   string
	useSSL     bool
}

func NewS3(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*S3Client, error) {
	client, err := newMinioClient(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &S3Client{client: client, bucketName: bucket, endpoint: endpoint, useSSL: useSSL}, nil
}

func (s *S3Client) Upload(r io.Reader, objectName string, size int64, contentType string) (string, error) {
	ctx := context.Background()
	opts := minio.PutObjectOptions{ContentType: contentType}
	_, err := s.client.PutObject(ctx, s.bucketName, objectName, r, size, opts)
	if err != nil {
		return "", err
	}

	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s/%s/%s", scheme, s.endpoint, s.bucketName, objectName)
	return url, nil
}
