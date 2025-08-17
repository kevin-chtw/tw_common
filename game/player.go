package game

import (
	"context"

	"github.com/kevin-chtw/tw_proto/cproto"
)

const (
	PlayerStatusUnEnter = "UnEnter" // 玩家状态：未进入
	PlayerStatusReady   = "ready"   // 玩家状态：准备
	PlayerStatusPlaying = "playing" // 玩家状态：游戏中
	PlayerStatusOffline = "offline" // 玩家状态：离线
)

// Player 表示游戏中的玩家
type Player struct {
	id     string // 玩家唯一ID
	Seat   int32  // 座位号
	Score  int64  // 玩家积分
	Status string // 玩家状态
}

// NewPlayer 创建新玩家实例
func NewPlayer(id string) *Player {
	return &Player{
		id:     id,
		Status: PlayerStatusOffline,
	}
}

// SetSeat 设置玩家座位号
func (p *Player) SetSeat(seatNum int32) {
	p.Seat = seatNum
}

// AddScore 增加玩家积分
func (p *Player) AddScore(delta int64) {
	p.Score += delta
}

// SetStatus 设置玩家状态
func (p *Player) SetStatus(status string) {
	p.Status = status
}

// HandleMessage 处理玩家消息
func (p *Player) HandleMessage(ctx context.Context, req *cproto.GameReq) {
	table := tableManager.Get(req.Matchid, req.Tableid)
	if nil == table {
		return // 桌子不存在
	}

	table.OnPlayerMsg(ctx, p, req)
}
