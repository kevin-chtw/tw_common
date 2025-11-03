package mahjong

import (
	"strconv"
	"strings"
)

var (
	TileNull   Tile = -1
	TileHun    Tile = MakeTile(ColorHun, 0)    // 混子
	TileInf    Tile = MakeTile(ColorEnd, 0)    // 无效牌
	TileZhong  Tile = MakeTile(ColorDragon, 0) // 中
	TileFa     Tile = MakeTile(ColorDragon, 1) // 发
	TileBai    Tile = MakeTile(ColorDragon, 2) // 白
	TileDong   Tile = MakeTile(ColorWind, 0)   // 东
	TileNan    Tile = MakeTile(ColorWind, 1)   // 南
	TileXi     Tile = MakeTile(ColorWind, 2)   // 西
	TileBei    Tile = MakeTile(ColorWind, 3)   // 北
	TileYaoJi  Tile = MakeTile(ColorBamboo, 0) // 幺鸡
	TileFlower Tile = MakeTile(ColorFlower, 0) // 花
	TileSpring Tile = MakeTile(ColorSeason, 0) // 春
)

type Tile int32

func MakeTile(color EColor, point int) Tile {
	return Tile((int(color)<<8 | (point << 4) | 1))
}

func MakeSpecialTile(color EColor, point int, flag int) Tile {
	return Tile((int(color)<<8 | (point << 4) | flag))
}

func (t Tile) Color() EColor {
	return EColor((t >> 8) & 0x0F)
}

func (t Tile) Point() int {
	return int((t >> 4) & 0x0F)
}

func (t Tile) Info() (EColor, int) {
	return t.Color(), t.Point()
}

func (t Tile) Flag() int {
	return int(t & 0x0F)
}

func (t Tile) IsValid() bool {
	return t > 0 && t < TileInf
}

func (t Tile) IsSuit() bool { // 数牌
	return t.IsValid() && t.Color() >= ColorCharacter && t.Color() <= ColorDot
}
func (t Tile) IsHonor() bool { // 字牌
	return t.IsValid() && t.Color() == ColorWind || t.Color() == ColorDragon
}

func (t Tile) Is258() bool { // 258牌
	return t.IsValid() && t.IsSuit() && (t.Point()%3 == 1)
}

func (t Tile) IsDragon() bool { // 箭牌
	return t.Color() == ColorDragon
}

func (t Tile) IsExtra() bool { // 花牌+季牌
	return t.IsValid() && t.Color() == ColorFlower || t.Color() == ColorSeason
}

func (t Tile) Name() string {
	c, p := t.Info()
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

func (t Tile) ToInt32() int32 {
	return int32(t)
}

func TilesName(tiles []Tile) string {
	var tileNames []string
	for _, tile := range tiles {
		tileNames = append(tileNames, tile.Name())
	}
	return strings.Join(tileNames, ", ")
}

func GetColorTile(tiles []Tile, color EColor) Tile {
	for _, t := range tiles {
		if t.Color() == color {
			return t
		}
	}
	return TileNull
}

func TilesInt32(tiles []Tile) []int32 {
	res := make([]int32, len(tiles))
	for i, t := range tiles {
		res[i] = int32(t)
	}
	return res
}

func Int32Tile(tiles []int32) []Tile {
	res := make([]Tile, len(tiles))
	for i, t := range tiles {
		res[i] = Tile(t)
	}
	return res
}
