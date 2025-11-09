package game

import (
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
)

const (
	SeatAll int32 = -2
)

var (
	gameCreator   GameCreator
	botCreator    BotCreator
	playerManager *PlayerManager
	tableManager  *TableManager
	botManager    *BotManager
)

type GameCreator func(*Table, int32) IGame
type BotCreator func(uid string) *BotPlayer

// Init 初始化游戏模块
func Init(app pitaya.Pitaya, gc GameCreator, bc BotCreator) {
	gameCreator = gc
	botCreator = bc
	playerManager = NewPlayerManager()
	tableManager = NewTableManager(app)
	botManager = NewBotManager()
}

// GetPlayerManager 获取玩家管理器实例
func GetPlayerManager() *PlayerManager {
	return playerManager
}

// GetTableManager 获取游戏桌管理器实例
func GetTableManager() *TableManager {
	return tableManager
}
