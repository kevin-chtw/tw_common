package mahjong

// 手牌风格类型
const (
	HandNone            EHandStyle = iota // 无特殊风格
	HandNormal                            // 普通手牌
	HandSevenPairs                        // 七对
	HandThirteenOrphans                   // 十三幺
)

// 听牌基础类
type TingBase struct {
	Tiles    []int32
	MinValue int
}

const (
	TileNull int32 = -1
	TileBack int32 = -1
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

var (
	TileHun    int32 = MakeTile(ColorHun, 0)    // 混子
	TileInf    int32 = MakeTile(ColorEnd, 0)    // 无效牌
	TileZhong  int32 = MakeTile(ColorDragon, 0) // 中
	TileFa     int32 = MakeTile(ColorDragon, 1) // 发
	TileBai    int32 = MakeTile(ColorDragon, 2) // 白
	TileDong   int32 = MakeTile(ColorWind, 0)   // 东
	TileNan    int32 = MakeTile(ColorWind, 1)   // 南
	TileXi     int32 = MakeTile(ColorWind, 2)   // 西
	TileBei    int32 = MakeTile(ColorWind, 3)   // 北
	TileYaoJi  int32 = MakeTile(ColorBamboo, 0) // 幺鸡
	TileFlower int32 = MakeTile(ColorFlower, 0) // 花
	TileSpring int32 = MakeTile(ColorSeason, 0) // 春
)

type ScoreReason int //算分原因

const (
	ScoreReasonHu   ScoreReason = iota // 胡
	ScoreReasonGang                    // 杠
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
	TipsPassHu   = iota // 过胡
	TipsPassPon         // 过碰
	TipsQiHuFan         // 起胡番
	TipsOnlyZiMo        // 只自摸
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

func MakeTile(color EColor, point int) int32 {
	// 默认flag为1
	return int32((int(color) << 8) | (point << 4) | 1)
}

func TileColor(tile int32) EColor {
	return EColor((tile >> 8) & 0x0F)
}

func TilePoint(tile int32) int {
	return int((tile >> 4) & 0x0F)
}

func TileFlag(tile int32) int {
	return int(tile & 0x0F)
}

func IsValidTile(tile int32) bool {
	return tile > 0 && tile < TileInf
}

func IsSuitColor(color EColor) bool {
	return color >= ColorCharacter && color <= ColorDot
}

func IsSuitTile(tile int32) bool {
	return IsValidTile(tile) && IsSuitColor(TileColor(tile))
}

func Is258Tile(tile int32) bool {
	return IsValidTile(tile) && IsSuitColor(TileColor(tile)) && (TilePoint(tile)%3 == 1)
}

func IsHonorColor(color EColor) bool {
	return color == ColorWind || color == ColorDragon
}

func IsDragonTile(tile int32) bool {
	return TileColor(tile) == ColorDragon
}

func IsHonorTile(tile int32) bool {
	return IsValidTile(tile) && IsHonorColor(TileColor(tile))
}

// NewTingEx14 创建14张牌的听牌工具
func NewTingEx14() *TingBase {
	return &TingBase{
		Tiles:    make([]int32, 0),
		MinValue: 0,
	}
}

func IsExtraColor(color EColor) bool {
	return color == ColorSeason || color == ColorFlower
}

func IsExtraTile(tile int32) bool {
	return IsValidTile(tile) && IsExtraColor(TileColor(tile))
}

func NextTileInSameColor(tile int32, step int) int32 {
	if !IsValidTile(tile) {
		return tile
	}
	color := TileColor(tile)
	count := PointCountByColor[color]
	step %= count
	point := (TilePoint(tile) + step + count) % count
	return MakeTile(color, point)
}

type Action struct {
	Seat    int32
	Tile    int32
	Operate int
	Extra   int
}
