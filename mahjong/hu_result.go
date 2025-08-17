package mahjong

type HuResult struct {
	HuTypes   []int32
	Extras    map[int32]int32
	TotalFan  int64
	TotalMuti int64
}

func (h *HuResult) GetExtra(extra int32) int32 {
	return h.Extras[extra]
}
