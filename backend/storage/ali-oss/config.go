package alioss

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// LoadAliOSSConfigFromEnv 从环境变量加载OSS配置
func LoadAliOSSConfigFromEnv() (*AliOSSConfig, error) {
	// 尝试加载.env文件，但不强制要求
	_ = godotenv.Load()

	config := &AliOSSConfig{
		AccessKeyID:     os.Getenv("OSS_ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("OSS_ACCESS_KEY_SECRET"),
		Region:          os.Getenv("OSS_REGION"),
		BucketName:      os.Getenv("OSS_BUCKET_NAME"),
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("OSS配置验证失败: %w", err)
	}

	return config, nil
}
