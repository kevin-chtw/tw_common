package matchbase

import "context"

// Player 表示游戏中的玩家
type Player struct {
	Ctx      context.Context
	ID       string // 玩家唯一ID
	IsOnline bool   // 玩家在线状态
	MatchId  int32
	TableId  int32
	Score    int64 // 玩家分数
}

// NewPlayer 创建新玩家实例
func NewPlayer(ctx context.Context, id string, matchId, tableId int32, score int64) *Player {
	p := &Player{
		Ctx:      ctx,
		ID:       id,
		IsOnline: true, // 默认在线状态
		MatchId:  matchId,
		TableId:  tableId,
		Score:    score,
	}
	return p
}
