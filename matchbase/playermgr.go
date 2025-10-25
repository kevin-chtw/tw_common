package matchbase

import (
	"sync"
)

// Playermgr 管理玩家
type Playermgr struct {
	mu      sync.RWMutex
	players map[string]*Player // tableID -> Table
}

// NewPlayermgr 创建玩家管理器
func NewPlayermgr() *Playermgr {
	return &Playermgr{
		players: make(map[string]*Player),
	}
}

// GetPlayer 获取玩家实例
func (p *Playermgr) Load(userID string) *Player {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.players[userID]
}

func (p *Playermgr) Store(player *Player) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.players[player.ID] = player
}

func (p *Playermgr) Delete(userID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.players, userID)
}

func (p *Playermgr) playerCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.players)
}
