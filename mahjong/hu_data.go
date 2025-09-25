package mahjong

import "slices"

type HuData struct {
	TilesInHand      []int32
	LaiCount         int
	tilesForChowLeft []int32
	tilesForPon      []int32
	tilesForKon      []int32
	ExtraHuTypes     []int32
	paoTile          int32
	countAnKon       int32
	isCall           bool
	canCall          bool
}

func NewCheckHuData(play *Play, playData *PlayData, self bool) *HuData {
	data := &HuData{
		tilesForChowLeft: playData.tilesForChowLeft(),
		tilesForPon:      playData.tilesForPon(),
		paoTile:          TileNull,
		isCall:           playData.call,
		canCall:          true,
	}

	data.TilesInHand, data.LaiCount = removeLaiZi(playData.handTiles, play.tilesLai...)
	if self {
		data.ExtraHuTypes = play.ExtraHuTypes.SelfExtraFans()
	} else {
		data.paoTile = play.GetCurTile()
		data.TilesInHand = append(data.TilesInHand, data.paoTile)
		data.ExtraHuTypes = play.ExtraHuTypes.PaoExtraFans()
	}
	data.tilesForKon, data.countAnKon = playData.tilesForKon()
	return data
}

func removeLaiZi(tiles []int32, laiTiles ...int32) (newTiles []int32, laiCount int) {
	laiSet := make(map[int32]struct{}, len(laiTiles))
	for _, t := range laiTiles {
		laiSet[t] = struct{}{}
	}
	newTiles = slices.DeleteFunc(tiles, func(t int32) bool {
		if _, ok := laiSet[t]; ok {
			laiCount++
			return true
		}
		return false
	})
	return
}
