package matchbase

type Config struct {
	GameType       string `yaml:"game_type"`
	Matchid        int32  `yaml:"matchid"`
	Name           string `yaml:"name"`
	PlayerPerTable int32  `yaml:"player_per_table"`
	SignCondition  string `yaml:"sign_condition"`
	InitialChips   int64  `yaml:"initial_chips"`
	ScoreBase      int64  `yaml:"score_base"`
	Property       string `yaml:"property"`
}
