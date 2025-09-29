package mahjong

import (
	"slices"

	"github.com/topfreegames/pitaya/v3/pkg/logger"
)

type Group struct {
	Tile  Tile
	From  int32
	Extra int32
}

type KonGroup struct {
	Tile          Tile
	From          int32
	Type          KonType
	HandPassBuKon bool
	Extra         int32
}

type ChowGroup struct {
	ChowTile Tile
	From     int32
	LeftTile Tile
}

type PlayData struct {
	play            *Play
	callDataMap     map[int]map[int]int
	call            bool
	tianTing        bool
	handTiles       []Tile
	outTiles        []Tile
	canGangTiles    []Tile
	tianDiHu        bool
	passPon         map[Tile]struct{}
	passHu          map[Tile]int32
	qiHuFanLimitTip bool
	chowGroups      []ChowGroup
	ponGroups       []Group
	konGroups       []KonGroup
	everPonCount    int
	everKonCount    int
	everChiCount    int
	minTingValue    int
	drawConfig      int
	drawRate        int
}

func NewPlayData(seat int32) *PlayData {
	return &PlayData{
		callDataMap:  make(map[int]map[int]int),
		handTiles:    make([]Tile, 0),
		outTiles:     make([]Tile, 0),
		canGangTiles: make([]Tile, 0),
		passPon:      make(map[Tile]struct{}),
		passHu:       make(map[Tile]int32),
		chowGroups:   make([]ChowGroup, 0),
		ponGroups:    make([]Group, 0),
		konGroups:    make([]KonGroup, 0),
		minTingValue: 17,
	}
}

func (p *PlayData) MutableCallDataMap() map[int]map[int]int {
	return p.callDataMap
}

func (p *PlayData) GetCallDataMap() map[int]map[int]int {
	return p.callDataMap
}

func (p *PlayData) Discard(tile Tile) bool {
	if !slices.Contains(p.handTiles, tile) {
		return false
	}
	p.handTiles = RemoveElements(p.handTiles, tile, 1)
	logger.Log.Info(p.handTiles)
	p.PutOutTile(tile)
	return true
}

func (p *PlayData) SetCall(tile int, tianTing bool) {
	p.call = true
	p.tianTing = tianTing
}

func (p *PlayData) PutHandTile(tile Tile) {
	p.handTiles = append(p.handTiles, tile)
	logger.Log.Info(p.handTiles)
}

func (p *PlayData) RemoveHandTile(tile Tile, count int) {
	p.handTiles = RemoveElements(p.handTiles, tile, count)
}

func (p *PlayData) PutOutTile(tile Tile) {
	p.outTiles = append(p.outTiles, tile)
}

func (p *PlayData) RemoveOutTile() {
	if len(p.outTiles) > 0 {
		p.outTiles = p.outTiles[:len(p.outTiles)-1]
	}
}

func (p *PlayData) canKon(tile Tile, konType KonType) bool {
	count := CountElement(p.handTiles, tile)
	switch konType {
	case KonTypeZhi:
		return count == 3
	case KonTypeAn:
		return count == 4
	case KonTypeBu:
		return count == 1 && p.HasPon(tile)
	default:
		return false
	}
}

func (p *PlayData) canPon(tile Tile, cantOnlyLaiAfterPon bool) bool {
	if CountElement(p.handTiles, tile) < 2 {
		return false
	}
	if cantOnlyLaiAfterPon && p.play.isAllLai(RemoveElements(p.handTiles, tile, 2)) {
		return false
	}
	return true
}

func (p *PlayData) canChow(tile Tile) bool {
	color, point := tile.Info()
	points := make([]int, PointCountByColor[color])

	for _, t := range p.handTiles {
		if t.Color() == color {
			points[t.Point()]++
		}
	}
	points[point]++
	leftPoint := max((point - 2), 0)
	maxLeftPoint := min(6, point)
	for i := leftPoint; i <= maxLeftPoint; i++ {
		if points[i] != 0 && points[i+1] != 0 && points[i+2] != 0 {
			return true
		}
	}
	return false
}

func (p *PlayData) GetHandTiles() []Tile {
	return p.handTiles
}

func (p *PlayData) GetHandTilesInt32() []int32 {
	return TilesInt32(p.handTiles)
}

func (p *PlayData) GetOutTiles() []Tile {
	return p.outTiles
}

func (p *PlayData) CloseTianDiHu() {
	p.tianDiHu = false
}

func (p *PlayData) TianDiHuState() bool {
	return p.tianDiHu
}

