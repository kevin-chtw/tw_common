package game

import (
	"context"
	"errors"
	"sync"

	"github.com/kevin-chtw/tw_common/utils"
	"github.com/kevin-chtw/tw_proto/cproto"
	"github.com/kevin-chtw/tw_proto/sproto"
	"github.com/sirupsen/logrus"

	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
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

// Table 表示一个游戏桌实例
type Table struct {
	MatchType      string             //
	matchID        int32              // 比赛ID
	tableID        int32              // 桌号
	matchServerId  string             // 匹配服务ID
	players        map[string]*Player // 玩家ID -> Player
	app            pitaya.Pitaya
	scoreBase      int64            // 分数基数
	gameCount      int32            // 游戏局数
	curGameCount   int32            // 当前局数
	playerCount    int32            // 玩家数量
	property       string           // 游戏配置
	creator        string           // 创建者ID
	description    string           // 房间描述
	fdproperty     map[string]int32 // 房间属性
	lastHandData   any
	handlers       map[string]func(*Player, *cproto.GameReq) error
	gameMutex      sync.Mutex // 保护game的对象锁
	game           IGame      // 游戏逻辑处理接口
	historyMsg     map[string][]*cproto.GameAck
	historyJsonMsg map[string][]*cproto.GameAck
	historyMutex   sync.Mutex // 保护historyMsg的锁
	gameOnce       sync.Once  // 确保每局游戏结束只执行一次
}

// NewTable 创建新的游戏桌实例
func NewTable(matchID, tableID int32, app pitaya.Pitaya) *Table {
	t := &Table{
		matchID:        matchID,
		tableID:        tableID,
		curGameCount:   0,
		matchServerId:  "",
		players:        make(map[string]*Player),
		fdproperty:     make(map[string]int32),
		app:            app,
		handlers:       make(map[string]func(*Player, *cproto.GameReq) error),
		gameMutex:      sync.Mutex{},
		game:           nil,
		historyMsg:     make(map[string][]*cproto.GameAck),
		historyJsonMsg: make(map[string][]*cproto.GameAck),
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
	if handler, ok := t.handlers[req.Req.TypeUrl]; ok {
		return handler(player, req)
	}

	return errors.New("unknown request type")
}

// handleEnterGame 处理玩家进入游戏请求
func (t *Table) handleEnterGame(player *Player, _ *cproto.GameReq) error {
	if !t.isOnTable(player.ack.Uid) {
		return errors.New("player not on table")
	}

	rsp, err := t.send2Account(&sproto.PlayerInfoReq{Uid: player.ack.Uid})
	if err != nil {
		return err
	}

	ack := rsp.(*sproto.AccountAck)
	msg, err := ack.Ack.UnmarshalNew()
	if err != nil {
		return err
	}

	player.enter = true
	t.sendEnterGame(player)
	player.setAck(msg.(*sproto.PlayerInfoAck))
	if player.entered {
		t.notifyTablePlayer(player, true)
		t.sendHisMsges(player)
	} else {
		player.entered = true
		player.ready = true
		t.broadcastTablePlayer(player)
		t.notifyTablePlayer(player, false)
		// 检查是否满足开赛条件
		if t.isAllPlayersReady() {
			t.gameBegin()
		}
	}

	return nil
}

func (t *Table) gameBegin() {
	t.gameMutex.Lock()
	defer t.gameMutex.Unlock()
	t.curGameCount++
	// 重置gameOnce以允许新一局游戏的NotifyGameOver执行
	t.gameOnce = sync.Once{}
	t.sendGameBegin()
	t.historyMsg = make(map[string][]*cproto.GameAck)
	t.game = CreateGame(t.app.GetServer().Type, t, t.curGameCount)
	t.game.OnGameBegin()
}

func (t *Table) handleTableMsg(player *Player, req *cproto.GameReq) error {
	t.gameMutex.Lock()
	defer t.gameMutex.Unlock()
	msg, err := req.Req.UnmarshalNew()
	if err != nil {
		return err
	}

	data := msg.(*cproto.TableMsgReq).GetMsg()
	if t.game != nil && data != nil {
		return t.game.OnPlayerMsg(player, data)
	}
	return errors.New("game not started")
}

func (t *Table) broadcastTablePlayer(player *Player) {
	msg := t.newMsg(player.ack)
	t.broadcast(msg)
	logger.Log.Infof("Player %s added to table %d", player.ack.Uid, t.tableID)
}

func (t *Table) sendGameBegin() {
	ack := &cproto.GameBeginAck{
		CurGameCount: t.curGameCount,
	}
	msg := t.newMsg(ack)
	t.broadcast(msg)
}
func (t *Table) sendGameOver() {
	ack := &cproto.GameOverAck{
		CurGameCount: t.curGameCount,
	}
	msg := t.newMsg(ack)
	t.broadcast(msg)

	gameOver := &sproto.GameOverReq{
		CurGameCount: t.curGameCount,
	}
	t.Send2Match(gameOver)
}

func (t *Table) notifyTablePlayer(player *Player, resume bool) {
	for _, p := range t.players {
		if p.ack.Uid == player.ack.Uid && !resume {
			continue
		}
		msg := t.newMsg(p.ack)
		t.sendMsg(msg, player)
	}
}

func (t *Table) sendEnterGame(player *Player) {
	ack := &cproto.EnterGameAck{
		Tableid:      t.tableID,
		ScoreBase:    t.scoreBase,
		GameCount:    t.gameCount,
		CurGameCount: t.curGameCount,
		PlayerCount:  t.playerCount,
		Property:     t.property,
		Creator:      t.creator,
		Desn:         t.description,
		Fdproperty:   t.fdproperty,
	}
	msg := t.newMsg(ack)
	t.sendMsg(msg, player)
}

func (t *Table) newMsg(ack proto.Message) *cproto.GameAck {
	data, err := anypb.New(ack)
	if err != nil {
		return nil
	}
	return &cproto.GameAck{
		Serverid: t.app.GetServerID(),
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
	for _, player := range t.players {
		if !player.ready {
			return false
		}
	}
	return true
}

// HandleStartGame 处理开始游戏请求
func (t *Table) HandleAddTable(ctx context.Context, msg proto.Message) (proto.Message, error) {
	req := msg.(*sproto.AddTableReq)
	t.MatchType = req.GetMatchType()
	t.scoreBase = int64(req.GetScoreBase())
	t.gameCount = req.GetGameCount()
	t.playerCount = req.GetPlayerCount()
	t.property = req.GetProperty()
	t.creator = req.GetCreator()
	t.description = req.GetDesn()
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
	player.Ctx = ctx
	player.SetSeat(req.Seat)
	player.AddScore(req.Score)
	t.players[req.Playerid] = player

	return &sproto.EmptyAck{}, nil
}

func (t *Table) HandleCancelTable(ctx context.Context, msg proto.Message) (proto.Message, error) {
	t.gameOver()
	return &sproto.EmptyAck{}, nil
}

func (t *Table) HandleNetState(ctx context.Context, msg proto.Message) (proto.Message, error) {
	req := msg.(*sproto.NetStateReq)
	player := playerManager.Get(req.Uid)
	if player == nil {
		return nil, errors.New("player not found")
	}

	logger.Log.Infof("Player %s online status changed to %v", req.Uid, req.Online)
	if player.online == req.Online {
		return nil, errors.New("player online status not changed")
	}
	player.online = req.Online
	if !req.Online {
		player.enter = false
	}

	if t.game != nil {
		t.game.OnNetChange(player, req.Online)
	}

	return &sproto.NetStateAck{Uid: req.Uid}, nil
}

func (t *Table) NotifyGameOver(gameId int32) {
	if gameId != t.curGameCount {
		return
	}

	t.gameOnce.Do(func() {
		result := &sproto.GameResultReq{
			CurGameCount: t.curGameCount,
			Players:      make([]*sproto.PlayerResult, 0),
		}

		for _, p := range t.players {
			result.Players = append(result.Players, &sproto.PlayerResult{
				Playerid: p.ack.Uid,
				Score:    p.score,
			})
		}
		t.Send2Match(result)

		if t.curGameCount >= t.gameCount {
			go t.gameOver()
		} else {
			go t.gameBegin()
		}
	})
}

func (t *Table) gameOver() {
	t.sendGameOver()
	for _, player := range t.players {
		playerManager.Delete(player.ack.Uid) // 从玩家管理器中删除玩家
	}
	tableManager.Delete(t.matchID, t.tableID) // 从桌子管理器中删除

	// 清理game对象
	t.gameMutex.Lock()
	t.game = nil
	t.gameMutex.Unlock()
}

func (t *Table) send2Account(msg proto.Message) (proto.Message, error) {
	logger.Log.Info(msg)
	data, err := anypb.New(msg)
	if err != nil {
		return nil, nil
	}
	req := &sproto.AccountReq{
		Req: data,
	}
	ack := &sproto.AccountAck{}
	if err := t.app.RPC(context.Background(), "account.remote.message", ack, req); err != nil {
		return nil, err
	}
	return ack, nil
}

func (t *Table) Send2Match(msg proto.Message) {
	logger.Log.Info(msg)
	data, err := anypb.New(msg)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	req := &sproto.MatchReq{
		Matchid: t.matchID,
		Req:     data,
	}
	ack := &sproto.MatchAck{}
	if err := t.app.RPCTo(context.Background(), t.matchServerId, t.MatchType+".remote.message", ack, req); err != nil {
		logger.Log.Error(err.Error())
	}
}

func (t *Table) Send2Player(ack proto.Message, seat int32) {
	logger.Log.Infof("seat: %d ack: %v", seat, ack)

	if seat != SeatAll {
		player := t.GetGamePlayer(seat)
		t.sendTableMsg(ack, player)
	} else {
		for _, player := range t.players {
			t.sendTableMsg(ack, player)
		}
	}
}

func (t *Table) sendTableMsg(ack proto.Message, player *Player) {
	pbData, err := proto.Marshal(ack)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	pbTableMsg := &cproto.TableMsgAck{
		Msg: pbData,
	}
	pbMsg := t.newMsg(pbTableMsg)

	jsonData, err := utils.JsonMarshal.Marshal(ack)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	jsonTableMsg := &cproto.TableMsgAck{
		Msg: jsonData,
	}
	jsonMsg := t.newMsg(jsonTableMsg)

	if utils.IsWebsocket(player.Ctx) {
		t.sendMsg(jsonMsg, player)
	} else {
		t.sendMsg(pbMsg, player)
	}
	t.addHisMsg(player.ack.Uid, pbMsg, jsonMsg)
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
		if p.ack.Seat == seat {
			return p
		}
	}
	return nil
}

func (t *Table) GetPlayerCount() int32 {
	return t.playerCount
}

func (t *Table) GetProperty() string {
	return t.property
}

func (t *Table) GetFdproperty() map[string]int32 {
	return t.fdproperty
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
	for _, player := range t.players {
		t.sendMsg(msg, player)
	}
}

func (t *Table) sendMsg(msg *cproto.GameAck, player *Player) {
	data, err := utils.Marshal(player.Ctx, msg)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	if !player.enter || !player.online {
		return
	}

	if _, err := t.app.SendPushToUsers(t.app.GetServer().Type, data, []string{player.ack.Uid}, "proxy"); err != nil {
		logger.Log.Errorf("player %v failed: %v", player.ack.Uid, err)
	}
}

func (t *Table) addHisMsg(uid string, pbAck, jsonAck *cproto.GameAck) {
	t.historyMutex.Lock()
	defer t.historyMutex.Unlock()
	if _, exists := t.historyMsg[uid]; !exists {
		t.historyMsg[uid] = make([]*cproto.GameAck, 0)
		t.historyJsonMsg[uid] = make([]*cproto.GameAck, 0)
	}
	t.historyMsg[uid] = append(t.historyMsg[uid], pbAck)
	t.historyJsonMsg[uid] = append(t.historyJsonMsg[uid], jsonAck)
}

func (t *Table) sendHisMsges(player *Player) {
	t.historyMutex.Lock()
	defer t.historyMutex.Unlock()

	if len(t.historyMsg[player.ack.Uid]) == 0 {
		return
	}

	t.sendMsg(t.newMsg(&cproto.HisBeginAck{}), player)
	historys := t.getHistoryMsg(player)
	for _, msg := range historys {
		t.sendMsg(msg, player)
	}
	t.sendMsg(t.newMsg(&cproto.HisEndAck{}), player)
}

// 辅助函数：获取历史消息
func (t *Table) getHistoryMsg(player *Player) []*cproto.GameAck {
	if utils.IsWebsocket(player.Ctx) {
		return t.historyJsonMsg[player.ack.Uid]
	}
	return t.historyMsg[player.ack.Uid]
}
