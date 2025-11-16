package mahjong

import (
	"bufio"
	"fmt"
	"os"
)

const tileIDStep = 0x10

// TingStepRecord mirrors MJ::TingStepRecord in C++.
type TingStepRecord struct {
	Double int8 // pair-like (two-tile) groups
	Single int8 // single tiles
}

func (t TingStepRecord) Sum() int {
	return int(t.Double) + int(t.Single)
}

// TingCode is the key in the ting table.
type TingCode uint32

// TingNormal is Go version of MJ::TingNormal.
type TingNormal struct {
	tingCountMap map[TingCode]TingStepRecord
	isInited     bool
}

func NewTingNormal() *TingNormal {
	return &TingNormal{
		tingCountMap: make(map[TingCode]TingStepRecord),
	}
}

// TableSize returns current number of entries in the internal tingCountMap.
func (t *TingNormal) TableSize() int {
	return len(t.tingCountMap)
}

// Init tries to load precomputed table from file, otherwise falls back to in-memory generation.
func (t *TingNormal) Init(tileCount int) {
	if t.isInited {
		return
	}
	t.isInited = true

	filename := fmt.Sprintf("MahjongTingCodes%d", tileCount)
	if f, err := os.Open(filename); err == nil {
		defer f.Close()
		reader := bufio.NewReader(f)
		var v int64
		for {
			_, err := fmt.Fscanf(reader, "%d", &v)
			if err != nil {
				break
			}
			code := TingCode(v >> 5)
			db := (v >> 2) & 0x7
			sg := v & 0x3
			t.tingCountMap[code] = TingStepRecord{
				Double: int8(db),
				Single: int8(sg),
			}
			// consume optional comma or whitespace; ignore errors
			_, _ = reader.ReadByte()
		}
	}

	// If file not loaded or counts not matching expected, rebuild in memory.
	if (tileCount == 14 && len(t.tingCountMap) != 140920) ||
		(tileCount == 17 && len(t.tingCountMap) != 417491) ||
		(tileCount != 14 && tileCount != 17) {
		t.tingCountMap = make(map[TingCode]TingStepRecord)
		t.initialize(tileCount)
	}
}

// CalcStepPingHu is Go port of MJ::TingNormal::CalcStepPingHu.
func (t *TingNormal) CalcStepPingHu(countsMap map[Tile]int, laiziCount, otherCount int) int {
	step := t.checkTingStep(countsMap, false, laiziCount, otherCount)
	for tile, cnt := range countsMap {
		if cnt >= 2 {
			countsMap[tile] -= 2
			step0 := t.checkTingStep(countsMap, true, laiziCount, otherCount)
			if step0 < step {
				step = step0
			}
			countsMap[tile] += 2
		}
	}
	return step
}

// ----- internal helpers (ported from MJTingNormal.im.hpp) -----

// countSequenceGenerator generates count sequences like "2111" whose total <= maxHandCount.
func countSequenceGenerator(maxHandCount int, onGen func(string)) {
	var temp string
	var run func(rest int)
	run = func(rest int) {
		onGen(temp)
		if rest == 0 || len(temp) >= 9 {
			return
		}
		for i := 1; i <= rest && i <= 4; i++ {
			temp += string(rune('0' + i))
			run(rest - i)
			temp = temp[:len(temp)-1]
		}
	}
	run(maxHandCount)
}

// stepSequenceGenerator generates step sequences like "0101" with constraints similar to C++.
func stepSequenceGenerator(onGen func(string)) {
	var temp string
	var run func(curLen int)
	run = func(curLen int) {
		onGen(temp)
		if len(temp) > 13 {
			return
		}
		for i := 0; i < 2; i++ {
			if i+curLen+1 > 8 {
				break
			}
			temp += string(rune('0' + i))
			run(curLen + 1 + i)
			temp = temp[:len(temp)-1]
		}
	}
	run(0)
}

// simpleCombineCalc was used for a simpler heuristic version in C++, but the final
// implementation uses BestCombineCalc; for clarity we only port BestCombineCalc and
// the CheckPartCounts wrapper actually used.

// bestCombineCalc finds the minimal Double/Single decomposition for a single suit.
type bestCombineCalc struct {
	best TingStepRecord
}

