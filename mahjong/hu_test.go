package mahjong_test

import (
	"testing"

	"github.com/kevin-chtw/tw_common/mahjong"
)

func TestCheckBasicHu(t *testing.T) {
	// 初始化HuCore
	hc := mahjong.NewHuCore(14) // 使用更大的手牌数限制
	if hc == nil {
		t.Fatal("Failed to create HuCore")
	}

	testCases := []struct {
		name  string
		cards []int32
		laiZi int
		want  bool
	}{
		{
			name: "Normal Hu",
			cards: []int32{
				mahjong.MakeTile(mahjong.ColorCharacter, 1), mahjong.MakeTile(mahjong.ColorCharacter, 1),
				mahjong.MakeTile(mahjong.ColorCharacter, 2), mahjong.MakeTile(mahjong.ColorCharacter, 2),
				mahjong.MakeTile(mahjong.ColorCharacter, 3), mahjong.MakeTile(mahjong.ColorCharacter, 3),
				mahjong.MakeTile(mahjong.ColorCharacter, 4), mahjong.MakeTile(mahjong.ColorCharacter, 4),
				mahjong.MakeTile(mahjong.ColorCharacter, 5), mahjong.MakeTile(mahjong.ColorCharacter, 5),
				mahjong.MakeTile(mahjong.ColorCharacter, 6), mahjong.MakeTile(mahjong.ColorCharacter, 6),
				mahjong.MakeTile(mahjong.ColorCharacter, 7), mahjong.MakeTile(mahjong.ColorCharacter, 7),
			},
			laiZi: 0,
			want:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := hc.CheckBasicHu(tc.cards, tc.laiZi)
			if got != tc.want {
				t.Errorf("CheckBasicHu(%v, %d) = %v, want %v",
					tc.cards, tc.laiZi, got, tc.want)
			}
		})
	}
}
