package game

import (
	"context"
	"errors"
	"sync"
	"time"

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
	MatchType     string //
	MatchID       int32  // 比赛ID
	App           pitaya.Pitaya
	tableID       int32              // 桌号
	matchServerId string             // 匹配服务ID
	players       map[string]*Player // 玩家ID -> Player
	scoreBase     int64              // 分数基数
	gameCount     int32              // 游戏局数
	curGameCount  int32              // 当前局数
	playerCount   int32              // 玩家数量
	property      string             // 游戏配置
	creator       string             // 创建者ID
	description   string             // 房间描述
	fdproperty    map[string]int32   // 房间属性
	lastHandData  any
	handlers      map[string]func(*Player, proto.Message) error
	gameMutex     sync.Mutex // 保护game的对象锁
	game          IGame      // 游戏逻辑处理接口

	historyMsg     map[string][]*cproto.GameAck
	historyJsonMsg map[string][]*cproto.GameAck
	historyMutex   sync.Mutex // 保护historyMsg的锁
	gameOnce       sync.Once  // 确保每局游戏结束只执行一次

	//dissolveMutex sync.Mutex // 保护dissovle的对象锁
	dissovle *cproto.GameDissolveAck
}

// NewTable 创建新的游戏桌实例
func NewTable(matchID, tableID int32, app pitaya.Pitaya) *Table {
	t := &Table{
		MatchID:        matchID,
		tableID:        tableID,
		curGameCount:   0,
		matchServerId:  "",
		players:        make(map[string]*Player),
		fdproperty:     make(map[string]int32),
		App:            app,
		handlers:       make(map[string]func(*Player, proto.Message) error),
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
	t.handlers[TypeUrl(&cproto.GameReadyReq{})] = t.handleGameReady
	t.handlers[TypeUrl(&cproto.GameDissolveReq{})] = t.handleGameDissolve
	t.handlers[TypeUrl(&cproto.TableMsgReq{})] = t.handleTableMsg
}

// OnPlayerMsg 处理玩家消息
func (t *Table) OnPlayerMsg(ctx context.Context, player *Player, req *cproto.GameReq) error {
	if req == nil || req.Req == nil {
		return errors.New("invalid request")
	}
	msg, err := req.Req.UnmarshalNew()
	if err != nil {
		return err
	}
	logger.Log.Infof("player %s recive msg %v", player.ack.Uid, utils.JsonMarshal.Format(msg))
	if handler, ok := t.handlers[req.Req.TypeUrl]; ok {
		return handler(player, msg)
	}

	return errors.New("unknown request type")
}

// handleEnterGame 处理玩家进入游戏请求
func (t *Table) handleEnterGame(player *Player, _ proto.Message) error {
	if !t.isOnTable(player.ack.Uid) {
		return errors.New("player not on table")
	}

	player.enter = true
	t.sendEnterGame(player)
	if player.entered {
		t.notifyTablePlayer(player, true)
		t.sendHisMsges(player)
	} else {
		player.entered = true
		t.broadcast(player.ack)
		t.notifyTablePlayer(player, false)
		t.checkBegin()
	}

	return nil
}

func (t *Table) handleGameReady(player *Player, msg proto.Message) error {
	if !t.isOnTable(player.ack.Uid) {
		return errors.New("player not on table")
	}
	req := msg.(*cproto.GameReadyReq)
	if player.ack.Ready == req.Ready {
		return errors.New("ready state not changed")
	}
	player.ack.Ready = req.Ready
	t.broadcastReady(player)
	if req.Ready {
		t.checkBegin()
	}
	return nil
}

func (t *Table) handleGameDissolve(player *Player, msg proto.Message) error {
	req := msg.(*cproto.GameDissolveReq)
	if !t.isOnTable(player.ack.Uid) {
		return errors.New("player not on table")
	}
	//t.dissolveMutex.Lock()
	if t.dissovle == nil {
		t.dissovle = &cproto.GameDissolveAck{
			Starttime: time.Now().Unix(),
			Endtime:   time.Now().Add(5 * time.Minute).Unix(),
			Seat:      player.GetSeat(),
			Agreed:    make(map[int32]bool),
		}
	}
	t.dissovle.Agreed[player.ack.Seat] = req.Agree
	t.broadcast(t.dissovle)
	if req.Agree {
		t.checkDissolve()
	} else {
		ack := &cproto.GameDissolveResultAck{
			Dissovle: false,
		}
		t.broadcast(ack)
		t.dissovle = nil
	}
	return nil
}

func (t *Table) checkDissolve() {
	//t.dissolveMutex.Lock()

	if t.dissovle == nil {
		return
	}
	if t.dissovle.Endtime >= time.Now().Unix() && len(t.dissovle.Agreed) < int(t.playerCount) {
		return
	}
	t.dissovle = nil
	//t.dissolveMutex.Unlock()
	for _, p := range t.players {
		p.ack.Ready = false
	}

	ack := &cproto.GameDissolveResultAck{
		Dissovle: true,
	}
	t.broadcast(ack)
	t.gameOver()
}

func (t *Table) gameBegin() {
	t.gameMutex.Lock()
	defer t.gameMutex.Unlock()
	t.curGameCount++
	// 重置gameOnce以允许新一局游戏的NotifyGameOver执行
	t.gameOnce = sync.Once{}
	t.sendGameBegin()
	t.historyMsg = make(map[string][]*cproto.GameAck)
	t.game = CreateGame(t.App.GetServer().Type, t, t.curGameCount)
	t.game.OnGameBegin()
}

func (t *Table) handleTableMsg(player *Player, msg proto.Message) error {
	t.gameMutex.Lock()
	defer t.gameMutex.Unlock()

	data := msg.(*cproto.TableMsgReq).GetMsg()
	if t.game != nil && data != nil {
		return t.game.OnPlayerMsg(player, data)
	}
	return errors.New("game not started")
}

func (t *Table) broadcastReady(player *Player) {
	ack := &cproto.GameReadyAck{
		Ready: player.ack.Ready,
		Seat:  player.ack.Seat,
	}
	t.broadcast(ack)
}

func (t *Table) sendGameBegin() {
	ack := &cproto.GameBeginAck{
		CurGameCount: t.curGameCount,
	}
	t.broadcast(ack)
}
func (t *Table) sendGameOver() {
	ack := &cproto.GameOverAck{
		CurGameCount: t.curGameCount,
		Ready:        make([]bool, 0),
	}
	for i := range t.players {
		t.players[i].ack.Ready = false
		ack.Ready = append(ack.Ready, false)
	}
	t.broadcast(ack)
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
		Serverid: t.App.GetServerID(),
		Tableid:  t.tableID,
		Matchid:  t.MatchID,
		Ack:      data,
	}
}

