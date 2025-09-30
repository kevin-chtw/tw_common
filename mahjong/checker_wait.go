package mahjong

import "slices"

// CheckerWait 定义检查接口
type CheckerWait interface {
	Check(play *Play, seat int32, opt *Operates, tips []int) []int
}

type CheckerPao struct{}      // 点炮检查器
type CheckerChow struct{}     // 吃牌检查器
type CheckerPon struct{}      // 碰牌检查器
type CheckerZhiKon struct{}   // 直杠检查器
type CheckerChowTing struct{} // 吃听检查器
type CheckerPonTing struct{}  // 碰听检查器

func (c *CheckerPao) Check(play *Play, seat int32, opt *Operates, tips []int) []int {
	if play.PlayConf.OnlyZimo {
		tips = append(tips, TipsOnlyZiMo)
	}

	data := NewCheckHuData(play, play.playData[seat], false)
	result, hu := Service.CheckHu(data, play.game.rule)
	if !hu {
		return tips
	}

	if play.PlayConf.MustHu {
		play.addHuOperate(opt, seat, result, true)
	} else if play.playData[seat].IsPassHuTile(play.curTile) && play.PlayConf.HuPass {
		tips = append(tips, TipsPassHu)
	} else if result.TotalMuti < play.PlayConf.MinMultipleLimit {
		tips = append(tips, TipsQiHuFan)
	} else {
		play.addHuOperate(opt, seat, result, true)
	}
	return tips
}

func (c *CheckerChow) Check(play *Play, seat int32, opt *Operates, tips []int) []int {
	if opt.IsMustHu {
		return tips
	}
	playData := play.playData[seat]
	if playData.ting {
		return tips
	}

	if GetNextSeat(play.curSeat, 1, play.game.GetPlayerCount()) != seat {
		return tips
	}

	if playData.canChow(play.curTile) {
		opt.AddOperate(OperateChow)
	}
	return tips
}

func (c *CheckerPon) Check(play *Play, seat int32, opt *Operates, tips []int) []int {
	if opt.IsMustHu {
		return tips
	}
	playData := play.playData[seat]
	if playData.ting {
		return tips
	}

	tmpOpr := &Operates{}
	if playData.canPon(play.curTile, play.PlayConf.CanotOnlyLaiAfterPon) {
		tmpOpr.AddOperate(OperatePon)
	}

	if !playData.IsPassPonTile(play.curTile) || !play.PlayConf.PonPass {
		opt.AddOperates(tmpOpr)
	} else if tmpOpr.Value != 0 {
		tips = append(tips, TipsPassPon)
	}
	return tips
}

func (c *CheckerZhiKon) Check(play *Play, seat int32, opt *Operates, tips []int) []int {
	if opt.IsMustHu {
		return tips
	}
	if play.dealer.GetRestCount() <= 0 {
		return tips
	}

	playData := play.playData[seat]
	if playData.canKon(play.curTile, KonTypeZhi) {
		opt.AddOperate(OperateKon)
	}
	return tips
}

func (c *CheckerChowTing) Check(play *Play, seat int32, opt *Operates, tips []int) []int {
	if opt.IsMustHu {
		return tips
	}
	playData := play.playData[seat]
	if playData.ting {
		return tips
	}

	if !playData.canChow(play.curTile) {
		return tips
	}

	huData := NewCheckHuData(play, play.playData[play.curSeat], true)
	leftPoint := max(0, play.curTile.Point()-2)
	color := play.curTile.Color()

	for p := leftPoint; p < leftPoint+3; p++ {
		tiles := make([]Tile, 0)
		for i := range 3 {
			tile := MakeTile(color, p+i)
			if tile != play.curTile && slices.Contains(playData.handTiles, tile) {
				huData.Tiles = RemoveElements(huData.Tiles, tile, 1)
				tiles = append(tiles, tile)
			}
		}
		if len(tiles) == 2 {
			callData := Service.CheckCall(huData, play.game.rule)
			if len(callData) > 0 {
				opt.AddOperate(OperateChowTing)
				return tips
			}
		}
		huData.Tiles = append(huData.Tiles, tiles...)
	}
	return tips
}

func (c *CheckerPonTing) Check(play *Play, seat int32, opt *Operates, tips []int) []int {
	if !opt.HasOperate(OperatePon) {
		return tips
	}

	huData := NewCheckHuData(play, play.playData[play.curSeat], true)
	huData.Tiles = RemoveElements(huData.Tiles, play.curTile, 2)
	callData := Service.CheckCall(huData, play.game.rule)
	if len(callData) > 0 {
		opt.AddOperate(OperatePonTing)
	}
	return tips
}
