package mahjong

import (
	"slices"

	"github.com/sirupsen/logrus"
)

type IExtraHuTypes interface {
	SelfExtraFans() []int32
	PaoExtraFans() []int32
}

type Play struct {
	ExtraHuTypes IExtraHuTypes
	PlayConf     *PlayConf
	game         *Game
	dealer       *Dealer
	curSeat      int32
	curTile      int32
	banker       int32
	tilesLai     []int32
	history      []Action
	playData     []*PlayData
	huSeats      []int32
	huResult     []*HuResult
	selfCheckers []SelfChecker
	waitcheckers []WaitChecker
}

func NewPlay(game *Game) *Play {
	return &Play{
		game:         game,
		dealer:       NewDealer(game),
		curSeat:      SeatNull,
		curTile:      TileNull,
		banker:       SeatNull,
		tilesLai:     make([]int32, 0),
		history:      make([]Action, 0),
		playData:     make([]*PlayData, game.GetPlayerCount()),
		huSeats:      make([]int32, 0),
		huResult:     make([]*HuResult, game.GetPlayerCount()),
		selfCheckers: make([]SelfChecker, 0),
		waitcheckers: make([]WaitChecker, 0),
	}
}

func (p *Play) RegisterSelfCheck(cks ...SelfChecker) {
	p.selfCheckers = append(p.selfCheckers, cks...)
}
func (p *Play) RegisterWaitCheck(cks ...WaitChecker) {
	p.waitcheckers = append(p.waitcheckers, cks...)
}

func (p *Play) Initialize(pdfn func(int32) *PlayData) {
	lgd := p.getLastGameData()
	p.banker = lgd.banker
	p.curSeat = p.banker
	p.dealer.Initialize()
	p.history = make([]Action, 0)
	for i := range p.game.GetPlayerCount() {
		p.playData[i] = pdfn(int32(i))
	}
}

func (p *Play) GetDealer() *Dealer {
	return p.dealer
}

func (p *Play) GetHuResult(seat int32) *HuResult {
	return p.huResult[seat]
}

func (p *Play) GetCurScores() []int64 {
	count := p.game.GetPlayerCount()
	scores := make([]int64, count)

	for i := range count {
		if player := p.game.GetPlayer(i); player != nil {
			scores[i] = player.GetCurScore()
		}
	}
	return scores
}

func (p *Play) Deal() {
	count := p.game.GetPlayerCount()
	for i := range count {
		p.playData[i].handTiles = p.dealer.Deal(Service.GetHandCount())
	}

	p.playData[p.banker].PutHandTile(p.dealer.DrawTile())
}

func (p *Play) GetPlayData(seat int32) *PlayData {
	return p.playData[seat]
}

func (p *Play) FetchSelfOperates() *Operates {
	opt := &Operates{}

	tips := make([]int, 0)
	for _, v := range p.selfCheckers {
		tips = v.Check(p, opt, tips)
	}

	if len(tips) > 0 {
		p.sendTips(tips[0], p.curSeat)
	}

	return opt
}

func (p *Play) FetchWaitOperates(seat int32) *Operates {
	opt := &Operates{}
	if p.game.GetPlayer(seat).isOut {
		return opt
	}

	tips := make([]int, 0)
	for _, v := range p.waitcheckers {
		tips = v.Check(p, seat, opt, tips)
	}

	if len(tips) > 0 {
		p.sendTips(tips[0], seat)
	}
	return opt
}

func (p *Play) sendTips(tips int, seat int32) {
	//TODO
}

func (p *Play) Discard(tile int32) {
	playData := p.playData[p.curSeat]
	if playData.call {
		tile = playData.handTiles[len(playData.handTiles)-1]
	}

	if tile == TileNull {
		panic("all tingyong in hand, cannot discard")
	}

	playData.Discard(tile)
	p.curTile = tile
	p.addHistory(p.curSeat, p.curTile, OperateDiscard, 0)
}

func (p *Play) ZhiKon(seat int32) {
	playData := p.playData[seat]
	if !playData.canKon(p.curTile, KonTypeZhi) {
		logrus.Error("player cannot zhi kon")
		return
	}
	playData.kon(p.curTile, p.curSeat, KonTypeZhi)
	p.playData[p.curSeat].RemoveOutTile()
	p.addHistory(seat, p.curTile, OperateKon, 0)
}

