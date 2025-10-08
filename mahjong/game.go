package mahjong

import (
	"errors"

	"github.com/kevin-chtw/tw_common/game"
)

type IGame interface {
	OnStart()
	OnReqMsg(seat int32, data []byte) error
}

type Game struct {
	IGame
	*game.Table
	id          int32
	timer       *Timer
	CurState    IState
	nextState   IState
	rule        *Rule
	players     []*Player
	increasedID int32   // 当前请求ID
	requestIDs  []int32 // 记录每个玩家的请求ID
}

func NewGame(subGame IGame, t *game.Table, id int32) *Game {
	g := &Game{
		IGame:       subGame,
		Table:       t,
		id:          id,
		timer:       NewTimer(),
		rule:        NewRule(),
		players:     make([]*Player, t.GetPlayerCount()),
		increasedID: 1,
		requestIDs:  make([]int32, t.GetPlayerCount()),
	}

	g.rule.LoadRule(t.GetProperty(), Service.GetDefaultRules())
	for i := int32(0); i < t.GetPlayerCount(); i++ {
		g.players[i] = NewPlayer(g, t.GetGamePlayer(i))
	}
	return g
}

func (g *Game) OnGameBegin() {
	g.IGame.OnStart()
	g.enterNextState()
}

func (g *Game) OnPlayerMsg(player *game.Player, data []byte) error {
	seat := player.Seat
	if !g.IsValidSeat(seat) {
		return errors.New("invalid seat")
	}

	if err := g.IGame.OnReqMsg(seat, data); err != nil {
		return err
	}
	g.enterNextState()
	return nil
}

func (g *Game) OnGameTimer() {
	g.timer.OnTick()
	g.enterNextState()
}

func (g *Game) OnGameOver() {
	for i := int32(0); i < g.GetPlayerCount(); i++ {
		g.GetPlayer(i).SyncGameResult()
	}
	g.NotifyGameOver(g.id)
}

func (g *Game) OnNetChange(player *game.Player, offline bool) {
	if p := g.GetPlayer(player.Seat); p != nil {
		p.isOffline = offline
		g.enterNextState()
	}
}

func (g *Game) GetRule() *Rule {
	return g.rule
}

func (g *Game) GetPlayer(seat int32) *Player {
	if g.IsValidSeat(seat) {
		return g.players[seat]
	}
	return nil
}

func (g *Game) SetNextState(newFn func(IGame, ...any) IState, args ...any) {
	g.nextState = newFn(g.IGame, args...)
}

func (g *Game) enterNextState() {
	for g.nextState != nil {
		g.CurState = g.nextState
		g.nextState = nil
		g.timer.Cancel()
		g.CurState.OnEnter()
	}
}

func (g *Game) GetRequestID(seat int32) int32 {
	g.increasedID++
	if g.IsValidSeat(seat) {
		g.requestIDs[seat] = g.increasedID
	} else {
		for i := range g.requestIDs {
			g.requestIDs[i] = g.increasedID
		}
	}
	return g.increasedID
}

func (g *Game) IsRequestID(seat, id int32) bool {
	if !g.IsValidSeat(seat) {
		return false
	}
	return g.requestIDs[seat] == id
}
