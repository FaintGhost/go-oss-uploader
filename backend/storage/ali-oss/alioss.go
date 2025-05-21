package alioss

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"

	"go-uploader/storage"
)

// AliOSSConfig 阿里云OSS配置
type AliOSSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	Region          string
	BucketName      string
}

// GetType 返回存储类型标识
func (c *AliOSSConfig) GetType() string {
	return "ali-oss"
}

// Validate 验证配置的合法性
func (c *AliOSSConfig) Validate() error {
	if c.AccessKeyID == "" || c.AccessKeySecret == "" ||
		c.Region == "" || c.BucketName == "" {
		return fmt.Errorf("缺少必要的OSS配置")
	}
	return nil
}

// AliOSSService 阿里云OSS存储服务实现
type AliOSSService struct {
	client   *oss.Client
	uploader *oss.Uploader
	config   *AliOSSConfig
}

// NewAliOSSService 创建新的阿里云OSS存储服务
func NewAliOSSService(config *AliOSSConfig) (*AliOSSService, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// 加载默认配置并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.AccessKeyID, config.AccessKeySecret)).
		WithRegion(config.Region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建上传管理器
	uploader := createUploader(client)

	return &AliOSSService{
		client:   client,
		uploader: uploader,
		config:   config,
	}, nil
}

// 创建上传管理器
func createUploader(client *oss.Client) *oss.Uploader {
	return client.NewUploader(func(uo *oss.UploaderOptions) {
		uo.PartSize = 5 * 1024 * 1024      // 设置分片大小为5MB
		uo.ParallelNum = 3                 // 设置并行上传数为3
		uo.EnableCheckpoint = true         // 启用断点续传
		uo.CheckpointDir = "./checkpoint/" // 断点记录文件保存路径
	})
}

// UploadFile 上传文件到OSS
func (s *AliOSSService) UploadFile(ctx context.Context, objectName string, localFile string, progressFn storage.ProgressCallback) (interface{}, error) {
	// 检查文件是否存在和可访问
	if _, err := os.Stat(localFile); err != nil {
		return nil, fmt.Errorf("文件访问错误: %v", err)
	}

	// 创建一个自定义进度回调，用于WebSocket进度更新
	progressCallback := func(increment, transferred, total int64) {
		if progressFn != nil {
			progressFn(increment, transferred, total)
		}
	}

	// 创建上传对象的请求
	putRequest := &oss.PutObjectRequest{
		Bucket:       oss.Ptr(s.config.BucketName),
		Key:          oss.Ptr(objectName),
		StorageClass: oss.StorageClassStandard,
		Acl:          oss.ObjectACLPrivate,
		ProgressFn:   progressCallback,
	}

	// 使用Uploader上传文件
	result, err := s.uploader.UploadFile(ctx, putRequest, localFile)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// IsObjectExist 检查对象是否存在于OSS
func (s *AliOSSService) IsObjectExist(ctx context.Context, objectName string) (bool, error) {
	// 调用OSS API检查对象是否存在
	exists, err := s.client.IsObjectExist(ctx, s.config.BucketName, objectName)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// GeneratePresignedURL 生成预签名上传URL
func (s *AliOSSService) GeneratePresignedURL(ctx context.Context, objectName string, expiration time.Duration) (string, map[string]string, error) {
	// 创建上传对象的请求
	putRequest := &oss.PutObjectRequest{
		Bucket:       oss.Ptr(s.config.BucketName), // 存储空间名称
		Key:          oss.Ptr(objectName),          // 对象名称
		StorageClass: oss.StorageClassStandard,     // 指定对象的存储类型为标准存储
		Acl:          oss.ObjectACLPrivate,         // 指定对象的访问权限为私有访问
	}

	// 生成预签名URL
	result, err := s.client.Presign(ctx, putRequest, oss.PresignExpires(expiration))
	if err != nil {
		return "", nil, err
	}

	return result.URL, result.SignedHeaders, nil
}

// GeneratePresignedDownloadURL 生成预签名下载URL
func (s *AliOSSService) GeneratePresignedDownloadURL(ctx context.Context, objectName string, expiration time.Duration) (string, map[string]string, error) {
	// 创建获取对象的请求
	getRequest := &oss.GetObjectRequest{
		Bucket: oss.Ptr(s.config.BucketName), // 存储空间名称
		Key:    oss.Ptr(objectName),          // 对象名称
	}

	// 生成预签名URL
	result, err := s.client.Presign(ctx, getRequest, oss.PresignExpires(expiration))
	if err != nil {
		return "", nil, err
	}

	return result.URL, result.SignedHeaders, nil
}

// GetBucketDomain 获取存储桶的域名
func (s *AliOSSService) GetBucketDomain() string {
	return fmt.Sprintf("%s.%s.aliyuncs.com", s.config.BucketName, s.config.Region)
}

// AliOSSFactory 阿里云OSS服务工厂
type AliOSSFactory struct{}

// NewAliOSSFactory 创建阿里云OSS服务工厂
func NewAliOSSFactory() *AliOSSFactory {
	return &AliOSSFactory{}
}

// CreateStorageService 根据配置创建存储服务
func (f *AliOSSFactory) CreateStorageService(config storage.StorageConfig) (storage.StorageService, error) {
	ossConfig, ok := config.(*AliOSSConfig)
	if !ok {
		return nil, fmt.Errorf("配置类型错误，需要 AliOSSConfig 类型")
	}

	return NewAliOSSService(ossConfig)
}
