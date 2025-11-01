package mahjong

type ScoreNode struct {
	scoreReason ScoreReason
	Scores      []int64
}

type ScorelatorMany struct {
	scorelator
	scores []*ScoreNode
}

func NewScorelatorMany(g *Game, scoreType ScoreType) *ScorelatorMany {
	return &ScorelatorMany{
		scorelator: *NewScorelator(g, scoreType),
		scores:     make([]*ScoreNode, 0),
	}
}

func (s *ScorelatorMany) CalcMulti(sr ScoreReason, multi []int64) []int64 {
	base := s.game.GetScoreBase()
	takeScores := make([]int64, 0)
	winScores := make([]int64, 0)
	for i, p := range s.game.players {
		takescore := p.GetCurScore()
		if multi[i] > 0 {
			takescore += p.GetTax()
		}
		takeScores = append(takeScores, takescore)
		winScores = append(winScores, multi[i]*base)
	}
	return s.calc(sr, takeScores, winScores)
}

func (s *ScorelatorMany) CalcScores(sr ScoreReason, scores []int64) []int64 {
	takeScores := make([]int64, 0)
	for i, p := range s.game.players {
		takescore := p.GetCurScore()
		if scores[i] > 0 {
			takescore += p.GetTax()
		}
		takeScores = append(takeScores, takescore)
	}
	return s.calc(sr, takeScores, scores)
}

func (s *ScorelatorMany) CalcKon(sr ScoreReason, win, loss int32, lossMulti, otherMulti int64) []int64 {
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
	return s.CalcMulti(sr, multi)
}

func (s *ScorelatorMany) RemoveLastScore() *ScoreNode {
	if len(s.scores) == 0 {
		return nil
	}
	sn := s.scores[len(s.scores)-1]
	s.scores = s.scores[:len(s.scores)-1]
	return sn
}

func (s *ScorelatorMany) addScores(sr ScoreReason, scores []int64) {
	sn := &ScoreNode{
		scoreReason: sr,
		Scores:      scores,
	}
	s.scores = append(s.scores, sn)
}

func (s *ScorelatorMany) calc(sr ScoreReason, takeScores, winScores []int64) []int64 {
	final := s.calculate(takeScores, winScores)
	for i, p := range s.game.players {
		p.AddScoreChange(final[i])
	}
	s.addScores(sr, final)
	return final
}
