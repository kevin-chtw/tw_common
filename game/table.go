package game

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/kevin-chtw/tw_proto/cproto"
	"github.com/kevin-chtw/tw_proto/sproto"
	"github.com/sirupsen/logrus"

	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type IGame interface {
	// OnGameBegin 游戏开始
	OnGameBegin()
	// OnPlayerMsg 处理玩家消息
	OnPlayerMsg(player *Player, data []byte) error
	// OnGameTimer 每秒调用一次
	OnGameTimer()
	// OnNetChange 处理玩家网络状态变化
	OnNetChange(player *Player, offline bool)
}

const (
	// TableStatusPreparing 桌子状态：准备中
	TableStatusPreparing = "preparing"
	// TableStatusPlaying 桌子状态：游戏中
	TableStatusPlaying = "playing"
	// TableStatusFinished 桌子状态：已结束
	TableStatusFinished = "finished"
)

// Table 表示一个游戏桌实例
type Table struct {
	gameID        int32              // 游戏ID
	matchID       int32              // 比赛ID
	tableID       int32              // 桌号
	matchServerId string             // 匹配服务ID
	players       map[string]*Player // 玩家ID -> Player
	status        string             // "preparing", "playing", "finished"
	app           pitaya.Pitaya
	matchType     string           //
	scoreBase     int64            // 分数基数
	gameCount     int32            // 游戏局数
	playerCount   int32            // 玩家数量
	property      string           // 游戏配置
	fdproperty    map[string]int32 // 房间属性
	lastHandData  any
	ticker        *time.Ticker // 定时器
	done          chan bool    // 停止信号
	handlers      map[string]func(*Player, *cproto.GameReq) error
	gameMutex     sync.Mutex // 保护game的对象锁
	game          IGame      // 游戏逻辑处理接口
	historyMsg    map[string][]*cproto.GameAck
	historyMutex  sync.Mutex // 保护historyMsg的锁
}

// NewTable 创建新的游戏桌实例
func NewTable(gameID, matchID, tableID int32, app pitaya.Pitaya) *Table {
	t := &Table{
		gameID:        gameID,
		matchID:       matchID,
		tableID:       tableID,
		matchServerId: "",
		players:       make(map[string]*Player),
		status:        TableStatusPreparing,
		fdproperty:    make(map[string]int32),
		app:           app,
		done:          make(chan bool),
		handlers:      make(map[string]func(*Player, *cproto.GameReq) error),
		gameMutex:     sync.Mutex{},
		game:          nil,
		historyMsg:    make(map[string][]*cproto.GameAck),
	}

	t.init()
	return t
}

func TypeUrl(src proto.Message) string {
	any, err := anypb.New(src)
	if err != nil {
		logrus.Error(err)
		return ""
	}
	return any.GetTypeUrl()
}

func (t *Table) init() {
	t.handlers[TypeUrl(&cproto.EnterGameReq{})] = t.handleEnterGame
	t.handlers[TypeUrl(&cproto.TableMsgReq{})] = t.handleTableMsg
}

// OnPlayerMsg 处理玩家消息
func (t *Table) OnPlayerMsg(ctx context.Context, player *Player, req *cproto.GameReq) error {
	if req == nil || req.Req == nil {
		return errors.New("invalid request")
	}
	logger.Log.Infof("Received player message: %v", req.Req.TypeUrl)
	if handler, ok := t.handlers[req.Req.TypeUrl]; ok {
		return handler(player, req)
	}

	return errors.New("unknown request type")
}

// handleEnterGame 处理玩家进入游戏请求
func (t *Table) handleEnterGame(player *Player, _ *cproto.GameReq) error {
	if !t.isOnTable(player.id) {
		return errors.New("player not on table")
	}

	if player.Status == PlayerStatusEnter {
		t.notifyTablePlayer(player, true)
		t.sendHisMsges(player)
	} else {
		player.Status = PlayerStatusEnter
		t.broadcastTablePlayer(player)
		t.notifyTablePlayer(player, false)
	}

	// 检查是否满足开赛条件
	if t.isAllPlayersReady() {
		t.status = TableStatusPlaying
		t.gameMutex.Lock()
		defer t.gameMutex.Unlock()
		t.game = CreateGame(t.gameID, t)
		t.game.OnGameBegin()
	}
	return nil
}