func (p *PlayData) IsPassHuTileWithFan(tile Tile, fan int32) bool {
	if f, ok := p.passHu[tile]; ok {
		return f == fan
	}
	return false
}

func (p *PlayData) IsPassHuTile(tile Tile) bool {
	_, ok := p.passHu[tile]
	return ok
}

func (p *PlayData) IsPassPonTile(tile Tile) bool {
	_, ok := p.passPon[tile]
	return ok
}

func (p *PlayData) ClearPass() {
	p.passPon = make(map[Tile]struct{})
	p.passHu = make(map[Tile]int32)
}

func (p *PlayData) PassPon(tile Tile) {
	p.passPon[tile] = struct{}{}
}

func (p *PlayData) PassHu(tile Tile, fan int32) {
	p.passHu[tile] = fan
}

func (p *PlayData) SetBanQiHuFanTip(flag bool) {
	p.qiHuFanLimitTip = flag
}

func (p *PlayData) IsBanQiHuFanTip() bool {
	return p.qiHuFanLimitTip
}

func (p *PlayData) TryChow(curTile, tile Tile, from int32) bool {
	color, point := tile.Info()
	if color != curTile.Color() || curTile.Point()-point >= 3 {
		return false
	}

	tiles := make([]Tile, 0)
	for i := range 3 {
		t := MakeTile(color, point+i)
		if t == curTile {
			continue
		}
		if !slices.Contains(p.handTiles, t) {
			return false
		}
		tiles = append(tiles, t)
	}

	for _, t := range tiles {
		p.RemoveHandTile(t, 1)
	}

	group := ChowGroup{
		ChowTile: curTile,
		From:     from,
		LeftTile: tile,
	}
	p.chowGroups = append(p.chowGroups, group)
	return true
}

func (p *PlayData) GetChowGroups() []ChowGroup {
	return p.chowGroups
}

func (p *PlayData) Pon(tile Tile, from int32) Tile {
	p.RemoveHandTile(tile, 2)
	group := Group{
		Tile: tile,
		From: from,
	}
	p.ponGroups = append(p.ponGroups, group)
	return tile
}

func (p *PlayData) HasPon(tile Tile) bool {
	for _, group := range p.ponGroups {
		if group.Tile == tile {
			return true
		}
	}
	return false
}

func (p *PlayData) kon(tile Tile, from int32, konType KonType) {
	if konType == KonTypeBu {
		p.buKon(tile, false, false)
	} else {
		p.anZhiKon(tile, from, konType)
	}
}

func (p *PlayData) buKon(tile Tile, buKonAfterPon, handPassBuKon bool) {
	p.RemoveHandTile(tile, 1)
	from := p.RemovePon(tile).From
	if buKonAfterPon {
		p.konGroups = append(p.konGroups, KonGroup{Tile: tile, From: from, Type: KonTypeZhi})
	} else {
		p.konGroups = append(p.konGroups, KonGroup{Tile: tile, From: from, Type: KonTypeBu, HandPassBuKon: handPassBuKon})
	}
}

func (p *PlayData) anZhiKon(tile Tile, from int32, konType KonType) {
	if konType == KonTypeAn {
		p.RemoveHandTile(tile, 4)
	} else {
		p.RemoveHandTile(tile, 3)
	}
	p.konGroups = append(p.konGroups, KonGroup{Tile: tile, From: from, Type: konType})
}

func (p *PlayData) HasKon(tile Tile) bool {
	for _, group := range p.konGroups {
		if group.Tile == tile {
			return true
		}
	}
	return false
}

func (p *PlayData) PushPon(group Group) {
	p.ponGroups = append(p.ponGroups, group)
}

func (p *PlayData) PushKon(group KonGroup) {
	p.konGroups = append(p.konGroups, group)
}

func (p *PlayData) GetKon(tile Tile) *KonGroup {
	for _, group := range p.konGroups {
		if group.Tile == tile {
			return &group
		}
	}
	return nil
}

func (p *PlayData) GetPon(tile Tile) *Group {
	for _, group := range p.ponGroups {
		if group.Tile == tile {
			return &group
		}
	}
	return nil
}

func (p *PlayData) RemovePon(tile Tile) Group {
	for i, group := range p.ponGroups {
		if group.Tile == tile {
			p.ponGroups = append(p.ponGroups[:i], p.ponGroups[i+1:]...)
			return group
		}
	}
	return Group{}
}

func (p *PlayData) RemoveKon(tile Tile) KonGroup {
	for i, group := range p.konGroups {
		if group.Tile == tile {
			p.konGroups = append(p.konGroups[:i], p.konGroups[i+1:]...)
			return group
		}
	}
	return KonGroup{}
}

