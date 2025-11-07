package mahjong

import (
	"slices"

	"github.com/kevin-chtw/tw_proto/game/pbmj"
)

type HuData struct {
	*PlayData
	HuCoreType   HuCoreType
	Tiles        []Tile
	ExtraHuTypes []int32 // 额外胡类型
	CurTile      Tile
	Self         bool
}

func NewHuData(playData *PlayData, self bool) *HuData {
	data := &HuData{
		PlayData:     playData,
		HuCoreType:   HU_NON,
		Tiles:        slices.Clone(playData.handTiles),
		CurTile:      playData.Play.curTile,
		ExtraHuTypes: playData.Play.PlayImp.GetExtraHuTypes(playData, self),
		Self:         self,
	}

	return data
}

func (h *HuData) CheckHu() (*pbmj.MJHuData, bool) {
	if len(h.Tiles)%3 != 2 {
		h.Tiles = append(h.Tiles, h.CurTile)
	}

	h.HuCoreType = h.Play.PlayImp.CheckHu(h)
	if h.HuCoreType == HU_NON {
		return nil, false
	}

	result := Service.GetHuResult(h)
	return result, true
}

func (h *HuData) InitHuResult() *pbmj.MJHuData {
	r := &pbmj.MJHuData{
		Seat:    h.PlayData.seat,
		HuTypes: make([]int32, 0),
	}
	r.HuTypes = append(r.HuTypes, h.ExtraHuTypes...)
	return r
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

func (h *HuData) CanHu() HuCoreType {
	tiles, laiCount := h.CountLaiZi(h.Tiles)
	return DefaultHuCore.CheckBasicHu(tiles, laiCount)
}

func (h *HuData) checkCalls() map[Tile]int64 {
	mutils := make(map[Tile]int64)
	testTiles := Service.GetAllTiles(h.Play.GetRule())
	originalTiles := slices.Clone(h.Tiles)
	for tile := range testTiles {
		h.CurTile = tile
		h.Tiles = append(h.Tiles, tile)
		if result, ok := h.CheckHu(); ok {
			mutils[tile] = result.Multi
		}
		h.Tiles = originalTiles
	}
	return mutils
}

func (h *HuData) CountLaiZi(tiles []Tile) ([]Tile, int) {
	if len(h.Play.tilesLai) == 0 {
		return tiles, 0
	}
	laiCount := 0
	newTiles := slices.DeleteFunc(tiles, func(t Tile) bool {
		if _, ok := h.Play.tilesLai[t]; ok {
			laiCount++
			return true
		}
		return false
	})
	return newTiles, laiCount
}

func (h *HuData) CheckShun(tile Tile, p1, p2 int) bool {
	requiredTiles := [2]Tile{
		MakeTile(tile.Color(), p1),
		MakeTile(tile.Color(), p2),
	}
	for _, reqTile := range requiredTiles {
		if !slices.Contains(h.Tiles, reqTile) {
			return false
		}
	}

	tiles := RemoveElements(h.Tiles, tile, 1)
	tiles = RemoveElements(tiles, requiredTiles[0], 1)
	tiles = RemoveElements(tiles, requiredTiles[1], 1)
	tiles, laiCount := h.CountLaiZi(tiles)
	return DefaultHuCore.CheckBasicHu(tiles, laiCount) != HU_NON
}
