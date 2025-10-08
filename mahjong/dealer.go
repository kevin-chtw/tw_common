package mahjong

import (
	"math/rand"
	"slices"
)

// Dealer 麻将发牌器接口
type Dealer struct {
	game     *Game
	tileWall []Tile
}

// NewDealer 创建新的发牌器
func NewDealer(game *Game) *Dealer {
	return &Dealer{
		game:     game,
		tileWall: make([]Tile, 0),
	}
}

// GetGame 获取关联的Game对象
func (d *Dealer) GetGame() *Game {
	return d.game
}

func (d *Dealer) Initialize() {
	tiles := Service.GetAllTiles(d.game.GetRule())
	// 预计算总牌数并一次性分配
	total := 0
	for _, count := range tiles {
		total += count
	}
	d.tileWall = make([]Tile, total)

	// 填充并同时随机化牌墙
	i := 0
	for tile, count := range tiles {
		for range count {
			// 随机插入位置
			pos := rand.Intn(i + 1)
			if pos != i {
				d.tileWall[i] = d.tileWall[pos]
			}
			d.tileWall[pos] = tile
			i++
		}
	}
}

// DrawTile 抽牌
func (d *Dealer) DrawTile() Tile {
	if len(d.tileWall) == 0 {
		return TileNull
	}
	tile := d.tileWall[0]
	d.tileWall = d.tileWall[1:]
	return tile
}
func (d *Dealer) Deal(count int) []Tile {
	tiles := make([]Tile, count)
	copy(tiles, d.tileWall[:count])
	d.tileWall = d.tileWall[count:]
	return tiles
}

// GetRestTileCount 获取剩余牌数
func (d *Dealer) GetRestCount() int32 {
	return int32(len(d.tileWall))
}

func (d *Dealer) HasTile(tile Tile) bool {
	return slices.Contains(d.tileWall, tile)
}

func (d *Dealer) LastTile() Tile {
	return d.tileWall[len(d.tileWall)-1]
}

func (d *Dealer) SwapLastTile() Tile {
	if len(d.tileWall) < 4 {
		return TileNull
	}
	i := len(d.tileWall) - 1
	for ; i >= 0; i-- {
		if d.Count(d.tileWall[i]) > 1 {
			break
		}
	}
	d.tileWall[len(d.tileWall)-1], d.tileWall[i] = d.tileWall[i], d.tileWall[len(d.tileWall)-1]
	return d.LastTile()
}

func (d *Dealer) Count(tile Tile) int {
	count := 0
	for _, t := range d.tileWall {
		if t == tile {
			count++
		}
	}
	return count
}
