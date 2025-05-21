package storage

import (
	"fmt"
	"sync"
)

// 存储服务类型常量
const (
	TypeAliOSS = "ali-oss"
	TypeMinIO  = "minio"
	// 将来可以添加更多存储类型，如:
	// TypeS3     = "s3"
)

// StorageFactoryFunc 存储服务工厂函数类型
type StorageFactoryFunc func(config StorageConfig) (StorageService, error)

// 全局存储服务工厂注册表
var (
	factoryMutex sync.RWMutex
	factories    = make(map[string]StorageFactoryFunc)
)

// RegisterStorageFactory 注册存储服务工厂
func RegisterStorageFactory(storageType string, factory StorageFactoryFunc) {
	factoryMutex.Lock()
	defer factoryMutex.Unlock()
	factories[storageType] = factory
}

// CreateStorageService 根据配置创建存储服务
func CreateStorageService(config StorageConfig) (StorageService, error) {
	factoryMutex.RLock()
	factory, exists := factories[config.GetType()]
	factoryMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("不支持的存储类型: %s", config.GetType())
	}

	return factory(config)
}
