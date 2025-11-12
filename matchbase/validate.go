package matchbase

import (
	"context"
	"errors"
)

// validateOptions 验证选项配置
type validateOptions struct {
	tableID               int32
	usePlayerTable        bool
	requirePlayerInTable  bool
	checkPlayerNotInMatch bool // 检查玩家不在比赛中
	allowCreateNewPlayer  bool
}

// ValidateOption 验证选项函数类型
type ValidateOption func(*validateOptions)

// WithTableID 指定具体的桌子ID进行验证
func WithTableID(id int32) ValidateOption {
	return func(o *validateOptions) {
		o.tableID = id
	}
}

// WithPlayerTable 使用玩家当前所在的桌子
func WithPlayerTable() ValidateOption {
	return func(o *validateOptions) {
		o.usePlayerTable = true
	}
}

// WithCheckPlayerNotInMatch 检查玩家不在比赛中
func WithCheckPlayerNotInMatch() ValidateOption {
	return func(o *validateOptions) {
		o.checkPlayerNotInMatch = true
	}
}

// WithCheckPlayerNotInMatch 检查玩家不在比赛中
func WithAllowCreateNewPlayer() ValidateOption {
	return func(o *validateOptions) {
		o.allowCreateNewPlayer = true
	}
}

// RequirePlayerInTable 要求玩家必须在指定的桌子中
func RequirePlayerInTable() ValidateOption {
	return func(o *validateOptions) {
		o.requirePlayerInTable = true
	}
}

// validatePlayer 验证玩家
func (m *Match) ValidatePlayer(ctx context.Context, opts ...ValidateOption) (*Player, error) {
	options := &validateOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// 验证用户登录
	uid := m.App.GetSessionFromCtx(ctx).UID()
	if uid == "" {
		return nil, errors.New("no logged in")
	}

	// 验证玩家
	player := m.playermgr.Load(uid)
	if options.checkPlayerNotInMatch && player != nil {
		return nil, errors.New("player is in match")
	}

	if player == nil {
		if !options.allowCreateNewPlayer {
			return nil, errors.New("player not found")
		}
		player = playerCreator(ctx, uid, m.Viper.GetInt32("matchid"), m.Viper.GetInt64("initial_chips"))
		m.playermgr.Store(player)
	}

	return player, nil
}

// validateTable 验证桌子
func (m *Match) ValidateTable(player *Player, opts ...ValidateOption) (*Table, error) {
	options := &validateOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// 确定要查找的桌子ID
	var tableID int32
	if options.tableID > 0 {
		tableID = options.tableID
	} else if options.usePlayerTable {
		tableID = player.TableId
	} else {
		return nil, errors.New("table identification not specified")
	}

	// 验证桌子存在
	table, ok := m.tables.Load(tableID)
	if !ok {
		return nil, errors.New("table not found")
	}
	t := table.(*Table)
	// 验证玩家在桌子中（如果需要）
	if options.requirePlayerInTable && !t.IsOnTable(player) {
		return nil, errors.New("player not in specified table")
	}

	return t, nil
}

// validateRequest 通用的请求验证（保持向后兼容）
func (m *Match) ValidateRequest(ctx context.Context, opts ...ValidateOption) (*Player, *Table, error) {
	player, err := m.ValidatePlayer(ctx, opts...)
	if err != nil {
		return nil, nil, err
	}

	table, err := m.ValidateTable(player, opts...)
	if err != nil {
		return nil, nil, err
	}

	return player, table, nil
}