func (p *Play) TryKon(tile int32, konType KonType) bool {
	playData := p.playData[p.curSeat]
	if !playData.canKon(tile, konType) {
		return false
	}
	playData.kon(tile, p.curSeat, konType)
	p.curTile = tile
	p.addHistory(p.curSeat, p.curTile, OperateKon, 0)
	return true
}

func (p *Play) Pon(seat int32) {
	playData := p.playData[seat]
	if !playData.canPon(p.curTile, p.PlayConf.CanotOnlyLaiAfterPon) {
		logrus.Error("player cannot pon")
		return
	}
	playData.Pon(p.curTile, p.curSeat)
	p.playData[p.curSeat].RemoveOutTile()
	p.addHistory(seat, p.curTile, OperatePon, 0)
}

func (p *Play) Zimo() (multiples []int64) {
	multiples = make([]int64, p.game.GetPlayerCount())
	huResult := p.huResult[p.curSeat]
	multi := p.PlayConf.GetRealMultiple(huResult.TotalMuti)
	for i := int32(0); i < p.game.GetPlayerCount(); i++ {
		if p.game.GetPlayer(i).IsOut() || i == p.curSeat {
			continue
		}
		multiples[i] = -multi
		multiples[p.curSeat] += multi
	}

	p.huSeats = append(p.huSeats, p.curSeat)
	p.addHistory(p.curSeat, p.curTile, OperateHu, 0)
	return
}

func (p *Play) Draw() int32 {
	tile := p.dealer.DrawTile()
	if tile == TileNull {
		p.playData[p.curSeat].PutHandTile(tile)
		p.addHistory(p.curSeat, tile, OperateDraw, 0)
	}
	return tile
}

func (p *Play) IsAfterPon() bool {
	return len(p.history) > 0 && p.history[len(p.history)-1].Operate == OperatePon
}

func (p *Play) IsAfterKon() bool {
	return len(p.history) > 0 && p.history[len(p.history)-1].Operate == OperateKon
}

func (p *Play) DoSwitchSeat(seat int32) {
	if seat == SeatNull {
		p.curSeat = GetNextSeat(p.curSeat, 1, p.game.GetPlayerCount())
	} else {
		p.curSeat = seat
	}
}

func (p *Play) GetCurSeat() int32 {
	return p.curSeat
}

func (p *Play) GetCurTile() int32 {
	return p.curTile
}

func (p *Play) GetBanker() int32 {
	return p.banker
}

func (p *Play) HasOperate(seat int32) bool {
	for _, action := range p.history {
		if action.Seat == seat {
			return true
		}
	}
	return false
}

func (p *Play) getLastGameData() *LastGameData {
	lastGameData := p.game.GetLastGameData()
	if lastGameData == nil {
		return NewLastGameData(int(p.game.GetPlayerCount()))
	}

	lgd, ok := lastGameData.(*LastGameData)
	if !ok {
		logrus.Errorf("invalid last game data type: %T", lastGameData)
		return NewLastGameData(int(p.game.GetPlayerCount()))
	}
	return lgd
}

func (p *Play) addHistory(seat int32, tile int32, operate int, extra int) {
	action := Action{
		Seat:    seat,
		Tile:    tile,
		Operate: operate,
		Extra:   extra,
	}
	p.history = append(p.history, action)
}

func (p *Play) addHuOperate(opt *Operates, seat int32, result *HuResult, mustHu bool) {
	opt.Capped = p.PlayConf.IsTopMultiple(result.TotalMuti)
	p.huResult[seat] = result
	opt.AddOperate(OperateHu)
	opt.IsMustHu = mustHu
}

func (p *Play) isKonAfterPon(tile int32) bool {
	if len(p.history) <= 0 {
		return false
	}
	action := p.history[len(p.history)-1]
	return action.Operate == OperatePon && action.Tile == tile
}

func (p *Play) checkMustHu(seat int32) bool {
	if p.PlayConf.MustHu {
		return true
	}

	if !p.PlayConf.MustHuIfOnlyLai {
		return false
	}
	playData := p.playData[seat]

	if playData.isAllLai() || playData.call && slices.Contains(p.tilesLai, playData.handTiles[len(playData.handTiles)-1]) {
		return true
	}
	return false
}
