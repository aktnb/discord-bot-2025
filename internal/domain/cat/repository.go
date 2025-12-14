package cat

import "context"

// CatImageRepository は猫画像取得のポートインターフェース
type CatImageRepository interface {
	FetchRandomImage(ctx context.Context) (*CatImage, error)
}
