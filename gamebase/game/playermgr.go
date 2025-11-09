package game

import (
	"errors"
	"sync"

	"github.com/kevin-chtw/tw_proto/sproto"
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

func (p *PlayerManager) Store(ack *sproto.PlayerInfoAck, bot bool, seat int32, score int64) (*Player, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.players[ack.Uid]; ok {
		return nil, errors.New("player is already in game")
	}

	player := newPlayer(ack, bot, seat, score)
	p.players[ack.Uid] = player
	return player, nil
}

// DeletePlayer 删除玩家实例
func (p *PlayerManager) Delete(userID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.players, userID)
}
