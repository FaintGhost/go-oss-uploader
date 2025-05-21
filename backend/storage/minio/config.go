package minio

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// MinioConfig MinIO存储配置
type MinioConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
	Region          string
}

// GetType 返回存储类型标识
func (c *MinioConfig) GetType() string {
	return "minio"
}

// Validate 验证配置的合法性
func (c *MinioConfig) Validate() error {
	if c.Endpoint == "" || c.AccessKeyID == "" ||
		c.SecretAccessKey == "" || c.BucketName == "" {
		return fmt.Errorf("缺少必要的MinIO配置")
	}
	return nil
}

// LoadMinioConfigFromEnv 从环境变量加载MinIO配置
func LoadMinioConfigFromEnv() (*MinioConfig, error) {
	// 尝试加载.env文件，但不强制要求
	_ = godotenv.Load()

	// 读取SSL配置，默认为true
	useSSL := true
	if sslStr := os.Getenv("MINIO_USE_SSL"); sslStr != "" {
		if boolVal, err := strconv.ParseBool(sslStr); err == nil {
			useSSL = boolVal
		}
	}

	config := &MinioConfig{
		Endpoint:        os.Getenv("MINIO_ENDPOINT"),
		AccessKeyID:     os.Getenv("MINIO_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("MINIO_SECRET_ACCESS_KEY"),
		UseSSL:          useSSL,
		BucketName:      os.Getenv("MINIO_BUCKET_NAME"),
		Region:          os.Getenv("MINIO_REGION"),
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("MinIO配置验证失败: %w", err)
	}

	return config, nil
}
