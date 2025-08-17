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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type IGame interface {
	// OnGameBegin 游戏开始
	OnGameBegin()
	// OnPlayerMsg 处理玩家消息
	OnPlayerMsg(player *Player, data []byte)
	// OnGameTimer 每秒调用一次
	OnGameTimer()
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
	matchType     int32  // 0: 普通匹配, 1: 房卡模式
	scoreBase     int64  // 分数基数
	gameCount     int32  // 游戏局数
	playerCount   int32  // 玩家数量
	gameRule      string // 游戏配置
	lastHandData  any
	ticker        *time.Ticker // 定时器
	done          chan bool    // 停止信号
	handlers      map[string]func(player *Player, req *cproto.GameReq)

	gameMutex sync.Mutex // 保护game的对象锁
	game      IGame      // 游戏逻辑处理接口
}

// NewTable 创建新的游戏桌实例
func NewTable(gameId, matchID, tableID int32, app pitaya.Pitaya) *Table {
	t := &Table{
		gameID:        gameId,
		matchID:       matchID,
		tableID:       tableID,
		matchServerId: "",
		players:       make(map[string]*Player),
		status:        TableStatusPreparing,
		app:           app,
		done:          make(chan bool),
		handlers:      make(map[string]func(player *Player, req *cproto.GameReq)),
		gameMutex:     sync.Mutex{},
		game:          nil,
	}

	// 启动定时器
	t.ticker = time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-t.ticker.C:
				t.onTick()
			case <-t.done:
				return
			}
		}
	}()

	return t
}

func TypeUrl(src proto.Message) string {
	any, err := anypb.New(&cproto.CreateRoomReq{})
	if err != nil {
		logrus.Error(err)
		return ""
	}
	return any.GetTypeUrl()
}

func (t *Table) Init() {
	logrus.Info("PlayerService initialized")

	t.handlers[TypeUrl(&cproto.EnterGameReq{})] = t.handleEnterGame
	t.handlers[TypeUrl(&cproto.TableMsgReq{})] = t.handleTableMsg
}

// OnPlayerMsg 处理玩家消息
func (t *Table) OnPlayerMsg(ctx context.Context, player *Player, req *cproto.GameReq) {
	if req == nil || req.Req == nil {
		return
	}

	if handler, ok := t.handlers[req.Req.TypeUrl]; ok {
		handler(player, req)
	}
}

// handleEnterGame 处理玩家进入游戏请求
func (t *Table) handleEnterGame(player *Player, _ *cproto.GameReq) {
	if !t.isOnTable(player.id) {
		return // 玩家不在桌上
	}

	// 添加玩家到桌中
	t.players[player.id] = player
	t.broadcastTablePlayer(player)
	t.SendTablePlayer(player)
	player.SetStatus(PlayerStatusReady)

	// 检查是否满足开赛条件
	if t.isAllPlayersReady() {
		t.status = TableStatusPlaying
		t.gameMutex.Lock()
		defer t.gameMutex.Unlock()
		t.game = CreateGame(t.gameID, t)
		t.game.OnGameBegin()
	}
}

func (t *Table) handleTableMsg(player *Player, req *cproto.GameReq) {
	t.gameMutex.Lock()
	defer t.gameMutex.Unlock()
	msg := &cproto.TableMsgReq{}
	if err := proto.Unmarshal(req.Req.Value, msg); err != nil {
		return
	}
	data := msg.GetMsg()
	if t.game != nil && data != nil {
		t.game.OnPlayerMsg(player, data)
	}
}

func (t *Table) broadcastTablePlayer(player *Player) {
	ack := &cproto.TablePlayerAck{
		Playerid: player.id,
		Seatnum:  player.Seat,
	}
	msg := t.newMsg(ack)
	t.broadcast(msg)
	logrus.Infof("Player %s added to table %d", player.id, t.tableID)
}

func (t *Table) SendTablePlayer(player *Player) {
	for _, p := range t.players {
		if p.id != player.id {
			ack := &cproto.TablePlayerAck{
				Playerid: p.id,
				Seatnum:  p.Seat,
			}
			msg := t.newMsg(ack)
			t.sendMsg(msg, []string{player.id})
		}
	}
}

func (t *Table) newMsg(ack proto.Message) *cproto.GameAck {
	if ack == nil {
		return nil
	}

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
	for _, player := range t.players {
		if player.Status != PlayerStatusReady {
			return false
		}
	}
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
	t.gameRule = req.GetGameConfig()
	return &sproto.AddTableAck{ErrorCode: int32(0)}, nil
}

func (t *Table) HandleAddPlayer(ctx context.Context, msg proto.Message) (proto.Message, error) {
	req := msg.(*sproto.AddPlayerReq)
	if t.isOnTable(req.Playerid) {
		return nil, errors.New("player already on table")
	}

	player := playerManager.GetPlayer(req.Playerid)
	player.SetSeat(req.Seat)
	player.AddScore(req.Score)
	t.players[req.Playerid] = player

	return &sproto.AddPlayerAck{ErrorCode: int32(0)}, nil
}

func (t *Table) HandleCancelTable(ctx context.Context, msg proto.Message) (proto.Message, error) {
	if t.status == TableStatusPlaying {
		return nil, errors.New("cannot cancel a playing table")
	}
	// 清理玩家状态
	for _, player := range t.players {
		player.SetStatus(PlayerStatusOffline)
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

	return &sproto.CancelTableAck{ErrorCode: int32(0)}, nil
}

func (t *Table) NotifyGameOver() {
	t.status = TableStatusFinished
	t.Send2Match()
}

func (t *Table) Send2Match() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := t.app.RPCTo(ctx, t.matchServerId, "match.game.message", nil, &sproto.Match2GameAck{}); err != nil {
		t.status = "preparing" // 回滚状态
		return err
	}
	return nil
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
	if seat == -1 {
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
	return t.gameRule
}

func (t *Table) GetScoreBase() int64 {
	return int64(t.scoreBase)
}

func (t *Table) onTick() {
	t.gameMutex.Lock()
	defer t.gameMutex.Unlock()
	if t.game != nil {
		t.game.OnGameTimer()
	}
}

func (t *Table) broadcast(msg *cproto.GameAck) {
	players := make([]string, 0, len(t.players))
	for _, player := range t.players {
		if player.Status != PlayerStatusOffline && player.Status != PlayerStatusUnEnter {
			players = append(players, player.id)
		}
	}
	t.sendMsg(msg, players)
}

func (t *Table) sendMsg(msg *cproto.GameAck, players []string) {
	if m, err := t.app.SendPushToUsers("gamemsg", msg, players, "proxy"); err != nil {
		logrus.Errorf("send game message to player %v failed: %v", players, err)
	} else {
		logrus.Infof("send game message to player %v: %v", players, m)
	}
}
