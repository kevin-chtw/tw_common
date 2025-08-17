package mahjong

func TilesToMap(cards []int32) map[int32]int {
	tileMap := make(map[int32]int)
	for _, tile := range cards {
		tileMap[tile]++
	}
	return tileMap
}

func CheckQiDui(cards []int32, countLaiZi int) (bool, int) {
	tileMap := TilesToMap(cards)
	luxury := 0
	pairs := 0

	for _, count := range tileMap {
		if count >= 4 {
			luxury++
		}
		if count >= 2 {
			pairs += count / 2
		}
	}

	totalPairs := pairs + countLaiZi
	return totalPairs >= 7, luxury
}

func CheckAllPairs(cards []int32, countLaiZi int) (bool, int) {
	tileMap := TilesToMap(cards)
	luxury := 0
	pairs := 0

	for _, count := range tileMap {
		if count >= 4 {
			luxury++
		}
		if count%2 != 0 {
			return false, 0
		}
		pairs += count / 2
	}

	return (pairs*2 + countLaiZi) == len(cards), luxury
}

func CheckBasicHu(cards []int32, countLaiZi int) (bool, *bool, *bool) {
	tileMap := TilesToMap(cards)

	// 检查碰碰胡
	allPon := CheckPonPonHu(cards, countLaiZi)
	if allPon {
		return true, &allPon, nil
	}

	// 检查全顺子
	allShun := CheckStraightHu(cards, countLaiZi)
	if allShun {
		return true, nil, &allShun
	}

	// 检查任意胡牌
	anyHu := _CheckAnyHu(tileMap, countLaiZi)
	return anyHu, nil, nil
}

func _CheckAnyHu(cards map[int32]int, countLaiZi int) bool {
	// 实现类似C++的任意胡牌检查算法
	// 首先尝试找将牌
	for tile, count := range cards {
		tempMap := make(map[int32]int)
		for k, v := range cards {
			tempMap[k] = v
		}

		needLaiZi := 0
		if count >= 2 {
			tempMap[tile] -= 2
		} else if count == 1 {
			tempMap[tile] = 0
			needLaiZi = 1
		} else {
			continue
		}

		if countLaiZi >= needLaiZi && _CheckAllMeld(tempMap, countLaiZi-needLaiZi) {
			return true
		}
	}

	// 无将牌但赖子足够的情况
	if countLaiZi >= 2 && _CheckAllMeld(cards, countLaiZi-2) {
		return true
	}

	return false
}

func _CheckAllMeld(cards map[int32]int, countLaiZi int) bool {
	// 实现类似C++的_CheckAllMeld算法
	// 检查所有牌是否能组成顺子或刻子
	for tile, count := range cards {
		if count == 0 {
			continue
		}

		if IsHonorTile(tile) {
			// 字牌只能组成刻子
			if count%3 != 0 {
				if countLaiZi >= (3 - count%3) {
					countLaiZi -= (3 - count%3)
				} else {
					return false
				}
			}
		} else {
			// 数牌可以组成顺子或刻子
			// 这里简化处理，实际需要更复杂的算法
			if count >= 3 {
				continue
			}
			if countLaiZi >= (3 - count) {
				countLaiZi -= (3 - count)
			} else {
				return false
			}
		}
	}
	return true
}

func CheckPonPonHu(cards []int32, countLaiZi int) bool {
	tileMap := TilesToMap(cards)
	return _CheckPonPonHu(tileMap, countLaiZi)
}

func _CheckPonPonHu(cards map[int32]int, countLaiZi int) bool {
	// 实现完整的碰碰胡算法
	hasJiang := false
	needLaiZi := 0

	for _, count := range cards {
		if count == 0 {
			continue
		}

		remain := count % 3
		if remain == 0 {
			continue
		}

		if hasJiang {
			needLaiZi += (3 - remain)
		} else {
			hasJiang = true
			needLaiZi += (2 - remain)
		}
	}

	if !hasJiang {
		needLaiZi += 2
	}

	return needLaiZi <= countLaiZi
}

func CheckStraightHu(cards []int32, countLaiZi int) bool {
	tileMap := TilesToMap(cards)
	return _CheckStraightHu(tileMap, countLaiZi)
}

func _CheckStraightHu(cards map[int32]int, countLaiZi int) bool {
	// 实现全顺子胡牌算法
	tempMap := make(map[int32]int)
	for tile, count := range cards {
		if !IsSuitTile(tile) {
			return false
		}
		tempMap[tile] = count
	}

	return _TestJiang(tempMap, countLaiZi, _CheckAllStraight)
}

func _CheckAllStraight(cards map[int32]int, countLaiZi int) bool {
	// 检查所有牌是否能组成顺子
	for tile, count := range cards {
		if count == 0 {
			continue
		}

		next1 := tile + 1
		next2 := tile + 2

		count1 := cards[next1]
		count2 := cards[next2]

		minCount := min(count, min(count1, count2))
		if minCount == 0 {
			if countLaiZi > 0 {
				countLaiZi--
			} else {
				return false
			}
		}

		cards[tile] -= minCount
		cards[next1] -= minCount
		cards[next2] -= minCount
	}

	return true
}

func _TestJiang(cards map[int32]int, countLaiZi int, checker func(map[int32]int, int) bool) bool {
	// 尝试各种可能的将牌组合
	for tile, count := range cards {
		tempMap := make(map[int32]int)
		for k, v := range cards {
			tempMap[k] = v
		}

		needLaiZi := 0
		if count >= 2 {
			tempMap[tile] -= 2
		} else if count == 1 {
			tempMap[tile] = 0
			needLaiZi = 1
		} else {
			continue
		}

		if countLaiZi >= needLaiZi && checker(tempMap, countLaiZi-needLaiZi) {
			return true
		}
	}

	// 无将牌但赖子足够的情况
	if countLaiZi >= 2 && checker(cards, countLaiZi-2) {
		return true
	}

	return false
}
