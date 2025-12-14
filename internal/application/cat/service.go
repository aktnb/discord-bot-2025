package cat

import (
	"context"

	"github.com/aktnb/discord-bot-go/internal/domain/cat"
)

type Service struct {
	repo cat.CatImageRepository
}

func NewCatService(repo cat.CatImageRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetRandomCatImage(ctx context.Context) (*cat.CatImage, error) {
	image, err := s.repo.FetchRandomImage(ctx)
	if err != nil {
		return nil, err
	}
	return image, nil
}
