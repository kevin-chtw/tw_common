package mahjong

import (
	"github.com/sirupsen/logrus"
)

type IPlay interface {
	CheckHu(data *HuData) bool
	GetExtraHuTypes(data *PlayData, self bool) []int32
}

type Play struct {
	PlayImp      IPlay
	PlayConf     *PlayConf
	game         *Game
	curSeat      int32
	curTile      Tile
	banker       int32
	dealer       *Dealer
	tilesLai     map[Tile]struct{}
	history      []Action
	playData     []*PlayData
	huSeats      []int32
	huResult     []*HuResult
	selfCheckers []CheckerSelf
	waitcheckers []CheckerWait
}

func NewPlay(playImp IPlay, game *Game, dealer *Dealer) *Play {
	return &Play{
		PlayImp:      playImp,
		game:         game,
		dealer:       dealer,
		curSeat:      SeatNull,
		curTile:      TileNull,
		banker:       SeatNull,
		tilesLai:     make(map[Tile]struct{}),
		history:      make([]Action, 0),
		playData:     make([]*PlayData, game.GetPlayerCount()),
		huSeats:      make([]int32, 0),
		huResult:     make([]*HuResult, game.GetPlayerCount()),
		selfCheckers: make([]CheckerSelf, 0),
		waitcheckers: make([]CheckerWait, 0),
	}
}

func (p *Play) RegisterSelfCheck(cks ...CheckerSelf) {
	p.selfCheckers = append(p.selfCheckers, cks...)
}
func (p *Play) RegisterWaitCheck(cks ...CheckerWait) {
	p.waitcheckers = append(p.waitcheckers, cks...)
}

func (p *Play) Initialize(pdfn func(*Play, int32) *PlayData) {
	lgd := p.getLastGameData()
	p.banker = lgd.banker
	p.curSeat = p.banker
	p.dealer.Initialize()
	p.history = make([]Action, 0)
	for i := range p.game.GetPlayerCount() {
		p.playData[i] = pdfn(p, int32(i))
	}
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
	for i := range p.game.GetPlayerCount() {
		p.playData[i].handTiles = p.dealer.Deal(Service.GetHandCount())
	}
	p.playData[p.banker].PutHandTile(p.dealer.DrawTile())
	for i := range p.game.GetPlayerCount() {
		p.freshCallData(i)
	}
}

func (p *Play) GetPlayData(seat int32) *PlayData {
	return p.playData[seat]
}

func (p *Play) FetchSelfOperates() *Operates {
	opt := &Operates{Value: OperateDiscard}

	tips := make([]int, 0)
	for _, v := range p.selfCheckers {
		tips = v.Check(opt, tips)
	}

	if len(tips) > 0 {
		p.sendTips(tips[0], p.curSeat)
	}

	return opt
}

func (p *Play) FetchWaitOperates(seat int32) *Operates {
	opt := &Operates{Value: OperatePass}
	if p.game.GetPlayer(seat).isOut {
		return opt
	}

	tips := make([]int, 0)
	for _, v := range p.waitcheckers {
		tips = v.Check(seat, opt, tips)
	}

	if len(tips) > 0 {
		p.sendTips(tips[0], seat)
	}
	return opt
}

func (p *Play) FetchAfterBuKonOperates(seat int32, checker CheckerWait) *Operates {
	opt := &Operates{Value: OperatePass}
	if p.game.GetPlayer(seat).isOut {
		return opt
	}

	tips := make([]int, 0)
	tips = checker.Check(seat, opt, tips)

	if len(tips) > 0 {
		p.sendTips(tips[0], seat)
	}
	return opt
}

func (p *Play) sendTips(tips int, seat int32) {
	//TODO
}

func (p *Play) Ting(tile Tile) bool {
	playData := p.playData[p.curSeat]
	if !playData.canTing(tile) {
		return false
	}
	if playData.Discard(tile) {
		playData.ting = true
		p.addHistory(p.curSeat, OperateTing, p.curTile, 0)
		p.freshTingMuti(p.curSeat)
		p.curTile = tile
		return true
	}
	return false
}

func (p *Play) Discard(tile Tile) bool {
	playData := p.playData[p.curSeat]
	if playData.ting {
		tile = playData.handTiles[len(playData.handTiles)-1]
	}

	if tile == TileNull {
		tile = playData.handTiles[len(playData.handTiles)-1]
	}

	if playData.Discard(tile) {
		p.addHistory(p.curSeat, OperateDiscard, p.curTile, 0)
		p.freshCallData(p.curSeat)
		p.curTile = tile
		return true
	}
	return false
}

func (p *Play) ZhiKon(seat int32) {
	playData := p.playData[seat]
	if !playData.canKon(p.curTile, KonTypeZhi) {
		logrus.Error("player cannot zhi kon")
		return
	}
	playData.kon(p.curTile, p.curSeat, KonTypeZhi)
	p.playData[p.curSeat].RemoveOutTile()
	p.addHistory(seat, OperateKon, p.curTile, 0)
	p.freshCallData(seat)
}

