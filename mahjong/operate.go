package mahjong

const (
	OperateNone = -1
	OperatePass = 0
	OperateChow = 1 << iota
	OperatePon
	OperateKon
	OperateTing
	OperateHu
	OperateDiscard
	OperateExchange
	OperateDraw
	OperateTianTing
	OperateFlower
	OperatePass10
	OperateInvalid = 0
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
	OperatePass10:   "Pass",
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
}

type Operates struct {
	Value    int32
	IsMustHu bool
	Capped   bool
}

func (o *Operates) AddOperate(op int32) {
	o.Value |= op
}

func (o *Operates) AddOperates(ops Operates) {
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
