package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort     string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3UseSSL    bool
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()

	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "waste_collection")
	viper.SetDefault("S3_ENDPOINT", "localhost:9000")
	viper.SetDefault("S3_ACCESS_KEY", "minioadmin")
	viper.SetDefault("S3_SECRET_KEY", "minioadmin")
	viper.SetDefault("S3_BUCKET", "payments")
	viper.SetDefault("S3_USE_SSL", "false")

	return &Config{
		AppPort:     viper.GetString("APP_PORT"),
		DBHost:      viper.GetString("DB_HOST"),
		DBPort:      viper.GetString("DB_PORT"),
		DBUser:      viper.GetString("DB_USER"),
		DBPassword:  viper.GetString("DB_PASSWORD"),
		DBName:      viper.GetString("DB_NAME"),
		S3Endpoint:  viper.GetString("S3_ENDPOINT"),
		S3AccessKey: viper.GetString("S3_ACCESS_KEY"),
		S3SecretKey: viper.GetString("S3_SECRET_KEY"),
		S3Bucket:    viper.GetString("S3_BUCKET"),
		S3UseSSL:    viper.GetString("S3_USE_SSL") == "true",
	}, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Jakarta",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}
