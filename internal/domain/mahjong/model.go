package mahjong

// MahjongStartingHand は麻雀の配牌を表すドメインエンティティ
type MahjongStartingHand struct {
	// ImageData は画像バイナリデータ
	ImageData []byte
	// ContentType は画像のMIMEタイプ（通常は"image/png"）
	ContentType string
}
