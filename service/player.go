package service

import (
	"context"
	"errors"

	"github.com/kevin-chtw/tw_common/game"
	"github.com/kevin-chtw/tw_proto/cproto"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/component"
)

// Player 独立的玩家服务
type Player struct {
	component.Base
	app pitaya.Pitaya
}

// NewPlayer 创建独立的玩家服务
func NewPlayer(app pitaya.Pitaya) *Player {
	return &Player{
		app: app,
	}
}

func (p *Player) Message(ctx context.Context, req *cproto.GameReq) (*cproto.GameAck, error) {
	userID := p.app.GetSessionFromCtx(ctx).UID()
	if userID == "" {
		return nil, errors.New("user ID not found in session")
	}

	player := game.GetPlayerManager().Get(userID)
	if player == nil {
		return nil, errors.New("Player not found in player manager")
	}
	if err := player.HandleMessage(ctx, req); err != nil {
		return nil, err
	}
	return &cproto.GameAck{}, nil
}
