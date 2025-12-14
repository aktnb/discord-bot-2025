package catapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aktnb/discord-bot-go/internal/domain/cat"
)

const (
	baseURL        = "https://api.thecatapi.com/v1/images/search"
	requestTimeout = 10 * time.Second
)

// apiResponse はCat APIのレスポンス構造
type apiResponse struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type CatAPIClient struct {
	httpClient *http.Client
}

func NewCatAPIClient() *CatAPIClient {
	return &CatAPIClient{
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

func (c *CatAPIClient) FetchRandomImage(ctx context.Context) (*cat.CatImage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, cat.ErrAPIUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var responses []apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
		return nil, cat.ErrInvalidResponse
	}

	if len(responses) == 0 {
		return nil, cat.ErrImageNotFound
	}

	apiResp := responses[0]
	return &cat.CatImage{
		ID:     apiResp.ID,
		URL:    apiResp.URL,
		Width:  apiResp.Width,
		Height: apiResp.Height,
	}, nil
}
