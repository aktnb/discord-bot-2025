package voicetext

import (
	"context"

	"github.com/aktnb/discord-bot-go/internal/interfaces/db"
	"github.com/aktnb/discord-bot-go/internal/shared/discordid"
)

type Repository interface {
	FindByVoiceChannel(ctx context.Context, guildId discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*VoiceTextLink, error)
	FindByTextChannel(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID) (*VoiceTextLink, error)
	Save(ctx context.Context, vtl *VoiceTextLink) error
	Delete(ctx context.Context, id VoiceTextID) error
}

type Repositories interface {
	VoiceTextLink(tx db.Tx) Repository
}
