package game

import (
	"context"
	"errors"

	"github.com/kevin-chtw/tw_proto/cproto"
	"github.com/kevin-chtw/tw_proto/sproto"
)

// Player 表示游戏中的玩家
type Player struct {
	Ctx     context.Context
	ack     *cproto.TablePlayerAck
	score   int64 // 玩家积分
	online  bool  // 玩家是否在线
	enter   bool  // 玩家是否进入游戏
	entered bool  // 玩家是否进入过游戏
}

// newPlayer 创建新玩家实例
func newPlayer(ack *sproto.PlayerInfoAck, seat int32, score int64) *Player {
	p := &Player{
		score:  score,
		online: true,
		enter:  false,
	}
	p.ack = &cproto.TablePlayerAck{
		Uid:      ack.Uid,
		Seat:     seat,
		Avatar:   ack.Avatar,
		Nickname: ack.Nickname,
		Vip:      ack.Vip,
		Diamond:  ack.Diamond,
		Ready:    false,
	}
	return p
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

// HandleMessage 处理玩家消息
func (p *Player) HandleMessage(ctx context.Context, req *cproto.GameReq) error {
	table := tableManager.Get(req.Matchid, req.Tableid)
	if nil == table {
		return errors.New("table not found")
	}
	p.Ctx = ctx
	return table.OnPlayerMsg(ctx, p, req)
}