func (p *Play) TryKon(tile Tile, konType KonType) bool {
	playData := p.playData[p.curSeat]
	if !playData.canKon(tile, konType) {
		return false
	}

	playData.kon(tile, p.curSeat, konType)
	p.addHistory(p.curSeat, OperateKon, p.curTile, 0)
	p.freshCallData(p.curSeat)
	p.curTile = tile
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
	p.addHistory(seat, OperatePon, p.curTile, 0)
	p.freshCallData(seat)
}

func (p *Play) PonTing(seat int32, disTile Tile) {
	playData := p.playData[seat]
	if !playData.canPon(p.curTile, p.PlayConf.CanotOnlyLaiAfterPon) {
		logrus.Error("player cannot pon")
		return
	}

	huData := NewHuData(playData, false)
	huData.Tiles = RemoveElements(huData.Tiles, p.curTile, 2)
	callData := huData.CheckCall()
	if _, ok := callData[disTile]; !ok {
		logrus.Error("player cannot ting")
		return
	}

	if playData.Discard(disTile) {
		playData.Pon(p.curTile, p.curSeat)
		p.addHistory(seat, OperatePonTing, p.curTile, disTile)
		p.freshTingMuti(seat)
		p.curTile = disTile
	}
}

func (p *Play) Chow(seat int32, leftTile Tile) {
	playData := p.playData[seat]
	tiles, ok := playData.tryChow(p.curTile, leftTile)
	if !ok {
		logrus.Error("player cannot chow")
	}

	playData.chow(tiles, p.curTile, leftTile, seat)
	p.playData[p.curSeat].RemoveOutTile()
	p.addHistory(seat, OperateChow, p.curTile, leftTile)
	p.freshCallData(seat)
}

func (p *Play) ChowTing(seat int32, leftTile, disTile Tile) {
	playData := p.playData[seat]
	tiles, ok := playData.tryChow(p.curTile, leftTile)
	if !ok {
		logrus.Error("player cannot chow")
	}

	huData := NewHuData(playData, false)
	for _, tile := range tiles {
		huData.Tiles = RemoveElements(huData.Tiles, tile, 1)
	}
	callData := huData.CheckCall()
	if _, ok := callData[disTile]; !ok {
		logrus.Error("player cannot ting")
		return
	}

	if playData.Discard(disTile) {
		playData.chow(tiles, p.curTile, leftTile, seat)
		p.addHistory(seat, OperatePonTing, p.curTile, disTile)
		p.freshTingMuti(seat)
		p.curTile = disTile
	}
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
	p.addHistory(p.curSeat, OperateHu, p.curTile, 0)
	return
}

func (p *Play) PaoHu(huSeats []int32) []int64 {
	p.playData[p.curSeat].RemoveOutTile()
	multiples := make([]int64, p.game.GetPlayerCount())
	for _, seat := range huSeats {
		huResult := p.huResult[seat]
		multi := p.PlayConf.GetRealMultiple(huResult.TotalMuti)
		multiples[p.curSeat] -= multi
		if !p.game.GetPlayer(seat).IsOut() {
			multiples[seat] = +multi
			p.addHistory(p.curSeat, OperateHu, p.curTile, 0)
		}
	}
	p.huSeats = append(p.huSeats, huSeats...)
	return multiples
}

func (p *Play) Draw() Tile {
	tile := p.dealer.DrawTile()
	if tile != TileNull {
		p.playData[p.curSeat].PutHandTile(tile)
		p.addHistory(p.curSeat, OperateDraw, tile, 0)
		p.freshCallData(p.curSeat)
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

func (p *Play) GetCurTile() Tile {
	return p.curTile
}

func (p *Play) GetRule() *Rule {
	return p.game.rule
}

func (p *Play) HasOperate(seat int32) bool {
	for _, action := range p.history {
		if action.Seat == seat {
			return true
		}
	}
	return false
}

func (p *Play) AddHuOperate(opt *Operates, seat int32, result *HuResult, mustHu bool) {
	opt.Capped = p.PlayConf.IsTopMultiple(result.TotalMuti)
	p.huResult[seat] = result
	opt.AddOperate(OperateHu)
	opt.IsMustHu = mustHu
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

func (p *Play) addHistory(seat int32, operate int, tile Tile, extra Tile) {
	action := Action{
		Seat:    seat,
		Operate: operate,
		Tile:    tile,
		Extra:   extra,
	}
	p.history = append(p.history, action)
}

func (p *Play) isKonAfterPon(tile Tile) bool {
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
	return p.isAllLai(playData.handTiles)
}

func (p *Play) isAllLai(tiles []Tile) bool {
	for _, tile := range tiles {
		if _, ok := p.tilesLai[tile]; ok {
			return false
		}
	}
	return true
}

func (p *Play) freshCallData(seat int32) {
	playData := p.GetPlayData(seat)
	if !playData.IsTing() {
		data := NewHuData(playData, false)
		playData.SetCallMap(data.CheckCall())
	}
}

func (p *Play) freshTingMuti(seat int32) {
	playData := p.GetPlayData(seat)
	data := NewHuData(playData, false)
	playData.SetCallData(data.checkCalls())
}
