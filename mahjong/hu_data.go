package mahjong

import "slices"

type HuData struct {
	Tiles            []Tile
	LaiCount         int
	tilesForChowLeft []Tile
	tilesForPon      []Tile
	tilesForKon      []Tile
	ExtraHuTypes     []int32
	paoTile          Tile
	countAnKon       int32
	isTing           bool
	canTing          bool
}

func NewCheckHuData(play *Play, playData *PlayData, self bool) *HuData {
	data := playData.MakeHuData()
	if self {
		data.ExtraHuTypes = play.ExtraHuTypes.SelfExtraFans()
	} else {
		data.paoTile = play.GetCurTile()
		data.Tiles = append(data.Tiles, data.paoTile)
		data.ExtraHuTypes = play.ExtraHuTypes.PaoExtraFans()
	}

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
