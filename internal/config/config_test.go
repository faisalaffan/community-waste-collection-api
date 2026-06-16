package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	os.Setenv("APP_PORT", "9090")
	os.Setenv("DB_HOST", "db-test")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "user-test")
	os.Setenv("DB_PASSWORD", "pass-test")
	os.Setenv("DB_NAME", "db-test")
	os.Setenv("S3_ENDPOINT", "s3-test:9000")
	os.Setenv("S3_ACCESS_KEY", "ak-test")
	os.Setenv("S3_SECRET_KEY", "sk-test")
	os.Setenv("S3_BUCKET", "bucket-test")
	os.Setenv("S3_USE_SSL", "true")
	defer os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.AppPort != "9090" {
		t.Errorf("AppPort = %s, want 9090", cfg.AppPort)
	}
	if cfg.DBHost != "db-test" {
		t.Errorf("DBHost = %s, want db-test", cfg.DBHost)
	}
	if cfg.DBPort != "5433" {
		t.Errorf("DBPort = %s, want 5433", cfg.DBPort)
	}
	if cfg.DBUser != "user-test" {
		t.Errorf("DBUser = %s, want user-test", cfg.DBUser)
	}
	if cfg.S3Endpoint != "s3-test:9000" {
		t.Errorf("S3Endpoint = %s, want s3-test:9000", cfg.S3Endpoint)
	}
	if cfg.S3AccessKey != "ak-test" {
		t.Errorf("S3AccessKey = %s, want ak-test", cfg.S3AccessKey)
	}
	if !cfg.S3UseSSL {
		t.Errorf("S3UseSSL = false, want true")
	}
	if cfg.S3Bucket != "bucket-test" {
		t.Errorf("S3Bucket = %s, want bucket-test", cfg.S3Bucket)
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.AppPort != "8080" {
		t.Errorf("default AppPort = %s, want 8080", cfg.AppPort)
	}
	if cfg.DBHost != "localhost" {
		t.Errorf("default DBHost = %s, want localhost", cfg.DBHost)
	}
	if cfg.DBPort != "5432" {
		t.Errorf("default DBPort = %s, want 5432", cfg.DBPort)
	}
	if cfg.S3Endpoint != "localhost:9000" {
		t.Errorf("default S3Endpoint = %s, want localhost:9000", cfg.S3Endpoint)
	}
	if cfg.S3UseSSL {
		t.Errorf("default S3UseSSL = true, want false")
	}
}

func TestDSN(t *testing.T) {
	cfg := &Config{
		DBHost:     "pg",
		DBPort:     "5432",
		DBUser:     "admin",
		DBPassword: "secret",
		DBName:     "waste",
	}
	dsn := cfg.DSN()
	expected := "host=pg port=5432 user=admin password=secret dbname=waste sslmode=disable TimeZone=Asia/Jakarta"
	if dsn != expected {
		t.Errorf("DSN = %s, want %s", dsn, expected)
	}
}
