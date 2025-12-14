package dog

import (
	"context"

	"github.com/aktnb/discord-bot-go/internal/domain/dog"
)

type Service struct {
	repo dog.DogImageRepository
}

func NewDogService(repo dog.DogImageRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetRandomDogImage(ctx context.Context) (*dog.DogImage, error) {
	image, err := s.repo.FetchRandomImage(ctx)
	if err != nil {
		return nil, err
	}
	return image, nil
}
