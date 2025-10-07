package mahjong

import "slices"

// CheckerWait 定义检查接口
type CheckerWait interface {
	Check(seat int32, opt *Operates, tips []int) []int
}

type CheckerPao struct{ play *Play } // 点炮检查器
func NewCheckerPao(play *Play) CheckerWait {
	return &CheckerPao{play: play}
}
func (c *CheckerPao) Check(seat int32, opt *Operates, tips []int) []int {
	if c.play.PlayConf.OnlyZimo {
		tips = append(tips, TipsOnlyZiMo)
	}

	data := NewHuData(c.play.playData[seat], false)
	result, hu := data.CheckHu()
	if !hu {
		return tips
	}

	if c.play.PlayConf.MustHu {
		c.play.AddHuOperate(opt, seat, result, true)
	} else if c.play.playData[seat].IsPassHuTile(c.play.curTile) && c.play.PlayConf.HuPass {
		tips = append(tips, TipsPassHu)
	} else if result.TotalMuti < c.play.PlayConf.MinMultipleLimit {
		tips = append(tips, TipsQiHuFan)
	} else {
		c.play.AddHuOperate(opt, seat, result, true)
	}
	return tips
}

type CheckerChow struct{ play *Play } // 吃牌检查器
func NewCheckerChow(play *Play) CheckerWait {
	return &CheckerChow{play: play}
}

func (c *CheckerChow) Check(seat int32, opt *Operates, tips []int) []int {
	if opt.IsMustHu {
		return tips
	}
	playData := c.play.playData[seat]
	if playData.ting {
		return tips
	}

	if GetNextSeat(c.play.curSeat, 1, c.play.game.GetPlayerCount()) != seat {
		return tips
	}

	if playData.canChow(c.play.curTile) {
		opt.AddOperate(OperateChow)
	}
	return tips
}

type CheckerPon struct{ play *Play } // 碰牌检查器
func NewCheckerPon(play *Play) CheckerWait {
	return &CheckerPon{play: play}
}
func (c *CheckerPon) Check(seat int32, opt *Operates, tips []int) []int {
	if opt.IsMustHu {
		return tips
	}
	playData := c.play.playData[seat]
	if playData.ting {
		return tips
	}

	tmpOpr := &Operates{}
	if playData.canPon(c.play.curTile, c.play.PlayConf.CanotOnlyLaiAfterPon) {
		tmpOpr.AddOperate(OperatePon)
	}

	if !playData.IsPassPonTile(c.play.curTile) || !c.play.PlayConf.PonPass {
		opt.AddOperates(tmpOpr)
	} else if tmpOpr.Value != 0 {
		tips = append(tips, TipsPassPon)
	}
	return tips
}

type CheckerZhiKon struct{ play *Play } // 直杠检查器
func NewCheckerZhiKon(play *Play) CheckerWait {
	return &CheckerZhiKon{play: play}
}
func (c *CheckerZhiKon) Check(seat int32, opt *Operates, tips []int) []int {
	if opt.IsMustHu {
		return tips
	}
	if c.play.dealer.GetRestCount() <= 0 {
		return tips
	}

	playData := c.play.playData[seat]
	if playData.canKon(c.play.curTile, KonTypeZhi) {
		opt.AddOperate(OperateKon)
	}
	return tips
}

type CheckerChowTing struct{ play *Play } // 吃听检查器
func NewCheckerChowTing(play *Play) CheckerWait {
	return &CheckerChowTing{play: play}
}
func (c *CheckerChowTing) Check(seat int32, opt *Operates, tips []int) []int {
	if opt.IsMustHu {
		return tips
	}
	playData := c.play.playData[seat]
	if playData.ting {
		return tips
	}

	if !playData.canChow(c.play.curTile) {
		return tips
	}

	huData := NewHuData(c.play.playData[c.play.curSeat], false)
	leftPoint := max(0, c.play.curTile.Point()-2)
	color := c.play.curTile.Color()

	for p := leftPoint; p < leftPoint+3; p++ {
		tiles := make([]Tile, 0)
		for i := range 3 {
			tile := MakeTile(color, p+i)
			if tile != c.play.curTile && slices.Contains(playData.handTiles, tile) {
				huData.Tiles = RemoveElements(huData.Tiles, tile, 1)
				tiles = append(tiles, tile)
			}
		}
		if len(tiles) == 2 {
			callData := huData.CheckCall()
			if len(callData) > 0 {
				opt.AddOperate(OperateChowTing)
				return tips
			}
		}
		huData.Tiles = append(huData.Tiles, tiles...)
	}
	return tips
}

type CheckerPonTing struct{ play *Play } // 碰听检查器
func NewCheckerPonTing(play *Play) CheckerWait {
	return &CheckerPonTing{play: play}
}

func (c *CheckerPonTing) Check(seat int32, opt *Operates, tips []int) []int {
	if !opt.HasOperate(OperatePon) {
		return tips
	}

	huData := NewHuData(c.play.playData[c.play.curSeat], false)
	huData.Tiles = RemoveElements(huData.Tiles, c.play.curTile, 2)
	callData := huData.CheckCall()
	if len(callData) > 0 {
		opt.AddOperate(OperatePonTing)
	}
	return tips
}
