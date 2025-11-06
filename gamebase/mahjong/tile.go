package mahjong

import (
	"strconv"
	"strings"
	"unicode/utf8"
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
	TileMei    Tile = MakeTile(ColorFlower, 0) // 梅
	TileLan    Tile = MakeTile(ColorFlower, 1) // 兰
	TileZhu    Tile = MakeTile(ColorFlower, 2) // 竹
	TileJu     Tile = MakeTile(ColorFlower, 3) // 菊
	TileSpring Tile = MakeTile(ColorSeason, 0) // 春
	TileSummer Tile = MakeTile(ColorSeason, 1) // 夏
	TileAutumn Tile = MakeTile(ColorSeason, 2) // 秋
	TileWinter Tile = MakeTile(ColorSeason, 3) // 冬
)

// 静态表
var singleTileMap = map[rune]Tile{
	// 风
	'东': TileDong,
	'南': TileNan,
	'西': TileXi,
	'北': TileBei,
	// 箭
	'中': TileZhong,
	'发': TileFa,
	'白': TileBai,
	// 花
	'梅': TileMei,
	'兰': TileLan,
	'竹': TileZhu,
	'菊': TileJu,
	// 季
	'春': TileSpring,
	'夏': TileSummer,
	'秋': TileAutumn,
	'冬': TileWinter,
}

// 静态表：最后一个 rune -> 颜色
var lastRuneToColor = map[rune]EColor{
	'万': ColorCharacter,
	'条': ColorBamboo,
	'筒': ColorDot,
}

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

func namesToTiles(names string) []Tile {
	parts := strings.Split(names, ",")
	res := make([]Tile, len(parts))
	for i, name := range parts {
		res[i] = nameToTile(name)
	}
	return res
}

func nameToTile(name string) Tile {
	if name == "" {
		return TileNull
	}

	if len(name) >= 2 {
		r, size := utf8.DecodeLastRuneInString(name)
		color, ok := lastRuneToColor[r]
		if !ok {
			return TileNull
		}
		prefix := name[:len(name)-size]
		num, err := strconv.Atoi(prefix)
		if err != nil || num < 1 || num > 9 {
			return TileNull
		}
		return MakeTile(color, num-1)
	}

	r, size := utf8.DecodeRuneInString(name)
	if size == len(name) {
		if t, ok := singleTileMap[r]; ok {
			return t
		}
	}
	return TileNull
}

func makeTiles(t Tile, count int) []Tile {
	if count <= 0 {
		return []Tile{}
	}
	res := make([]Tile, count)
	for i := range res {
		res[i] = t
	}
	return res
}
