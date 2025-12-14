package dogapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aktnb/discord-bot-go/internal/domain/dog"
)

const (
	baseURL        = "https://dog.ceo/api/breeds/image/random"
	requestTimeout = 10 * time.Second
)

// apiResponse はDog APIのレスポンス構造
type apiResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

type DogAPIClient struct {
	httpClient *http.Client
}

func NewDogAPIClient() *DogAPIClient {
	return &DogAPIClient{
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

func (c *DogAPIClient) FetchRandomImage(ctx context.Context) (*dog.DogImage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, dog.ErrAPIUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, dog.ErrAPIUnavailable
	}

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, dog.ErrInvalidResponse
	}

	if apiResp.Status != "success" || apiResp.Message == "" {
		return nil, dog.ErrImageNotFound
	}

	return &dog.DogImage{
		URL: apiResp.Message,
	}, nil
}
