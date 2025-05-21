package utils

import (
	"io"
	"log"
	"os"
)

// Logger 结构体，封装日志功能
type Logger struct {
	*log.Logger
	logFile   *os.File
	isVerbose bool
}

// NewLogger 创建一个新的日志记录器
func NewLogger(verbose bool) (*Logger, error) {
	// 创建日志目录（如果不存在）
	logDir := "./logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, err
		}
	}

	// 创建日志文件
	logFile, err := os.OpenFile("./logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	// 根据verbose标志决定日志输出位置
	var output io.Writer
	if verbose {
		// 同时输出到控制台和文件
		output = io.MultiWriter(os.Stdout, logFile)
	} else {
		// 仅输出到文件
		output = logFile
	}

	// 创建日志记录器
	logger := &Logger{
		Logger:    log.New(output, "", log.Ldate|log.Ltime|log.Lshortfile),
		logFile:   logFile,
		isVerbose: verbose,
	}

	return logger, nil
}

// IsVerbose 返回日志是否处于详细模式
func (l *Logger) IsVerbose() bool {
	return l.isVerbose
}

// Close 关闭日志文件
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// SetupGinLogger 设置 Gin 框架的日志输出
func (l *Logger) SetupGinLogger() (*os.File, error) {
	// 创建日志文件用于Gin的HTTP请求日志
	f, err := os.Create("./logs/gin.log")
	if err != nil {
		l.Fatalf("创建Gin日志文件失败: %v", err)
		return nil, err
	}

	return f, nil
}
