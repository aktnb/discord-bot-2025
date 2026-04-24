package jeffdean

import (
	"context"

	domainjeffdean "github.com/aktnb/discord-bot-go/internal/domain/jeffdean"
	"github.com/aktnb/discord-bot-go/internal/domain/legend"
)

type Service struct{}

func NewJeffDeanService() *Service {
	return &Service{}
}

func (s *Service) GetRandomFact(ctx context.Context) (legend.Episode, error) {
	return domainjeffdean.RandomFact(), nil
}
