package mahjong

import (
	"errors"
	"math"
	"slices"
)

// 分数计算器
type scorelator struct {
	game      *Game
	scoreType ScoreType
}

func NewScorelator(g *Game, scoreType ScoreType) *scorelator {
	return &scorelator{
		game:      g,
		scoreType: scoreType,
	}
}

func (s *scorelator) calculate(takeScores, winScores []int64) (res []int64) {
	res = slices.Clone(winScores)
	switch s.scoreType {
	case ScoreTypeMinScore:
		res, _ = s.calcMinScore(takeScores, winScores)
	case ScoreTypePositive:
		for i := 0; i < len(res); i++ {
			if takeScores[i] < 0 {
				res[i] = 0
			} else if winScores[i]+takeScores[i] < 0 {
				res[i] = -takeScores[i]
			}
		}
	case ScoreTypeJustWin:
		for i := range res {
			if res[i] < 0 {
				res[i] = 0
			}
		}
	}
	return
}

func (s *scorelator) calcMinScore(takeScores, winScores []int64) ([]int64, error) {
	return s.minScore(takeScores, winScores)
}

func (s *scorelator) minScore(takeScores, winScores []int64) ([]int64, error) {
	if err := s.checkArgs(takeScores, winScores); err != nil {
		return winScores, errors.New("param error")
	}
	count := len(takeScores)
	var winners, losers []int
	res := make([]int64, count)

	// 分离赢家/输家索引
	for i := range count {
		if winScores[i] == 0 || takeScores[i] == 0 {
			continue
		}
		if winScores[i] > 0 {
			winners = append(winners, i)
		} else {
			losers = append(losers, i)
		}
	}

	if len(losers) == 0 || len(winners) == 0 {
		return res, nil
	}

	// 计算每个赢家对每个输家最多可赢
	var winAll int64
	for _, i := range winners {
		for _, k := range losers {
			temp := slices.Min([]int64{takeScores[i], winScores[i], takeScores[k], -winScores[k]})
			res[i] += temp
			res[k] += temp // 暂存输家最多输分
		}
		// 赢家最终最多可赢
		res[i] = slices.Min([]int64{res[i], winScores[i]})
		winAll += res[i]
	}

	// 输家总输
	var loseAll int64
	for _, k := range losers {
		res[k] = slices.Min([]int64{res[k], takeScores[k], -winScores[k]})
		loseAll += res[k]
	}

	minAll := slices.Min([]int64{loseAll, winAll})
	winRate := float64(minAll) / float64(winAll)
	loseRate := float64(minAll) / float64(loseAll)

	// 缩放赢家
	for _, i := range winners {
		res[i] = int64(math.Round(winRate * float64(res[i])))
	}
	// 缩放输家并取负
	for _, k := range losers {
		res[k] = -int64(math.Round(loseRate * float64(res[k])))
	}

	return res, nil
}

func (s *scorelator) checkArgs(takeScores, winScores []int64) error {
	if len(takeScores) != len(winScores) {
		return errors.New("param error")
	}

	for _, v := range takeScores {
		if v < 0 {
			return errors.New("takeScores must >= 0")
		}
	}
	var sumWin int64
	for _, v := range winScores {
		sumWin += v
	}
	if sumWin != 0 {
		return errors.New("winScores must all 0")
	}
	return nil
}