func (t *Table) handleTableMsg(player *Player, req *cproto.GameReq) error {
	t.gameMutex.Lock()
	defer t.gameMutex.Unlock()
	msg := &cproto.TableMsgReq{}
	if err := proto.Unmarshal(req.Req.Value, msg); err != nil {
		return err
	}
	data := msg.GetMsg()
	if t.game != nil && data != nil {
		return t.game.OnPlayerMsg(player, data)
	}
	return errors.New("game not started")
}

func (t *Table) broadcastTablePlayer(player *Player) {
	ack := &cproto.TablePlayerAck{
		Playerid: player.id,
		Seatnum:  player.Seat,
	}
	msg := t.newMsg(ack)
	t.broadcast(msg)
	logger.Log.Infof("Player %s added to table %d", player.id, t.tableID)
}

func (t *Table) notifyTablePlayer(player *Player, resume bool) {
	for _, p := range t.players {
		if p.id == player.id && !resume {
			continue
		}
		ack := &cproto.TablePlayerAck{
			Playerid: p.id,
			Seatnum:  p.Seat,
		}
		msg := t.newMsg(ack)
		t.sendMsg(msg, []string{player.id})
	}
}

func (t *Table) newMsg(ack proto.Message) *cproto.GameAck {
	data, err := anypb.New(ack)
	if err != nil {
		return nil
	}
	return &cproto.GameAck{
		Serverid: t.app.GetServerID(),
		Gameid:   t.gameID,
		Tableid:  t.tableID,
		Matchid:  t.matchID,
		Ack:      data,
	}
}

func (t *Table) isOnTable(playerID string) bool {
	if _, ok := t.players[playerID]; ok {
		return true
	}
	return false
}

func (t *Table) isAllPlayersReady() bool {
	if len(t.players) != int(t.playerCount) {
		return false
	}
	// for _, player := range t.players {
	// 	if player.Status != PlayerStatusReady {
	// 		return false
	// 	}
	// }
	return true
}

// HandleStartGame 处理开始游戏请求
func (t *Table) HandleAddTable(ctx context.Context, msg proto.Message) (proto.Message, error) {
	req := msg.(*sproto.AddTableReq)

	t.status = TableStatusPreparing
	t.matchType = req.GetMatchType()
	t.scoreBase = int64(req.GetScoreBase())
	t.gameCount = req.GetGameCount()
	t.playerCount = req.GetPlayerCount()
	t.property = req.GetProperty()
	t.fdproperty = req.GetFdproperty()
	return &sproto.EmptyAck{}, nil
}

func (t *Table) HandleAddPlayer(ctx context.Context, msg proto.Message) (proto.Message, error) {
	req := msg.(*sproto.AddPlayerReq)
	if t.isOnTable(req.Playerid) {
		return nil, errors.New("player already on table")
	}

	player, err := playerManager.Store(req.Playerid)
	if err != nil {
		return nil, err
	}
	player.SetSeat(req.Seat)
	player.AddScore(req.Score)
	t.players[req.Playerid] = player

	return &sproto.EmptyAck{}, nil
}

func (t *Table) HandleCancelTable(ctx context.Context, msg proto.Message) (proto.Message, error) {
	if t.status == TableStatusPlaying {
		return nil, errors.New("cannot cancel a playing table")
	}
	for _, player := range t.players {
		playerManager.Delete(player.id) // 从玩家管理器中删除玩家
	}
	tableManager.Delete(t.matchID, t.tableID) // 从桌子管理器中删除

	// 停止定时器
	if t.ticker != nil {
		t.ticker.Stop()
		t.done <- true
	}

	// 清理game对象
	t.gameMutex.Lock()
	t.game = nil
	t.gameMutex.Unlock()

	return &sproto.EmptyAck{}, nil
}

func (t *Table) HandleNetState(ctx context.Context, msg proto.Message) (proto.Message, error) {
	req := msg.(*sproto.NetStateReq)
	player := playerManager.Get(req.Uid)
	if player == nil {
		return nil, errors.New("player not found")
	}

	if player.online == req.Online {
		return nil, errors.New("player online status not changed")
	}
	player.online = req.Online

	if t.game != nil {
		t.game.OnNetChange(player, req.Online)
	}

	return &sproto.NetStateAck{Uid: req.Uid}, nil
}

