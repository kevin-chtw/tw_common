package matchbase

import (
	"context"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/kevin-chtw/tw_proto/sproto"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type MatchCreator func(pitaya.Pitaya, string) *Match
type PlayerCreator func(context.Context, string, int32, int64) *Player

var (
	matchCreator    MatchCreator
	playerCreator   PlayerCreator
	defaultMatchmgr *Matchmgr
)

// Init 初始化游戏模块
func Init(app pitaya.Pitaya, mc MatchCreator, pc PlayerCreator) {
	matchCreator = mc
	playerCreator = pc
	defaultMatchmgr = NewMatchmgr(app)
}

func GetMatch(matchid int32) *Match {
	return defaultMatchmgr.Get(matchid)
}

// Matchmgr 管理玩家
type Matchmgr struct {
	App    pitaya.Pitaya
	Matchs map[int32]*Match
	ticker *time.Ticker
}

// NewMatchmgr 创建玩家管理器
func NewMatchmgr(app pitaya.Pitaya) *Matchmgr {
	m := &Matchmgr{
		App:    app,
		Matchs: make(map[int32]*Match),
		ticker: time.NewTicker(time.Second),
	}
	if err := m.LoadMatchs(); err != nil {
		logger.Log.Panicf("加载比赛配置失败: %v", err)
		return nil
	}
	// 启动40秒定时上报match人数
	go m.startReportPlayerCount()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Log.Errorf("panic recovered %s\n %s", r, string(debug.Stack()))
			}
		}()
		for range m.ticker.C {
			m.tick()
		}
	}()
	return m
}

func (m *Matchmgr) tick() {
	for _, match := range m.Matchs {
		match.Sub.Tick()
	}
}

func (m *Matchmgr) LoadMatchs() error {
	files, err := filepath.Glob(filepath.Join("etc", m.App.GetServer().Type, "*.yaml"))
	if err != nil {
		return err
	}
	for _, file := range files {
		logger.Log.Infof("加载比赛配置: %s", file)
		match := matchCreator(m.App, file)
		m.Add(match)
	}
	return nil
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
			Name:          match.Viper.GetString("name"),
			GameType:      match.Viper.GetString("game_type"),
			MatchType:     m.App.GetServer().Type,
			Serverid:      m.App.GetServerID(),
			SignCondition: match.Viper.GetString("sign_condition"),
			Online:        int32(match.playermgr.playerCount()),
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
	m.Matchs[match.Viper.GetInt32("matchid")] = match
}

func (m *Matchmgr) Get(matchId int32) *Match {
	return m.Matchs[matchId]
}
