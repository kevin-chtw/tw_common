package mahjong

import (
	"time"
)

const (
	TimerIntervalMS = 250
	GameTimeoutMS   = 120 * 1000
)

// Timer 麻将游戏定时器
type Timer struct {
	triggerTime time.Time
	delay       time.Duration
	callback    func()
	isLongLive  bool
}

// NewTimer 创建新的定时器实例
func NewTimer() *Timer {
	return &Timer{}
}

// Schedule 安排定时任务
// delay: 延迟时间
// callback: 回调函数
func (t *Timer) Schedule(delay time.Duration, callback func()) {
	t.triggerTime = time.Now().Add(delay)
	t.delay = delay
	t.callback = callback
	t.isLongLive = false
}

// Cancel 取消定时任务
func (t *Timer) Cancel() {
	t.callback = nil
	t.isLongLive = false
}

// SetLongLive 设置定时器为长期存活
// infinite: 是否长期存活
func (t *Timer) SetLongLive(infinite bool) {
	t.isLongLive = infinite
}

// OnTick 定时器触发时的处理
func (t *Timer) OnTick() {
	if t.callback == nil {
		return
	}

	if time.Now().After(t.triggerTime) {
		t.callback()
		if !t.isLongLive {
			t.callback = nil
		} else {
			// 长期存活的定时器需要重新设置触发时间
			t.triggerTime = time.Now().Add(t.delay)
		}
	}
}
