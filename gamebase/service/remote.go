package service

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/kevin-chtw/tw_common/gamebase/game"
	"github.com/kevin-chtw/tw_common/utils"
	"github.com/kevin-chtw/tw_proto/sproto"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/component"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Remote 独立的匹配服务
type Remote struct {
	component.Base
	app      pitaya.Pitaya
	handlers map[string]func(*game.Table, context.Context, proto.Message) (proto.Message, error)
}

// NewRemote 创建独立的匹配服务
func NewRemote(app pitaya.Pitaya) *Remote {
	return &Remote{
		app:      app,
		handlers: make(map[string]func(*game.Table, context.Context, proto.Message) (proto.Message, error)),
	}
}

// Init 组件初始化
func (m *Remote) Init() {
	m.handlers[utils.TypeUrl(&sproto.AddTableReq{})] = (*game.Table).HandleAddTable
	m.handlers[utils.TypeUrl(&sproto.AddPlayerReq{})] = (*game.Table).HandleAddPlayer
	m.handlers[utils.TypeUrl(&sproto.CancelTableReq{})] = (*game.Table).HandleCancelTable
	m.handlers[utils.TypeUrl(&sproto.NetStateReq{})] = (*game.Table).HandleNetState
}

// Message 处理匹配服务消息
func (m *Remote) Message(ctx context.Context, req *sproto.GameReq) (*sproto.GameAck, error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Errorf("panic recovered %s\n %s", r, string(debug.Stack()))
		}
	}()
	if req == nil {
		return nil, errors.New("nil request")
	}
	logger.Log.Info(req.String(), req.Req.TypeUrl)

	msg, err := req.Req.UnmarshalNew()
	if err != nil {
		return nil, err
	}

	table := game.GetTableManager().Get(req.Matchid, req.Tableid)
	if req.Req.TypeUrl == utils.TypeUrl(&sproto.AddTableReq{}) {
		table = game.GetTableManager().LoadOrStore(req.Matchid, req.Tableid)
	}
	if table == nil {
		return nil, fmt.Errorf("table not found for match %d, table %d", req.Matchid, req.Tableid)
	}

	if handler, ok := m.handlers[req.Req.TypeUrl]; ok {
		rsp, err := handler(table, ctx, msg)
		if err != nil {
			return nil, err
		}
		return m.newGameAck(req, rsp)
	}
	return nil, errors.New("invalid request type")
}

func (m *Remote) newGameAck(req *sproto.GameReq, ack proto.Message) (*sproto.GameAck, error) {
	data, err := anypb.New(ack)
	if err != nil {
		return nil, err
	}
	return &sproto.GameAck{
		Matchid: req.GetMatchid(),
		Tableid: req.GetTableid(),
		Ack:     data,
	}, nil
}
