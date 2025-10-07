package mahjong

import (
	"github.com/kevin-chtw/tw_common/game"
	"github.com/kevin-chtw/tw_proto/game/pbmj"
	"google.golang.org/protobuf/proto"
)

type MsgPacker interface {
	PackMsg(msg proto.Message) (proto.Message, error)
}

type Sender struct {
	game   *Game
	play   *Play
	packer MsgPacker
}

func ToCallData(callData map[Tile]map[Tile]int64) map[int32]*pbmj.CallData {
	result := make(map[int32]*pbmj.CallData)
	for tile, callMap := range callData {
		callPb := &pbmj.CallData{
			CallTiles: make(map[int32]int64),
		}
		for tile, fan := range callMap {
			callPb.CallTiles[int32(tile)] = fan
		}
		result[int32(tile)] = callPb
	}
	return result
}

func NewSender(game *Game, play *Play, packer MsgPacker) *Sender {
	return &Sender{
		game:   game,
		play:   play,
		packer: packer,
	}
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
			Seat:     i,
			Tiles:    s.play.GetPlayData(i).GetHandTilesInt32(),
			CallData: ToCallData(s.play.GetPlayData(i).GetCallMap()),
		}
		s.SendMsg(openDoor, i)
	}
}

func (s *Sender) SendAnimationAck() {
	animationAck := &pbmj.MJAnimationAck{
		Requestid: s.game.GetRequestID(game.SeatAll),
	}
	s.SendMsg(animationAck, game.SeatAll)
}

func (s *Sender) SendRequestAck(seat int32, operates *Operates) {
	requestAck := &pbmj.MJRequestAck{
		Seat:        seat,
		RequestType: int32(operates.Value),
		Requestid:   s.game.GetRequestID(seat),
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
