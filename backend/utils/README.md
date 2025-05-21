# 工具模块

本目录包含项目中使用的各种工具模块，用于降低主程序与功能模块之间的耦合度。

## 短链接模块 (shortlink.go)

短链接模块提供了生成和管理短链接的功能，可以将较长的URL（如预签名下载链接）转换为短链接形式，方便分享和使用。

### 主要功能：

1. **短链接生成**：将长URL转换为短链接
2. **链接管理**：存储和管理已生成的短链接
3. **过期处理**：自动清理过期的短链接
4. **链接重定向**：当访问短链接时，重定向到原始长链接

### 使用方法：

```go
// 创建短链接管理器实例
shortLinkManager := utils.NewShortLinkManager()

// 创建短链接
id, err := shortLinkManager.CreateShortLink(longURL, fileName, expiration)

// 获取短链接对应的长链接
longURL, fileName, exists := shortLinkManager.GetLongURL(shortLinkID)

// 清理过期短链接
shortLinkManager.CleanupExpiredLinks()
```

此模块通过定时任务自动清理过期链接，减少内存占用和资源消耗。

## 日志模块 (logger.go)

日志模块提供了统一的日志记录功能，支持详细模式和普通模式，可以输出到控制台和文件。

### 主要功能：

1. **日志记录**：统一的日志记录接口
2. **详细模式**：支持详细和普通两种日志记录模式
3. **双重输出**：可以同时输出到控制台和文件
4. **Gin框架集成**：为Gin框架提供日志支持

### 使用方法：

```go
// 创建日志记录器，参数为是否启用详细模式
logger, err := utils.NewLogger(verbose)
if err != nil {
    log.Fatalf("初始化日志记录器失败: %v", err)
}

// 记录普通日志
logger.Printf("这是一条普通日志")

// 记录错误日志
logger.Fatalf("这是一条致命错误: %v", err)

// 为Gin框架设置日志
ginLogFile, err := logger.SetupGinLogger()
```

日志模块自动创建所需的目录和文件，使用标准的日志格式，便于阅读和分析。

## 上传进度管理器 (progress.go)

上传进度管理器提供了文件上传过程中的进度跟踪和实时反馈功能，通过WebSocket向客户端推送上传状态。

### 主要功能：

1. **进度跟踪**：跟踪文件上传的进度和已传输的字节数
2. **速度计算**：计算实时上传速度
3. **WebSocket推送**：通过WebSocket向客户端推送进度信息
4. **多文件管理**：支持同时追踪多个文件的上传进度

### 使用方法：

```go
// 创建进度管理器实例
progressManager := utils.NewProgressManager(logger)

// 注册WebSocket连接
progressManager.RegisterConn(uploadID, websocketConn)

// 在上传过程中更新进度
progressManager.UpdateProgress(uploadID, fileName, increment, transferred, total)

// 上传完成后清除连接
progressManager.RemoveConn(uploadID)

// 获取特定上传任务的进度
progress, exists := progressManager.GetProgress(uploadID)

// 清除所有连接和进度信息
progressManager.ClearAll()
```

通过与WebSocket结合，该模块能够为用户提供实时的上传进度反馈，提升用户体验。 