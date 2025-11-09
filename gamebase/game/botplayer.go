package game

import (
	"errors"

	"google.golang.org/protobuf/proto"
)

type IBot interface {
	OnBotMsg(msg proto.Message) error
}

// BotPlayer 表示游戏中的bot玩家
type BotPlayer struct {
	Bot IBot
	Uid string
}

// NewBotPlayer 创建新的bot玩家实例
func NewBotPlayer(uid string) *BotPlayer {
	b := &BotPlayer{
		Uid: uid,
	}
	return b
}

// OnBotMsg 处理bot收到的消息
func (b *BotPlayer) OnBotMsg(msg proto.Message) error {
	if b.Bot == nil {
		return errors.New("bot player has no bot")
	}
	return b.Bot.OnBotMsg(msg)
}
