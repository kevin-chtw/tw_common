package game

import (
	"runtime/debug"
	"sync"
	"time"

	"github.com/kevin-chtw/tw_proto/cproto"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"google.golang.org/protobuf/proto"
)

// BotManager 全局管理所有的bot玩家
type BotManager struct {
	bots       map[string]*BotPlayer // bot ID -> BotPlayer
	msgInChan  chan *BotMessage      // 接收消息通道
	msgOutChan chan *BotMessage      // 发送消息通道
	mu         sync.RWMutex
	ticker     *time.Ticker
}

// BotMessage 定义bot消息结构
type BotMessage struct {
	BotID string
	Msg   interface{}
}

// GetBotManager 获取BotManager单例
func NewBotManager() *BotManager {
	bm := &BotManager{
		bots:       make(map[string]*BotPlayer),
		msgInChan:  make(chan *BotMessage, 100),
		msgOutChan: make(chan *BotMessage, 100),
		mu:         sync.RWMutex{},
		ticker:     time.NewTicker(time.Second),
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Log.Errorf("panic recovered %s\n %s", r, string(debug.Stack()))
			}
		}()
		for range bm.ticker.C {
			bm.tick()
		}
	}()
	go bm.processIncomingMessages()
	go bm.processOutgoingMessages()
	return bm
}

func (b *BotManager) tick() {
	b.mu.RLock()
	bots := make([]*BotPlayer, 0, len(b.bots))
	for _, bot := range b.bots {
		bots = append(bots, bot)
	}
	b.mu.RUnlock()

	for _, bot := range bots {
		bot.Bot.OnTimer()
	}
}

// AddBot 添加一个新的bot玩家
func (m *BotManager) AddBot(uid string, matchid, tableid int32) *BotPlayer {
	bot := botCreator(uid, matchid, tableid)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bots[bot.Uid] = bot
	return bot
}

// RemoveBot 移除一个bot玩家
func (m *BotManager) RemoveBot(botID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.bots, botID)
}

// GetBot 获取指定bot玩家
func (m *BotManager) GetBot(botID string) *BotPlayer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.bots[botID]
}

// processIncomingMessages 处理接收到的消息
func (m *BotManager) processIncomingMessages() {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Errorf("panic recovered %s\n %s", r, string(debug.Stack()))
		}
	}()
	for msg := range m.msgInChan {
		bot := m.GetBot(msg.BotID)
		ack := msg.Msg.(*cproto.GameAck)
		if bot == nil {
			bot = m.AddBot(msg.BotID, ack.Matchid, ack.Tableid)
		}
		bot.OnBotMsg(ack)
	}
}

// processOutgoingMessages 处理要发送的消息
func (m *BotManager) processOutgoingMessages() {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Errorf("panic recovered %s\n %s", r, string(debug.Stack()))
		}
	}()
	for msg := range m.msgOutChan {
		tableManager.OnBotMsg(msg.BotID, msg.Msg.(proto.Message))
	}
}

// SendToTable 发送消息到指定table
func (m *BotManager) SendToTable(botID string, msg proto.Message) {
	m.msgOutChan <- &BotMessage{
		BotID: botID,
		Msg:   msg,
	}
}

// OnBotMessage 处理接收到的bot消息
func (m *BotManager) OnBotMessage(botID string, msg *cproto.GameAck) {
	m.msgInChan <- &BotMessage{
		BotID: botID,
		Msg:   msg,
	}
}
