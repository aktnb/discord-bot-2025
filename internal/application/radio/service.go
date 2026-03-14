package radio

import (
	"context"
	"sync"

	"github.com/aktnb/discord-bot-go/internal/domain/radio"
)

// RadikoClient は Radiko API クライアントのインターフェース
type RadikoClient interface {
	Authenticate(ctx context.Context) (authToken string, areaID string, err error)
	FetchStations(ctx context.Context, areaID string) ([]*radio.Station, error)
	GetStreamURL(stationID string) string
}

// Service は Radiko ラジオ機能のアプリケーションサービス
type Service struct {
	client   RadikoClient
	mu       sync.Mutex
	authToken string
	areaID    string
	stations  []*radio.Station
}

func NewService(client RadikoClient) *Service {
	return &Service{client: client}
}

func (s *Service) ensureAuth(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.authToken != "" {
		return nil
	}
	token, area, err := s.client.Authenticate(ctx)
	if err != nil {
		return err
	}
	s.authToken = token
	s.areaID = area
	return nil
}

// GetStations はラジオ局一覧を返す（初回のみ API から取得してキャッシュ）
func (s *Service) GetStations(ctx context.Context) ([]*radio.Station, error) {
	if err := s.ensureAuth(ctx); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stations != nil {
		return s.stations, nil
	}
	stations, err := s.client.FetchStations(ctx, s.areaID)
	if err != nil {
		return nil, err
	}
	s.stations = stations
	return stations, nil
}

// GetStreamInfo は指定ラジオ局のストリーム URL と認証トークンを返す
func (s *Service) GetStreamInfo(ctx context.Context, stationID string) (streamURL string, authToken string, err error) {
	if err := s.ensureAuth(ctx); err != nil {
		return "", "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.client.GetStreamURL(stationID), s.authToken, nil
}

// RefreshAuth は認証トークンを再取得する
func (s *Service) RefreshAuth(ctx context.Context) error {
	s.mu.Lock()
	s.authToken = ""
	s.mu.Unlock()
	return s.ensureAuth(ctx)
}
