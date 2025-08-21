package game

import (
	"errors"
	"sync"
)

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

// Get 获取玩家实例
func (p *PlayerManager) Get(userID string) (player *Player) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	player = p.players[userID]
	return
}

func (p *PlayerManager) Store(userId string) (*Player, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.players[userId]; ok {
		return nil, errors.New("player is already in game")
	}

	player := NewPlayer(userId)
	p.players[userId] = player
	return player, nil
}

// DeletePlayer 删除玩家实例
func (p *PlayerManager) Delete(userID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.players, userID)
}
