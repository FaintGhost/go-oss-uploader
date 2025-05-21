package minio

import (
	"go-uploader/storage"
)

// 在init函数中注册MinIO存储服务工厂
func init() {
	// 创建工厂函数
	factory := func(config storage.StorageConfig) (storage.StorageService, error) {
		minioConfig, ok := config.(*MinioConfig)
		if !ok {
			return nil, storage.ErrInvalidConfig
		}
		return NewMinioService(minioConfig)
	}

	// 注册到全局工厂
	storage.RegisterStorageFactory("minio", factory)
}
