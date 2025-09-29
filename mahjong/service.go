package mahjong

var Service IService

type IService interface {
	GetAllTiles(conf *Rule) map[Tile]int
	GetHandCount() int
	GetDefaultRules() []int
	CheckHu(data *HuData, rule *Rule) (*HuResult, bool)
	CheckCall(data *HuData, rule *Rule) map[Tile]map[Tile]int64
}
