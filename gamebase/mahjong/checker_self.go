package mahjong

type CheckerSelf interface {
	Check(opt *Operates)
}

// 胡检查器
type checkerHu struct {
	play *Play
}

func NewCheckerHu(play *Play) CheckerSelf {
	return &checkerHu{play: play}
}

func (c *checkerHu) Check(opt *Operates) {
	if c.play.IsAfterPon() {
		return
	}

	data := NewHuData(c.play.playData[c.play.curSeat], true)
	result, hu := data.CheckHu()
	if !hu {
		return
	}

	if c.play.checkMustHu(c.play.curSeat) {
		opt.RemoveOperate(OperateDiscard)
		c.play.AddHuOperate(opt, c.play.curSeat, result, true)
	} else if result.Multi < c.play.PlayConf.MinMultipleLimit {
		opt.Tips = append(opt.Tips, TipsQiHuFan)
	} else {
		c.play.AddHuOperate(opt, c.play.curSeat, result, false)
	}
}

// 杠检查器
type checkerKon struct {
	play *Play
}

func NewCheckerKon(play *Play) CheckerSelf {
	return &checkerKon{play: play}
}
func (c *checkerKon) Check(opt *Operates) {
	if opt.IsMustHu {
		return
	}
	if c.play.playData[c.play.curSeat].canSelfKon(c.play.tilesLai) {
		opt.AddOperate(OperateKon)
	}
}

// 听检查器
type checkerTing struct {
	play *Play
}

func NewCheckerTing(play *Play) CheckerSelf {
	return &checkerTing{play: play}
}
func (c *checkerTing) Check(opt *Operates) {
	if opt.IsMustHu || c.play.playData[c.play.curSeat].ting {
		return
	}

	huData := NewHuData(c.play.playData[c.play.curSeat], false)
	callData := huData.CheckCall()
	if len(callData) <= 0 {
		return
	}

	if c.play.PlayConf.TianTing && !c.play.HasOperate(c.play.curSeat) {
		opt.AddOperate(OperateTianTing)
	} else {
		opt.AddOperate(OperateTing)
	}
}
