package mahjong

import (
	"slices"
	"sort"
	"sync"
)

const MaxTing = 99

// tileListToMap is Go version of MJ::TileListToMap.
// It splits tiles into normal counts, laizi count, and "other" (extra) count.
func tileListToMap(tiles []Tile, laizi []Tile, laiziCount *int, otherCount *int) map[Tile]int {
	*laiziCount = 0
	*otherCount = 0
	m := make(map[Tile]int)
	for _, card := range tiles {
		if slices.Contains(laizi, card) {
			*laiziCount++
		} else if card.IsExtra() {
			*otherCount++
		} else {
			m[card]++
		}
	}
	return m
}

// calcStepTo13Yao is Go port of MJ::CalcStepTo13Yao.
func calcStepTo13Yao(cards []Tile, laiziCards []Tile) int {
	if len(cards) < 13 {
		return MaxTing
	}
	laiziCount := 0
	targets := []Tile{
		MakeTile(ColorCharacter, 0),
		MakeTile(ColorCharacter, 8),
		MakeTile(ColorBamboo, 0),
		MakeTile(ColorBamboo, 8),
		MakeTile(ColorDot, 0),
		MakeTile(ColorDot, 8),
		MakeTile(ColorWind, 0),
		MakeTile(ColorWind, 1),
		MakeTile(ColorWind, 2),
		MakeTile(ColorWind, 3),
		MakeTile(ColorDragon, 0),
		MakeTile(ColorDragon, 1),
		MakeTile(ColorDragon, 2),
	}

	var selected []Tile
	extraCount := 0
	for _, card := range cards {
		if slices.Contains(laiziCards, card) {
			laiziCount++
		} else if slices.Contains(targets, card) {
			if slices.Contains(selected, card) {
				extraCount++
			} else {
				selected = append(selected, card)
			}
		}
	}
	usedCount := len(selected)
	if extraCount > 0 {
		usedCount++
	}
	usedCount += laiziCount
	if usedCount >= 14 {
		return 0
	}
	return 14 - usedCount
}