func (t *Table) NotifyGameOver() {
	t.status = TableStatusFinished
	result := &sproto.TableResultAck{}
	t.Send2Match(result)
}

func (t *Table) Send2Match(msg proto.Message) {
	data, err := anypb.New(msg)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	ack := &sproto.Match2GameAck{
		Gameid:  t.gameID,
		Matchid: t.matchID,
		Tableid: t.tableID,
		Ack:     data,
	}
	req := &sproto.Match2GameReq{}
	if err := t.app.RPCTo(context.Background(), t.matchServerId, t.matchType+".game.message", req, ack); err != nil {
		logger.Log.Error(err.Error())
	}
}

func (t *Table) Send2Player(ack proto.Message, seat int32) {
	data, err := protojson.Marshal(ack)
	if err != nil {
		logrus.Error(err.Error())
	}
	tablemsg := &cproto.TableMsgAck{
		Msg: data,
	}
	msg := t.newMsg(tablemsg)
	if seat == SeatAll {
		t.broadcast(msg)
	} else {
		p := t.GetGamePlayer(seat)
		t.sendMsg(msg, []string{p.id})
	}
}

func (t *Table) GetLastGameData() any {
	return t.lastHandData
}

func (t *Table) SetLastGameData(data any) {
	t.lastHandData = data
}

func (g *Table) IsValidSeat(seat int32) bool {
	return seat >= 0 && seat < g.playerCount
}

func (t *Table) GetGamePlayer(seat int32) *Player {
	for _, p := range t.players {
		if p.Seat == seat {
			return p
		}
	}
	return nil
}

func (t *Table) GetPlayerCount() int32 {
	return t.playerCount
}

func (t *Table) GetGameRule() string {
	return t.property
}

func (t *Table) GetScoreBase() int64 {
	return int64(t.scoreBase)
}

func (t *Table) tick() {
	t.gameMutex.Lock()
	defer t.gameMutex.Unlock()
	if t.game != nil {
		t.game.OnGameTimer()
	}
}

func (t *Table) broadcast(msg *cproto.GameAck) {
	uids := make([]string, 0, len(t.players))
	for _, player := range t.players {
		if player.online && player.Status != PlayerStatusUnEnter {
			uids = append(uids, player.id)
		}
	}
	t.sendMsg(msg, uids)
}

func (t *Table) sendMsg(msg *cproto.GameAck, uids []string) {
	logger.Log.Infof("broadcast game message to player %v: %v", uids, msg)
	t.addHisMsg(uids, msg)
	if m, err := t.app.SendPushToUsers(t.app.GetServer().Type, msg, uids, "proxy"); err != nil {
		logger.Log.Errorf("send game message to player %v failed: %v", uids, err)
	} else {
		logger.Log.Debugf("send game message to player %v success: %v", uids, m)
	}
}

func (t *Table) addHisMsg(uids []string, gameAck *cproto.GameAck) {
	t.historyMutex.Lock()
	defer t.historyMutex.Unlock()
	for _, uid := range uids {
		if _, exists := t.historyMsg[uid]; !exists {
			t.historyMsg[uid] = []*cproto.GameAck{}
		}
		t.historyMsg[uid] = append(t.historyMsg[uid], gameAck)
	}
}

func (t *Table) sendHisMsges(player *Player) {
	t.historyMutex.Lock()
	defer t.historyMutex.Unlock()

	if len(t.historyMsg[player.id]) == 0 {
		return
	}

	t.app.SendPushToUsers(t.app.GetServer().Type, t.newMsg(&cproto.HisBeginAck{}), []string{player.id}, "proxy")
	if msgs, exists := t.historyMsg[player.id]; exists {
		for _, msg := range msgs {
			t.app.SendPushToUsers(t.app.GetServer().Type, msg, []string{player.id}, "proxy")
		}
	}
	t.app.SendPushToUsers(t.app.GetServer().Type, t.newMsg(&cproto.HisEndAck{}), []string{player.id}, "proxy")
}
