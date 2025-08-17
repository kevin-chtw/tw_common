package mahjong

type Group struct {
	Tile  int32
	From  int32
	Extra int32
}

type KonGroup struct {
	Tile          int32
	From          int32
	Type          KonType
	HandPassBuKon bool
	Extra         int32
}

type ChowGroup struct {
	ChowTile int32
	From     int32
	LeftTile int32
}

type PlayData struct {
	play             *Play
	callDataMap      map[int]map[int]int
	currentDrawTiles map[int]struct{}
	call             bool
	tianTing         bool
	handTiles        []int32
	outTiles         []int32
	canGangTiles     []int32
	tianDiHu         bool
	passPon          map[int]struct{}
	passHu           map[int]int
	qiHuFanLimitTip  bool
	chowGroups       []ChowGroup
	ponGroups        []Group
	konGroups        []KonGroup
	everPonCount     int
	everKonCount     int
	everChiCount     int
	minTingValue     int
	drawConfig       int
	drawRate         int
}

func NewPlayData(seat int) *PlayData {
	return &PlayData{
		callDataMap:      make(map[int]map[int]int),
		currentDrawTiles: make(map[int]struct{}),
		handTiles:        make([]int32, 0),
		outTiles:         make([]int32, 0),
		canGangTiles:     make([]int32, 0),
		passPon:          make(map[int]struct{}),
		passHu:           make(map[int]int),
		chowGroups:       make([]ChowGroup, 0),
		ponGroups:        make([]Group, 0),
		konGroups:        make([]KonGroup, 0),
		minTingValue:     17,
	}
}

func (p *PlayData) MutableCallDataMap() map[int]map[int]int {
	return p.callDataMap
}

func (p *PlayData) GetCallDataMap() map[int]map[int]int {
	return p.callDataMap
}

func (p *PlayData) Draw(tile int) {
	p.currentDrawTiles[tile] = struct{}{}
}

func (p *PlayData) Discard(tile int32) {
	RemoveElement(&p.handTiles, tile)
	p.PutOutTile(tile)
}

func (p *PlayData) SetCall(tile int, tianTing bool) {
	p.call = true
	p.tianTing = tianTing
}

func (p *PlayData) PutHandTile(tile int32) {
	p.handTiles = append(p.handTiles, tile)
}

func (p *PlayData) RemoveHandTile(tile int32, count int) {
	// 实现移除多张手牌逻辑
}

func (p *PlayData) PutOutTile(tile int32) {
	p.outTiles = append(p.outTiles, tile)
}

func (p *PlayData) RemoveOutTile() {
	if len(p.outTiles) > 0 {
		p.outTiles = p.outTiles[:len(p.outTiles)-1]
	}
}

func (p *PlayData) HasTile(tile int32) bool {
	for _, t := range p.handTiles {
		if t == tile {
			return true
		}
	}
	return false
}