func (t *Table) isOnTable(playerID string) bool {
	if _, ok := t.players[playerID]; ok {
		return true
	}
	return false
}

func (t *Table) checkBegin() {
	if len(t.players) < int(t.playerCount) {
		return
	}

	for _, player := range t.players {
		if !player.enter || (!player.ack.Ready && t.MatchType == "fdtable") {
			return
		}
	}
	t.gameBegin()
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

	rsp, err := t.send2Account(&sproto.PlayerInfoReq{Uid: req.Playerid})
	if err != nil {
		return nil, err
	}
	ack := rsp.(*sproto.AccountAck)
	account, err := ack.Ack.UnmarshalNew()
	if err != nil {
		return nil, err
	}
	player, err := playerManager.Store(account.(*sproto.PlayerInfoAck), req.Seat, req.Score)
	if err != nil {
		return nil, err
	}
	player.Ctx = ctx
	t.players[req.Playerid] = player

	return &sproto.EmptyAck{}, nil
}

func (t *Table) HandleCancelTable(ctx context.Context, msg proto.Message) (proto.Message, error) {
	ack := &cproto.GameDissolveResultAck{
		Dissovle: true,
	}
	t.broadcast(ack)
	t.gameOver()
	return &sproto.EmptyAck{}, nil
}

func (t *Table) HandleExitTable(ctx context.Context, msg proto.Message) (proto.Message, error) {
	req := msg.(*sproto.ExitTableReq)
	if !t.isOnTable(req.Playerid) {
		return nil, errors.New("player not on table")
	}

	if t.curGameCount > 0 {
		return nil, errors.New("game is not over")
	}

	delete(t.players, req.Playerid)
	playerManager.Delete(req.Playerid) // 从玩家管理器中删除玩家
	ack := &cproto.GameExitAck{
		Uid: req.Playerid,
	}
	t.broadcast(ack)
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

	t.gameMutex.Lock()
	defer t.gameMutex.Unlock()
	if t.game != nil {
		t.game.OnNetChange(player, req.Online)
	}

	return &sproto.NetStateAck{Uid: req.Uid}, nil
}

func (t *Table) NotifyGameOver(gameId int32, roundData string) {
	if gameId != t.curGameCount {
		return
	}

	t.gameOnce.Do(func() {
		result := &sproto.GameResultReq{
			Tableid:      t.tableID,
			CurGameCount: t.curGameCount,
			Scores:       make(map[string]int64),
			PlayerData:   make(map[string]string),
			RoundData:    roundData,
		}

		for _, p := range t.players {
			result.Scores[p.ack.Uid] = p.score
			result.PlayerData[p.ack.Uid] = p.GetDatas()
		}
		t.Send2Match(result)
		t.sendGameOver()
		if t.curGameCount >= t.gameCount {
			go t.gameOver()
		} else {
			go t.checkBegin()
		}
	})
}

func (t *Table) gameOver() {
	gameOver := &sproto.GameOverReq{
		CurGameCount: t.curGameCount,
		Tableid:      t.tableID,
	}
	t.Send2Match(gameOver)
	for _, player := range t.players {
		playerManager.Delete(player.ack.Uid) // 从玩家管理器中删除玩家
	}
	tableManager.Delete(t.MatchID, t.tableID) // 从桌子管理器中删除

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
	if err := t.App.RPC(context.Background(), "account.remote.message", ack, req); err != nil {
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
		Matchid: t.MatchID,
		Req:     data,
	}
	ack := &sproto.MatchAck{}
	if err := t.App.RPCTo(context.Background(), t.matchServerId, t.MatchType+".remote.message", ack, req); err != nil {
		logger.Log.Error(err.Error())
	}
}

func (t *Table) Send2Player(ack proto.Message, seat int32) {
	logger.Log.Infof("seat: %d ack: %v", seat, utils.JsonMarshal.Format(ack))

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
	t.checkDissolve()
	t.gameMutex.Lock()
	defer t.gameMutex.Unlock()
	if t.game != nil {
		t.game.OnGameTimer()
	}
}

func (t *Table) broadcast(ack proto.Message) {
	logger.Log.Info(ack)
	msg := t.newMsg(ack)
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

	if _, err := t.App.SendPushToUsers(t.App.GetServer().Type, data, []string{player.ack.Uid}, "proxy"); err != nil {
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
	if t.dissovle != nil {
		t.sendMsg(t.newMsg(t.dissovle), player)
	}
}

// 辅助函数：获取历史消息
func (t *Table) getHistoryMsg(player *Player) []*cproto.GameAck {
	if utils.IsWebsocket(player.Ctx) {
		return t.historyJsonMsg[player.ack.Uid]
	}
	return t.historyMsg[player.ack.Uid]
}
