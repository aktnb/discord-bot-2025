package faker

import (
	"context"

	domainfaker "github.com/aktnb/discord-bot-go/internal/domain/faker"
	"github.com/aktnb/discord-bot-go/internal/domain/legend"
)

type Service struct{}

func NewFakerService() *Service {
	return &Service{}
}

func (s *Service) GetRandomEpisode(ctx context.Context) (legend.Episode, error) {
	return domainfaker.RandomEpisode(), nil
}
