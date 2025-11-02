package mahjong

import (
	"encoding/json"
	"math/rand"
)

type LastGameData struct {
	banker int32
	data   map[string]int32
}

func NewLastGameData(playerCount int) *LastGameData {
	return &LastGameData{
		banker: int32(rand.Intn(playerCount)),
		data:   make(map[string]int32),
	}
}

func (lgd *LastGameData) Set(key string, value int32) {
	lgd.data[key] += value
}

func (lgd *LastGameData) Get(key string) int32 {
	return lgd.data[key]
}

func (lgd *LastGameData) String() string {
	data, err := json.Marshal(lgd.data)
	if err != nil {
		return ""
	}
	return string(data)
}
