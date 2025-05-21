package utils

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ProgressInfo 进度信息结构体
type ProgressInfo struct {
	FileName    string  `json:"fileName"`
	Increment   int64   `json:"increment"`
	Transferred int64   `json:"transferred"`
	Total       int64   `json:"total"`
	Percentage  int     `json:"percentage"`
	Speed       float64 `json:"speed"` // 上传速度 (KB/s)
}

// ProgressManager 上传进度管理器
type ProgressManager struct {
	mutex      sync.Mutex
	conns      map[string]*websocket.Conn
	progresses map[string]*ProgressInfo
	lastTimes  map[string]time.Time // 上次更新时间，用于计算速度
	logger     *Logger              // 使用我们的自定义日志器
}

// NewProgressManager 创建新的进度管理器
func NewProgressManager(logger *Logger) *ProgressManager {
	// 如果没有提供日志器，使用空实现避免nil指针
	if logger == nil {
		stdLogger := log.New(log.Writer(), "", log.LstdFlags)
		logger = &Logger{Logger: stdLogger}
	}

	return &ProgressManager{
		conns:      make(map[string]*websocket.Conn),
		progresses: make(map[string]*ProgressInfo),
		lastTimes:  make(map[string]time.Time),
		logger:     logger,
	}
}

// RegisterConn 注册新连接
func (m *ProgressManager) RegisterConn(id string, conn *websocket.Conn) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.conns[id] = conn
}

// RemoveConn 删除连接
func (m *ProgressManager) RemoveConn(id string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.conns, id)
	delete(m.progresses, id)
	delete(m.lastTimes, id)
}

// UpdateProgress 更新进度并发送给客户端
func (m *ProgressManager) UpdateProgress(id string, fileName string, increment, transferred, total int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 计算百分比
	percentage := 0
	if total > 0 {
		percentage = int(transferred * 100 / total)
	}

	// 计算上传速度 (KB/s)
	var speed float64 = 0
	now := time.Now()
	if lastTime, ok := m.lastTimes[id]; ok && increment > 0 {
		elapsed := now.Sub(lastTime).Seconds()
		if elapsed > 0 {
			speed = float64(increment) / elapsed / 1024 // 转换为KB/s
		}
	}
	m.lastTimes[id] = now

	// 更新进度信息
	progress := &ProgressInfo{
		FileName:    fileName,
		Increment:   increment,
		Transferred: transferred,
		Total:       total,
		Percentage:  percentage,
		Speed:       speed,
	}

	m.progresses[id] = progress

	// 如果有连接，发送进度更新
	if conn, ok := m.conns[id]; ok {
		if err := conn.WriteJSON(progress); err != nil {
			m.logger.Printf("发送进度信息失败: %v", err)
		}
	}
}

// GetProgress 获取进度信息
func (m *ProgressManager) GetProgress(id string) (*ProgressInfo, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	progress, ok := m.progresses[id]
	return progress, ok
}

// ClearAll 清除所有连接和进度信息
func (m *ProgressManager) ClearAll() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 关闭所有WebSocket连接
	for _, conn := range m.conns {
		_ = conn.Close()
	}

	// 清空所有映射
	m.conns = make(map[string]*websocket.Conn)
	m.progresses = make(map[string]*ProgressInfo)
	m.lastTimes = make(map[string]time.Time)
}
