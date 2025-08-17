package mahjong

var Service IService

type IService interface {
	GetAllTiles(conf *Rule) map[int32]int
	GetHandCount() int
	GetDefaultRules() []int
	CheckHu(data *HuData, rule *Rule) (*HuResult, bool)
	CheckCall(data *HuData, rule *Rule) map[int32]map[int32]int64
}
