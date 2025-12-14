package mahjong

import "context"

// MahjongRepository はランダムな麻雀配牌取得のポートインターフェース
type MahjongRepository interface {
	FetchRandomStartingHand(ctx context.Context) (*MahjongStartingHand, error)
}
