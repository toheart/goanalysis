package chanMgr

import (
	"fmt"
	"sync"
)

type ChannelManager struct {
	channels map[string]chan []byte
	mu       sync.RWMutex
}

func NewChannelManager() *ChannelManager {
	return &ChannelManager{
		channels: make(map[string]chan []byte),
	}
}

func (cm *ChannelManager) Get(key string) (chan []byte, error) {
	cm.mu.RLock()
	ch, exists := cm.channels[key]
	cm.mu.RUnlock()

	if exists {
		return ch, nil
	}

	// 如果通道不存在，返回错误
	return nil, fmt.Errorf("channel not found for key: %s", key)
}

func (cm *ChannelManager) Set(key string, ch chan []byte) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.channels[key] = ch
}

// GetAll 返回所有通道
func (cm *ChannelManager) GetAll() map[string]chan []byte {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// 创建一个副本以避免并发问题
	result := make(map[string]chan []byte, len(cm.channels))
	for k, v := range cm.channels {
		result[k] = v
	}

	return result
}

// Close 关闭指定的通道
func (cm *ChannelManager) Close(key string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if ch, exists := cm.channels[key]; exists {
		close(ch)
		delete(cm.channels, key)
	}
}

// CloseAll 关闭所有通道
func (cm *ChannelManager) CloseAll() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for key, ch := range cm.channels {
		close(ch)
		delete(cm.channels, key)
	}
}
