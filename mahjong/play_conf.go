package mahjong

// PlayConf 功能配置
type PlayConf struct {
	HasChow              bool  // 是否吃
	HasPon               bool  // 是否允许碰
	PonPass              bool  // 是否过碰不碰
	HuPass               bool  // 过胡不胡
	MustHu               bool  // 有胡必胡
	OnlyZimo             bool  // 只能自摸
	CanotOnlyLaiAfterPon bool  // 不允许碰后全是赖子
	MustHuIfOnlyLai      bool  // 全赖子必胡
	CanotDiscardLai      bool  // 赖子不可打出
	TianTing             bool  // 有天听玩法
	EnableCall           bool  // 有报叫玩法
	BuKonPass            bool  // 补杠区分过手杠
	ZhiKonAfterPon       bool  // 碰后补杠算直杠
	MinMultipleLimit     int64 // 起胡倍数
	MaxMultipleLimit     int64 // 封顶倍数
}

// GetRealMultiple 获取倍数
func (p *PlayConf) GetRealMultiple(mult int64) int64 {
	if p.MaxMultipleLimit <= 0 {
		return mult
	}
	if mult >= p.MaxMultipleLimit {
		return p.MaxMultipleLimit
	}
	return mult
}

// IsTopMultiple 是否达到最大倍数
func (p *PlayConf) IsTopMultiple(mult int64) bool {
	if p.MaxMultipleLimit <= 0 {
		return false
	}
	return mult >= p.MaxMultipleLimit
}
