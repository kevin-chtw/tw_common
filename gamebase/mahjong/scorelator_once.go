package mahjong

import "github.com/topfreegames/pitaya/v3/pkg/logger"

type ScorelatorOnce struct {
	scorelator
	multiples      map[ScoreReason][]int64 // 每种分数类型的倍数
	totalMultiples []int64                 // 总倍数
}

func NewScorelatorOnce(g *Game, scoreType ScoreType) *ScorelatorOnce {
	return &ScorelatorOnce{
		scorelator:     *NewScorelator(g, scoreType),
		multiples:      make(map[ScoreReason][]int64),
		totalMultiples: make([]int64, g.GetPlayerCount()),
	}
}

func (s *ScorelatorOnce) AddMultiple(t ScoreReason, multiple []int64) {
	if len(multiple) != int(s.game.GetPlayerCount()) {
		logger.Log.Errorf("multiple size not equal player count, type: %d", t)
		return
	}

	if ref, ok := s.multiples[t]; ok {
		for i, v := range ref {
			v += multiple[i]
		}
	} else {
		s.multiples[t] = multiple
	}
	for i := 0; i < int(s.game.GetPlayerCount()); i++ {
		s.totalMultiples[i] += multiple[i]
	}
}

func (s *ScorelatorOnce) Calculate() (scores []int64) {
	winScore := make([]int64, len(s.totalMultiples))
	takeScores := make([]int64, len(s.totalMultiples))
	for i := int32(0); i < s.game.GetPlayerCount(); i++ {
		player := s.game.GetPlayer(i)
		takeScores[i] = player.GetCurScore()
		if s.totalMultiples[i] > 0 {
			takeScores[i] += player.GetTax()
		}
		winScore[i] = s.totalMultiples[i] * s.game.GetScoreBase()
	}
	scores = s.calculate(takeScores, winScore)
	for i := range scores {
		player := s.game.GetPlayer(int32(i))
		player.AddScoreChange(scores[i])
	}
	return
}
