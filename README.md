# 阿里云OSS文件上传服务

这是一个使用Go语言和Gin框架实现的阿里云OSS文件上传服务。

## 功能特点

- 使用Gin框架实现Web API
- 支持两种上传方式：普通上传和预签名上传
- 支持生成可分享的预签名下载链接
- 支持选择文件夹批量上传文件
- 实时显示文件上传进度
- 使用WebSocket实现进度实时更新
- 从环境变量加载配置，更安全
- 提供简单的Web界面用于测试
- 支持详细日志记录模式
- 上传性能对比统计
- 支持断点续传功能，适合大文件上传
- 智能检测重复文件，避免重复上传
- 完整的Docker支持，便于部署

## 上传方式对比

| 特性 | 普通上传 | 预签名上传 |
|------|---------|-----------|
| 路径 | 客户端 → 服务器 → OSS | 客户端 → OSS |
| 服务器负载 | 较高 | 较低 |
| 适用场景 | 小文件，需要服务器处理 | 大文件，无需服务器处理 |
| 进度跟踪 | WebSocket实时反馈 | 浏览器原生进度反馈 |
| 安全性 | 服务端完全控制 | 通过临时签名授权 |
| 文件夹上传 | 支持批量上传文件夹中所有文件 | 支持批量上传文件夹中所有文件 |
| 分享下载 | 支持生成预签名下载链接 | 支持生成预签名下载链接 |

## 配置说明

在根目录创建`.env`文件，包含以下配置：

```
OSS_ACCESS_KEY_ID=您的AccessKeyId
OSS_ACCESS_KEY_SECRET=您的AccessKeySecret
OSS_REGION=您的OSS区域（例如：oss-cn-hangzhou）
OSS_BUCKET_NAME=您的存储桶名称
```

## 运行方法

### 原生方式

1. 确保已安装Go 1.16+
2. 填写正确的`.env`配置
3. 运行服务：

```bash
# 标准模式运行
go run main.go

# 详细日志模式运行
go run main.go --verbose
```

### Docker方式

#### 使用Docker

```bash
# 构建Docker镜像
docker build -t go-uploader .

# 运行Docker容器
docker run -d --name go-uploader \
  -p 5050:5050 \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/temp:/app/temp \
  -v $(pwd)/checkpoint:/app/checkpoint \
  --env-file .env \
  go-uploader
```

#### 使用Docker Compose

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

4. 访问 http://localhost:5050 使用Web界面测试上传功能

## Docker卷

Docker配置定义了以下卷挂载，用于数据持久化：

- `./logs:/app/logs` - 存储应用程序日志
- `./temp:/app/temp` - 存储临时上传的文件
- `./checkpoint:/app/checkpoint` - 存储断点续传的检查点信息

## 日志记录

应用程序会生成两个日志文件：

- `logs/app.log`: 包含应用程序的详细运行日志
- `logs/gin.log`: 包含HTTP请求的处理日志

使用 `--verbose` 参数启动程序（或使用默认Docker配置）时，日志信息会同时输出到控制台，便于调试。

## 健康检查

API包含一个健康检查端点：

- **URL**: `/health`
- **方法**: `GET`
- **响应**: 返回服务状态和当前时间

Docker Compose配置中已包含健康检查设置。

## API接口

### 普通上传文件

- **URL**: `/upload`
- **方法**: `POST`
- **Content-Type**: `multipart/form-data`
- **参数**:
  - `file`: 要上传的文件
  - `uploadID`: 上传任务的唯一标识符，用于WebSocket进度追踪

#### 响应示例：

```json
{
  "message": "File uploaded successfully",
  "filename": "example.jpg",
  "size": 1024,
  "url": "https://your-bucket.oss-region.aliyuncs.com/example.jpg",
  "result": {}
}
```

### 生成预签名下载URL

- **URL**: `/download/:filename`
- **方法**: `GET`
- **参数**:
  - `:filename`: 文件名称（OSS中的对象键）
  - `expiration`: 链接有效期，例如：`1h`、`24h`、`7d`，默认为`24h`

#### 响应示例：

```json
{
  "url": "https://your-bucket.oss-region.aliyuncs.com/example.jpg?Expires=1675456789&Signature=xxx",
  "headers": {
    "Host": "your-bucket.oss-region.aliyuncs.com"
  },
  "expiration": "2023-01-01T00:10:00Z",
  "method": "GET",
  "fileName": "example.jpg"
}
```

### 生成预签名URL

- **URL**: `/presign`
- **方法**: `POST`
- **Content-Type**: `multipart/form-data`
- **参数**:
  - `fileName`: 文件名称（用于生成OSS对象键）

#### 响应示例：

```json
{
  "url": "https://your-bucket.oss-region.aliyuncs.com/example.jpg?Expires=1675456789&Signature=xxx",
  "headers": {
    "Host": "your-bucket.oss-region.aliyuncs.com",
    "x-oss-date": "20230101T000000Z",
    "x-oss-content-sha256": "UNSIGNED-PAYLOAD"
  },
  "expiration": "2023-01-01T00:10:00Z",
  "method": "PUT",
  "contentType": "application/octet-stream"
}
```

### WebSocket接口

- **URL**: `/ws/progress/:id`
- **协议**: `WebSocket`
- **参数**:
  - `:id`: 上传任务的唯一标识符，与上传请求中的`uploadID`一致

#### 进度消息格式：

```json
{
  "fileName": "example.jpg",
  "increment": 1024,
  "transferred": 10240,
  "total": 102400,
  "percentage": 10
}
```

## 部署提示

### Docker容器内部结构

应用程序在Docker容器内的工作目录结构如下：

```
/app
├── uploader (Go编译后的二进制文件)
├── test.html (Web界面文件)
├── logs/ (日志目录)
├── temp/ (临时文件目录)
└── checkpoint/ (断点续传检查点目录)
```

### 配置OSS跨域访问

若要使预签名上传功能正常工作，需要在OSS控制台配置CORS规则：

1. 登录[OSS控制台](https://oss.console.aliyun.com/)
2. 选择您的存储桶，点击"权限管理"
3. 点击"跨域设置"，添加以下规则：
   - 来源：您的应用域名，例如`http://localhost:5050`
   - 允许Methods：PUT, GET, POST, DELETE, HEAD
   - 允许Headers：*, Content-Type, Content-MD5, Authorization
   - 暴露Headers：ETag
   - 缓存时间：86400秒

## 常见问题排查

如果在部署环境中遇到问题：

1. 使用 `--verbose` 模式启动，查看详细日志输出
2. 检查 `logs/app.log` 和 `logs/gin.log` 文件
3. 确认OSS配置信息是否正确
4. 检查服务器防火墙是否开放5050端口
5. 确认OSS存储桶的CORS设置已正确配置，允许浏览器直接上传
6. 对于预签名上传和下载链接，确保客户端时钟准确（时间偏差太大会导致签名验证失败）
7. 预签名下载链接最长有效期为7天，超过这个时间需要重新生成链接
8. 使用 `/health` 端点确认服务是否正常运行
9. 使用 `docker ps` 和 `docker logs go-uploader` 查看容器状态和日志

## Docker相关故障排除

如果Docker容器启动失败：

1. 检查容器日志：`docker logs go-uploader`
2. 确保`.env`文件存在并包含正确的配置
3. 确保Docker有足够的资源（内存、CPU等）
4. 检查卷挂载权限是否正确
5. 如果容器崩溃，可以尝试禁用健康检查：在`docker-compose.yml`中注释掉healthcheck部分