package storage

import (
	"errors"
)

// 存储服务相关错误定义
var (
	// ErrInvalidConfig 无效配置错误
	ErrInvalidConfig = errors.New("无效的存储配置")

	// ErrStorageNotFound 存储服务不存在错误
	ErrStorageNotFound = errors.New("未找到指定的存储服务")

	// ErrObjectExists 对象已存在错误
	ErrObjectExists = errors.New("对象已存在")

	// ErrObjectNotExists 对象不存在错误
	ErrObjectNotExists = errors.New("对象不存在")
)
