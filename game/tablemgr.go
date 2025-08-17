package game

import (
	"strconv"
	"sync"

	"github.com/kevin-chtw/tw_common/utils"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
)

// TableManager 管理游戏桌
type TableManager struct {
	mu     sync.RWMutex
	tables map[string]*Table // tableID -> Table
	app    pitaya.Pitaya
}

// NewTableManager 创建游戏桌管理器
func NewTableManager(app pitaya.Pitaya) *TableManager {
	return &TableManager{
		tables: make(map[string]*Table),
		app:    app,
	}
}

// GetTable 获取指定比赛和桌号的游戏桌
func (tm *TableManager) Get(matchID, tableID int32) *Table {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	key := strconv.FormatInt(int64(matchID), 10) + ":" + strconv.FormatInt(int64(tableID), 10)
	return tm.tables[key]
}

// LoadOrStore 加载或存储游戏桌
func (tm *TableManager) LoadOrStore(matchID, tableID int32) *Table {
	key := strconv.FormatInt(int64(matchID), 10) + ":" + strconv.FormatInt(int64(tableID), 10)

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 检查是否已存在
	if table, ok := tm.tables[key]; ok {
		return table
	}

	// 创建新表
	table := NewTable(utils.GameID_MahjongSC, matchID, tableID, tm.app)
	tm.tables[key] = table
	return table
}

// Delete 删除指定比赛和桌号的游戏桌
func (tm *TableManager) Delete(matchID, tableID int32) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	key := strconv.FormatInt(int64(matchID), 10) + ":" + strconv.FormatInt(int64(tableID), 10)
	delete(tm.tables, key)
}
