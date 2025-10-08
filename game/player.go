package game

import (
	"context"
	"errors"

	"github.com/kevin-chtw/tw_proto/cproto"
)

const (
	PlayerStatusUnEnter = iota // 玩家状态：未进入
	PlayerStatusEnter          // 玩家状态：进入
	PlayerStatusReady          // 玩家状态：准备
	PlayerStatusPlaying        // 玩家状态：游戏中
)

// Player 表示游戏中的玩家
type Player struct {
	id     string // 玩家唯一ID
	Seat   int32  // 座位号
	Status int    // 玩家状态
	score  int64  // 玩家积分
	online bool   // 玩家是否在线
}

// NewPlayer 创建新玩家实例
func NewPlayer(id string) *Player {
	return &Player{
		id:     id,
		Status: PlayerStatusUnEnter,
		online: true,
	}
}

// SetSeat 设置玩家座位号
func (p *Player) SetSeat(seatNum int32) {
	p.Seat = seatNum
}

// AddScore 增加玩家积分
func (p *Player) AddScore(delta int64) {
	p.score += delta
}

func (p *Player) GetScore() int64 {
	return p.score
}

// HandleMessage 处理玩家消息
func (p *Player) HandleMessage(ctx context.Context, req *cproto.GameReq) error {
	table := tableManager.Get(req.Matchid, req.Tableid)
	if nil == table {
		return errors.New("table not found")
	}

	return table.OnPlayerMsg(ctx, p, req)
}
