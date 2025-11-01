package mahjong

import (
	"github.com/kevin-chtw/tw_common/gamebase/game"
	"github.com/kevin-chtw/tw_proto/game/pbmj"
	"google.golang.org/protobuf/proto"
)

type MsgPacker interface {
	PackMsg(msg proto.Message) (proto.Message, error)
}

type Sender struct {
	game        *Game
	play        *Play
	packer      MsgPacker
	increasedID int32   // 当前请求ID
	requestIDs  []int32 // 记录每个玩家的请求ID
}

func ToCallData(callData map[Tile]map[Tile]int64) map[int32]*pbmj.CallData {
	result := make(map[int32]*pbmj.CallData)
	for tile, callMap := range callData {
		callPb := &pbmj.CallData{
			CallTiles: make(map[int32]*pbmj.CallInfo),
		}
		for tile, multi := range callMap {
			callPb.CallTiles[int32(tile)] = &pbmj.CallInfo{Multi: multi}
		}
		result[int32(tile)] = callPb
	}
	return result
}

func NewSender(game *Game, play *Play, packer MsgPacker) *Sender {
	return &Sender{
		game:        game,
		play:        play,
		packer:      packer,
		increasedID: 1,
		requestIDs:  make([]int32, game.GetPlayerCount()),
	}
}

func (s *Sender) GetRequestID(seat int32) int32 {
	s.increasedID++
	if s.game.IsValidSeat(seat) {
		s.requestIDs[seat] = s.increasedID
	} else {
		for i := range s.requestIDs {
			s.requestIDs[i] = s.increasedID
		}
	}
	return s.increasedID
}

func (s *Sender) IsRequestID(seat, id int32) bool {
	if !s.game.IsValidSeat(seat) {
		return false
	}
	return s.requestIDs[seat] == id
}

func (s *Sender) SendMsg(msg proto.Message, seat int32) error {
	ack, err := s.packer.PackMsg(msg)
	if err != nil {
		return err
	}
	s.game.Send2Player(ack, seat)
	return nil
}

func (s *Sender) SendGameStartAck() {
	startAck := &pbmj.MJGameStartAck{
		Banker:    s.play.banker,
		TileCount: s.play.dealer.GetRestCount(),
		Scores:    s.play.GetCurScores(),
		Property:  s.game.GetRule().ToString(),
	}
	s.SendMsg(startAck, game.SeatAll)
}

func (s *Sender) SendOpenDoorAck() {
	count := s.game.GetPlayerCount()
	for i := range count {
		openDoor := &pbmj.MJOpenDoorAck{
			Seat:  i,
			Tiles: s.play.GetPlayData(i).GetHandTilesInt32(),
		}
		s.SendMsg(openDoor, i)
	}
}

func (s *Sender) SendAnimationAck() {
	animationAck := &pbmj.MJAnimationAck{
		Requestid: s.GetRequestID(game.SeatAll),
	}
	s.SendMsg(animationAck, game.SeatAll)
}

func (s *Sender) SendRequestAck(seat int32, operates *Operates) {
	requestAck := &pbmj.MJRequestAck{
		Seat:        seat,
		RequestType: int32(operates.Value),
		Requestid:   s.GetRequestID(seat),
		HuMulti:     operates.HuMulti,
		ChowLpoints: operates.ChowLPoints,
	}
	s.SendMsg(requestAck, seat)
}

func (s *Sender) SendDiscardAck() {
	discardAck := &pbmj.MJDiscardAck{
		Seat: s.play.GetCurSeat(),
		Tile: s.play.GetCurTile().ToInt32(),
	}
	s.SendMsg(discardAck, game.SeatAll)
}

func (s *Sender) SendTingAck(seat int32, tile Tile) {
	tingAck := &pbmj.MJTingAck{
		Seat:     seat,
		Tile:     tile.ToInt32(),
		TianTing: s.play.GetPlayData(seat).IsTianTing(),
	}
	s.SendMsg(tingAck, game.SeatAll)
}

func (s *Sender) SendChowAck(seat int32, tile, leftTile Tile) {
	chowAck := &pbmj.MJChowAck{
		Seat:     seat,
		From:     s.play.GetCurSeat(),
		Tile:     tile.ToInt32(),
		LeftTile: leftTile.ToInt32(),
		CallData: ToCallData(s.play.GetPlayData(seat).GetCallMap()),
	}
	s.SendMsg(chowAck, game.SeatAll)
}

