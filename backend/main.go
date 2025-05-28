package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"

	storage "go-uploader/storage"
	alioss "go-uploader/storage/ali-oss"
	"go-uploader/storage/minio"
	"go-uploader/utils"
)

// 全局变量用于控制日志级别
var (
	verbose     bool   // 详细日志模式
	storageType string // 存储类型: ali-oss, minio
	logger      *utils.Logger
)

// WebSocket升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有跨域请求
	},
}

// 初始化短链接管理器
var shortLinkManager = utils.NewShortLinkManager()

// 初始化上传进度管理器
var progressManager *utils.ProgressManager

// 初始化解析命令行参数
func init() {
	// 添加verbose标志
	flag.BoolVar(&verbose, "verbose", false, "启用详细日志输出模式")
	// 添加存储类型标志
	flag.StringVar(&storageType, "storage", "", "存储类型: ali-oss, minio")
	flag.Parse()
}

func main() {
	// 加载环境变量
	_ = godotenv.Load()

	// 如果命令行未指定存储类型，则尝试从环境变量读取
	if storageType == "" {
		envStorageType := os.Getenv("OSS")
		if envStorageType != "" {
			storageType = envStorageType
		} else {
			// 默认使用阿里云OSS
			storageType = "ali-oss"
		}
	}

	// 初始化日志记录器
	var err error
	logger, err = utils.NewLogger(verbose)
	if err != nil {
		log.Fatalf("初始化日志记录器失败: %v", err)
	}
	logger.Printf("应用启动, 详细日志模式: %v", verbose)
	logger.Printf("使用存储类型: %s", storageType)

	// 初始化上传进度管理器
	progressManager = utils.NewProgressManager(logger)
	logger.Printf("上传进度管理器初始化成功")

	// 如果是详细日志模式，设置Gin为调试模式
	if verbose {
		gin.SetMode(gin.DebugMode)
		logger.Printf("Gin设置为调试模式")
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建日志目录（如果不存在）
	logDir := "./logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		logger.Printf("创建日志目录: %s", logDir)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			logger.Fatalf("创建日志目录失败: %v", err)
		}
	}

	// 创建日志文件用于Gin的HTTP请求日志
	f, err := logger.SetupGinLogger()
	if err != nil {
		logger.Fatalf("创建Gin日志文件失败: %v", err)
	}

	// 根据verbose设置Gin的日志输出
	if verbose {
		// 同时输出到控制台和文件
		gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	} else {
		gin.DefaultWriter = f
	}

	// 加载存储配置并创建存储服务
	var storageConfig storage.StorageConfig
	logger.Printf("加载存储服务配置中，存储类型: %s...", storageType)

	switch storageType {
	case "ali-oss":
		ossConfig, err := alioss.LoadAliOSSConfigFromEnv()
		if err != nil {
			logger.Fatalf("加载阿里云OSS配置失败: %v", err)
		}
		storageConfig = ossConfig
		logger.Printf("阿里云OSS配置加载成功, 区域: %s, 存储桶: %s",
			ossConfig.Region, ossConfig.BucketName)
	case "minio":
		minioConfig, err := minio.LoadMinioConfigFromEnv()
		if err != nil {
			logger.Fatalf("加载MinIO配置失败: %v", err)
		}
		storageConfig = minioConfig
		logger.Printf("MinIO配置加载成功, 端点: %s, 存储桶: %s",
			minioConfig.Endpoint, minioConfig.BucketName)
	default:
		logger.Fatalf("不支持的存储类型: %s", storageType)
	}

	// 初始化存储服务
	logger.Printf("初始化存储服务...")
	storageService, err := storage.CreateStorageService(storageConfig)
	if err != nil {
		logger.Fatalf("初始化存储服务失败: %v", err)
	}
	logger.Printf("存储服务初始化成功")

	// 确保checkpoint目录存在
	checkpointDir := "./checkpoint"
	if _, err := os.Stat(checkpointDir); os.IsNotExist(err) {
		logger.Printf("创建断点续传目录: %s", checkpointDir)
		if err := os.MkdirAll(checkpointDir, 0755); err != nil {
			logger.Printf("创建断点续传目录失败: %v", err)
		}
	}

	// 创建Gin路由
	r := gin.New() // 使用New而不是Default以便自定义中间件

	// 添加Recovery中间件
	r.Use(gin.Recovery())

	// 添加自定义日志中间件
	r.Use(func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方式
		reqMethod := c.Request.Method

		// 请求路由
		reqUri := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		// 请求IP
		clientIP := c.ClientIP()

		// 日志格式
		logger.Printf("[GIN] %v | %3d | %13v | %15s | %s | %s",
			endTime.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latencyTime,
			clientIP,
			reqMethod,
			reqUri,
		)
	})

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 设置文件大小限制 (默认为32MB)
	r.MaxMultipartMemory = 8 << 20 // 8 MB
	logger.Printf("设置最大上传文件大小为: %d MB", (r.MaxMultipartMemory / 1024 / 1024))

	// 设置静态文件服务
	r.StaticFile("/", "./index.html")
	logger.Printf("静态文件路由设置完成")

	// WebSocket处理上传进度
	r.GET("/api/ws/progress/:id", func(c *gin.Context) {
		id := c.Param("id")
		logger.Printf("WebSocket连接请求: ID=%s", id)

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Printf("WebSocket升级失败: %v", err)
			return
		}
		defer conn.Close()

		// 注册连接
		progressManager.RegisterConn(id, conn)
		logger.Printf("WebSocket连接已注册: ID=%s", id)
		defer progressManager.RemoveConn(id)

		// 保持WebSocket连接直到客户端断开
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				logger.Printf("WebSocket连接关闭: ID=%s, 错误: %v", id, err)
				break
			}
		}
	})

	// 设置文件上传路由
	r.POST("/api/upload", func(c *gin.Context) {
		logger.Printf("收到文件上传请求")

		// 获取上传ID
		uploadID := c.PostForm("uploadID")
		if uploadID == "" {
			logger.Printf("上传ID为空")
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No upload ID provided",
			})
			return
		}
		logger.Printf("上传ID: %s", uploadID)

		// 获取原始文件名，如果有的话（用于文件夹上传）
		originalFileName := c.PostForm("originalFileName")

		// 获取上传的文件
		file, err := c.FormFile("file")
		if err != nil {
			logger.Printf("获取上传文件失败: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  "No file uploaded",
				"detail": err.Error(),
			})
			return
		}
		logger.Printf("文件信息: 名称=%s, 大小=%d bytes", file.Filename, file.Size)

		// 如果提供了原始文件名，使用它作为对象名
		objectName := file.Filename
		if originalFileName != "" {
			logger.Printf("使用原始文件名: %s", originalFileName)
			objectName = originalFileName
		}

		// 检查文件是否已存在于存储服务中
		exists, err := storageService.IsObjectExist(context.Background(), objectName)
		if err != nil {
			logger.Printf("检查文件是否存在失败: %v", err)
			// 继续处理，不中断上传流程
		}

		// 如果文件已存在，告知用户并提供分享链接选项
		if exists {
			logger.Printf("文件 %s 已存在于存储中，不需要重新上传", objectName)
			bucketDomain := storageService.GetBucketDomain()
			url := "https://" + bucketDomain + "/" + objectName
			c.JSON(http.StatusOK, gin.H{
				"message":       "File already exists",
				"filename":      objectName,
				"size":          file.Size,
				"url":           url,
				"alreadyExists": true,
				"skipUpload":    true,
			})
			return
		}

		// 文件不存在，继续上传流程
		// 创建临时目录保存上传的文件
		tempDir := "./temp"
		if _, err := os.Stat(tempDir); os.IsNotExist(err) {
			logger.Printf("创建临时目录: %s", tempDir)
			if err := os.Mkdir(tempDir, 0755); err != nil {
				logger.Printf("创建临时目录失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":  "Failed to create temp directory",
					"detail": err.Error(),
				})
				return
			}
		}

		// 生成临时文件路径
		tempFilePath := filepath.Join(tempDir, file.Filename)
		logger.Printf("临时文件路径: %s", tempFilePath)

		// 保存上传的文件到临时目录
		logger.Printf("保存文件到临时目录...")
		if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
			logger.Printf("保存文件到临时目录失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "Failed to save file",
				"detail": err.Error(),
			})
			return
		}
		logger.Printf("文件已保存到临时目录")

		// 创建进度回调函数
		progressCallback := func(increment, transferred, total int64) {
			progressManager.UpdateProgress(uploadID, objectName, increment, transferred, total)

			if verbose && (transferred == total || transferred%(total/10) < increment) {
				percentage := int(float64(transferred) / float64(total) * 100)
				logger.Printf("上传进度: %s - %d%% (%d/%d 字节)", objectName, percentage, transferred, total)
			}
		}

		// 上传文件到存储服务
		logger.Printf("开始上传文件到存储服务...")
		startTime := time.Now()
		result, err := storageService.UploadFile(context.Background(), objectName, tempFilePath, progressCallback)
		if err != nil {
			logger.Printf("上传文件到存储服务失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "Failed to upload file",
				"detail": err.Error(),
			})
			return
		}
		elapsedTime := time.Since(startTime)
		logger.Printf("文件上传完成: %s, 用时: %v", objectName, elapsedTime)

		// 删除临时文件
		logger.Printf("删除临时文件: %s", tempFilePath)
		if err := os.Remove(tempFilePath); err != nil {
			logger.Printf("删除临时文件失败: %v", err)
		}

		// 返回上传结果
		bucketDomain := storageService.GetBucketDomain()
		url := "https://" + bucketDomain + "/" + objectName
		logger.Printf("上传成功, 文件URL: %s", url)
		c.JSON(http.StatusOK, gin.H{
			"message":  "File uploaded successfully",
			"filename": objectName,
			"size":     file.Size,
			"url":      url,
			"result":   result,
		})
	})

	// 添加一个健康检查路由
	r.GET("/api/health", func(c *gin.Context) {
		logger.Printf("收到健康检查请求")
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 添加预签名下载URL接口
	r.GET("/api/download/:filename", func(c *gin.Context) {
		logger.Printf("收到预签名下载URL请求")

		// 获取文件名
		fileName := c.Param("filename")
		if fileName == "" {
			logger.Printf("文件名为空")
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No file name provided",
			})
			return
		}
		logger.Printf("请求生成预签名下载URL的文件名: %s", fileName)

		// 获取过期时间参数，默认24小时
		expirationStr := c.DefaultQuery("expiration", "24h")
		expiration, err := time.ParseDuration(expirationStr)
		if err != nil {
			logger.Printf("解析过期时间失败: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid expiration format, use like 1h, 24h, 7d",
			})
			return
		}

		// 限制最大过期时间为7天（阿里云OSS的限制）
		maxExpiration := 7 * 24 * time.Hour
		if expiration > maxExpiration {
			logger.Printf("过期时间超过最大限制，已调整为7天")
			expiration = maxExpiration
		}

		// 生成预签名下载URL
		url, headers, err := storageService.GeneratePresignedDownloadURL(context.Background(), fileName, expiration)
		if err != nil {
			logger.Printf("生成预签名下载URL失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "Failed to generate presigned download URL",
				"detail": err.Error(),
			})
			return
		}

		// 返回预签名下载URL和必要的请求头
		c.JSON(http.StatusOK, gin.H{
			"url":        url,
			"headers":    headers,
			"expiration": time.Now().Add(expiration).Format(time.RFC3339),
			"method":     "GET",
			"fileName":   fileName,
		})
	})

	// 添加预签名URL接口
	r.POST("/api/presign", func(c *gin.Context) {
		logger.Printf("收到预签名URL请求")

		// 获取文件名和文件大小
		fileName := c.PostForm("fileName")
		fileSizeStr := c.PostForm("fileSize")

		if fileName == "" {
			logger.Printf("文件名为空")
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No file name provided",
			})
			return
		}
		logger.Printf("请求生成预签名URL的文件名: %s", fileName)

		// 根据文件大小动态调整过期时间
		expiration := 10 * time.Minute // 默认10分钟

		// 如果提供了文件大小，则根据大小调整过期时间
		if fileSizeStr != "" {
			fileSize, err := strconv.ParseInt(fileSizeStr, 10, 64)
			if err == nil {
				// 100MB以上的文件
				if fileSize > 100*1024*1024 {
					expiration = 30 * time.Minute
				}
				// 500MB以上的文件
				if fileSize > 500*1024*1024 {
					expiration = 60 * time.Minute
				}
				// 1GB以上的文件
				if fileSize > 1*1024*1024*1024 {
					expiration = 3 * time.Hour
				}
				logger.Printf("文件大小: %d 字节, 设置过期时间: %v", fileSize, expiration)
			}
		}

		// 生成预签名URL
		url, headers, err := storageService.GeneratePresignedURL(context.Background(), fileName, expiration)
		if err != nil {
			logger.Printf("生成预签名URL失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "Failed to generate presigned URL",
				"detail": err.Error(),
			})
			return
		}

		// 返回预签名URL和必要的请求头
		c.JSON(http.StatusOK, gin.H{
			"url":         url,
			"headers":     headers,
			"expiration":  time.Now().Add(expiration).Format(time.RFC3339),
			"method":      "PUT",
			"contentType": "application/octet-stream",
		})
	})

	// 添加API端点获取短链接
	r.POST("/api/short-link", func(c *gin.Context) {
		logger.Printf("收到生成短链接请求")

		// 获取必要参数
		longURL := c.PostForm("url")
		fileName := c.PostForm("fileName")
		expirationStr := c.PostForm("expiration")

		if longURL == "" || fileName == "" || expirationStr == "" {
			logger.Printf("缺少必要参数")
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Missing required parameters",
			})
			return
		}

		// 解析过期时间
		expiration, err := time.Parse(time.RFC3339, expirationStr)
		if err != nil {
			logger.Printf("解析过期时间失败: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid expiration format",
			})
			return
		}

		// 生成短链接
		encodedID, err := shortLinkManager.CreateShortLink(longURL, fileName, expiration)
		if err != nil {
			logger.Printf("生成短链接失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate short link",
			})
			return
		}

		// 构建短链接URL (encodedID已经通过CreateShortLink进行了URL编码)
		protocol := "http"
		if c.Request.TLS != nil {
			protocol = "https"
		}
		shortURL := fmt.Sprintf("%s://%s/s/%s", protocol, c.Request.Host, encodedID)

		logger.Printf("生成短链接成功: %s -> %s", shortURL, longURL)

		// 返回短链接
		c.JSON(http.StatusOK, gin.H{
			"shortURL":   shortURL,
			"id":         encodedID,
			"fileName":   fileName,
			"expiration": expiration.Format(time.RFC3339),
		})
	})

	// 处理短链接访问 - 使用编码后的文件名作为路径参数
	r.GET("/s/:uniqueID/*fileName", func(c *gin.Context) {
		// 获取唯一ID
		uniqueID := c.Param("uniqueID")

		// 获取文件名参数
		fileName := c.Param("fileName")
		if len(fileName) > 0 && fileName[0] == '/' {
			fileName = fileName[1:] // 移除开头的斜杠
		}

		logger.Printf("收到短链接访问请求: ID=%s, 文件名=%s", uniqueID, fileName)

		// 从唯一ID获取长URL
		longURL, fileName, exists := shortLinkManager.GetLongURLByUniqueID(uniqueID)

		// 如果找不到，尝试使用完整路径（兼容旧版格式）
		if !exists {
			completePath := uniqueID
			if fileName != "" {
				completePath = uniqueID + "/" + fileName
			}
			logger.Printf("尝试使用完整路径重新查找: %s", completePath)
			longURL, fileName, exists = shortLinkManager.GetLongURL(completePath)
		}

		if !exists {
			logger.Printf("短链接不存在或已过期: %s", uniqueID)
			c.JSON(http.StatusNotFound, gin.H{
				"error": "链接不存在或已过期",
			})
			return
		}

		logger.Printf("短链接重定向: %s -> %s", uniqueID, longURL)

		// 设置Content-Disposition头以提示浏览器下载文件
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

		// 重定向到长链接
		c.Redirect(http.StatusTemporaryRedirect, longURL)
	})

	// 启动定时任务清理过期短链接
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			logger.Printf("开始清理过期短链接")
			shortLinkManager.CleanupExpiredLinks()
		}
	}()

	// 启动服务器
	port := ":5050"
	logger.Printf("启动HTTP服务器, 监听端口%s", port)
	if err := r.Run(port); err != nil {
		logger.Fatalf("启动服务器失败: %v", err)
	}
}
