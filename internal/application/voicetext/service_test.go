package voicetext

import (
	"context"
	"errors"
	"fmt"
	"testing"

	domain "github.com/aktnb/discord-bot-go/internal/domain/voicetext"
	"github.com/aktnb/discord-bot-go/internal/interfaces/db"
	"github.com/aktnb/discord-bot-go/internal/interfaces/discord"
	"github.com/aktnb/discord-bot-go/internal/shared/discordid"
)

type immediateTxManager struct{ db.TxManager }

func (immediateTxManager) WithTx(ctx context.Context, fn func(context.Context, db.Tx) error) error {
	return fn(ctx, nil)
}

type lookupRepository struct {
	domain.Repository
	err error
}

func (r lookupRepository) FindByVoiceChannel(context.Context, discordid.GuildID, discordid.VoiceChannelID) (*domain.VoiceTextLink, error) {
	return nil, r.err
}

type createFailingDiscord struct {
	discord.DiscordPort
	err error
}

func (d createFailingDiscord) CreateTextChannelForVoice(context.Context, discordid.GuildID, discordid.VoiceChannelID) (discordid.TextChannelID, error) {
	return "", d.err
}

func TestJoinVoiceRecognizesWrappedNotFound(t *testing.T) {
	createErr := errors.New("channel creation failed")
	lookupErr := errors.New("database unavailable")
	for _, tt := range []struct {
		name            string
		lookupErr, want error
	}{
		{"not found", domain.ErrVoiceTextLinkNotFound, createErr},
		{"wrapped not found", fmt.Errorf("lookup: %w", domain.ErrVoiceTextLinkNotFound), createErr},
		{"other error", lookupErr, lookupErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			service := NewVoiceTextService(func(db.Tx) domain.Repository {
				return lookupRepository{err: tt.lookupErr}
			}, immediateTxManager{}, createFailingDiscord{err: createErr})
			err := service.JoinVoice(context.Background(), JoinVoiceCommand{
				GuildID: "guild", VoiceChannelID: "voice", UserID: "user",
			})
			if !errors.Is(err, tt.want) {
				t.Fatalf("JoinVoice() error = %v, want %v", err, tt.want)
			}
		})
	}
}
