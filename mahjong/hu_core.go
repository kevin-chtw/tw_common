package mahjong

import (
	"sort"
)

type HuCore struct {
	maxHandCount int
	mapHuAll     []map[int]TileStyle // 普通牌型表
	mapHuAllFZ   []map[int]TileStyle // 风牌表
	mapHu4All    map[int]TileStyle   // 普通牌型4个以上表
	mapHu4AllFZ  map[int]TileStyle   // 风牌4个以上表
	byArray      []TileStyle         // 普通牌型快速查询表
	byArrayFZ    []TileStyle         // 风牌快速查询表
	initialized  bool
}

func NewHuCore(maxHandCount int) *HuCore {
	hc := &HuCore{
		maxHandCount: maxHandCount,
		mapHuAll:     make([]map[int]TileStyle, maxHandCount+1),
		mapHuAllFZ:   make([]map[int]TileStyle, maxHandCount+1),
		mapHu4All:    make(map[int]TileStyle),
		mapHu4AllFZ:  make(map[int]TileStyle),
		byArray:      make([]TileStyle, 1<<(MAX_VAL_NUM*2)),
		byArrayFZ:    make([]TileStyle, 1<<(MAX_VAL_NUM*2)),
	}

	for i := 0; i <= maxHandCount; i++ {
		hc.mapHuAll[i] = make(map[int]TileStyle)
		hc.mapHuAllFZ[i] = make(map[int]TileStyle)
	}

	hc.initialize()
	return hc
}

func (hc *HuCore) initialize() {
	if hc.initialized {
		return
	}

	// 初始化单刻子、顺子组合
	single := make(map[int]int)
	singleFZ := make(map[int]int)
	singleJiang := make(map[int]int)
	singleJiangFZ := make(map[int]int)

	hc.trainSingleKe(single, singleFZ, singleJiang, singleJiangFZ)
	// 训练所有组合
	hc.trainAllComb(single, hc.mapHuAll)
	hc.trainAllCombJiang(singleJiang, hc.mapHuAll)
	hc.trainAllComb(singleFZ, hc.mapHuAllFZ)
	hc.trainAllCombJiang(singleJiangFZ, hc.mapHuAllFZ)

	// 构建快速查询表
	hc.buildQuickTable()

	hc.initialized = true
}
func (hc *HuCore) trainSingleKe(single, singleFZ, singleJiang, singleJiangFZ map[int]int) {
	tiles := []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 3}
	single[hc.getKeyByIndex(tiles)] = 0
	singleFZ[hc.getKeyByIndex(tiles)] = 0

	// 1.1 刻子
	for i := 0; i < MAX_VAL_NUM; i++ {
		tiles = make([]int, MAX_KEY_NUM)
		for n := 0; n < 3; n++ {
			tiles[i] = 3 - n
			tiles[9] = n // 赖子数
			single[hc.getKeyByIndex(tiles)] = 0
			if i < 7 {
				singleFZ[hc.getKeyByIndex(tiles)] = 0
			}
		}
	}

	// 1.2 顺子 没赖子
	for i := 0; i < MAX_VAL_NUM-2; i++ {
		tiles = make([]int, MAX_KEY_NUM)
		tiles[i] = 1
		tiles[i+1] = 1
		tiles[i+2] = 1
		single[hc.getKeyByIndex(tiles)] = 1
	}

	// 1.3 顺子 1个赖子 (2个赖子时也就是刻子)
	for i := 0; i < MAX_VAL_NUM-2; i++ {
		for n := 0; n < 3; n++ {
			tiles = make([]int, MAX_KEY_NUM)
			tiles[i] = 1
			tiles[i+1] = 1
			tiles[i+2] = 1
			tiles[i+n] = 0
			tiles[9] = 1
			single[hc.getKeyByIndex(tiles)] = 1
		}
	}
	// 2.1 将牌
	tiles = make([]int, MAX_KEY_NUM)
	tiles[9] = 2 // 赖子数
	singleJiang[hc.getKeyByIndex(tiles)] = 0
	singleJiangFZ[hc.getKeyByIndex(tiles)] = 0

	for i := 0; i < MAX_VAL_NUM; i++ {
		tiles := make([]int, MAX_KEY_NUM)
		for n := 0; n < 2; n++ {
			tiles[i] = 2 - n
			tiles[9] = n // 赖子数
			singleJiang[hc.getKeyByIndex(tiles)] = 0
			if i < 7 {
				singleJiangFZ[hc.getKeyByIndex(tiles)] = 0
			}
		}
	}
}

