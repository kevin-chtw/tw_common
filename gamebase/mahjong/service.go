package mahjong

const defaultHandCount = 14

var (
	Service       IService
	DefaultHuCore = NewHuCore(defaultHandCount)
)

type IService interface {
	GetAllTiles(conf *Rule) map[Tile]int
	GetHandCount() int
	GetDefaultRules() []int
	GetFdRules() map[string]int32
	GetHuTypes(data *HuData) []int32
	TotalMuti(types []int32, conf *Rule) int64
}
