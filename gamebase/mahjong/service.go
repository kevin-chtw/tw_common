package mahjong

import "github.com/kevin-chtw/tw_proto/game/pbmj"

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
	GetHuResult(data *HuData) *pbmj.MJHuData
}