func (b *bestCombineCalc) doCheck(counts []int) TingStepRecord {
	// upper bound for singles
	b.best.Single = int8(len(counts) * 4)
	b.doPick(counts, TingStepRecord{}, 0)
	return b.best
}

func (b *bestCombineCalc) doPick(counts []int, st TingStepRecord, index int) {
	hasKe := b.pickKe(counts, st, index)
	hasShun := b.pickShun(counts, st, index)
	k2 := b.pickKe2(counts, st, index)
	s2 := b.pickShun2(counts, st, index)
	if !hasKe && !hasShun && !k2 && !s2 {
		b.end(counts, st)
	}
}

func (b *bestCombineCalc) pickKe(counts []int, st TingStepRecord, index int) bool {
	for i := index; i < len(counts); i++ {
		if counts[i] >= 3 {
			cc := append([]int(nil), counts...)
			cc[i] -= 3
			b.doPick(cc, st, i)
			return true
		}
	}
	return false
}

func (b *bestCombineCalc) pickShun(counts []int, st TingStepRecord, index int) bool {
	if len(counts) < 3 {
		return false
	}
	for ; index < len(counts)-2; index++ {
		if counts[index] > 0 && counts[index+1] > 0 && counts[index+2] > 0 {
			cc := append([]int(nil), counts...)
			cc[index]--
			cc[index+1]--
			cc[index+2]--
			b.doPick(cc, st, index)
			return true
		}
	}
	return false
}

func (b *bestCombineCalc) pickKe2(counts []int, st TingStepRecord, index int) bool {
	for i := index; i < len(counts); i++ {
		if counts[i] >= 2 {
			cc := append([]int(nil), counts...)
			cc[i] -= 2
			st2 := st
			st2.Double++
			b.doPick(cc, st2, i)
			return true
		}
	}
	return false
}

func (b *bestCombineCalc) pickShun2(counts []int, st TingStepRecord, index int) bool {
	if len(counts) < 2 {
		return false
	}
	for ; index < len(counts)-1; index++ {
		n0 := counts[index]
		for i := 1; i <= 2 && n0 > 0 && index+i < len(counts); i++ {
			n1 := counts[index+i]
			n := n0
			if n1 < n {
				n = n1
			}
			if n > 0 {
				cc := append([]int(nil), counts...)
				cc[index] -= n
				cc[index+i] -= n
				st2 := st
				st2.Double += int8(n)
				b.doPick(cc, st2, index)
				return true
			}
		}
	}
	return false
}

func (b *bestCombineCalc) end(counts []int, st TingStepRecord) {
	for _, n := range counts {
		st.Single += int8(n)
	}
	if st.Sum() < b.best.Sum() {
		b.best = st
	}
}

// checkPartCounts is Go equivalent of CheckPartCounts(...) in C++.
func checkPartCounts(cardsCount []int) TingStepRecord {
	b := &bestCombineCalc{}
	return b.doCheck(cardsCount)
}

func (t *TingNormal) initialize(tileCount int) {
	if len(t.tingCountMap) > 0 {
		return
	}
	stepSeqs := make(map[int][]string)
	stepSequenceGenerator(func(s string) {
		stepSeqs[len(s)] = append(stepSeqs[len(s)], s)
	})

	countSequenceGenerator(tileCount, func(cs string) {
		if cs == "" {
			return
		}
		for _, ss := range stepSeqs[len(cs)-1] {
			t.addItemToMap(cs, ss)
		}
	})
}

// encode expression to TingCode (ExpressionToCode in C++).
func (t *TingNormal) expressionToCode(cardCounts, cardSteps string) TingCode {
	if len(cardSteps)+1 != len(cardCounts) {
		panic(fmt.Sprintf("invalid expression: %d %d", len(cardSteps), len(cardCounts)))
	}
	code := TingCode(1<<2) | TingCode(cardCounts[0]-'1')
	for i := 0; i < len(cardSteps); i++ {
		code <<= 1
		if cardSteps[i] != '0' {
			code |= 1
		}
		code <<= 2
		code |= TingCode(cardCounts[i+1] - '1')
	}
	return code
}

