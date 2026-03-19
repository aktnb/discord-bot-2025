package ichiro

import (
	"context"

	"github.com/aktnb/discord-bot-go/internal/domain/ichiro"
)

type Service struct{}

func NewIchiroService() *Service {
	return &Service{}
}

// GetRandomEpisode は全盛期のイチロー伝説エピソードをランダムに1つ返す
func (s *Service) GetRandomEpisode(ctx context.Context) (ichiro.Episode, error) {
	return ichiro.RandomEpisode(), nil
}
