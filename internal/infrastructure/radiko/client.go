package radiko

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	// Radikoの認証キー（公開されている固定値）
	authKey = "bcd151073c03b352e1ef2fd66c32209da9ca0afa"

	auth1URL   = "https://radiko.jp/v2/api/auth1"
	auth2URL   = "https://radiko.jp/v2/api/auth2"
	streamBase = "https://f-radiko.smartstream.ne.jp/%s/_definst_/simul-stream.stream/playlist.m3u8"
)

// Client はRadiko APIのクライアント
type Client struct {
	httpClient *http.Client
}

// NewClient はRadikoクライアントを生成する
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
	}
}

// Authenticate はRadikoの認証を行い、認証トークンを返す
func (c *Client) Authenticate(ctx context.Context) (string, error) {
	// Step 1: auth1 リクエスト
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, auth1URL, nil)
	if err != nil {
		return "", fmt.Errorf("auth1リクエスト作成に失敗しました: %w", err)
	}
	req.Header.Set("X-Radiko-App", "pc_html5")
	req.Header.Set("X-Radiko-App-Version", "0.0.1")
	req.Header.Set("X-Radiko-User", "test-stream")
	req.Header.Set("X-Radiko-Device", "pc")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth1リクエストに失敗しました: %w", err)
	}
	defer resp.Body.Close()
	// レスポンスボディを読み捨てて接続を解放する
	_, _ = io.Copy(io.Discard, resp.Body)

	authToken := resp.Header.Get("X-Radiko-AuthToken")
	keyLengthStr := resp.Header.Get("X-Radiko-KeyLength")
	keyOffsetStr := resp.Header.Get("X-Radiko-KeyOffset")

	if authToken == "" {
		return "", fmt.Errorf("認証トークンがレスポンスに含まれていません")
	}

	keyLength, err := strconv.Atoi(keyLengthStr)
	if err != nil {
		return "", fmt.Errorf("KeyLengthの解析に失敗しました: %w", err)
	}
	keyOffset, err := strconv.Atoi(keyOffsetStr)
	if err != nil {
		return "", fmt.Errorf("KeyOffsetの解析に失敗しました: %w", err)
	}

	// パーシャルキーを計算する
	partialKey := base64.StdEncoding.EncodeToString([]byte(authKey)[keyOffset : keyOffset+keyLength])

	// Step 2: auth2 リクエスト
	req2, err := http.NewRequestWithContext(ctx, http.MethodGet, auth2URL, nil)
	if err != nil {
		return "", fmt.Errorf("auth2リクエスト作成に失敗しました: %w", err)
	}
	req2.Header.Set("X-Radiko-AuthToken", authToken)
	req2.Header.Set("X-Radiko-PartialKey", partialKey)
	req2.Header.Set("X-Radiko-Device", "pc")
	req2.Header.Set("X-Radiko-User", "test-stream")

	resp2, err := c.httpClient.Do(req2)
	if err != nil {
		return "", fmt.Errorf("auth2リクエストに失敗しました: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp2.Body)
		return "", fmt.Errorf("auth2がステータス%dで失敗しました: %s", resp2.StatusCode, string(body))
	}
	_, _ = io.Copy(io.Discard, resp2.Body)

	return authToken, nil
}

// GetStreamURL はラジオ局のHLSストリームURLを返す
func (c *Client) GetStreamURL(stationID string) string {
	return fmt.Sprintf(streamBase, stationID)
}
