package yamada

import (
	"context"

	domainyamada "github.com/aktnb/discord-bot-go/internal/domain/yamada"
	"github.com/aktnb/discord-bot-go/internal/domain/legend"
)

type Service struct{}

func NewYamadaService() *Service {
	return &Service{}
}

func (s *Service) GetRandomEpisode(ctx context.Context) (legend.Episode, error) {
	return domainyamada.RandomEpisode(), nil
}