func (s *Sender) SendPonAck(seat int32, tile Tile) {
	ponAck := &pbmj.MJPonAck{
		Seat:     seat,
		From:     s.play.GetCurSeat(),
		Tile:     tile.ToInt32(),
		CallData: ToCallData(s.play.GetPlayData(seat).GetCallMap()),
	}
	s.SendMsg(ponAck, game.SeatAll)
}

func (s *Sender) SendKonAck(seat int32, tile Tile, konType KonType) {
	konAck := &pbmj.MJKonAck{
		Seat:    seat,
		From:    s.play.GetCurSeat(),
		Tile:    tile.ToInt32(),
		KonType: int32(konType),
	}
	s.SendMsg(konAck, game.SeatAll)
}

func (s *Sender) SendHuAck(huSeats []int32, paoSeat int32) {
	huAck := &pbmj.MJHuAck{
		PaoSeat: paoSeat,
		Tile:    s.play.GetCurTile().ToInt32(),
		HuData:  make([]*pbmj.MJHuData, len(huSeats)),
	}
	for i := range huSeats {
		huAck.HuData[i] = &pbmj.MJHuData{
			Seat:    huSeats[i],
			Multi:   s.play.huResult[huSeats[i]].Multi,
			Gen:     s.play.huResult[huSeats[i]].Gen,
			HuTypes: s.play.huResult[huSeats[i]].HuTypes,
		}
	}
	s.SendMsg(huAck, game.SeatAll)
}

func (s *Sender) SendDrawAck(tile Tile) {
	drawAck := &pbmj.MJDrawAck{
		Seat:     s.play.GetCurSeat(),
		Tile:     tile.ToInt32(),
		CallData: ToCallData(s.play.GetPlayData(s.play.GetCurSeat()).GetCallMap()),
	}
	s.SendMsg(drawAck, drawAck.Seat)
	drawAck.Tile = TileNull.ToInt32()
	drawAck.CallData = nil
	for i := range s.game.GetPlayerCount() {
		if i != drawAck.Seat {
			s.SendMsg(drawAck, i)
		}
	}
}

func (s *Sender) SendCallDataAck(seat int32) {
	ack := &pbmj.MJCallDataAck{
		Seat:     seat,
		CallData: ToCallData(s.play.GetPlayData(seat).GetCallMap()),
	}
	s.SendMsg(ack, seat)
}

func (s *Sender) SendTrustAck(seat int32, trust bool) {
	player := s.game.GetPlayer(seat)
	if player.IsTrusted() == trust {
		return
	}
	player.SetTrusted(trust)
	ack := &pbmj.MJTrustAck{
		Seat:  seat,
		Trust: trust,
	}
	s.SendMsg(ack, seat)
}

func (s *Sender) SendScoreChangeAck(sr ScoreReason, scores []int64, tile Tile, paoSeat int32, huSeats []int32) {
	scoreChangeAck := &pbmj.MJScoreChangeAck{
		ScoreReason: int32(sr),
		Scores:      scores,
		Tile:        tile.ToInt32(),
		PaoSeat:     paoSeat,
		HuData:      make([]*pbmj.MJHuData, len(huSeats)),
	}
	for i := range huSeats {
		scoreChangeAck.HuData[i] = &pbmj.MJHuData{
			Seat:    huSeats[i],
			HuTypes: s.play.huResult[huSeats[i]].HuTypes,
		}
	}
	s.SendMsg(scoreChangeAck, game.SeatAll)
}

func (s *Sender) SendResult(liuju bool) {
	resultAck := &pbmj.MJResultAck{
		Liuju:         liuju,
		PlayerResults: make([]*pbmj.MJPlayerResult, s.game.GetPlayerCount()),
	}
	for i := range resultAck.PlayerResults {
		resultAck.PlayerResults[i] = &pbmj.MJPlayerResult{
			Seat:     int32(i),
			CurScore: s.game.GetPlayer(int32(i)).GetCurScore(),
			WinScore: s.game.GetPlayer(int32(i)).GetScoreChangeWithTax(),
			Tiles:    s.play.GetPlayData(int32(i)).GetHandTilesInt32(),
		}
	}
	s.SendMsg(resultAck, game.SeatAll)
}
