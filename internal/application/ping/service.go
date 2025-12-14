package ping

import (
	"context"
)

type Service struct{}

func NewPingService() *Service {
	return &Service{}
}

func (s *Service) Ping(ctx context.Context) (string, error) {
	return "Pong!", nil
}
