package matchbase

import (
	"context"
	"time"

	"github.com/kevin-chtw/tw_proto/sproto"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Matchmgr 管理玩家
type Matchmgr struct {
	App    pitaya.Pitaya
	Matchs map[int32]*Match
}

// NewMatchmgr 创建玩家管理器
func NewMatchmgr(app pitaya.Pitaya) *Matchmgr {
	matchmgr := &Matchmgr{
		App:    app,
		Matchs: make(map[int32]*Match),
	}
	// 启动40秒定时上报match人数
	go matchmgr.startReportPlayerCount()
	return matchmgr
}

// startReportPlayerCount 启动定时上报match人数
func (m *Matchmgr) startReportPlayerCount() {
	ticker := time.NewTicker(40 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.reportPlayerCount()
	}
}

// reportPlayerCount 上报所有match的玩家数量
func (m *Matchmgr) reportPlayerCount() {
	req := &sproto.TourneyUpdateReq{}
	for matchID, match := range m.Matchs {
		info := &sproto.TourneyInfo{
			Id:            matchID,
			Name:          match.Conf.Name,
			GameType:      match.Conf.GameType,
			MatchType:     "fdtable",
			Serverid:      m.App.GetServerID(),
			SignCondition: match.Conf.SignCondition,
			Online:        int32(match.Playermgr.playerCount()),
		}
		req.Infos = append(req.Infos, info)
	}
	m.sendTourneyReq(req)
}

func (m *Matchmgr) sendTourneyReq(msg proto.Message) {
	data, err := anypb.New(msg)
	if err != nil {
		logger.Log.Errorf("failed to create anypb: %v", err)
		return
	}

	req := &sproto.TourneyReq{
		Req: data,
	}
	ack := &sproto.TourneyAck{}
	if err = m.App.RPC(context.Background(), "tourney.remote.message", ack, req); err != nil {
		logger.Log.Errorf("failed to register match to tourney: %v", err)
	}
}

func (m *Matchmgr) Add(match *Match) {
	m.Matchs[match.Conf.Matchid] = match
}

func (m *Matchmgr) Get(matchId int32) *Match {
	return m.Matchs[matchId]
}
