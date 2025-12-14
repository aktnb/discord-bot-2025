package omikuji

// FortuneLevel はおみくじの結果レベル
type FortuneLevel int

const (
	UltraGreatBlessing FortuneLevel = iota // 超大吉
	GreatBlessing                           // 大吉
	MiddleBlessing                          // 中吉
	SmallBlessing                           // 小吉
	Blessing                                // 吉
	BadLuck                                 // 凶
	GreatBadLuck                            // 大凶
)

// String はFortuneLevel型を日本語文字列に変換
func (f FortuneLevel) String() string {
	switch f {
	case UltraGreatBlessing:
		return "超大吉"
	case GreatBlessing:
		return "大吉"
	case MiddleBlessing:
		return "中吉"
	case SmallBlessing:
		return "小吉"
	case Blessing:
		return "吉"
	case BadLuck:
		return "凶"
	case GreatBadLuck:
		return "大凶"
	default:
		return "unknown"
	}
}

// Fortune はおみくじの結果を表現するドメインエンティティ
type Fortune struct {
	Level   FortuneLevel
	Message string
}

// NewFortune はFortune型のコンストラクタ
func NewFortune(level FortuneLevel) *Fortune {
	return &Fortune{
		Level:   level,
		Message: generateMessage(level),
	}
}

// generateMessage はレベルに応じたメッセージを生成
func generateMessage(level FortuneLevel) string {
	switch level {
	case UltraGreatBlessing:
		return "素晴らしい運勢です！今日は何をやっても上手くいきそう！"
	case GreatBlessing:
		return "とても良い運勢です！"
	case MiddleBlessing:
		return "良い運勢です！"
	case SmallBlessing:
		return "まずまずの運勢です！"
	case Blessing:
		return "普通の運勢です！"
	case BadLuck:
		return "少し注意が必要かもしれません..."
	case GreatBadLuck:
		return "今日は慎重に行動しましょう..."
	default:
		return ""
	}
}
