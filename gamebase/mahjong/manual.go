package mahjong

import (
	"fmt"
	"maps"
	"math/rand"
	"path/filepath"

	"github.com/spf13/viper"
)

// Manual 对应 C++ Manual
type Manual struct {
	vp *viper.Viper
}

// newManual 构造函数
func newManual(name string, matchId int32) *Manual {
	m := &Manual{
		vp: viper.New(),
	}
	m.vp.SetConfigType("yaml")
	m.vp.SetConfigFile(filepath.Join(".", "initcard", fmt.Sprintf("%s_%d.yaml", name, matchId)))
	if err := m.vp.ReadInConfig(); err != nil {
		//fmt.Fprintf(os.Stderr, "read config file error: %v\n", err)
		return nil
	}
	return m
}

func (m *Manual) enabled() bool {
	if m == nil {
		return false
	}
	return m.vp.GetBool("enable")
}

func (m *Manual) load(tiles map[Tile]int, playerCount, handCount int) ([]Tile, error) {
	cards := m.vp.GetStringSlice("cards")
	groups := make([][]Tile, len(cards))
	for i := range cards {
		groups[i] = namesToTiles(cards[i])
	}

	tmp := make(map[Tile]int)
	maps.Copy(tmp, tiles)
	for _, g := range groups {
		for _, t := range g {
			tmp[t]--
			if tmp[t] < 0 {
				return nil, fmt.Errorf("tile %d overflow", t)
			}
		}
	}

	var rests []Tile
	for t, count := range tmp {
		if count > 0 {
			rests = append(rests, MakeTiles(t, count)...)
		}
	}

	m.shuffle(rests)
	var out []Tile
	for i := range len(groups) {
		out = append(out, groups[i]...)
		more := handCount - len(groups[i])
		if i < playerCount {
			out = append(out, rests[:more]...)
			rests = rests[more:]
		}
	}
	out = append(out, rests...)
	return out, nil
}

func (m *Manual) shuffle(s []Tile) {
	for i := len(s) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		s[i], s[j] = s[j], s[i]
	}
}