func (t *TingNormal) addItemToMap(cardCounts, cardSteps string) {
	if len(cardSteps)+1 != len(cardCounts) {
		panic(fmt.Sprintf("invalid expression: %d %d", len(cardSteps), len(cardCounts)))
	}
	var cardsCount []int
	cardsCount = append(cardsCount, int(cardCounts[0]-'0'))
	for i := 0; i < len(cardSteps); i++ {
		if cardSteps[i] != '0' {
			cardsCount = append(cardsCount, 0)
		}
		cardsCount = append(cardsCount, int(cardCounts[i+1]-'0'))
	}
	code := t.expressionToCode(cardCounts, cardSteps)
	t.tingCountMap[code] = checkPartCounts(cardsCount)
}

func (t *TingNormal) checkCount(cardCounts, cardSteps string, res *TingStepRecord) {
	code := t.expressionToCode(cardCounts, cardSteps)
	val, ok := t.tingCountMap[code]
	if !ok {
		panic(fmt.Sprintf("code not found, size=%d, cardCounts=%s, cardSteps=%s",
			len(t.tingCountMap), cardCounts, cardSteps))
	}
	res.Double += val.Double
	res.Single += val.Single
}

func (t *TingNormal) checkTingStep(countsMap map[Tile]int, hasJiang bool, laiziCount, otherCount int) int {
	var cardCounts, cardSteps string
	var preCard Tile
	counts := TingStepRecord{}
	first := true

	// countsMap is a map; in C++ it's ordered map. We need deterministic order by Tile.
	keys := make([]Tile, 0, len(countsMap))
	for k := range countsMap {
		keys = append(keys, k)
	}
	// simple ascending sort
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	for _, tile := range keys {
		cnt := countsMap[tile]
		if cnt < 0 {
			panic(fmt.Sprintf("negative count: %d", cnt))
		}
		if cnt == 0 {
			continue
		}
		if !first {
			step := int(tile-preCard) / tileIDStep
			if !tile.IsSuit() {
				step = 3
			}
			if step > 2 {
				t.checkCount(cardCounts, cardSteps, &counts)
				cardCounts = ""
				cardSteps = ""
			} else if step == 2 {
				cardSteps += "1"
			} else {
				cardSteps += "0"
			}
		}
		cardCounts += string(rune('0' + cnt))
		preCard = tile
		first = false
	}
	if cardCounts != "" {
		t.checkCount(cardCounts, cardSteps, &counts)
	}
	counts.Single += int8(otherCount)
	return t.calcStep(counts, hasJiang, laiziCount)
}

func (t *TingNormal) combineLaizi(record *TingStepRecord, laiziCount *int) {
	// every Double needs 1 laizi; every Single needs 2 laizi
	if *laiziCount < int(record.Double) {
		record.Double -= int8(*laiziCount)
		*laiziCount = 0
	} else {
		*laiziCount -= int(record.Double)
		record.Double = 0
		need := int(record.Single) * 2
		if *laiziCount >= need {
			*laiziCount -= need
			record.Single = 0
		} else {
			usedPairs := *laiziCount / 2
			record.Single -= int8(usedPairs)
			*laiziCount = *laiziCount % 2
		}
	}
}

func (t *TingNormal) calcStep(record TingStepRecord, hasJiang bool, laiziCount int) int {
	t.combineLaizi(&record, &laiziCount)
	step := 0
	if record.Double <= record.Single {
		step += int(record.Double)
		restSingle := int(record.Single - record.Double)
		if hasJiang {
			step += ((restSingle + 2) / 3) * 2
		} else {
			if restSingle > 2 {
				restSingle -= 2
				step++
				step += ((restSingle + 2) / 3) * 2
			} else if restSingle > 0 {
				step++
			} else {
				step += 2
			}
		}
	} else {
		restKS2 := int(record.Double - record.Single)
		step += int(record.Single)
		if hasJiang {
			step += (restKS2 / 3) * 2
			step += restKS2 % 3
		} else {
			step += (restKS2 / 3) * 2
			n := []int{2, 1, 2}
			step += n[restKS2%3]
		}
	}
	if step >= laiziCount {
		return step - laiziCount
	}
	return 0
}
