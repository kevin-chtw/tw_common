package mahjong_test

import (
	"slices"
	"strconv"
	"testing"

	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
)

type Case struct {
	cards []mahjong.Tile
	laiZi int
	want  mahjong.HuCoreType
}

func Test_Hu(t *testing.T) {
	// 初始化HuCore
	hc := mahjong.NewHuCore(14) // 使用更大的手牌数限制
	if hc == nil {
		t.Fatal("Failed to create HuCore")
	}

	testCases := []Case{
		{
			cards: []mahjong.Tile{17, 17, 33, 33, 49, 49, 65, 65, 81, 81, 97, 97, 113, 113},
			laiZi: 0,
			want:  mahjong.HU_PIN,
		},
		{
			cards: []mahjong.Tile{17, 17, 33, 49, 65, 65, 65, 81, 81, 97, 97},
			laiZi: 0,
			want:  mahjong.HU_PIN,
		},
		{
			cards: []mahjong.Tile{625, 97, 513, 577, 513, 273, 529, 561, 257, 273, 609, 641, 625, 113},
			laiZi: 0,
			want:  mahjong.HU_NON,
		},
	}

	// 测试每个案例
	for i, tc := range testCases {
		t.Run("case"+strconv.FormatInt(int64(i), 10), func(t *testing.T) {
			slices.Sort(tc.cards)
			t.Log(mahjong.TilesName(tc.cards))
			got := hc.CheckBasicHu(tc.cards, tc.laiZi)
			if got != tc.want {
				t.Errorf("CheckBasicHu(%v, %d) = %v, want %v", tc.cards, tc.laiZi, got, tc.want)
			}
		})
	}
}
