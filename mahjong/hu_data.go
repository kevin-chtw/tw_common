package mahjong

import "slices"

type HuData struct {
	TilesInHand      []Tile
	LaiCount         int
	tilesForChowLeft []Tile
	tilesForPon      []Tile
	tilesForKon      []Tile
	ExtraHuTypes     []int32
	paoTile          Tile
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

func removeLaiZi(tiles []Tile, laiTiles ...Tile) (newTiles []Tile, laiCount int) {
	laiSet := make(map[Tile]struct{}, len(laiTiles))
	for _, t := range laiTiles {
		laiSet[t] = struct{}{}
	}
	newTiles = slices.DeleteFunc(tiles, func(t Tile) bool {
		if _, ok := laiSet[t]; ok {
			laiCount++
			return true
		}
		return false
	})
	return
}
