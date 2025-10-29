package mahjong

// 手牌风格类型
const (
	HandNone            EHandStyle = iota // 无特殊风格
	HandNormal                            // 普通手牌
	HandSevenPairs                        // 七对
	HandThirteenOrphans                   // 十三幺
)

const (
	SeatNull int32 = -1
)

const (
	NP4 = 4
	NP3 = 3
	NP2 = 2
)

const (
	TileCountInitBanker = 14
	TileCountInitNormal = 13
)

type EColor int

const (
	ColorUndefined EColor = -1
	ColorCharacter EColor = iota - 1 // 万
	ColorBamboo                      // 条
	ColorDot                         // 筒
	ColorWind                        // 风牌
	ColorDragon                      // 箭牌
	ColorFlower                      // 花牌
	ColorSeason                      // 季牌
	ColorHun                         // 混子
	ColorEnd
	ColorBegin = ColorCharacter
)

var PointCountByColor = [ColorEnd]int{9, 9, 9, 4, 3, 4, 4, 0}
var SameTileCountByColor = [ColorEnd]int{4, 4, 4, 4, 4, 1, 1, 0}
var SEQ_BEGIN_BY_COLOR = [ColorEnd]int{0, 9, 18, 27, 31, 34, 38, 42}

type ScoreReason int //算分原因

const (
	ScoreReasonHu      ScoreReason = iota // 胡 0
	ScoreReasonAnKon                      // 暗杠 1
	ScoreReasonBuKon                      // 补杠 2
	ScoreReasonZhiKon                     // 直杠 3
	ScoreReasonTuiKon                     // 退杠 4
	ScoreReasonChaJiao                    // 查叫 5
)

type ScoreType int //算分方式

const (
	ScoreTypeNatural  ScoreType = iota // 自然分
	ScoreTypeMinScore                  // 积分最小化
	ScoreTypePositive                  // 超出玩家带入的输分由系统支出
	ScoreTypeJustWin                   // 只赢不输
)

type EHandStyle int

const (
	HandStyleNone EHandStyle = -1 + iota
	HandStyleTianHu
	HandStyleTianTing
	HandStyleYSYT
)

type ETrustType int

const (
	TrustTypeUntrust      ETrustType = iota
	TrustTypeTimeout                 = 2
	TrustTypeFDTNetBreak             = 5
	TrustTypeFDTNetResume            = 6
)

const (
	TipsPassHu   = iota // 过胡 0
	TipsPassPon         // 过碰 1
	TipsQiHuFan         // 起胡番 2
	TipsOnlyZiMo        // 只自摸 3
	TipsMenQin          // 未开门 4
)

type EPlayerType int

const (
	PlayerTypeNone EPlayerType = iota
	PlayerTypeNewbie
	PlayerTypeUnusual
	PlayerTypeNormal
	PlayerTypeNeedhelp
)

type EDecisionStage int

const (
	DecisionStageStart EDecisionStage = 1 + iota
	DecisionStageAck
	DecisionStageResult
)

type EGameOverStatus int

const (
	GameOverNormal EGameOverStatus = iota
	GameOverTimeout
	GameOverException
)

type KonType int

const (
	KonTypeNone KonType = -1 + iota
	KonTypeZhi
	KonTypeAn
	KonTypeBu
	KonTypeBuDelay
)

type EGroupType int

const (
	GroupTypeNone EGroupType = iota
	GroupTypeChow
	GroupTypePon
	GroupTypeZhiKon
	GroupTypeAnKon
	GroupTypeBuKon
)

type TileStyle struct {
	ShunCount  int
	NaiZiCount int
	Enable     bool
}

const (
	MAX_VAL_NUM   = 9
	MAX_KEY_NUM   = 10
	MAX_FENZI_NUM = 7
	BIT_VAL_NUM   = 3
	MAX_NAI_NUM   = 4
	BIT_VAL_FLAG  = 0x07
)

func GetNextSeat(seat, step, seatCount int32) int32 {
	return (seat + step) % seatCount
}

type Action struct {
	Seat    int32
	Operate int
	Tile    Tile
	Extra   Tile
}
