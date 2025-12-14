package omikuji

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/aktnb/discord-bot-go/internal/domain/omikuji"
)

type Service struct{}

func NewOmikujiService() *Service {
	return &Service{}
}

// DrawFortune はユーザーIDと日付に基づいて決定的におみくじを引く
func (s *Service) DrawFortune(ctx context.Context, userID string) (*omikuji.Fortune, error) {
	// 今日の日付を取得（JST）
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	today := time.Now().In(jst).Format("2006-01-02")

	// ユーザーID + 日付でシード値を生成（決定性を保証）
	seed := generateSeed(userID, today)

	// シードから運勢レベルを決定
	level := determineFortuneLevel(seed)

	// Fortuneエンティティを生成
	fortune := omikuji.NewFortune(level)

	return fortune, nil
}

// generateSeed はユーザーIDと日付からシード値を生成
func generateSeed(userID string, date string) uint64 {
	// SHA256でハッシュ化
	input := fmt.Sprintf("%s:%s", userID, date)
	hash := sha256.Sum256([]byte(input))

	// ハッシュの最初の8バイトをuint64に変換
	seed := binary.BigEndian.Uint64(hash[:8])

	return seed
}

// determineFortuneLevel はシード値から運勢レベルを決定
// 確率分布：
// - 超大吉: 1%  (0-9)
// - 大吉:  10%  (10-109)
// - 中吉:  20%  (110-309)
// - 小吉:  20%  (310-509)
// - 吉:    25%  (510-759)
// - 凶:    20%  (760-959)
// - 大凶:  4%   (960-999)
func determineFortuneLevel(seed uint64) omikuji.FortuneLevel {
	// 0-999の範囲に正規化
	value := seed % 1000

	switch {
	case value < 10:
		return omikuji.UltraGreatBlessing
	case value < 110:
		return omikuji.GreatBlessing
	case value < 310:
		return omikuji.MiddleBlessing
	case value < 510:
		return omikuji.SmallBlessing
	case value < 760:
		return omikuji.Blessing
	case value < 960:
		return omikuji.BadLuck
	default:
		return omikuji.GreatBadLuck
	}
}
