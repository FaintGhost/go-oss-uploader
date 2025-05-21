package utils

import (
	"fmt"
	mathrand "math/rand"
	"net/url"
	"os"
	"sync"
	"time"
)

// ShortLink 短链接结构体
type ShortLink struct {
	ID         string    // 短链接ID (存储uniqueID部分)
	UniqueID   string    // 唯一标识符
	LongURL    string    // 长URL（预签名URL）
	FileName   string    // 文件名
	Expiration time.Time // 过期时间
}

// ShortLinkManager 短链接管理器
type ShortLinkManager struct {
	mutex sync.RWMutex
	links map[string]*ShortLink // UniqueID -> ShortLink
}

// NewShortLinkManager 创建新的短链接管理器
func NewShortLinkManager() *ShortLinkManager {
	// 初始化随机种子
	mathrand.Seed(time.Now().UnixNano())
	return &ShortLinkManager{
		links: make(map[string]*ShortLink),
	}
}

// GenerateUniqueID 生成唯一ID
func GenerateUniqueID() string {
	// 使用多个随机源组合增加唯一性
	// 1. 时间戳（纳秒级）
	now := time.Now().UnixNano()

	// 2. 随机数（更大范围）
	random := mathrand.Intn(100000000)

	// 3. 特殊字符生成
	characters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	randomChar := characters[mathrand.Intn(len(characters))]

	// 4. 进程ID（如果在同一进程中，进程ID是恒定的，但与其他因素组合可增加唯一性）
	pid := os.Getpid()

	// 5. 静态计数器（确保即使其他所有因素都相同，ID也会不同）
	counter := getAndIncrementCounter()

	// 组合所有因素生成唯一ID
	uniqueID := fmt.Sprintf("%x%x%c%x%d", now, random, randomChar, pid, counter)

	// 取前12位，提供足够的唯一性空间
	if len(uniqueID) > 12 {
		return uniqueID[:12]
	}
	return uniqueID
}

// 用于生成唯一ID的静态计数器
var (
	counterMutex sync.Mutex
	idCounter    uint64 = 0
)

// getAndIncrementCounter 获取并增加计数器
func getAndIncrementCounter() uint64 {
	counterMutex.Lock()
	defer counterMutex.Unlock()
	current := idCounter
	idCounter++
	return current
}

// CreateShortLink 创建新的短链接
func (m *ShortLinkManager) CreateShortLink(longURL string, fileName string, expiration time.Time) (string, error) {
	// 生成唯一ID
	uniqueID := GenerateUniqueID()

	// 将文件名URL编码
	encodedFileName := url.PathEscape(fileName)

	// 完整的短链接ID格式为: uniqueID/encodedFileName
	linkID := uniqueID + "/" + encodedFileName

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.links[uniqueID] = &ShortLink{
		ID:         linkID,
		UniqueID:   uniqueID,
		LongURL:    longURL,
		FileName:   fileName,
		Expiration: expiration,
	}

	return linkID, nil
}

// GetLongURL 获取短链接对应的长链接 (通过完整ID查找)
func (m *ShortLinkManager) GetLongURL(id string) (string, string, bool) {
	return m.getLongURLByID(id)
}

// GetLongURLByUniqueID 通过唯一ID获取长链接
func (m *ShortLinkManager) GetLongURLByUniqueID(uniqueID string) (string, string, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	link, exists := m.links[uniqueID]
	if !exists {
		return "", "", false
	}

	// 检查是否已过期
	if time.Now().After(link.Expiration) {
		// 过期了，删除并返回不存在
		delete(m.links, uniqueID)
		return "", "", false
	}

	return link.LongURL, link.FileName, true
}

// getLongURLByID 通过完整ID查找 (兼容旧版本)
func (m *ShortLinkManager) getLongURLByID(id string) (string, string, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 先尝试直接查找完整ID
	for _, link := range m.links {
		if link.ID == id {
			// 检查是否已过期
			if time.Now().After(link.Expiration) {
				// 过期了，删除并返回不存在
				delete(m.links, link.UniqueID)
				return "", "", false
			}

			return link.LongURL, link.FileName, true
		}
	}

	// 为了兼容直接使用唯一ID的情况
	if link, exists := m.links[id]; exists {
		// 检查是否已过期
		if time.Now().After(link.Expiration) {
			// 过期了，删除并返回不存在
			delete(m.links, id)
			return "", "", false
		}

		return link.LongURL, link.FileName, true
	}

	return "", "", false
}

// CleanupExpiredLinks 清理过期的短链接
func (m *ShortLinkManager) CleanupExpiredLinks() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	for id, link := range m.links {
		if now.After(link.Expiration) {
			delete(m.links, id)
		}
	}
}