func (hc *HuCore) trainAllComb(single map[int]int, outMap []map[int]TileStyle) {
	// 获取并排序map的键
	keys := make([]int, 0, len(single))
	for k := range single {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	var dfs func(start, sumKey, sumVal int, depth int)
	dfs = func(start, sumKey, sumVal int, depth int) {
		if depth > 0 { // 长度 1~5 才需要写入
			if hc.isValidKey(sumKey) {
				hc.addMap(outMap, sumKey, sumVal)
			}
			if depth == 5 { // 已到最大长度，无需继续
				return
			}
		}
		for i := start; i < len(keys); i++ {
			dfs(i, sumKey+keys[i], sumVal+single[keys[i]], depth+1)
		}
	}
	dfs(0, 0, 0, 0)
}

func (hc *HuCore) trainAllCombJiang(single map[int]int, outMap []map[int]TileStyle) {
	tempMap := make([]map[int]TileStyle, hc.maxHandCount+1)
	for i := range tempMap {
		tempMap[i] = make(map[int]TileStyle)
		for k, v := range outMap[i] {
			tempMap[i][k] = v
		}
	}

	for key := range single {
		hc.addMap(outMap, key, single[key])

		for j := 0; j <= hc.maxHandCount; j++ {
			for tempKey, style := range tempMap[j] {
				newKey := key + tempKey + (int(style.NaiZiCount&BIT_VAL_FLAG) << 27)
				if hc.isValidKey(newKey) {
					hc.addMap(outMap, newKey, single[key]+style.ShunCount)
				}
			}
		}
	}
}

func (hc *HuCore) buildQuickTable() {
	var numAll int = 0
	for _, m := range hc.mapHuAll {
		numAll += len(m)
	}
	for _, m := range hc.mapHuAllFZ {
		numAll += len(m)
	}

	for i := 0; i < (hc.maxHandCount + 1); i++ {
		for key, style := range hc.mapHuAll[i] {
			nArrayIndex, localMax := hc.getArrayAndMax(key)
			if localMax < 4 {
				hc.byArray[nArrayIndex] = style
			} else {
				hc.mapHu4All[key] = style
			}
		}

		for key, style := range hc.mapHuAllFZ[i] {
			nArrayIndex, localMax := hc.getArrayAndMax(key)
			if localMax < 4 {
				hc.byArrayFZ[nArrayIndex] = style
			} else {
				hc.mapHu4AllFZ[key] = style
			}
		}
	}
}

func (hc *HuCore) CheckBasicHu(cards []int32, countLaiZi int) bool {
	if !hc.initialized {
		return false
	}

	tiles := make([]int, 42)
	for _, tile := range cards {
		color := int(TileColor(tile))
		point := TilePoint(tile)
		pos := SEQ_BEGIN_BY_COLOR[color] + point
		if tiles[pos]++; tiles[pos] > 4 {
			return false
		}
	}

	return hc.checkBasicHu(tiles, countLaiZi)
}

func (hc *HuCore) checkBasicHu(tiles []int, countLaiZi int) bool {
	var countShun int = 0
	var byJiangNum uint8 = 0
	for cor := ColorCharacter; cor <= ColorWind; cor++ {
		var byArray []TileStyle
		var hu4All map[int]TileStyle

		if cor == ColorWind {
			byArray = hc.byArrayFZ
			hu4All = hc.mapHu4AllFZ
		} else {
			byArray = hc.byArray
			hu4All = hc.mapHu4All
		}

		nMax := MAX_VAL_NUM
		if cor == ColorWind {
			nMax = MAX_FENZI_NUM
		}

		isArrayFlag, nNum, nVal := hc.toNumVal(tiles[MAX_VAL_NUM*cor:], nMax)
		if nNum == 0 {
			continue
		}

		// 预判断
		if isArrayFlag && !byArray[nVal].Enable {
			return false
		}

		nVal4 := hc.getKey(tiles[MAX_VAL_NUM*cor:], MAX_VAL_NUM)
		if cor == ColorWind {
			nVal4 &= 0x1FFFFF
		}

		nNaiTry := uint8(0xFF)
		if isArrayFlag {
			nNaiTry = uint8(byArray[nVal].NaiZiCount)
			countShun += byArray[nVal].ShunCount
		} else {
			if style, ok := hu4All[nVal4]; ok {
				nNaiTry = uint8(style.NaiZiCount)
				countShun += style.ShunCount
			}
		}

		if nNaiTry == 0xFF {
			return false
		}

		if (nNum+int(nNaiTry))%3 == 2 {
			byJiangNum++
		}

		countLaiZi -= int(nNaiTry)

		if int(byJiangNum) > countLaiZi+1 {
			return false
		}
	}

	return byJiangNum > 0 || countLaiZi >= 2
}

func (hc *HuCore) isValidKey(key int) bool {
	tiles := make([]int, MAX_KEY_NUM)
	for i := 0; i < MAX_KEY_NUM; i++ {
		tiles[i] = (key >> (BIT_VAL_NUM * i)) & BIT_VAL_FLAG
	}

	if tiles[MAX_KEY_NUM-1] > MAX_NAI_NUM {
		return false
	}
	total := 0
	for _, cnt := range tiles {
		total += cnt
		if cnt > 4 || total > hc.maxHandCount {
			return false
		}
	}

	return total > 0 && total <= hc.maxHandCount
}

func (hc *HuCore) getNumByKey(llVal int, byNum uint8) uint8 {
	var byIndex [MAX_KEY_NUM]uint8
	for i := 0; i < len(byIndex); i++ {
		byIndex[i] = uint8((llVal >> (BIT_VAL_NUM * i)) & BIT_VAL_FLAG)
	}

	var nNum uint8 = 0
	for i := 0; i < int(byNum); i++ {
		nNum += byIndex[i]
	}
	return nNum
}

func (hc *HuCore) addMap(outMap []map[int]TileStyle, key int, shunCount int) {
	nNum := hc.getNumByKey(key, MAX_VAL_NUM)
	byNum := (key >> (BIT_VAL_NUM * 9)) & BIT_VAL_FLAG
	val := (key & 0x7FFFFFF)
	if style, exists := outMap[nNum][val]; exists {
		//优先对对胡
		if byNum < style.NaiZiCount || (byNum == style.NaiZiCount && style.ShunCount > shunCount) {
			style.NaiZiCount = int(byNum)
			style.ShunCount = shunCount
			outMap[nNum][val] = style
		}
	} else {
		outMap[nNum][val] = TileStyle{
			NaiZiCount: int(byNum),
			ShunCount:  shunCount,
			Enable:     true,
		}
	}
}

func (hc *HuCore) getArrayAndMax(llVal int) (int, uint8) {
	var byMax uint8 = 0
	byIndex := make([]uint8, MAX_VAL_NUM)
	for i := 0; i < MAX_VAL_NUM; i++ {
		byIndex[i] = uint8((llVal >> (BIT_VAL_NUM * i)) & BIT_VAL_FLAG)
		if byIndex[i] > byMax {
			byMax = byIndex[i]
		}
	}
	return hc.getArrayIndex(byIndex, MAX_VAL_NUM), byMax
}

func (hc *HuCore) getArrayIndex(byIndex []uint8, byNum int) int {
	nKey := 0
	for i := 0; i < byNum; i++ {
		if (byIndex[i] & BIT_VAL_FLAG) > 3 {
			byIndex[i] -= 3 //为节约内存，最大支持3张
		}
		nKey |= int(byIndex[i]&0x03) << (2 * i) //因为小于等于3，所以只需要2bit
	}
	return nKey
}

func (hc *HuCore) toNumVal(tiles []int, max int) (flag bool, num, val int) {
	flag = true
	for i := 0; i < max; i++ {
		num += tiles[i]
		val |= tiles[i] << (2 * i)
		if tiles[i] > 3 {
			flag = false
		}
	}
	return
}

func (hc *HuCore) getKeyByIndex(tiles []int) int {
	return hc.getKey(tiles, MAX_KEY_NUM)
}

func (hc *HuCore) getKey(tiles []int, max int) int {
	key := 0
	for i := 0; i < max; i++ {
		key |= (int)(tiles[i]&BIT_VAL_FLAG) << (BIT_VAL_NUM * i)
	}
	return key
}
