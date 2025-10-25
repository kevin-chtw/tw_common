package mahjong

const (
	OperateNone     = 0               // 无操作
	OperatePass     = 1 << (iota - 1) // 过  1<<0 = 1
	OperateChow                       // 吃  1<<1 = 2
	OperatePon                        // 碰  1<<2 = 4
	OperateKon                        // 杠  1<<3 = 8
	OperateTing                       // 听  1<<4 = 16
	OperateHu                         // 胡  1<<5 = 32
	OperateDiscard                    // 出牌  1<<6 = 64
	OperateExchange                   // 换牌  1<<7 = 128
	OperateDraw                       // 摸牌  1<<8 = 256
	OperateTianTing                   // 天听  1<<9 = 512
	OperateFlower                     // 换花  1<<10 = 1024
	OperateChowTing                   // 吃听  1<<11 = 2048
	OperatePonTing                    // 碰听  1<<12 = 4096
)

var OperateNames = map[int]string{
	OperatePass:     "Pass",
	OperateChow:     "Chow",
	OperatePon:      "Pon",
	OperateKon:      "Kon",
	OperateTing:     "Ting",
	OperateHu:       "Win",
	OperateDiscard:  "Discard",
	OperateExchange: "Exchange",
	OperateDraw:     "Draw",
	OperateTianTing: "TianTing",
	OperateFlower:   "Flower",
	OperateChowTing: "ChowTing",
	OperatePonTing:  "PonTing",
}

var OperateIDs = map[string]int{
	"Pass":     OperatePass,
	"Chow":     OperateChow,
	"Pon":      OperatePon,
	"Kon":      OperateKon,
	"Ting":     OperateTing,
	"Win":      OperateHu,
	"Discard":  OperateDiscard,
	"Exchange": OperateExchange,
	"Draw":     OperateDraw,
	"TianTing": OperateTianTing,
	"Flower":   OperateFlower,
	"ChowTing": OperateChowTing,
	"PonTing":  OperatePonTing,
}

type Operates struct {
	Value    int32
	IsMustHu bool
	Capped   bool
}

func (o *Operates) AddOperate(op int32) {
	o.Value |= op
}

func (o *Operates) AddOperates(ops *Operates) {
	o.Value |= ops.Value
}

func (o *Operates) RemoveOperate(op int32) {
	o.Value &= ^op
}

func (o *Operates) HasOperate(op int32) bool {
	return (o.Value & op) != 0
}

func (o *Operates) Reset() {
	o.Value = 0
}

func GetOperateName(operate int, names map[int]string) string {
	if name, ok := names[operate]; ok {
		return name
	}
	return ""
}

func GetOperateID(name string, ids map[string]int) int {
	if id, ok := ids[name]; ok {
		return id
	}
	return OperateNone
}