func (p *PlayData) canKon(tile int32, konType KonType) bool {
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

func (p *PlayData) canPon(tile int32) bool {
	return CountElement(p.handTiles, tile) >= 2
}

func (p *PlayData) GetHandTiles() []int32 {
	return p.handTiles
}

func (p *PlayData) GetOutTiles() []int32 {
	return p.outTiles
}

func (p *PlayData) CloseTianDiHu() {
	p.tianDiHu = false
}

func (p *PlayData) TianDiHuState() bool {
	return p.tianDiHu
}

func (p *PlayData) IsPassHuTile(tile, fan int) bool {
	if f, ok := p.passHu[tile]; ok {
		return f == fan
	}
	return false
}

func (p *PlayData) IsPassHuTileWithoutFan(tile int) bool {
	_, ok := p.passHu[tile]
	return ok
}

func (p *PlayData) IsPassPonTile(tile int) bool {
	_, ok := p.passPon[tile]
	return ok
}

func (p *PlayData) ClearPass() {
	p.passPon = make(map[int]struct{})
	p.passHu = make(map[int]int)
}

func (p *PlayData) PassPon(tile int) {
	p.passPon[tile] = struct{}{}
}

func (p *PlayData) PassHu(tile, fan int) {
	p.passHu[tile] = fan
}

func (p *PlayData) SetBanQiHuFanTip(flag bool) {
	p.qiHuFanLimitTip = flag
}

func (p *PlayData) IsBanQiHuFanTip() bool {
	return p.qiHuFanLimitTip
}

func (p *PlayData) Chow(curTile, tile, from int32) int32 {
	group := ChowGroup{
		ChowTile: curTile,
		From:     from,
		LeftTile: tile,
	}
	p.chowGroups = append(p.chowGroups, group)
	return curTile
}

func (p *PlayData) GetChowGroups() []ChowGroup {
	return p.chowGroups
}

func (p *PlayData) Pon(tile, from int32) int32 {
	group := Group{
		Tile: tile,
		From: from,
	}
	p.ponGroups = append(p.ponGroups, group)
	return tile
}

func (p *PlayData) HasPon(tile int32) bool {
	for _, group := range p.ponGroups {
		if group.Tile == tile {
			return true
		}
	}
	return false
}

func (p *PlayData) kon(tile, from int32, konType KonType) {
	if konType == KonTypeBu {
		p.buKon(tile, false, false)
	} else {
		p.anZhiKon(tile, from, konType)
	}
}

func (p *PlayData) buKon(tile int32, buKonAfterPon, handPassBuKon bool) {
	p.RemoveHandTile(tile, 1)
	from := p.RemovePon(tile).From
	if buKonAfterPon {
		p.konGroups = append(p.konGroups, KonGroup{Tile: tile, From: from, Type: KonTypeZhi})
	} else {
		p.konGroups = append(p.konGroups, KonGroup{Tile: tile, From: from, Type: KonTypeBu, HandPassBuKon: handPassBuKon})
	}
}

func (p *PlayData) anZhiKon(tile, from int32, konType KonType) {
	if konType == KonTypeAn {
		p.RemoveHandTile(tile, 4)
	} else {
		p.RemoveHandTile(tile, 3)
	}
	p.konGroups = append(p.konGroups, KonGroup{Tile: tile, From: from, Type: konType})
}

func (p *PlayData) HasKon(tile int32) bool {
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

func (p *PlayData) GetKon(tile int32) *KonGroup {
	for _, group := range p.konGroups {
		if group.Tile == tile {
			return &group
		}
	}
	return nil
}

func (p *PlayData) GetPon(tile int32) *Group {
	for _, group := range p.ponGroups {
		if group.Tile == tile {
			return &group
		}
	}
	return nil
}

func (p *PlayData) RemovePon(tile int32) Group {
	for i, group := range p.ponGroups {
		if group.Tile == tile {
			p.ponGroups = append(p.ponGroups[:i], p.ponGroups[i+1:]...)
			return group
		}
	}
	return Group{}
}

func (p *PlayData) RemoveKon(tile int32) KonGroup {
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

func (p *PlayData) tilesForChowLeft() []int32 {
	tiles := make([]int32, len(p.chowGroups))
	for i, group := range p.chowGroups {
		tiles[i] = int32(group.LeftTile)
	}
	return tiles
}

func (p *PlayData) tilesForPon() []int32 {
	tiles := make([]int32, len(p.ponGroups))
	for i, group := range p.ponGroups {
		tiles[i] = group.Tile
	}
	return tiles
}

func (p *PlayData) tilesForKon() (tiles []int32, countAnKon int32) {
	tiles = make([]int32, len(p.konGroups))
	for i, group := range p.konGroups {
		tiles[i] = int32(group.Tile)
		if group.Type == KonTypeAn {
			countAnKon++
		}
	}
	return
}

// CanSelfKon 判断是否可以自杠
func (p *PlayData) canSelfKon(rule *Rule, ignoreTiles []int32) bool {
	p.canGangTiles = make([]int32, 0)
	counts := make(map[int32]int)
	for _, tile := range p.handTiles {
		if !HasElement(ignoreTiles, tile) {
			counts[tile]++
		}
	}

	if !p.call {
		for _, pon := range p.ponGroups {
			if p.HasTile(pon.Tile) {
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

func (p *PlayData) canKonAfterCall(tile int32, konType KonType, rule *Rule) bool {
	if KonTypeZhi != konType && tile != p.handTiles[len(p.handTiles)-1] {
		return false
	}

	hudata := NewCheckHuData(p.play, p, false)
	if KonTypeZhi != konType {
		hudata.TilesInHand = hudata.TilesInHand[:len(hudata.TilesInHand)-1]
	}
	call0 := Service.CheckCall(hudata, rule)
	RemoveAllElement(&hudata.TilesInHand, tile)
	call1 := Service.CheckCall(hudata, rule)
	if len(call0) != 1 || len(call1) != 1 {
		return false
	}
	return HasSameKeys(call0[TileNull], call1[TileNull])
}
