package matchbase

import (
	"context"

	"github.com/kevin-chtw/tw_proto/cproto"
	"github.com/kevin-chtw/tw_proto/sproto"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Table struct {
	Match *Match
	ID    int32
}

func NewTable(m *Match) *Table {
	return &Table{
		Match: m,
		ID:    m.nextTableID(),
	}
}

func (t *Table) SendAddTableReq(scorebase int64, gameCount int32, property string, fdproperty map[string]int32) {
	req := &sproto.AddTableReq{
		Property:    property,
		ScoreBase:   scorebase,
		MatchType:   t.Match.App.GetServer().Type,
		GameCount:   gameCount,
		PlayerCount: t.Match.Conf.PlayerPerTable,
		Fdproperty:  fdproperty,
	}
	t.send2Game(req)
}

func (t *Table) SendAddPlayer(player *Player, seat int32) {
	req := &sproto.AddPlayerReq{
		Playerid: player.ID,
		Seat:     seat,
		Score:    player.Score,
	}
	t.send2Game(req)
}

func (t *Table) SendExitTableReq(player *Player) *sproto.ExitTableAck {
	req := &sproto.ExitTableReq{
		Playerid: player.ID,
	}
	rsp := t.send2Game(req)
	if rsp.Ack == nil {
		return nil
	}
	ack, err := rsp.Ack.UnmarshalNew()
	if err != nil {
		logger.Log.Error(err.Error())
		return nil
	}
	return ack.(*sproto.ExitTableAck)
}

func (t *Table) SendStartClient(p *Player) {
	startClientAck := &cproto.StartClientAck{
		MatchType: t.Match.App.GetServer().Type,
		GameType:  t.Match.Conf.GameType,
		ServerId:  t.Match.App.GetServerID(),
		MatchId:   t.Match.Conf.Matchid,
		TableId:   t.ID,
	}
	data, err := t.Match.NewMatchAck(p.Ctx, startClientAck)
	if err != nil {
		logger.Log.Errorf("Failed to send start client ack: %v", err)
		return
	}
	t.Match.App.SendPushToUsers(t.Match.App.GetServer().Type, data, []string{p.ID}, "proxy")
}

func (t *Table) SendNetState(player *Player, online bool) {
	req := &sproto.NetStateReq{
		Uid:    player.ID,
		Online: online,
	}
	t.send2Game(req)
}

func (t *Table) send2Game(msg proto.Message) *sproto.GameAck {
	data, err := anypb.New(msg)
	if err != nil {
		logger.Log.Errorf("Failed to encode message: %v", err)
		return nil
	}

	req := &sproto.GameReq{
		Matchid: t.Match.Conf.Matchid,
		Tableid: t.ID,
		Req:     data,
	}
	rsp := &sproto.GameAck{}
	if err = t.Match.App.RPC(context.Background(), t.Match.Conf.GameType+".remote.message", rsp, req); err != nil {
		logger.Log.Errorf("Failed to send message: %v", err)
	}
	return rsp
}
