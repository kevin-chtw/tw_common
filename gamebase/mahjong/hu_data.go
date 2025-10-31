package mahjong

import (
	"slices"

	"github.com/kevin-chtw/tw_proto/game/pbmj"
)

type HuData struct {
	*PlayData
	Tiles        []Tile
	ExtraHuTypes []int32 // 额外胡类型
	curTile      Tile
}

func NewHuData(playData *PlayData, self bool) *HuData {
	data := &HuData{
		PlayData:     playData,
		Tiles:        slices.Clone(playData.handTiles),
		curTile:      playData.Play.curTile,
		ExtraHuTypes: playData.Play.PlayImp.GetExtraHuTypes(playData, self),
	}

	return data
}

func (h *HuData) GetCurTile() Tile {
	return h.curTile
}

func (h *HuData) CheckHu() (*pbmj.MJHuData, bool) {
	if len(h.Tiles)%3 != 2 {
		h.Tiles = append(h.Tiles, h.curTile)
	}
	if !h.Play.PlayImp.CheckHu(h) {
		return nil, false
	}

	hyTypes := Service.GetHuTypes(h)
	hyTypes = append(hyTypes, h.ExtraHuTypes...)
	result := &pbmj.MJHuData{
		HuTypes: hyTypes,
		Multi:   Service.TotalMuti(hyTypes),
	}
	return result, true
}

func (h *HuData) CheckCall() map[Tile]map[Tile]int64 {
	callMap := make(map[Tile]map[Tile]int64)
	count := len(h.Tiles) % 3
	switch count {
	case 2:
		tileSet := make(map[Tile]bool)
		for _, tile := range h.Tiles {
			tileSet[tile] = true
		}

		tempData := *h
		for tile := range tileSet {
			tempData.Tiles = RemoveElements(h.Tiles, tile, 1)
			fans := tempData.checkCalls()
			if len(fans) > 0 {
				callMap[tile] = fans
			}
		}
	case 1:
		// 直接检查叫牌
		callData := h.checkCalls()
		if len(callData) > 0 {
			callMap[TileNull] = callData
		}
	}

	return callMap
}

func (h *HuData) CanHu() bool {
	tiles, laiCount := h.countLaiZi()
	return DefaultHuCore.CheckBasicHu(tiles, laiCount)
}

func (h *HuData) checkCalls() map[Tile]int64 {
	mutils := make(map[Tile]int64)
	testTiles := Service.GetAllTiles(h.Play.GetRule())
	originalTiles := slices.Clone(h.Tiles)
	for tile := range testTiles {
		h.curTile = tile
		h.Tiles = append(h.Tiles, tile)
		if result, ok := h.CheckHu(); ok {
			mutils[tile] = result.Multi
		}
		h.Tiles = originalTiles
	}
	return mutils
}

func (h *HuData) countLaiZi() ([]Tile, int) {
	if len(h.Play.tilesLai) == 0 {
		return h.Tiles, 0
	}
	laiCount := 0
	newTiles := slices.DeleteFunc(h.Tiles, func(t Tile) bool {
		if _, ok := h.Play.tilesLai[t]; ok {
			laiCount++
			return true
		}
		return false
	})
	return newTiles, laiCount
}
