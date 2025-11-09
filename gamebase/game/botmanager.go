package game

import (
	"sync"

	"github.com/kevin-chtw/tw_proto/cproto"
	"google.golang.org/protobuf/proto"
)

// BotManager 全局管理所有的bot玩家
type BotManager struct {
	bots       map[string]*BotPlayer // bot ID -> BotPlayer
	msgInChan  chan *BotMessage      // 接收消息通道
	msgOutChan chan *BotMessage      // 发送消息通道
	mu         sync.RWMutex
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
	}
	go bm.processIncomingMessages()
	go bm.processOutgoingMessages()
	return bm
}

// AddBot 添加一个新的bot玩家
func (m *BotManager) AddBot(uid string) *BotPlayer {
	bot := botCreator(uid)
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
	for msg := range m.msgInChan {
		bot := m.GetBot(msg.BotID)
		if bot == nil {
			bot = m.AddBot(msg.BotID)
		}
		bot.OnBotMsg(msg.Msg.(proto.Message))
	}
}

// processOutgoingMessages 处理要发送的消息
func (m *BotManager) processOutgoingMessages() {
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
