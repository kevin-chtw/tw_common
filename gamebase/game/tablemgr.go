package game

import (
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/kevin-chtw/tw_proto/cproto"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"google.golang.org/protobuf/proto"
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
		defer func() {
			if r := recover(); r != nil {
				logger.Log.Errorf("panic recovered %s\n %s", r, string(debug.Stack()))
			}
		}()
		for range t.ticker.C {
			t.tick()
		}
	}()

	return t
}

func (t *TableManager) tick() {
	t.mu.RLock()
	tables := make([]*Table, 0, len(t.tables))
	for _, table := range t.tables {
		tables = append(tables, table)
	}
	t.mu.RUnlock()

	for _, table := range tables {
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
func (t *TableManager) LoadOrStore(matchId, tableId int32) *Table {
	key := getTableKey(matchId, tableId)

	t.mu.Lock()
	defer t.mu.Unlock()

	// 检查是否已存在
	if table, ok := t.tables[key]; ok {
		return table
	}

	// 创建新表
	table := NewTable(matchId, tableId, t.app)
	t.tables[key] = table
	return table
}

func (t *TableManager) OnBotMsg(uid string, msg proto.Message) {
	req := msg.(*cproto.GameReq)
	table := t.Get(req.Matchid, req.Tableid)
	player := playerManager.Get(uid)
	if table != nil && player != nil {
		table.OnPlayerMsg(player, req)
	}
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
