package mahjongapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aktnb/discord-bot-go/internal/domain/mahjong"
)

const (
	baseURL        = "https://mahjong-api.vercel.app/api/starting-hand"
	requestTimeout = 10 * time.Second
	maxImageSize   = 8 << 20 // 8 MiB
)

type MahjongAPIClient struct {
	httpClient *http.Client
}

func NewMahjongAPIClient() *MahjongAPIClient {
	return &MahjongAPIClient{
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

func (c *MahjongAPIClient) FetchRandomStartingHand(ctx context.Context) (*mahjong.MahjongStartingHand, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", mahjong.ErrAPIUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 上限より1バイト多く読み、巨大なレスポンスによるメモリ消費を抑える
	imageData, err := io.ReadAll(io.LimitReader(resp.Body, maxImageSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if len(imageData) > maxImageSize {
		return nil, fmt.Errorf("image response exceeds %d bytes", maxImageSize)
	}

	if len(imageData) == 0 {
		return nil, mahjong.ErrImageNotFound
	}

	// Content-Typeヘッダーから MIME タイプを取得
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png" // デフォルト値
	}

	return &mahjong.MahjongStartingHand{
		ImageData:   imageData,
		ContentType: contentType,
	}, nil
}
