package service

import (
	"context"
	"runtime/debug"

	"github.com/kevin-chtw/tw_common/gamebase/game"
	"github.com/kevin-chtw/tw_common/utils"
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

func (p *Player) Message(ctx context.Context, data []byte) {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Errorf("panic recovered %s\n %s", r, string(debug.Stack()))
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
	req := &cproto.GameReq{}
	if err := utils.Unmarshal(ctx, data, req); err != nil {
		logger.Log.Error(err.Error())
		return
	}
	if err := player.HandleMessage(ctx, req); err != nil {
		logger.Log.Error(err.Error())
	}
}
