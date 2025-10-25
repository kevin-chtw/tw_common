package mahjong

import "math/rand"

type LastGameData struct {
	banker int32
}

func NewLastGameData(playerCount int) *LastGameData {
	return &LastGameData{
		banker: int32(rand.Intn(playerCount)),
	}
}
