package matchbase

import (
	"context"

	"github.com/kevin-chtw/tw_common/utils"
	"github.com/kevin-chtw/tw_proto/cproto"
	"github.com/spf13/viper"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type IMatch interface {
	Tick()
}

type Match struct {
	Sub       IMatch
	App       pitaya.Pitaya
	Viper     *viper.Viper
	Playermgr *Playermgr
	tableIds  *TableIDs
}

func NewMatch(app pitaya.Pitaya, file string, sub IMatch) *Match {
	m := &Match{
		Sub:       sub,
		App:       app,
		Viper:     viper.New(),
		Playermgr: NewPlayermgr(),
		tableIds:  NewTableIDs(),
	}
	m.initConfig(file)
	return m
}

func (m *Match) initConfig(file string) error {
	m.Viper.SetConfigType("yaml")
	m.Viper.SetConfigFile(file)
	return m.Viper.ReadInConfig()
}

func (m *Match) NewMatchAck(ctx context.Context, msg proto.Message) ([]byte, error) {
	data, err := anypb.New(msg)
	if err != nil {
		return nil, err
	}
	out := &cproto.MatchAck{
		Serverid: m.App.GetServerID(),
		Matchid:  m.Viper.GetInt32("matchid"),
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

func (m *Match) NewStartClientAck(p *Player) *cproto.StartClientAck {
	return &cproto.StartClientAck{
		MatchType: m.App.GetServer().Type,
		GameType:  m.Viper.GetString("game_type"),
		ServerId:  m.App.GetServerID(),
		MatchId:   m.Viper.GetInt32("matchid"),
		TableId:   p.TableId,
	}
}

func (m *Match) PushMsg(p *Player, msg proto.Message) error {
	data, err := m.NewMatchAck(p.Ctx, msg)
	if err != nil {
		logger.Log.Errorf("Failed to send start client ack: %v", err)
		return err
	}
	_, err = m.App.SendPushToUsers(m.App.GetServer().Type, data, []string{p.ID}, "proxy")
	return err
}
