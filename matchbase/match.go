package matchbase

import (
	"context"

	"github.com/kevin-chtw/tw_common/utils"
	"github.com/kevin-chtw/tw_proto/cproto"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Match struct {
	Sub       any
	App       pitaya.Pitaya
	Conf      *Config
	Playermgr *Playermgr
	tableIds  *TableIDs
}

func NewMatch(app pitaya.Pitaya, conf *Config, sub any) *Match {
	return &Match{
		Sub:       sub,
		App:       app,
		Conf:      conf,
		tableIds:  NewTableIDs(),
		Playermgr: NewPlayermgr(),
	}
}

func (m *Match) NewMatchAck(ctx context.Context, msg proto.Message) ([]byte, error) {
	data, err := anypb.New(msg)
	if err != nil {
		return nil, err
	}
	out := &cproto.MatchAck{
		Serverid: m.App.GetServerID(),
		Matchid:  m.Conf.Matchid,
		Ack:      data,
	}
	return utils.Marshal(ctx, out)
}

func (m *Match) nextTableID() int32 {
	id, err := m.tableIds.Take()
	if err != nil {
		logger.Log.Error(err.Error())
		return 0
	}
	return id
}

func (m *Match) PutBackTableId(id int32) {
	m.tableIds.PutBack(id)
}
