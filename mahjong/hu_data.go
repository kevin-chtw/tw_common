package mahjong

type HuData struct {
	TilesInHand      []int32
	tilesForChowLeft []int32
	tilesForPon      []int32
	tilesForKon      []int32
	tilesLai         []int32
	ExtraHuTypes     []int32
	paoTile          int32
	countAnKon       int32
	isCall           bool
	canCall          bool
}

func NewCheckHuData(play *Play, playData *PlayData, self bool) *HuData {
	data := &HuData{
		TilesInHand:      make([]int32, len(playData.handTiles)),
		tilesForChowLeft: playData.tilesForChowLeft(),
		tilesForPon:      playData.tilesForPon(),
		tilesLai:         play.tilesLai,
		paoTile:          TileNull,
		isCall:           playData.call,
		canCall:          true,
	}

	copy(data.TilesInHand, playData.handTiles)
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
