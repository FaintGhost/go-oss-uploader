package storage

import (
	"context"
	"time"
)

// StorageConfig 存储配置接口
type StorageConfig interface {
	// GetType 返回存储类型标识
	GetType() string

	// Validate 验证配置的合法性
	Validate() error
}

// ProgressCallback 上传进度回调函数
type ProgressCallback func(increment, transferred, total int64)

// StorageService 存储服务接口
type StorageService interface {
	// UploadFile 上传文件
	UploadFile(ctx context.Context, objectName string, localFile string, progressFn ProgressCallback) (interface{}, error)

	// IsObjectExist 检查对象是否存在
	IsObjectExist(ctx context.Context, objectName string) (bool, error)

	// GeneratePresignedURL 生成预签名上传URL
	GeneratePresignedURL(ctx context.Context, objectName string, expiration time.Duration) (string, map[string]string, error)

	// GeneratePresignedDownloadURL 生成预签名下载URL
	GeneratePresignedDownloadURL(ctx context.Context, objectName string, expiration time.Duration) (string, map[string]string, error)

	// GetBucketDomain 获取存储桶的域名
	GetBucketDomain() string
}

// StorageServiceFactory 存储服务工厂接口
type StorageServiceFactory interface {
	// CreateStorageService 根据配置创建存储服务
	CreateStorageService(config StorageConfig) (StorageService, error)
}
