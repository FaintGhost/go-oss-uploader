# 存储服务配置说明

go-uploader 支持多种对象存储服务，目前已实现：

1. 阿里云OSS
2. MinIO

## 配置方式

存储服务配置通过环境变量设置，可以通过 `.env` 文件或系统环境变量设置。

### 选择存储服务

您可以通过以下两种方式选择存储服务类型：

1. 通过命令行参数（优先级高）：
```bash
# 使用阿里云OSS
./go-uploader --storage=ali-oss

# 使用MinIO
./go-uploader --storage=minio
```

2. 通过环境变量（当命令行参数未指定时使用）：
```
# 在.env文件中设置
OSS=ali-oss
# 或
OSS=minio
```

如果两种方式都未指定，默认使用阿里云OSS。

### 阿里云OSS配置

```
OSS_ACCESS_KEY_ID=您的AccessKey ID
OSS_ACCESS_KEY_SECRET=您的AccessKey Secret
OSS_REGION=oss-cn-hangzhou
OSS_BUCKET_NAME=您的存储桶名称
```

### MinIO配置

```
MINIO_ENDPOINT=play.min.io
MINIO_ACCESS_KEY_ID=您的AccessKey ID
MINIO_SECRET_ACCESS_KEY=您的Secret Key
MINIO_USE_SSL=true
MINIO_BUCKET_NAME=您的存储桶名称
MINIO_REGION=us-east-1
```

## MinIO服务搭建

如果需要自行搭建MinIO服务，可以使用Docker快速启动：

```bash
docker run -p 9000:9000 -p 9001:9001 \
  --name minio \
  -v /path/to/data:/data \
  -e "MINIO_ROOT_USER=minioadmin" \
  -e "MINIO_ROOT_PASSWORD=minioadmin" \
  quay.io/minio/minio server /data --console-address ":9001"
```

访问 `http://localhost:9001` 可打开MinIO控制台，使用以下配置连接到本地MinIO服务：

```
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY_ID=minioadmin
MINIO_SECRET_ACCESS_KEY=minioadmin
MINIO_USE_SSL=false
MINIO_BUCKET_NAME=uploads
MINIO_REGION=us-east-1
```

## 添加新的存储服务

如需添加新的存储服务支持，请按照以下步骤操作：

1. 在 `storage` 目录下创建新的子目录，例如 `storage/s3`
2. 实现 `StorageConfig` 和 `StorageService` 接口
3. 在 `init.go` 中注册存储服务工厂
4. 在 `main.go` 中添加对应的配置加载逻辑

请参考现有的阿里云OSS或MinIO实现作为参考。 