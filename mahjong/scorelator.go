package mahjong

// 分数计算器
type Scorelator struct {
	game *Game
}

func NewScroelator(g *Game) *Scorelator {
	return &Scorelator{
		game: g,
	}
}

func (s *Scorelator) Calculate(multiples []int64) {
	if int32(len(multiples)) != s.game.GetPlayerCount() {
		return
	}

	winScore := make([]int64, len(multiples))
	takeScores := make([]int64, len(multiples))
	for i := int32(0); i < s.game.GetPlayerCount(); i++ {
		player := s.game.GetPlayer(i)
		takeScores[i] = player.GetCurScore()
		if multiples[i] > 0 {
			takeScores[i] += player.GetTax()
		}
		winScore[i] = multiples[i] * s.game.GetScoreBase()
	}

	result := s.calculate(takeScores, winScore)
	for i := 0; i < len(result); i++ {
		player := s.game.GetPlayer(int32(i))
		player.AddScoreChange(result[i])
	}
}

func (s *Scorelator) calculate(_, winScore []int64) []int64 {
	//TODO 其他算分方式
	return winScore
}
