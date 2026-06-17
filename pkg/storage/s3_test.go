package storage

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
)

type mockMinioClient struct {
	bucketExistsFn func(ctx context.Context, bucketName string) (bool, error)
	makeBucketFn   func(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error
	putObjectFn    func(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	getObjectFn    func(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error)
}

func (m *mockMinioClient) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return m.bucketExistsFn(ctx, bucketName)
}

func (m *mockMinioClient) MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	return m.makeBucketFn(ctx, bucketName, opts)
}

func (m *mockMinioClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return m.putObjectFn(ctx, bucketName, objectName, reader, objectSize, opts)
}

func (m *mockMinioClient) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	if m.getObjectFn != nil {
		return m.getObjectFn(ctx, bucketName, objectName, opts)
	}
	return nil, errors.New("not mocked")
}

func TestNewS3_InvalidEndpoint(t *testing.T) {
	_, err := NewS3("", "", "", "", false)
	assert.Error(t, err)
}

func TestNewS3_BucketExistsError(t *testing.T) {
	mc := &mockMinioClient{
		bucketExistsFn: func(ctx context.Context, bucketName string) (bool, error) {
			return false, errors.New("bucket check failed")
		},
	}
	old := newMinioClient
	newMinioClient = func(endpoint string, opts *minio.Options) (minioClient, error) {
		return mc, nil
	}
	defer func() { newMinioClient = old }()

	_, err := NewS3("s3.example.com", "access", "secret", "bucket", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bucket check failed")
}

func TestNewS3_MakeBucketError(t *testing.T) {
	mc := &mockMinioClient{
		bucketExistsFn: func(ctx context.Context, bucketName string) (bool, error) {
			return false, nil
		},
		makeBucketFn: func(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
			return errors.New("make bucket failed")
		},
	}
	old := newMinioClient
	newMinioClient = func(endpoint string, opts *minio.Options) (minioClient, error) {
		return mc, nil
	}
	defer func() { newMinioClient = old }()

	_, err := NewS3("s3.example.com", "access", "secret", "bucket", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "make bucket failed")
}

func TestNewS3_Success_WithExistingBucket(t *testing.T) {
	mc := &mockMinioClient{
		bucketExistsFn: func(ctx context.Context, bucketName string) (bool, error) {
			assert.Equal(t, "bucket", bucketName)
			return true, nil
		},
	}
	old := newMinioClient
	newMinioClient = func(endpoint string, opts *minio.Options) (minioClient, error) {
		return mc, nil
	}
	defer func() { newMinioClient = old }()

	c, err := NewS3("s3.example.com", "access", "secret", "bucket", true)
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, "bucket", c.bucketName)
	assert.Equal(t, "s3.example.com", c.endpoint)
	assert.True(t, c.useSSL)
}

func TestNewS3_Success_CreatesBucket(t *testing.T) {
	var makeBucketCalled bool
	mc := &mockMinioClient{
		bucketExistsFn: func(ctx context.Context, bucketName string) (bool, error) {
			return false, nil
		},
		makeBucketFn: func(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
			makeBucketCalled = true
			return nil
		},
	}
	old := newMinioClient
	newMinioClient = func(endpoint string, opts *minio.Options) (minioClient, error) {
		return mc, nil
	}
	defer func() { newMinioClient = old }()

	c, err := NewS3("s3.example.com", "access", "secret", "bucket", false)
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.True(t, makeBucketCalled)
}

func TestS3Client_Upload_Error(t *testing.T) {
	mc := &mockMinioClient{
		putObjectFn: func(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
			return minio.UploadInfo{}, errors.New("put failed")
		},
	}
	s3 := &S3Client{client: mc, bucketName: "bucket", endpoint: "s3.example.com", useSSL: false}
	url, err := s3.Upload(strings.NewReader("data"), "test.txt", 4, "text/plain")
	assert.Error(t, err)
	assert.Empty(t, url)
	assert.Contains(t, err.Error(), "put failed")
}

func TestS3Client_Upload_Success(t *testing.T) {
	mc := &mockMinioClient{
		putObjectFn: func(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
			return minio.UploadInfo{}, nil
		},
	}
	s3 := &S3Client{client: mc, bucketName: "bucket", endpoint: "s3.example.com", useSSL: true}
	url, err := s3.Upload(strings.NewReader("data"), "test.txt", 4, "text/plain")
	assert.NoError(t, err)
	assert.Equal(t, "https://s3.example.com/bucket/test.txt", url)
}

func TestS3Client_Upload_SuccessNoSSL(t *testing.T) {
	mc := &mockMinioClient{
		putObjectFn: func(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
			return minio.UploadInfo{}, nil
		},
	}
	s3 := &S3Client{client: mc, bucketName: "bucket", endpoint: "s3.example.com", useSSL: false}
	url, err := s3.Upload(strings.NewReader("data"), "test.txt", 4, "text/plain")
	assert.NoError(t, err)
	assert.Equal(t, "http://s3.example.com/bucket/test.txt", url)
}

func TestS3Client_Download_Error(t *testing.T) {
	mc := &mockMinioClient{
		getObjectFn: func(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
			return nil, errors.New("download failed")
		},
	}
	s3 := &S3Client{client: mc, bucketName: "bucket", endpoint: "s3.example.com", useSSL: false}
	_, err := s3.Download("test.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "download failed")
}
