package mahjong

import (
	"time"

	"google.golang.org/protobuf/proto"
)

type StateOpts struct {
	// 自定义选项
}

type IState interface {
	OnEnter()
	OnPlayerMsg(seat int32, req proto.Message)
}

func CreateState(newFn func(IGame, ...any) IState, g IGame, args ...any) IState {
	return newFn(g, args)
}

// State 麻将游戏状态基类
type State struct {
	game       *Game
	msgHandler func(seat int32, req proto.Message)
}

// NewState 创建新的游戏状态
func NewState(game *Game) *State {
	return &State{
		game:       game,
		msgHandler: nil,
	}
}

// AsyncMsgTimer 设置异步消息定时器
func (s *State) AsyncMsgTimer(
	handler func(seat int32, req proto.Message),
	timeout time.Duration,
	onTimeout func(),
) {
	s.msgHandler = handler
	s.game.GetTimer().Schedule(timeout, onTimeout)
}

// AsyncTimer 设置异步定时器
func (s *State) AsyncTimer(timeout time.Duration, onTimeout func()) {
	s.msgHandler = nil
	s.game.GetTimer().Schedule(timeout, onTimeout)
}

// HandlePlayerMsg 处理玩家消息
func (s *State) OnPlayerMsg(seat int32, req proto.Message) {
	if s.msgHandler != nil {
		s.msgHandler(seat, req)
	}
}
