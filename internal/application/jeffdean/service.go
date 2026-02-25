package jeffdean

import (
	"context"

	"github.com/aktnb/discord-bot-go/internal/domain/jeffdean"
)

type Service struct{}

func NewJeffDeanService() *Service {
	return &Service{}
}

// GetRandomFact は Jeff Dean の伝説をランダムに1つ返す
func (s *Service) GetRandomFact(ctx context.Context) (jeffdean.Fact, error) {
	return jeffdean.RandomFact(), nil
}
