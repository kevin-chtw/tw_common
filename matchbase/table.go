package matchbase

import (
	"context"
	"errors"

	"github.com/kevin-chtw/tw_proto/cproto"
	"github.com/kevin-chtw/tw_proto/sproto"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Table struct {
	Sub         any
	Match       *Match
	ID          int32
	PlayerCount int32
	Players     map[string]*Player
}

func NewTable(m *Match, sub any) *Table {
	return &Table{
		Sub:         sub,
		Match:       m,
		ID:          m.nextTableID(),
		PlayerCount: m.Viper.GetInt32("player_per_table"),
		Players:     make(map[string]*Player),
	}
}

func (t *Table) IsOnTable(player *Player) bool {
	for _, p := range t.Players {
		if p.ID == player.ID {
			return true
		}
	}
	return false
}

func (t *Table) NetChange(player *Player, online bool) error {
	if !t.IsOnTable(player) {
		return errors.New("player is not on table")
	}
	if err := t.SendNetState(player, online); err != nil {
		return err
	}
	if online {
		t.SendStartClient(player)
	}
	return nil
}

func (t *Table) AddPlayer(player *Player) error {
	if len(t.Players) >= int(t.PlayerCount) {
		return errors.New("table is full")
	}

	if t.IsOnTable(player) {
		return errors.New("player already exists on table")
	}

	player.Seat = t.getSeat()
	player.TableId = t.ID
	t.Players[player.ID] = player
	if err := t.SendAddPlayer(player); err != nil {
		// 发送失败时清理本地状态，避免不一致
		delete(t.Players, player.ID)
		player.Seat = -1
		player.TableId = 0
		return err
	}
	return nil
}

func (t *Table) SendAddTableReq(gameCount int32, creator string, fdproperty map[string]int32) error {
	req := &sproto.AddTableReq{
		Property:    t.Match.Viper.GetString("property"),
		ScoreBase:   t.Match.Viper.GetInt64("score_base"),
		MatchType:   t.Match.App.GetServer().Type,
		GameCount:   gameCount,
		PlayerCount: t.PlayerCount,
		Fdproperty:  fdproperty,
		Creator:     creator,
	}
	_, err := t.send2Game(req)
	if err != nil {
		logger.Log.Errorf("Failed to send add table request: %v", err)
		return err
	}
	return nil
}

func (t *Table) SendAddPlayer(player *Player) error {
	req := &sproto.AddPlayerReq{
		Playerid: player.ID,
		Seat:     player.Seat,
		Score:    player.Score,
		Bot:      player.Bot,
	}
	_, err := t.send2Game(req)
	if err != nil {
		logger.Log.Errorf("Failed to send add player request: %v", err)
		return err
	}
	return nil
}

func (t *Table) SendExitTableReq(player *Player) error {
	req := &sproto.ExitTableReq{
		Playerid: player.ID,
	}
	rsp, err := t.send2Game(req)
	if err != nil {
		// RPC 失败，直接返回错误，避免"假成功"
		return errors.New("failed to send exit table request: " + err.Error())
	}
	if rsp.Ack == nil {
		// 游戏服返回空响应，视为成功（兼容某些情况下的空应答）
		return nil
	}
	ack, err := rsp.Ack.UnmarshalNew()
	if err != nil {
		logger.Log.Errorf("Failed to unmarshal exit table ack: %v", err)
		return errors.New("failed to unmarshal exit table response")
	}
	exitAck := ack.(*sproto.ExitTableAck)
	if exitAck.Result != 0 {
		return errors.New("player cannot exit table")
	}
	return nil
}

func (t *Table) SendStartClient(p *Player) {
	startClientAck := &cproto.StartClientAck{
		MatchType: t.Match.App.GetServer().Type,
		GameType:  t.Match.Viper.GetString("game_type"),
		ServerId:  t.Match.App.GetServerID(),
		MatchId:   t.Match.Viper.GetInt32("matchid"),
		TableId:   t.ID,
	}
	data, err := t.Match.NewMatchAck(p.Ctx, startClientAck)
	if err != nil {
		logger.Log.Errorf("Failed to send start client ack: %v", err)
		return
	}
	t.Match.App.SendPushToUsers(t.Match.App.GetServer().Type, data, []string{p.ID}, "proxy")
}

func (t *Table) SendNetState(player *Player, online bool) error {
	req := &sproto.NetStateReq{
		Uid:    player.ID,
		Online: online,
	}
	_, err := t.send2Game(req)
	if err != nil {
		logger.Log.Errorf("Failed to send net state: %v", err)
		return err
	}
	return nil
}

func (t *Table) SendCancelTableReq() error {
	req := &sproto.CancelTableReq{
		Reason: 1,
	}
	_, err := t.send2Game(req)
	if err != nil {
		logger.Log.Errorf("Failed to send cancel table request: %v", err)
		return err
	}
	return nil
}

func (t *Table) send2Game(msg proto.Message) (*sproto.GameAck, error) {
	data, err := anypb.New(msg)
	if err != nil {
		logger.Log.Errorf("Failed to encode message: %v", err)
		return nil, err
	}

	req := &sproto.GameReq{
		Matchid: t.Match.Viper.GetInt32("matchid"),
		Tableid: t.ID,
		Req:     data,
	}
	rsp := &sproto.GameAck{}
	if err = t.Match.App.RPC(context.Background(), t.Match.Viper.GetString("game_type")+".remote.message", rsp, req); err != nil {
		logger.Log.Errorf("Failed to send message to game server: %v", err)
		return nil, err
	}
	return rsp, nil
}

func (t *Table) getSeat() int32 {
	for i := range t.PlayerCount {
		if !t.isUsed(int32(i)) {
			return int32(i)
		}
	}
	return -1
}

func (t *Table) isUsed(seat int32) bool {
	for _, p := range t.Players {
		if p.Seat == seat {
			return true
		}
	}
	return false
}
