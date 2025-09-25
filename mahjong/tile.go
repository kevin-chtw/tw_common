package mahjong

import (
	"strconv"
	"strings"
)

func GetTilesName(tiles []int32) string {
	var tileNames []string
	for _, tile := range tiles {
		tileNames = append(tileNames, GetTileName(tile))
	}
	return strings.Join(tileNames, ", ")
}

func GetTileName(tile int32) string {
	c := TileColor(tile)
	p := TilePoint(tile)
	switch c {
	case ColorCharacter:
		return strconv.Itoa(p+1) + "万"
	case ColorBamboo:
		return strconv.Itoa(p+1) + "条"
	case ColorDot:
		return strconv.Itoa(p+1) + "筒"
	case ColorWind:
		names := []string{"东", "南", "西", "北"}
		return names[p]
	case ColorDragon:
		names := []string{"中", "发", "白"}
		return names[p]
	case ColorFlower:
		names := []string{"梅", "兰", "竹", "菊"}
		return names[p]
	case ColorSeason:
		names := []string{"春", "夏", "秋", "冬"}
		return names[p]
	default:
		return ""
	}
}
