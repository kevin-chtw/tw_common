package matchbase

import "context"

// Player 表示游戏中的玩家
type Player struct {
	Sub     any
	Ctx     context.Context
	ID      string // 玩家唯一ID
	Online  bool   // 玩家在线状态
	Exit    bool   // 玩家是否退出
	MatchId int32
	TableId int32
	Score   int64 // 玩家分数
	Seat    int32 // 玩家座位号
}

// NewPlayer 创建新玩家实例
func NewPlayer(sub any, ctx context.Context, id string, matchId, tableId int32, score int64) *Player {
	p := &Player{
		Sub:     sub,
		Ctx:     ctx,
		ID:      id,
		Online:  true,
		Exit:    false,
		MatchId: matchId,
		TableId: tableId,
		Score:   score,
		Seat:    -1,
	}
	return p
}
