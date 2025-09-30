package mahjong

type CheckerSelf interface {
	Check(play *Play, opt *Operates, tips []int) []int
}

type CheckerHu struct{}   // 胡检查器
type CheckerKon struct{}  // 杠检查器
type CheckerTing struct{} // 听检查器

func (c *CheckerHu) Check(play *Play, opt *Operates, tips []int) []int {
	if play.IsAfterPon() {
		return tips
	}

	data := NewCheckHuData(play, play.playData[play.curSeat], true)
	result, hu := Service.CheckHu(data, play.game.rule)
	if !hu {
		return tips
	}

	if play.checkMustHu(play.curSeat) {
		opt.RemoveOperate(OperateDiscard)
		play.addHuOperate(opt, play.curSeat, result, true)
	} else if result.TotalMuti < play.PlayConf.MinMultipleLimit {
		tips = append(tips, TipsQiHuFan)
	} else {
		play.addHuOperate(opt, play.curSeat, result, false)
	}
	return tips
}

func (c *CheckerKon) Check(play *Play, opt *Operates, tips []int) []int {
	if opt.IsMustHu {
		return tips
	}
	if play.playData[play.curSeat].canSelfKon(play.game.rule, play.tilesLai) {
		opt.AddOperate(OperateKon)
	}
	return tips
}

func (c *CheckerTing) Check(play *Play, opt *Operates, tips []int) []int {
	if opt.IsMustHu {
		return tips
	}

	huData := NewCheckHuData(play, play.playData[play.curSeat], true)
	callData := Service.CheckCall(huData, play.game.rule)
	if len(callData) <= 0 {
		return tips
	}

	if play.PlayConf.TianTing && !play.HasOperate(play.curSeat) {
		opt.AddOperate(OperateTianTing)
	} else {
		opt.AddOperate(OperateTing)
	}
	return tips
}
