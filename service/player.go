package service

import (
	"context"

	"github.com/kevin-chtw/tw_common/game"
	"github.com/kevin-chtw/tw_proto/cproto"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/component"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
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

func (p *Player) Message(ctx context.Context, req *cproto.GameReq) {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Errorf("panic recovered: %v", r)
		}
	}()

	userID := p.app.GetSessionFromCtx(ctx).UID()
	if userID == "" {
		logger.Log.Error("user ID not found in session")
		return
	}

	player := game.GetPlayerManager().Get(userID)
	if player == nil {
		logger.Log.Error("player not found in player manager")
		return
	}
	if err := player.HandleMessage(ctx, req); err != nil {
		logger.Log.Error(err.Error())
	}
}
