package radiko

import (
	"context"
	"fmt"

	domain "github.com/aktnb/discord-bot-go/internal/domain/radiko"
)

// VoicePlayerPort はボイスストリーミングのインターフェース
type VoicePlayerPort interface {
	Play(guildID, channelID, streamURL, authToken string) error
	Stop(guildID string) error
	IsPlaying(guildID string) bool
}

// RadikoClientPort はRadiko APIのインターフェース
type RadikoClientPort interface {
	Authenticate(ctx context.Context) (string, error)
	GetStreamURL(stationID string) string
}

// Service はRadikoストリーミングのアプリケーションサービス
type Service struct {
	player VoicePlayerPort
	radiko RadikoClientPort
}

// NewService はServiceを生成する
func NewService(player VoicePlayerPort, radiko RadikoClientPort) *Service {
	return &Service{player: player, radiko: radiko}
}

// Play は指定したラジオ局をボイスチャンネルで再生する
// ユーザーがボイスチャンネルに参加していない場合はエラーを返す
func (s *Service) Play(ctx context.Context, guildID, channelID, stationID string) error {
	station, ok := domain.FindStation(stationID)
	if !ok {
		return fmt.Errorf("不明なラジオ局: %s", stationID)
	}

	authToken, err := s.radiko.Authenticate(ctx)
	if err != nil {
		return fmt.Errorf("Radiko認証に失敗しました: %w", err)
	}

	streamURL := s.radiko.GetStreamURL(stationID)

	if err := s.player.Play(guildID, channelID, streamURL, authToken); err != nil {
		return fmt.Errorf("%sの再生を開始できませんでした: %w", station.Name, err)
	}

	return nil
}

// Stop は指定したギルドのラジオ再生を停止する
func (s *Service) Stop(guildID string) error {
	return s.player.Stop(guildID)
}

// IsPlaying は指定したギルドでラジオが再生中かどうかを返す
func (s *Service) IsPlaying(guildID string) bool {
	return s.player.IsPlaying(guildID)
}
