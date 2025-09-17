package session

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/gorilla/sessions"
)

const (
	SessionName   = "goanalysis-session"
	DBPathKey     = "x-db-path"
	SessionIDKey  = "x-session-id"
	DefaultMaxAge = 86400 * 7 // 7天
)

// Manager session管理器，基于gorilla/sessions
type Manager struct {
	store sessions.Store
	log   *log.Helper
}

// NewManager 创建新的session管理器
func NewManager(secretKey string, logger log.Logger) *Manager {
	// 使用Cookie存储，生产环境可以考虑使用Redis存储
	store := sessions.NewCookieStore([]byte(secretKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   DefaultMaxAge,
		HttpOnly: true,
		Secure:   false, // 开发环境设为false，生产环境应设为true
		SameSite: http.SameSiteLaxMode,
	}

	return &Manager{
		store: store,
		log:   log.NewHelper(logger),
	}
}

// generateSessionID 生成一个新的session ID
// 格式: sess_<timestamp>_<random_hex>
// 例如: sess_20240101123456_a1b2c3d4e5f67890
func generateSessionID() (string, error) {
	// 生成8字节随机数
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// 使用时间戳确保唯一性，随机数确保不可预测性
	timestamp := time.Now().Format("20060102150405")
	randomHex := hex.EncodeToString(randomBytes)

	return fmt.Sprintf("sess_%s_%s", timestamp, randomHex), nil
}

// SetAnalysisPath 设置分析路径到session中
// 注意：这个方法假设session已经存在（通过EnsureSessionID创建），只设置路径不保存session
func (m *Manager) SetAnalysisPath(w http.ResponseWriter, r *http.Request, dbPath string) error {
	if dbPath == "" {
		return errors.New("database path cannot be empty")
	}

	session, err := m.store.Get(r, SessionName)
	if err != nil {
		m.log.Errorf("Failed to get session: %v", err)
		return err
	}

	// 确保session有ID，如果没有则生成一个
	if _, exists := session.Values[SessionIDKey]; !exists {
		sessionID, err := generateSessionID()
		if err != nil {
			m.log.Errorf("Failed to generate session ID: %v", err)
			return err
		}
		session.Values[SessionIDKey] = sessionID
		m.log.Infof("Generated new session ID: %s", sessionID)
	}

	session.Values[DBPathKey] = dbPath

	if err := session.Save(r, w); err != nil {
		m.log.Errorf("Failed to save session: %v", err)
		return err
	}

	sessionID := session.Values[SessionIDKey].(string)
	m.log.Infof("Set analysis path in session %s: %s", sessionID, dbPath)
	return nil
}

// GetAnalysisPath 从session中获取分析路径
func (m *Manager) GetAnalysisPath(r *http.Request) (string, bool, error) {
	session, err := m.store.Get(r, SessionName)
	if err != nil {
		m.log.Errorf("Failed to get session: %v", err)
		return "", false, err
	}

	dbPath, ok := session.Values[DBPathKey].(string)
	if !ok || dbPath == "" {
		return "", false, nil
	}

	return dbPath, true, nil
}

// ClearSession 清除session
func (m *Manager) ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, err := m.store.Get(r, SessionName)
	if err != nil {
		m.log.Errorf("Failed to get session: %v", err)
		return err
	}

	// 设置MaxAge为-1来删除session
	session.Options.MaxAge = -1

	if err := session.Save(r, w); err != nil {
		m.log.Errorf("Failed to clear session: %v", err)
		return err
	}

	m.log.Info("Session cleared")
	return nil
}

// GetSessionID 获取session ID（用于日志记录）
// 返回空字符串表示没有有效的session ID
func (m *Manager) GetSessionID(r *http.Request) string {
	session, err := m.store.Get(r, SessionName)
	if err != nil {
		return ""
	}

	// 从session中获取我们自己生成的session ID
	if sessionID, ok := session.Values[SessionIDKey].(string); ok && sessionID != "" {
		return sessionID
	}

	return ""
}

// EnsureSessionID 确保session有ID，如果没有则创建一个（用于中间件）
func (m *Manager) EnsureSessionID(w http.ResponseWriter, r *http.Request) (string, error) {
	session, err := m.store.Get(r, SessionName)
	if err != nil {
		return "", err
	}

	// 检查是否已有session ID
	if sessionID, ok := session.Values[SessionIDKey].(string); ok && sessionID != "" {
		// 已存在session，不需要保存
		return sessionID, nil
	}

	// 生成新的session ID
	sessionID, err := generateSessionID()
	if err != nil {
		return "", err
	}

	session.Values[SessionIDKey] = sessionID
	if err := session.Save(r, w); err != nil {
		return "", err
	}

	m.log.Infof("Created new session: %s", sessionID)
	return sessionID, nil
}

// EnsureSessionIDWithoutSave 确保session有ID，如果没有则创建一个，但不立即保存
// 返回 (sessionID, needsSave, error)
func (m *Manager) EnsureSessionIDWithoutSave(r *http.Request) (string, bool, error) {
	session, err := m.store.Get(r, SessionName)
	if err != nil {
		return "", false, err
	}

	// 检查是否已有session ID
	if sessionID, ok := session.Values[SessionIDKey].(string); ok && sessionID != "" {
		return sessionID, false, nil
	}

	// 生成新的session ID
	sessionID, err := generateSessionID()
	if err != nil {
		return "", false, err
	}

	session.Values[SessionIDKey] = sessionID
	m.log.Infof("Created new session (not saved yet): %s", sessionID)
	return sessionID, true, nil
}

// SaveSession 保存session
func (m *Manager) SaveSession(w http.ResponseWriter, r *http.Request) error {
	session, err := m.store.Get(r, SessionName)
	if err != nil {
		return err
	}

	return session.Save(r, w)
}

// SessionNeedsSave 检查session是否需要保存（用于检测session内容是否有变化）
func (m *Manager) SessionNeedsSave(r *http.Request) bool {
	// 这里我们简化处理，对于verify/path请求总是需要保存
	// 因为它可能修改了session中的路径
	return r.URL.Path == "/api/runtime/verify/path"
}
