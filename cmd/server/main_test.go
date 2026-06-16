package main

import (
	"errors"
	"os/signal"
	"syscall"
	"testing"

	"github.com/faisalaffan/community-waste-collection-api/internal/config"
	"github.com/faisalaffan/community-waste-collection-api/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRun_ConfigError(t *testing.T) {
	err := run(
		func() (*config.Config, error) {
			return nil, errors.New("config load failed")
		},
		nil, // not reached
		nil, // not reached
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load config")
}

func TestRun_DBOpenError(t *testing.T) {
	err := run(
		func() (*config.Config, error) {
			return &config.Config{}, nil
		},
		func(dsn string) (*gorm.DB, error) {
			return nil, errors.New("db connection failed")
		},
		nil, // not reached
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

func TestRun_ListenError(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	defer signal.Reset(syscall.SIGINT, syscall.SIGTERM)

	err = run(
		func() (*config.Config, error) {
			return &config.Config{AppPort: "-1"}, nil
		},
		func(dsn string) (*gorm.DB, error) {
			return db, nil
		},
		func(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*storage.S3Client, error) {
			return &storage.S3Client{}, nil
		},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server error")
}

func TestRun_S3Error(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = run(
		func() (*config.Config, error) {
			return &config.Config{}, nil
		},
		func(dsn string) (*gorm.DB, error) {
			return db, nil
		},
		func(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*storage.S3Client, error) {
			return nil, errors.New("s3 connection failed")
		},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to S3")
}
