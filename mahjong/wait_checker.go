package mahjong

import "slices"

// WaitChecker 定义检查接口
type WaitChecker interface {
	Check(play *Play, seat int32, opt *Operates, tips []int) []int
}

type PaoChecker struct{}      // 点炮检查器
type ChowChecker struct{}     // 吃牌检查器
type PonChecker struct{}      // 碰牌检查器
type ZhiKonChecker struct{}   // 直杠检查器
type ChowTingChecker struct{} // 吃听检查器
type PonTingChecker struct{}  // 碰听检查器

func (c *PaoChecker) Check(play *Play, seat int32, opt *Operates, tips []int) []int {
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

func (c *ChowChecker) Check(play *Play, seat int32, opt *Operates, tips []int) []int {
	if opt.IsMustHu {
		return tips
	}
	playData := play.playData[seat]
	if playData.call {
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

func (c *PonChecker) Check(play *Play, seat int32, opt *Operates, tips []int) []int {
	if opt.IsMustHu {
		return tips
	}
	playData := play.playData[seat]
	if playData.call {
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

func (c *ZhiKonChecker) Check(play *Play, seat int32, opt *Operates, tips []int) []int {
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

func (c *ChowTingChecker) Check(play *Play, seat int32, opt *Operates, tips []int) []int {
	if opt.IsMustHu {
		return tips
	}
	playData := play.playData[seat]
	if playData.call {
		return tips
	}

	if !playData.canChow(play.curTile) {
		return tips
	}

	huData := NewCheckHuData(play, play.playData[play.curSeat], true)
	leftPoint := max(0, TilePoint(play.curTile)-2)
	color := TileColor(play.curTile)

	for p := leftPoint; p < leftPoint+3; p++ {
		tiles := make([]int32, 0)
		for i := range 3 {
			tile := MakeTile(color, p+i)
			if tile != play.curTile && slices.Contains(playData.handTiles, tile) {
				huData.TilesInHand = RemoveElements(huData.TilesInHand, tile, 1)
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
		huData.TilesInHand = append(huData.TilesInHand, tiles...)
	}
	return tips
}

func (c *PonTingChecker) Check(play *Play, seat int32, opt *Operates, tips []int) []int {
	if !opt.HasOperate(OperatePon) {
		return tips
	}

	huData := NewCheckHuData(play, play.playData[play.curSeat], true)
	huData.TilesInHand = RemoveElements(huData.TilesInHand, play.curTile, 2)
	callData := Service.CheckCall(huData, play.game.rule)
	if len(callData) > 0 {
		opt.AddOperate(OperatePonTing)
	}
	return tips
}