// internal helper for QiDui.
func checkQiDuiStep(countsMap map[Tile]int, laiziCount int) int {
	pairCount := laiziCount
	// Ensure deterministic order.
	keys := make([]Tile, 0, len(countsMap))
	for k := range countsMap {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	for _, k := range keys {
		if !k.IsExtra() { // mimic !MJIsExtraTile in C++
			pairCount += countsMap[k] / 2
		}
	}
	if pairCount >= 7 {
		return 0
	}
	return 7 - pairCount
}

// MJTingType corresponds to MJTingBase::Type.
type MJTingType int

const (
	TingTypeNormal MJTingType = iota
	TingTypeQiDui
	TingType13Yao
)

// MJTingBase is Go version of MJTingBase.
type MJTingBase struct {
	normalTool *TingNormal
}

func newMJTingBase(tileCount int) *MJTingBase {
	t := NewTingNormal()
	t.Init(tileCount)
	return &MJTingBase{normalTool: t}
}

// fixHandCount makes hand size congruent to 1 mod 3, similar to _FixHandCount.
func fixHandCount(cardCount int) int {
	return (122 - cardCount) % 3
}

// CalcStep computes generic step count; by default only normal pinghu is considered.
func (m *MJTingBase) CalcStep(cards []Tile, laiziCards []Tile, extraTypes []int, typ *int) int {
	defer func() {
		if r := recover(); r != nil {
			panic(r)
		}
	}()
	var laiziCount, otherCount int
	countsMap := tileListToMap(cards, laiziCards, &laiziCount, &otherCount)
	otherCount += fixHandCount(len(cards))
	return m.normalTool.CalcStepPingHu(countsMap, laiziCount, otherCount)
}

// CalcTing = CalcStep - 1, but not less than 0.
func (m *MJTingBase) CalcTing(cards []Tile, laiziCards []Tile, extraTypes []int, typ *int) int {
	step := m.CalcStep(cards, laiziCards, extraTypes, typ) - 1
	if step < 0 {
		return 0
	}
	return step
}

// MJTingEx14 is a thin wrapper for 14-tile hands, supporting extra types (QiDui, 13Yao).
type MJTingEx14 struct {
	MJTingBase
}

func newMJTingEx14() *MJTingEx14 {
	return &MJTingEx14{MJTingBase: *newMJTingBase(14)}
}

// CalcStep for 14 tiles, considering extra types (QiDui, 13Yao).
func (m *MJTingEx14) CalcStep(cards []Tile, laiziCards []Tile, extraTypes []int, typ *int) int {
	var laiziCount, otherCount int
	countsMap := tileListToMap(cards, laiziCards, &laiziCount, &otherCount)
	otherCount += fixHandCount(len(cards))

	step := m.normalTool.CalcStepPingHu(countsMap, laiziCount, otherCount)
	SetValueIfNotNull(typ, int(TingTypeNormal))

	if slices.Contains(extraTypes, int(TingTypeQiDui)) && len(cards) > 12 {
		step1 := checkQiDuiStep(countsMap, laiziCount)
		if step1 < step {
			step = step1
			SetValueIfNotNull(typ, int(TingTypeQiDui))
		}
	}
	if slices.Contains(extraTypes, int(TingType13Yao)) && len(cards) > 12 {
		step1 := calcStepTo13Yao(cards, laiziCards)
		if step1 < step {
			step = step1
			SetValueIfNotNull(typ, int(TingType13Yao))
		}
	}
	return step
}

// ----- single core + CalcTing method (you only init once) -----

// tingCore 是内部使用的“核心对象”。外部不要直接访问。
type tingCore struct {
	tileCount int
	base      *MJTingBase
	ex14      *MJTingEx14
}

// newTingCore creates a core for given tileCount.
// 内部会根据牌张数选择对应的计算器，14 张时支持 平胡+七对+十三幺。
func newTingCore(tileCount int) *tingCore {
	core := &tingCore{tileCount: tileCount}
	if tileCount == 14 {
		core.ex14 = newMJTingEx14()
	} else {
		core.base = newMJTingBase(tileCount)
	}
	return core
}

var (
	coreOnce sync.Once
	coreInst *tingCore
)

// Init(tileCount) 只在第一次调用时用 tileCount 初始化核心，以后再传入其他值会被忽略。
// 也就是说，你在业务里只需调用一次 Init(14)（或 11/17），然后就可以一直用 CalcTing。
func InitTingCore(tileCount int) {
	coreOnce.Do(func() {
		coreInst = newTingCore(tileCount)
	})
}

// CalcTing 是对外唯一需要关心的计算接口：给手牌、赖子、（可选）牌型列表，返回听牌向听数和选中的牌型。
// 在调用 CalcTing 之前必须先调用一次 Init(...)。
func CalcTing(cards []Tile, laiziCards []Tile, extraTypes []MJTingType) (int, MJTingType) {
	if coreInst == nil {
		panic("mjting: Init must be called before CalcTing")
	}
	return coreInst.calcTing(cards, laiziCards, extraTypes)
}

// 内部实现：真正的计算逻辑。
func (c *tingCore) calcTing(cards []Tile, laiziCards []Tile, extraTypes []MJTingType) (int, MJTingType) {
	switch c.tileCount {
	case 14:
		// 14 张：用 MJTingEx14，支持平胡+七对+十三幺
		intTypes := make([]int, len(extraTypes))
		for i, et := range extraTypes {
			intTypes[i] = int(et)
		}
		var typInt int
		stepToHu := c.ex14.CalcStep(cards, laiziCards, intTypes, &typInt)
		stepToTing := stepToHu - 1
		if stepToTing < 0 {
			stepToTing = 0
		}
		return stepToTing, MJTingType(typInt)
	default:
		// 其他张数：只算普通平胡
		var typInt int
		stepToTing := c.base.CalcTing(cards, laiziCards, nil, &typInt)
		return stepToTing, MJTingType(typInt)
	}
}
