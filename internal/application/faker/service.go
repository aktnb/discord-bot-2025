package faker

import (
	"context"

	"github.com/aktnb/discord-bot-go/internal/domain/faker"
)

type Service struct{}

func NewFakerService() *Service {
	return &Service{}
}

// GetRandomEpisode は Faker の伝説エピソードをランダムに1つ返す
func (s *Service) GetRandomEpisode(ctx context.Context) (string, error) {
	return faker.RandomEpisode(), nil
}
