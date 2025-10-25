package mahjong

type CheckerSelf interface {
	Check(opt *Operates, tips []int) []int
}

// 胡检查器
type checkerHu struct {
	play *Play
}

func NewCheckerHu(play *Play) CheckerSelf {
	return &checkerHu{play: play}
}

func (c *checkerHu) Check(opt *Operates, tips []int) []int {
	if c.play.IsAfterPon() {
		return tips
	}

	data := NewHuData(c.play.playData[c.play.curSeat], true)
	result, hu := data.CheckHu()
	if !hu {
		return tips
	}

	if c.play.checkMustHu(c.play.curSeat) {
		opt.RemoveOperate(OperateDiscard)
		c.play.AddHuOperate(opt, c.play.curSeat, result, true)
	} else if result.TotalMuti < c.play.PlayConf.MinMultipleLimit {
		tips = append(tips, TipsQiHuFan)
	} else {
		c.play.AddHuOperate(opt, c.play.curSeat, result, false)
	}
	return tips
}

// 杠检查器
type checkerKon struct {
	play *Play
}

func NewCheckerKon(play *Play) CheckerSelf {
	return &checkerKon{play: play}
}
func (c *checkerKon) Check(opt *Operates, tips []int) []int {
	if opt.IsMustHu {
		return tips
	}
	if c.play.playData[c.play.curSeat].canSelfKon(c.play.tilesLai) {
		opt.AddOperate(OperateKon)
	}
	return tips
}

// 听检查器
type checkerTing struct {
	play *Play
}

func NewCheckerTing(play *Play) CheckerSelf {
	return &checkerTing{play: play}
}
func (c *checkerTing) Check(opt *Operates, tips []int) []int {
	if opt.IsMustHu {
		return tips
	}

	huData := NewHuData(c.play.playData[c.play.curSeat], false)
	callData := huData.CheckCall()
	if len(callData) <= 0 {
		return tips
	}

	if c.play.PlayConf.TianTing && !c.play.HasOperate(c.play.curSeat) {
		opt.AddOperate(OperateTianTing)
	} else {
		opt.AddOperate(OperateTing)
	}
	return tips
}
