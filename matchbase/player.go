package matchbase

import "context"

// Player 表示游戏中的玩家
type Player struct {
	Sub     any
	Ctx     context.Context
	ID      string // 玩家唯一ID
	State   int32  //玩家状态 0-在线 1-离线，2-退出
	MatchId int32
	TableId int32
	Score   int64 // 玩家分数
}

// NewPlayer 创建新玩家实例
func NewPlayer(sub any, ctx context.Context, id string, matchId, tableId int32, score int64) *Player {
	p := &Player{
		Sub:     sub,
		Ctx:     ctx,
		ID:      id,
		State:   0, // 默认在线状态
		MatchId: matchId,
		TableId: tableId,
		Score:   score,
	}
	return p
}

func (p *Player) SetState(online bool) bool {
	state := int32(0)
	if !online {
		p.State = 1 // 离线状态
	}
	if p.State == state {
		return false
	}
	p.State = state
	return true
}
