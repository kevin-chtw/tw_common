package service

import (
	"context"

	"github.com/kevin-chtw/tw_common/game"
	"github.com/kevin-chtw/tw_proto/cproto"
	"github.com/sirupsen/logrus"
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

func (p *Player) Message(ctx context.Context, req *cproto.GameReq) {
	logrus.Infof("Received player message: %v", req)
	userID := p.app.GetSessionFromCtx(ctx).UID()
	if userID == "" {
		logrus.Error("Received player message with empty user ID")
		return
	}

	player := game.GetPlayerManager().GetPlayer(userID)
	player.HandleMessage(ctx, req)
}
