package game

import (
	"errors"

	"github.com/kevin-chtw/tw_proto/cproto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type IBot interface {
	OnBotMsg(msg proto.Message) error
	OnTimer() error
}

type PendingReq struct {
	Req   proto.Message // 待发送的请求
	Delay int           // 剩余延迟(ms)
}

// BotPlayer 表示游戏中的bot玩家
type BotPlayer struct {
	Bot     IBot
	Uid     string
	matchid int32
	tableid int32
	Seat    int32
}

// NewBotPlayer 创建新的bot玩家实例
func NewBotPlayer(uid string, matchid, tableid int32) *BotPlayer {
	b := &BotPlayer{
		Uid:     uid,
		matchid: matchid,
		tableid: tableid,
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

func (b *BotPlayer) SendMsg(msg proto.Message) error {
	pbData, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	pbTableMsg := &cproto.TableMsgReq{
		Msg: pbData,
	}
	data, err := anypb.New(pbTableMsg)
	if err != nil {
		return err
	}
	req := &cproto.GameReq{
		Tableid: b.tableid,
		Matchid: b.matchid,
		Req:     data,
	}
	botManager.SendToTable(b.Uid, req)
	return nil
}
