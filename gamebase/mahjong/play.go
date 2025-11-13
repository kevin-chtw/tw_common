package mahjong

import (
	"github.com/kevin-chtw/tw_proto/game/pbmj"
	"github.com/sirupsen/logrus"
)

type IPlay interface {
	CheckHu(data *HuData) HuCoreType
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
	huResult     []*pbmj.MJHuData
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
		huResult:     make([]*pbmj.MJHuData, game.GetPlayerCount()),
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

func (p *Play) GetPlayerCount() int32 {
	return p.game.GetPlayerCount()
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
}

func (p *Play) GetPlayData(seat int32) *PlayData {
	return p.playData[seat]
}

func (p *Play) FetchSelfOperates(sender *Sender) *Operates {
	opt := NewOperates(OperateDiscard)

	for _, v := range p.selfCheckers {
		v.Check(opt)
	}

	if len(opt.Tips) > 0 {
		sender.SendTipAck(p.curSeat, opt.Tips[0])
	}

	return opt
}

func (p *Play) FetchWaitOperates(seat int32, sender *Sender) *Operates {
	opt := NewOperates(OperatePass)
	if p.game.GetPlayer(seat).isOut {
		return opt
	}

	for _, v := range p.waitcheckers {
		v.Check(seat, opt)
	}

	if len(opt.Tips) > 0 {
		sender.SendTipAck(seat, opt.Tips[0])
	}

	return opt
}

func (p *Play) FetchAfterBuKonOperates(seat int32, checker CheckerWait, sender *Sender) *Operates {
	opt := NewOperates(OperatePass)
	if p.game.GetPlayer(seat).isOut {
		return opt
	}

	checker.Check(seat, opt)
	if len(opt.Tips) > 0 {
		sender.SendTipAck(seat, opt.Tips[0])
	}
	return opt
}

func (p *Play) Ting(tile Tile) bool {
	playData := p.playData[p.curSeat]
	if !playData.canTing(tile) {
		return false
	}
	if playData.Discard(tile) {
		playData.SetTing(false)
		p.addHistory(p.curSeat, p.curSeat, OperateTing, p.curTile, 0)
		p.freshTingMuti(p.curSeat)
		p.curTile = tile
		p.game.GetGamePlayer(p.curSeat).AddData("ting", 1)
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
		p.addHistory(p.curSeat, p.curSeat, OperateDiscard, p.curTile, 0)
		p.FreshCallData(p.curSeat)
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
	p.addHistory(seat, p.curSeat, OperateKon, p.curTile, 0)
	p.game.GetGamePlayer(seat).AddData("kon", 1)
	p.FreshCallData(seat)
}

func (p *Play) TryKon(tile Tile, konType KonType) bool {
	playData := p.playData[p.curSeat]
	if !playData.canKon(tile, konType) {
		return false
	}

	playData.kon(tile, p.curSeat, konType)
	p.addHistory(p.curSeat, p.curSeat, OperateKon, tile, 0)
	p.game.GetGamePlayer(p.curSeat).AddData("kon", 1)
	p.FreshCallData(p.curSeat)
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
	p.addHistory(seat, p.curSeat, OperatePon, p.curTile, 0)
	p.game.GetGamePlayer(seat).AddData("pon", 1)
	p.FreshCallData(seat)
}

func (p *Play) Chow(seat int32, leftTile Tile) {
	playData := p.playData[seat]
	tiles, ok := playData.tryChow(p.curTile, leftTile)
	if !ok {
		logrus.Error("player cannot chow")
	}

	playData.chow(tiles, p.curTile, leftTile, seat)
	p.playData[p.curSeat].RemoveOutTile()
	p.addHistory(seat, p.curSeat, OperateChow, p.curTile, leftTile)
	p.game.GetGamePlayer(seat).AddData("chow", 1)
	p.FreshCallData(seat)
}

func (p *Play) Zimo() (multiples []int64) {
	multiples = make([]int64, p.game.GetPlayerCount())
	huResult := p.huResult[p.curSeat]
	multi := p.PlayConf.GetRealMultiple(huResult.Multi)
	for i := int32(0); i < p.game.GetPlayerCount(); i++ {
		if p.game.GetPlayer(i).IsOut() || i == p.curSeat {
			continue
		}
		multiples[i] = -multi
		multiples[p.curSeat] += multi
	}

	p.addHistory(p.curSeat, p.curSeat, OperateHu, p.curTile, 0)
	p.game.GetGamePlayer(p.curSeat).AddData("hu", 1)
	return
}

func (p *Play) PaoHu(huSeats []int32) []int64 {
	p.playData[p.curSeat].RemoveOutTile()
	multiples := make([]int64, p.game.GetPlayerCount())
	for _, seat := range huSeats {
		huResult := p.huResult[seat]
		multi := p.PlayConf.GetRealMultiple(huResult.Multi)
		multiples[p.curSeat] -= multi
		if !p.game.GetPlayer(seat).IsOut() {
			multiples[seat] = +multi
			p.addHistory(seat, p.curSeat, OperateHu, p.curTile, 0)
			p.game.GetGamePlayer(seat).AddData("hu", 1)
		}
	}
	p.game.GetGamePlayer(p.curSeat).AddData("dianpao", 1)
	return multiples
}

func (p *Play) DianKonHua(paoSeat int32) []int64 {
	multiples := make([]int64, p.game.GetPlayerCount())
	huResult := p.huResult[p.curSeat]
	multi := p.PlayConf.GetRealMultiple(huResult.Multi)
	multiples[p.curSeat] += multi
	multiples[paoSeat] = -multi
	p.addHistory(p.curSeat, paoSeat, OperateHu, p.curTile, 0)
	p.game.GetGamePlayer(p.curSeat).AddData("hu", 1)
	p.game.GetGamePlayer(paoSeat).AddData("diankh", 1)
	return multiples
}

func (p *Play) Draw() Tile {
	tile := p.dealer.DrawTile()
	if tile != TileNull {
		p.curTile = tile
		p.playData[p.curSeat].PutHandTile(tile)
		p.addHistory(p.curSeat, p.curSeat, OperateDraw, tile, 0)
		p.FreshCallData(p.curSeat)
	}
	return tile
}

func (p *Play) IsAfterPon() bool {
	return len(p.history) > 0 && p.history[len(p.history)-1].Operate == OperatePon
}

func (p *Play) IsAfterKon() bool {
	return len(p.history) > 0 && p.history[len(p.history)-1].Operate == OperateKon
}

func (p *Play) IsAfterZhiKon() int32 {
	if len(p.history) == 0 {
		return SeatNull
	}

	lastAction := p.history[len(p.history)-1]
	if lastAction.Operate != OperateKon || lastAction.Seat == lastAction.From {
		return SeatNull
	}
	return lastAction.From
}

func (p *Play) DoSwitchSeat(seat int32) {
	if seat == SeatNull {
		seat = GetNextSeat(p.curSeat, 1, p.game.GetPlayerCount())
	}

	startSeat := seat
	for {
		if !p.game.GetPlayer(seat).IsOut() {
			p.curSeat = seat
			return
		}
		seat = GetNextSeat(seat, 1, p.game.GetPlayerCount())
		if seat == startSeat {
			break
		}
	}
}

func (p *Play) GetCurSeat() int32 {
	return p.curSeat
}

func (p *Play) GetBanker() int32 {
	return p.banker
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

func (p *Play) AddHuOperate(opt *Operates, seat int32, result *pbmj.MJHuData, mustHu bool) {
	opt.Capped = p.PlayConf.IsTopMultiple(result.Multi)
	p.huResult[seat] = result
	opt.AddOperate(OperateHu)
	opt.IsMustHu = mustHu
	opt.HuMulti = result.Multi
}

func (p *Play) getLastGameData() *LastGameData {
	lastGameData := p.game.GetLastGameData()
	if lastGameData == nil {
		data := NewLastGameData(int(p.game.GetPlayerCount()))
		p.game.SetLastGameData(data)
		return data
	}

	lgd, ok := lastGameData.(*LastGameData)
	if !ok {
		logrus.Errorf("invalid last game data type: %T", lastGameData)
		return nil
	}
	return lgd
}

func (p *Play) addHistory(seat, from int32, operate int, tile Tile, extra Tile) {
	action := Action{
		Seat:    seat,
		From:    from,
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

func (p *Play) FreshCallData(seat int32) {
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