func (p *PlayData) RevertKon(tile int) {
	// 实现杠牌回退逻辑
}

func (p *PlayData) GetPonGroups() []Group {
	return p.ponGroups
}

func (p *PlayData) GetKonGroups() []KonGroup {
	return p.konGroups
}

func (p *PlayData) GetExchangeRecommend() []int {
	// 实现交换推荐逻辑
	return nil
}

func (p *PlayData) CanExchangeOut(tiles []int) bool {
	// 实现能否交换出牌逻辑
	return false
}

func (p *PlayData) ExchangeOut(outs []int) {
	// 实现交换出牌逻辑
}

func (p *PlayData) ExchangeIn(ines []int) {
	// 实现交换进牌逻辑
}

func (p *PlayData) Exchange(outs, ines []int) {
	p.ExchangeOut(outs)
	p.ExchangeIn(ines)
}

func (p *PlayData) IncEverPonCount() {
	p.everPonCount++
}

func (p *PlayData) IncEverKonCount() {
	p.everKonCount++
}

func (p *PlayData) IncEverChiCount() {
	p.everChiCount++
}

func (p *PlayData) GetEverPonCount() int {
	return p.everPonCount
}

func (p *PlayData) GetEverKonCount() int {
	return p.everKonCount
}

func (p *PlayData) GetEverChiCount() int {
	return p.everChiCount
}

func (p *PlayData) GetMinTing() int {
	return p.minTingValue
}

func (p *PlayData) SetDrawConfig(drawConfig, drawRate int) {
	p.drawConfig = drawConfig
	p.drawRate = drawRate
}

func (p *PlayData) GetDrawConfig() int {
	return p.drawConfig
}

func (p *PlayData) GetDrawRate() int {
	return p.drawRate
}

func (p *PlayData) tilesForChowLeft() []Tile {
	tiles := make([]Tile, len(p.chowGroups))
	for i, group := range p.chowGroups {
		tiles[i] = group.LeftTile
	}
	return tiles
}

func (p *PlayData) tilesForPon() []Tile {
	tiles := make([]Tile, len(p.ponGroups))
	for i, group := range p.ponGroups {
		tiles[i] = group.Tile
	}
	return tiles
}

func (p *PlayData) tilesForKon() (tiles []Tile, countAnKon int32) {
	tiles = make([]Tile, len(p.konGroups))
	for i, group := range p.konGroups {
		tiles[i] = group.Tile
		if group.Type == KonTypeAn {
			countAnKon++
		}
	}
	return
}

// CanSelfKon 判断是否可以自杠
func (p *PlayData) canSelfKon(rule *Rule, ignoreTiles []Tile) bool {
	p.canGangTiles = make([]Tile, 0)
	counts := make(map[Tile]int)
	for _, tile := range p.handTiles {
		if !slices.Contains(ignoreTiles, tile) {
			counts[tile]++
		}
	}

	if !p.call {
		for _, pon := range p.ponGroups {
			if slices.Contains(p.handTiles, pon.Tile) {
				p.canGangTiles = append(p.canGangTiles, pon.Tile)
			}
		}
		for tile, count := range counts {
			if count == 4 {
				p.canGangTiles = append(p.canGangTiles, tile)
			}
		}
		return len(p.canGangTiles) > 0
	}

	// 新开杠判断
	lastTile := p.handTiles[len(p.handTiles)-1]
	for _, pon := range p.ponGroups {
		if pon.Tile == lastTile {
			p.canGangTiles = append(p.canGangTiles, pon.Tile)
			return true
		}
	}

	if counts[lastTile] == 4 && p.canKonAfterCall(lastTile, KonTypeAn, rule) {
		p.canGangTiles = append(p.canGangTiles, lastTile)
		return true
	}
	return false
}

func (p *PlayData) canKonAfterCall(tile Tile, konType KonType, rule *Rule) bool {
	if KonTypeZhi != konType && tile != p.handTiles[len(p.handTiles)-1] {
		return false
	}

	hudata := NewCheckHuData(p.play, p, false)
	if KonTypeZhi != konType {
		hudata.TilesInHand = hudata.TilesInHand[:len(hudata.TilesInHand)-1]
	}
	call0 := Service.CheckCall(hudata, rule)
	hudata.TilesInHand = RemoveAllElement(hudata.TilesInHand, tile)
	call1 := Service.CheckCall(hudata, rule)
	if len(call0) != 1 || len(call1) != 1 {
		return false
	}
	return HasSameKeys(call0[TileNull], call1[TileNull])
}
