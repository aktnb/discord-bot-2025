package mahjong

import (
	"context"

	"github.com/aktnb/discord-bot-go/internal/domain/mahjong"
)

type Service struct {
	repo mahjong.MahjongRepository
}

func NewMahjongService(repo mahjong.MahjongRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetRandomStartingHand(ctx context.Context) (*mahjong.MahjongStartingHand, error) {
	hand, err := s.repo.FetchRandomStartingHand(ctx)
	if err != nil {
		return nil, err
	}
	return hand, nil
}
