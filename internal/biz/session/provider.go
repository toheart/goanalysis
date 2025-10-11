package session

import (
	"github.com/go-kratos/kratos/v2/log"
)

// NewManager 创建session管理器的工厂函数
func NewSessionManager(logger log.Logger) *Manager {
	// 使用一个默认的secret key，生产环境应该从配置文件读取
	secretKey := "goanalysis-session-secret-key-change-in-production"
	return NewManager(secretKey, logger)
}
