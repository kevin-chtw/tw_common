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
	logger.Log.Infof("Received player message: %v", req)
	userID := p.app.GetSessionFromCtx(ctx).UID()
	if userID == "" {
		logger.Log.Error("Received player message with empty user ID")
		return
	}

	player := game.GetPlayerManager().Get(userID)
	if player == nil {
		logger.Log.Errorf("Player not found: %s", userID)
		return
	}
	if err := player.HandleMessage(ctx, req); err != nil {
		logger.Log.Errorf("Error handling player message: %v", err)
	}
}
