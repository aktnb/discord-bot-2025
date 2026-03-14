package radiko

import (
	"context"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aktnb/discord-bot-go/internal/domain/radio"
)

const (
	// authKey は Radiko 認証に使用する共通キー
	authKey = "bcd151073c03b352e1ef2fd66c32209da9ca0afa"

	auth1URL       = "https://radiko.jp/v2/api/auth1"
	auth2URL       = "https://radiko.jp/v2/api/auth2"
	stationListURL = "https://radiko.jp/v3/station/list/%s.xml"
	streamURLFmt   = "https://f-radiko.smartstream.ne.jp/%s/_definst_/simul-stream.stream/playlist.m3u8"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
	}
}

// Authenticate は Radiko の認証を行い、authToken と areaID を返す
func (c *Client) Authenticate(ctx context.Context) (authToken string, areaID string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, auth1URL, nil)
	if err != nil {
		return "", "", fmt.Errorf("auth1 request: %w", err)
	}
	req.Header.Set("X-Radiko-App", "pc_html5")
	req.Header.Set("X-Radiko-App-Version", "0.0.1")
	req.Header.Set("X-Radiko-User", "dummy_user")
	req.Header.Set("X-Radiko-Device", "pc")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("auth1: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	authToken = resp.Header.Get("X-Radiko-Authtoken")
	if authToken == "" {
		return "", "", fmt.Errorf("auth1: empty auth token")
	}

	var keyLength, keyOffset int
	fmt.Sscanf(resp.Header.Get("X-Radiko-Keylength"), "%d", &keyLength)
	fmt.Sscanf(resp.Header.Get("X-Radiko-Keyoffset"), "%d", &keyOffset)

	if keyLength == 0 {
		return "", "", fmt.Errorf("auth1: invalid key length")
	}

	keyBytes := []byte(authKey)
	if keyOffset+keyLength > len(keyBytes) {
		return "", "", fmt.Errorf("auth1: key offset/length out of range")
	}
	partialKey := base64.StdEncoding.EncodeToString(keyBytes[keyOffset : keyOffset+keyLength])

	req2, err := http.NewRequestWithContext(ctx, http.MethodGet, auth2URL, nil)
	if err != nil {
		return "", "", fmt.Errorf("auth2 request: %w", err)
	}
	req2.Header.Set("X-Radiko-App", "pc_html5")
	req2.Header.Set("X-Radiko-App-Version", "0.0.1")
	req2.Header.Set("X-Radiko-User", "dummy_user")
	req2.Header.Set("X-Radiko-Device", "pc")
	req2.Header.Set("X-Radiko-Authtoken", authToken)
	req2.Header.Set("X-Radiko-PartialKey", partialKey)

	resp2, err := c.httpClient.Do(req2)
	if err != nil {
		return "", "", fmt.Errorf("auth2: %w", err)
	}
	defer resp2.Body.Close()

	body, err := io.ReadAll(resp2.Body)
	if err != nil {
		return "", "", fmt.Errorf("auth2 read body: %w", err)
	}

	// レスポンスは "JP13,東京都" のような形式
	parts := strings.SplitN(strings.TrimSpace(string(body)), ",", 2)
	if len(parts) < 1 || parts[0] == "" {
		return "", "", fmt.Errorf("auth2: invalid response: %q", string(body))
	}
	areaID = strings.TrimSpace(parts[0])

	return authToken, areaID, nil
}

type stationsResponse struct {
	XMLName  xml.Name         `xml:"stations"`
	Stations []stationXMLElem `xml:"station"`
}

type stationXMLElem struct {
	ID   string `xml:"id"`
	Name string `xml:"name"`
}

// FetchStations は指定エリアのラジオ局一覧を取得する
func (c *Client) FetchStations(ctx context.Context, areaID string) ([]*radio.Station, error) {
	url := fmt.Sprintf(stationListURL, areaID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("station list request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("station list: %w", err)
	}
	defer resp.Body.Close()

	var xmlData stationsResponse
	if err := xml.NewDecoder(resp.Body).Decode(&xmlData); err != nil {
		return nil, fmt.Errorf("station list decode: %w", err)
	}

	stations := make([]*radio.Station, 0, len(xmlData.Stations))
	for _, s := range xmlData.Stations {
		if s.ID == "" {
			continue
		}
		stations = append(stations, &radio.Station{
			ID:   s.ID,
			Name: s.Name,
		})
	}
	return stations, nil
}

// GetStreamURL はラジオ局のライブ HLS ストリーム URL を返す
func (c *Client) GetStreamURL(stationID string) string {
	return fmt.Sprintf(streamURLFmt, stationID)
}
