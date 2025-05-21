package alioss

import (
	"go-uploader/storage"
)

// 注册阿里云OSS工厂
func init() {
	// 创建工厂函数
	factory := func(config storage.StorageConfig) (storage.StorageService, error) {
		ossConfig, ok := config.(*AliOSSConfig)
		if !ok {
			return nil, storage.ErrInvalidConfig
		}
		return NewAliOSSService(ossConfig)
	}

	// 注册到全局工厂
	storage.RegisterStorageFactory("ali-oss", factory)
}
