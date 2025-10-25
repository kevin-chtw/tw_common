package game

import (
	"context"
	"errors"

	"github.com/kevin-chtw/tw_proto/cproto"
	"github.com/kevin-chtw/tw_proto/sproto"
)

const (
	PlayerStatusUnEnter = iota // 玩家状态：未进入
	PlayerStatusEnter          // 玩家状态：进入
	PlayerStatusReady          // 玩家状态：准备
	PlayerStatusPlaying        // 玩家状态：游戏中
)

// Player 表示游戏中的玩家
type Player struct {
	Ctx    context.Context
	ack    *cproto.TablePlayerAck
	Status int   // 玩家状态
	score  int64 // 玩家积分
	online bool  // 玩家是否在线
}

// NewPlayer 创建新玩家实例
func NewPlayer(id string) *Player {
	return &Player{
		ack:    &cproto.TablePlayerAck{Uid: id},
		Status: PlayerStatusUnEnter,
		online: true,
	}
}

// SetSeat 设置玩家座位号
func (p *Player) SetSeat(seat int32) {
	p.ack.Seat = seat
}

func (p *Player) GetSeat() int32 {
	return p.ack.Seat
}

// AddScore 增加玩家积分
func (p *Player) AddScore(score int64) {
	p.score += score
}

func (p *Player) GetScore() int64 {
	return p.score
}

func (p *Player) setAck(ack *sproto.PlayerInfoAck) {
	p.ack.Avatar = ack.Avatar
	p.ack.Nickname = ack.Nickname
	p.ack.Vip = ack.Vip
	p.ack.Diamond = ack.Diamond
}

// HandleMessage 处理玩家消息
func (p *Player) HandleMessage(ctx context.Context, req *cproto.GameReq) error {
	table := tableManager.Get(req.Matchid, req.Tableid)
	if nil == table {
		return errors.New("table not found")
	}
	p.Ctx = ctx
	return table.OnPlayerMsg(ctx, p, req)
}
