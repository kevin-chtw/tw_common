package mahjong

import (
	"github.com/kevin-chtw/tw_common/gamebase/game"
)

type IGame interface {
	OnStart()
	OnReqMsg(player *game.Player, data []byte) error
}

type Game struct {
	IGame
	*game.Table
	id        int32
	timer     *Timer
	CurState  IState
	nextState IState
	rule      *Rule
	players   []*Player
}

func NewGame(subGame IGame, t *game.Table, id int32) *Game {
	g := &Game{
		IGame:   subGame,
		Table:   t,
		id:      id,
		timer:   NewTimer(),
		rule:    NewRule(),
		players: make([]*Player, t.GetPlayerCount()),
	}

	g.rule.LoadRule(t.GetProperty(), Service.GetDefaultRules())
	if t.MatchType == "fdtable" {
		g.rule.LoadFdRule(t.GetFdproperty(), Service.GetFdRules())
	}
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
	if err := g.IGame.OnReqMsg(player, data); err != nil {
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
	if p := g.GetPlayer(player.GetSeat()); p != nil {
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

func (g *Game) GetRestCount() int {
	count := 0
	for _, p := range g.players {
		if !p.IsOut() {
			count++
		}
	}
	return count
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
