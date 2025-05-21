package minio

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"go-uploader/storage"
)

// MinioService MinIO存储服务实现
type MinioService struct {
	client *minio.Client
	config *MinioConfig
}

// NewMinioService 创建新的MinIO存储服务
func NewMinioService(config *MinioConfig) (*MinioService, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// 初始化MinIO客户端
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("初始化MinIO客户端失败: %w", err)
	}

	// 检查并确保存储桶存在
	exists, err := client.BucketExists(context.Background(), config.BucketName)
	if err != nil {
		return nil, fmt.Errorf("检查存储桶失败: %w", err)
	}

	if !exists {
		// 尝试创建存储桶
		err = client.MakeBucket(context.Background(), config.BucketName, minio.MakeBucketOptions{
			Region: config.Region,
		})
		if err != nil {
			return nil, fmt.Errorf("创建存储桶失败: %w", err)
		}
	}

	return &MinioService{
		client: client,
		config: config,
	}, nil
}

// 创建一个自定义的读取器以支持进度回调
type progressReader struct {
	io.Reader
	progressFn storage.ProgressCallback
	total      int64
	current    int64
	lastUpdate int64
}

func (r *progressReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if n > 0 && r.progressFn != nil {
		r.current += int64(n)
		// 更新进度，但避免过于频繁的更新
		if r.current-r.lastUpdate > r.total/100 || r.current == r.total {
			increment := r.current - r.lastUpdate
			r.progressFn(increment, r.current, r.total)
			r.lastUpdate = r.current
		}
	}
	return
}

// UploadFile 上传文件到MinIO
func (s *MinioService) UploadFile(ctx context.Context, objectName string, localFile string, progressFn storage.ProgressCallback) (interface{}, error) {
	// 检查文件是否存在和可访问
	fileInfo, err := os.Stat(localFile)
	if err != nil {
		return nil, fmt.Errorf("文件访问错误: %w", err)
	}

	// 打开文件
	file, err := os.Open(localFile)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %w", err)
	}
	defer file.Close()

	// 创建进度读取器
	reader := &progressReader{
		Reader:     file,
		progressFn: progressFn,
		total:      fileInfo.Size(),
	}

	// 确保对象名称没有前导斜杠
	objectName = strings.TrimPrefix(objectName, "/")

	// 获取文件的MIME类型
	contentType := "application/octet-stream"
	ext := filepath.Ext(localFile)
	if ext != "" {
		switch strings.ToLower(ext) {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		case ".pdf":
			contentType = "application/pdf"
		case ".txt":
			contentType = "text/plain"
		case ".html", ".htm":
			contentType = "text/html"
		case ".mp4":
			contentType = "video/mp4"
		case ".mp3":
			contentType = "audio/mpeg"
		}
	}

	// 上传文件
	info, err := s.client.PutObject(ctx, s.config.BucketName, objectName, reader, fileInfo.Size(),
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	return info, nil
}

// IsObjectExist 检查对象是否存在于MinIO
func (s *MinioService) IsObjectExist(ctx context.Context, objectName string) (bool, error) {
	// 确保对象名称没有前导斜杠
	objectName = strings.TrimPrefix(objectName, "/")

	// 尝试获取对象信息
	_, err := s.client.StatObject(ctx, s.config.BucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		// 检查错误类型
		if errResp, ok := err.(minio.ErrorResponse); ok {
			if errResp.Code == "NoSuchKey" || errResp.Code == "NotFound" {
				return false, nil // 对象不存在，但不是错误
			}
		}
		return false, fmt.Errorf("检查对象是否存在失败: %w", err)
	}

	return true, nil // 对象存在
}

// GeneratePresignedURL 生成预签名上传URL
func (s *MinioService) GeneratePresignedURL(ctx context.Context, objectName string, expiration time.Duration) (string, map[string]string, error) {
	// 确保对象名称没有前导斜杠
	objectName = strings.TrimPrefix(objectName, "/")

	// 生成预签名上传URL
	presignedURL, err := s.client.PresignedPutObject(ctx, s.config.BucketName, objectName, expiration)
	if err != nil {
		return "", nil, fmt.Errorf("生成预签名上传URL失败: %w", err)
	}

	// MinIO不像阿里云OSS那样提供签名头，返回一个空的头部映射
	headers := make(map[string]string)

	return presignedURL.String(), headers, nil
}

// GeneratePresignedDownloadURL 生成预签名下载URL
func (s *MinioService) GeneratePresignedDownloadURL(ctx context.Context, objectName string, expiration time.Duration) (string, map[string]string, error) {
	// 确保对象名称没有前导斜杠
	objectName = strings.TrimPrefix(objectName, "/")

	// 设置请求参数，包括下载时的文件名
	reqParams := make(url.Values)
	fileName := filepath.Base(objectName)
	reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

	// 生成预签名下载URL
	presignedURL, err := s.client.PresignedGetObject(ctx, s.config.BucketName, objectName, expiration, reqParams)
	if err != nil {
		return "", nil, fmt.Errorf("生成预签名下载URL失败: %w", err)
	}

	// MinIO不像阿里云OSS那样提供签名头，返回一个空的头部映射
	headers := make(map[string]string)

	return presignedURL.String(), headers, nil
}

// GetBucketDomain 获取存储桶的域名
func (s *MinioService) GetBucketDomain() string {
	protocol := "http"
	if s.config.UseSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s", protocol, s.config.Endpoint, s.config.BucketName)
}

// MinioFactory MinIO服务工厂
type MinioFactory struct{}

// NewMinioFactory 创建MinIO服务工厂
func NewMinioFactory() *MinioFactory {
	return &MinioFactory{}
}

// CreateStorageService 根据配置创建存储服务
func (f *MinioFactory) CreateStorageService(config storage.StorageConfig) (storage.StorageService, error) {
	minioConfig, ok := config.(*MinioConfig)
	if !ok {
		return nil, fmt.Errorf("配置类型错误，需要 MinioConfig 类型")
	}

	return NewMinioService(minioConfig)
}
