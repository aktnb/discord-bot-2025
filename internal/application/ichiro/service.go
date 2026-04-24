package ichiro

import (
	"context"

	domainiichiro "github.com/aktnb/discord-bot-go/internal/domain/ichiro"
	"github.com/aktnb/discord-bot-go/internal/domain/legend"
)

type Service struct{}

func NewIchiroService() *Service {
	return &Service{}
}

func (s *Service) GetRandomEpisode(ctx context.Context) (legend.Episode, error) {
	return domainiichiro.RandomEpisode(), nil
}
