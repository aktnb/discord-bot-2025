package discord

import (
	"context"

	"github.com/aktnb/discord-bot-go/internal/shared/discordid"
)

type DiscordPort interface {
	CreateTextChannelForVoice(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (textChannelID discordid.TextChannelID, err error)
	DeleteTextChannel(ctx context.Context, textChannelID discordid.TextChannelID) error

	IsVoiceChannelExists(ctx context.Context, channelID discordid.VoiceChannelID) (bool, error)
	IsTextChannelExists(ctx context.Context, channelID discordid.TextChannelID) (bool, error)

	AddMemberToTextChannel(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID, userID discordid.UserID) error
	RemoveMemberFromTextChannel(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID, userID discordid.UserID) error

	GetVoiceChannelMemberCount(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (int, error)

	GetGuilds(ctx context.Context) ([]discordid.GuildID, error)
	GetGuildVoiceStates(ctx context.Context, guildID discordid.GuildID) (map[discordid.VoiceChannelID][]discordid.UserID, error)
	GetTextChannelMembers(ctx context.Context, textChannelID discordid.TextChannelID) ([]discordid.UserID, error)
}
