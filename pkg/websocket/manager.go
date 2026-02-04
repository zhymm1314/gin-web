package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/olahol/melody"
	"go.uber.org/zap"
)

// Message WebSocket 消息结构
type Message struct {
	Type    string      `json:"type"`
	To      string      `json:"to,omitempty"`
	From    string      `json:"from,omitempty"`
	Content interface{} `json:"content"`
}

// Manager WebSocket 管理器
type Manager struct {
	melody       *melody.Melody
	userSessions map[string]map[*melody.Session]bool
	mu           sync.RWMutex
	log          *zap.Logger
}

// NewManager 创建 WebSocket 管理器
func NewManager(log *zap.Logger) *Manager {
	m := &Manager{
		melody:       melody.New(),
		userSessions: make(map[string]map[*melody.Session]bool),
		log:          log,
	}

	// 配置 Melody
	m.melody.Config.MaxMessageSize = 512 * 1024 // 512KB
	m.melody.Config.MessageBufferSize = 256

	// 注册事件处理器
	m.setupHandlers()

	return m
}

func (m *Manager) setupHandlers() {
	// 连接建立
	m.melody.HandleConnect(func(s *melody.Session) {
		userID, _ := s.Get("user_id")
		m.log.Info("websocket client connected", zap.Any("user_id", userID))

		if uid, ok := userID.(string); ok && uid != "" {
			m.mu.Lock()
			if m.userSessions[uid] == nil {
				m.userSessions[uid] = make(map[*melody.Session]bool)
			}
			m.userSessions[uid][s] = true
			m.mu.Unlock()
		}
	})

	// 连接断开
	m.melody.HandleDisconnect(func(s *melody.Session) {
		userID, _ := s.Get("user_id")
		m.log.Info("websocket client disconnected", zap.Any("user_id", userID))

		if uid, ok := userID.(string); ok && uid != "" {
			m.mu.Lock()
			delete(m.userSessions[uid], s)
			if len(m.userSessions[uid]) == 0 {
				delete(m.userSessions, uid)
			}
			m.mu.Unlock()
		}
	})

	// 收到消息
	m.melody.HandleMessage(func(s *melody.Session, msg []byte) {
		var message Message
		if err := json.Unmarshal(msg, &message); err != nil {
			m.log.Error("invalid message format", zap.Error(err))
			return
		}

		userID, _ := s.Get("user_id")
		if uid, ok := userID.(string); ok {
			message.From = uid
		}

		// 处理消息路由
		if message.To != "" {
			m.SendToUser(message.To, &message)
		} else {
			m.Broadcast(&message)
		}
	})

	// 错误处理
	m.melody.HandleError(func(s *melody.Session, err error) {
		m.log.Error("websocket error", zap.Error(err))
	})
}

// HandleRequest 处理 WebSocket 升级请求
func (m *Manager) HandleRequest(w http.ResponseWriter, r *http.Request, userID string) error {
	return m.melody.HandleRequestWithKeys(w, r, map[string]interface{}{
		"user_id": userID,
	})
}

// Broadcast 广播消息给所有客户端
func (m *Manager) Broadcast(message *Message) {
	data, _ := json.Marshal(message)
	m.melody.Broadcast(data)
}

// BroadcastFilter 按条件广播
func (m *Manager) BroadcastFilter(message *Message, filter func(s *melody.Session) bool) {
	data, _ := json.Marshal(message)
	m.melody.BroadcastFilter(data, filter)
}

// SendToUser 发送消息给指定用户
func (m *Manager) SendToUser(userID string, message *Message) {
	m.mu.RLock()
	sessions := m.userSessions[userID]
	m.mu.RUnlock()

	if len(sessions) == 0 {
		return
	}

	data, _ := json.Marshal(message)
	for session := range sessions {
		session.Write(data)
	}
}

// OnlineCount 获取在线人数
func (m *Manager) OnlineCount() int {
	return m.melody.Len()
}

// OnlineUsers 获取在线用户列表
func (m *Manager) OnlineUsers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]string, 0, len(m.userSessions))
	for userID := range m.userSessions {
		users = append(users, userID)
	}
	return users
}

// Close 关闭管理器
func (m *Manager) Close() {
	m.melody.Close()
	m.log.Info("websocket manager closed")
}
