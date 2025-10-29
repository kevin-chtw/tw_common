package mahjong

type ScorelatorMany struct {
	scorelator
}

func NewScorelatorMany(g *Game, scoreType ScoreType) *ScorelatorMany {
	return &ScorelatorMany{
		scorelator: *NewScorelator(g, scoreType),
	}
}

func (s *ScorelatorMany) Calculate(multi []int64) []int64 {
	base := s.game.GetScoreBase()
	takeScores := make([]int64, 0)
	originScores := make([]int64, 0)
	for i, p := range s.game.players {
		takescore := p.GetCurScore()
		if multi[i] > 0 {
			takescore += p.GetTax()
		}
		takeScores = append(takeScores, takescore)
		originScores = append(originScores, multi[i]*base)
	}

	final := s.calculate(takeScores, originScores)
	for i, p := range s.game.players {
		p.AddScoreChange(final[i])
	}
	return final
}

func (s *ScorelatorMany) Check(win, loss int32, lossMulti, otherMulti int64) []int64 {
	multi := make([]int64, s.game.GetPlayerCount())
	for i := range s.game.GetPlayerCount() {
		if i == loss {
			multi[i] = lossMulti
			multi[win] -= lossMulti
			continue
		}
		if !s.game.GetPlayer(i).isOut {
			multi[i] = otherMulti
			multi[win] -= otherMulti
		}
	}
	return s.Calculate(multi)
}
