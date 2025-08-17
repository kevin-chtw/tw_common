package game

import "sync"

// TableManager 管理游戏桌
type PlayerManager struct {
	mu      sync.RWMutex
	players map[string]*Player // tableID -> Table
}

// NewPlayerManager 创建玩家管理器
func NewPlayerManager() *PlayerManager {
	return &PlayerManager{
		players: make(map[string]*Player),
	}
}

// GetPlayer 获取玩家实例
func (pm *PlayerManager) GetPlayer(userID string) *Player {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	player, ok := pm.players[userID]
	if !ok {
		player = NewPlayer(userID)
		pm.players[userID] = player
	}
	return player
}

// DeletePlayer 删除玩家实例
func (pm *PlayerManager) Delete(userID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.players, userID)
}
