package game

import (
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
)

const (
	SeatAll int32 = -2
)

var playerManager *PlayerManager
var tableManager *TableManager

type NewGame func(*Table, int32) IGame

var fn = make(map[string]NewGame)

func Register(serverType string, f NewGame) {
	fn[serverType] = f
}

func CreateGame(serverType string, t *Table, id int32) IGame {
	if f, ok := fn[serverType]; ok {
		return f(t, id)
	}
	return nil
}

// InitGame 初始化游戏模块
func InitGame(app pitaya.Pitaya) {
	playerManager = NewPlayerManager()
	tableManager = NewTableManager(app)
}

// GetPlayerManager 获取玩家管理器实例
func GetPlayerManager() *PlayerManager {
	return playerManager
}

// GetTableManager 获取游戏桌管理器实例
func GetTableManager() *TableManager {
	return tableManager
}
