package game

import (
	"strconv"
	"sync"
	"time"

	pitaya "github.com/topfreegames/pitaya/v3/pkg"
)

// TableManager 管理游戏桌
type TableManager struct {
	mu     sync.RWMutex
	tables map[string]*Table // tableID -> Table
	app    pitaya.Pitaya
	ticker *time.Ticker
}

// NewTableManager 创建游戏桌管理器
func NewTableManager(app pitaya.Pitaya) *TableManager {
	t := &TableManager{
		tables: make(map[string]*Table),
		app:    app,
		ticker: time.NewTicker(time.Second),
	}
	go func() {
		for range t.ticker.C {
			t.tick()
		}
	}()

	return t
}

func (t *TableManager) tick() {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, table := range t.tables {
		table.tick()
	}
}

// GetTable 获取指定比赛和桌号的游戏桌
func (t *TableManager) Get(matchID, tableID int32) *Table {
	t.mu.RLock()
	defer t.mu.RUnlock()
	key := getTableKey(matchID, tableID)
	return t.tables[key]
}

// LoadOrStore 加载或存储游戏桌
func (t *TableManager) LoadOrStore(gameId, matchId, tableId int32) *Table {
	key := getTableKey(matchId, tableId)

	t.mu.Lock()
	defer t.mu.Unlock()

	// 检查是否已存在
	if table, ok := t.tables[key]; ok {
		return table
	}

	// 创建新表
	table := NewTable(gameId, matchId, tableId, t.app)
	t.tables[key] = table
	return table
}

// Delete 删除指定比赛和桌号的游戏桌
func (t *TableManager) Delete(matchID, tableID int32) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.tables, getTableKey(matchID, tableID))
}

func getTableKey(matchID, tableID int32) string {
	return strconv.FormatInt(int64(matchID), 10) + ":" + strconv.FormatInt(int64(tableID), 10)
}
