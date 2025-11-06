package mahjong

import (
	"fmt"
	"maps"
	"math/rand"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 对应 YAML 根结构
type configFile struct {
	Enable bool     `yaml:"enable"`
	Cards  []string `yaml:"cards"`
}

// Manual 对应 C++ Manual
type Manual struct {
	InitCardFile string
}

// newManual 构造函数
func newManual(name string, matchId int32) *Manual {
	return &Manual{
		InitCardFile: filepath.Join(".", "initCard", fmt.Sprintf("%s_%d.yaml", name, matchId)),
	}
}

func (m *Manual) enabled() bool {
	var c configFile
	if err := m.loadFile(&c); err != nil {
		return false
	}
	return c.Enable
}

func (m *Manual) load(tiles map[Tile]int, playerCount, handCount int) ([]Tile, error) {
	var c configFile
	if err := m.loadFile(&c); err != nil {
		return nil, err
	}
	// 1. 解析牌组
	groups := make([][]Tile, playerCount+2)
	fields := []string{}
	for i := 0; i < playerCount+2; i++ {
		groups[i] = namesToTiles(fields[i])
	}

	// 2. 校验牌池充足
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

	// 3. 分离花牌与普通剩余牌
	var rests []Tile
	for t, count := range tmp {
		if count > 0 {
			rests = append(rests, makeTiles(t, count)...)
		}
	}

	m.shuffle(rests)

	// 4. 补满各家到指定张数
	var out []Tile
	for i := range len(groups) {
		out = append(out, groups[i]...)
		more := handCount - len(groups[i])
		if i < playerCount {
			out = append(out, rests[:more]...)
			rests = rests[more:]
		}
	}

	out = append(out, rests...) // 剩余
	return out, nil
}

func (m *Manual) loadFile(cfg *configFile) error {
	f, err := os.Open(m.InitCardFile)
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewDecoder(f).Decode(cfg)
}

func (m *Manual) shuffle(s []Tile) {
	for i := len(s) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		s[i], s[j] = s[j], s[i]
	}
}
