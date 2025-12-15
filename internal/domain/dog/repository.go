package dog

import "context"

// DogImageRepository は犬画像取得のポートインターフェース
type DogImageRepository interface {
	FetchRandomImage(ctx context.Context) (*DogImage, error)
}
