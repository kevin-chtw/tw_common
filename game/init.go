package game

import (
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
)

const (
	SeatAll int32 = -2
)

var playerManager *PlayerManager
var tableManager *TableManager

type NewGame func(*Table) IGame

var reg = make(map[int32]NewGame)

func Register(id int32, f NewGame) {
	reg[id] = f
}

func CreateGame(id int32, t *Table) IGame {
	if f, ok := reg[id]; ok {
		return f(t)
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
