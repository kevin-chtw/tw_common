package game

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kevin-chtw/tw_proto/cproto"
	"github.com/kevin-chtw/tw_proto/sproto"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
)

// Player 表示游戏中的玩家
type Player struct {
	Ctx     context.Context
	ack     *cproto.TablePlayerAck
	datas   map[string]int32 //玩家数据
	score   int64            // 玩家积分
	online  bool             // 玩家是否在线
	enter   bool             // 玩家是否进入游戏
	entered bool             // 玩家是否进入过游戏
	isBot   bool             // 是否是bot玩家
}

// newPlayer 创建新玩家实例
func newPlayer(ack *sproto.PlayerInfoAck, seat int32, score int64) *Player {
	p := &Player{
		datas:  make(map[string]int32),
		score:  score,
		online: true,
		enter:  false,
		isBot:  strings.HasPrefix(ack.Uid, "bot_"), // 判断是否是bot玩家
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

func (p *Player) AddData(key string, value int32) {
	p.datas[key] += value
}

func (p *Player) GetScore() int64 {
	return p.score
}

func (p *Player) GetDatas() string {
	if datas, err := json.Marshal(p.datas); err != nil {
		logger.Log.Error(err.Error())
		return ""
	} else {
		return string(datas)
	}
}

// HandleMessage 处理玩家消息
func (p *Player) HandleMessage(ctx context.Context, req *cproto.GameReq) error {
	table := tableManager.Get(req.Matchid, req.Tableid)
	if nil == table {
		return fmt.Errorf("table not found %d", req.Tableid)
	}
	p.Ctx = ctx
	return table.OnPlayerMsg(p, req)
}
