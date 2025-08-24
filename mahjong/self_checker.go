package mahjong

type SelfChecker interface {
	Check(play *Play, opt *Operates, tips []int) []int
}

type HuChecker struct{}   // 胡检查器
type KonChecker struct{}  // 杠检查器
type CallChecker struct{} // 听检查器

func (c *HuChecker) Check(play *Play, opt *Operates, tips []int) []int {
	opt.AddOperate(OperateDiscard)
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

func (c *KonChecker) Check(play *Play, opt *Operates, tips []int) []int {
	if opt.IsMustHu {
		return tips
	}
	if play.playData[play.curSeat].canSelfKon(play.game.rule, play.tilesLai) {
		opt.AddOperate(OperateKon)
	}
	return tips
}

func (c *CallChecker) Check(play *Play, opt *Operates, tips []int) []int {
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
