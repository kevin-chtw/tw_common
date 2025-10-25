package mahjong

import (
	"time"
)

const (
	RecordPBFlag = 512 // f=<0:begin,1:end>
)

type RecordItem struct {
	ActionID   int
	ActionName string
	Properties map[string]string
}

func NewRecordItem(actionID int) *RecordItem {
	return &RecordItem{
		ActionID:   actionID,
		Properties: make(map[string]string),
	}
}

func NewRecordItemByName(actionName string) *RecordItem {
	return &RecordItem{
		ActionName: actionName,
		Properties: make(map[string]string),
	}
}

func (i *RecordItem) ResetAction(actionID int) {
	i.ActionID = actionID
	i.ActionName = ""
	i.Properties = make(map[string]string)
}

func (i *RecordItem) ResetActionByName(actionName string) {
	i.ActionID = 0
	i.ActionName = actionName
	i.Properties = make(map[string]string)
}

func (i *RecordItem) AddProperty(name string, value interface{}) {
	i.Properties[name] = ToString(value)
}

type Record struct {
	startTime   time.Time
	currentItem *RecordItem
	hasPBRecord bool
	game        *Game
}

func NewRecord(game *Game) *Record {
	return &Record{
		game:        game,
		startTime:   time.Now(),
		currentItem: NewRecordItem(0),
	}
}

func (r *Record) Initialize() {
	r.startTime = time.Now()
	r.currentItem = NewRecordItem(0)
	r.hasPBRecord = false
}

func (r *Record) Finish() {
	// 游戏结束时调用
}

func (r *Record) RecordAction(ack interface{}, seat int32) {
	// 记录动作
}

func (r *Record) AddPlayerNetBreak(seat int32) {
	// 记录玩家断网
}

func (r *Record) AddPlayerNetResume(seat int32) {
	// 记录玩家恢复网络
}

func (r *Record) AddPlayerExitMatch(seat int32) {
	// 记录玩家退出比赛
}

func (r *Record) HasPBRecord() bool {
	return r.hasPBRecord
}

func (r *Record) AddItem(item *RecordItem) {
	// 添加记录项
}

func (r *Record) RecordLog(str string) {
	// 记录日志
}

func (r *Record) RecordPBBegin() {
	// 记录PB开始
	r.hasPBRecord = true
}

func (r *Record) RecordPBAction(ack interface{}, seat int32) {
	// 记录PB动作
}

func (r *Record) RecordPBResult(winners []int32) {
	// 记录PB结果
}

func (r *Record) RecordPBEnd() {
	// 记录PB结束
}
