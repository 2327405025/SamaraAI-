package aihelper

import (
	"SamaraAI/internal/config"
	"context"
	"sync"
)

const defaultMaxHelpersPerUser = 50

var ctx = context.Background()

// AIHelperManager AI助手管理器，管理用户-会话-AIHelper的映射关系
type AIHelperManager struct {
	helpers map[string]map[string]*AIHelper // map[用户账号（唯一）]map[会话ID]*AIHelper
	mu      sync.RWMutex
}

// NewAIHelperManager 创建新的管理器实例
func NewAIHelperManager() *AIHelperManager {
	return &AIHelperManager{
		helpers: make(map[string]map[string]*AIHelper),
	}
}

// 获取或创建AIHelper
func (m *AIHelperManager) GetOrCreateAIHelper(userName string, sessionID string, modelType string, config map[string]interface{}) (*AIHelper, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取用户的会话映射
	userHelpers, exists := m.helpers[userName]
	if !exists {
		userHelpers = make(map[string]*AIHelper)
		m.helpers[userName] = userHelpers
	}

	helper, exists := userHelpers[sessionID]
	if exists {
		if helper.GetModelType() != modelType {
			factory := GetGlobalFactory()
			newHelper, err := factory.CreateAIHelper(ctx, modelType, sessionID, config)
			if err != nil {
				return nil, err
			}
			for _, msg := range helper.GetMessages() {
				newHelper.AddMessage(msg.Content, msg.UserName, msg.IsUser, false)
			}
			userHelpers[sessionID] = newHelper
			return newHelper, nil
		}
		return helper, nil
	}

	m.evictIfNeeded(userName, userHelpers)

	// 创建新的AIHelper
	factory := GetGlobalFactory()
	helper, err := factory.CreateAIHelper(ctx, modelType, sessionID, config)
	if err != nil {
		return nil, err
	}

	userHelpers[sessionID] = helper
	return helper, nil
}

// 获取指定用户的指定会话的AIHelper
func (m *AIHelperManager) GetAIHelper(userName string, sessionID string) (*AIHelper, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userHelpers, exists := m.helpers[userName]
	if !exists {
		return nil, false
	}

	helper, exists := userHelpers[sessionID]
	return helper, exists
}

// 移除指定用户的指定会话的AIHelper
func (m *AIHelperManager) RemoveAIHelper(userName string, sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	userHelpers, exists := m.helpers[userName]
	if !exists {
		return
	}

	delete(userHelpers, sessionID)

	// 如果用户没有会话了，清理用户映射
	if len(userHelpers) == 0 {
		delete(m.helpers, userName)
	}
}

func (m *AIHelperManager) evictIfNeeded(userName string, userHelpers map[string]*AIHelper) {
	max := defaultMaxHelpersPerUser
	if cfg := config.Get(); cfg != nil && cfg.AiHelper.MaxSessionsPerUser > 0 {
		max = cfg.AiHelper.MaxSessionsPerUser
	}
	for len(userHelpers) >= max {
		for sid := range userHelpers {
			delete(userHelpers, sid)
			break
		}
	}
	if len(userHelpers) == 0 {
		delete(m.helpers, userName)
	}
}

// 全局管理器实例
var globalManager *AIHelperManager
var once sync.Once

// GetGlobalManager 获取全局管理器实例
func GetGlobalManager() *AIHelperManager {
	once.Do(func() {
		globalManager = NewAIHelperManager()
	})
	return globalManager
}
