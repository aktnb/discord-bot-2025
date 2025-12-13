package discord

import (
	"context"
	"fmt"

	"github.com/aktnb/discord-bot-go/internal/shared/discordid"
	"github.com/bwmarrin/discordgo"
)

type DiscordAdapter struct {
	session *discordgo.Session
}

func NewDiscordAdapter(session *discordgo.Session) *DiscordAdapter {
	return &DiscordAdapter{session: session}
}

func (a *DiscordAdapter) CreateTextChannelForVoice(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (discordid.TextChannelID, error) {
	// ボイスチャンネル情報を取得
	voiceChannel, err := a.session.Channel(string(voiceChannelID))
	if err != nil {
		return discordid.TextChannelID(""), fmt.Errorf("failed to get voice channel: %w", err)
	}

	// チャンネル名を生成
	textChannelName := "txt-" + voiceChannel.Name

	// @everyone に対する権限設定（ViewChannel を拒否）
	permissionOverwrites := []*discordgo.PermissionOverwrite{
		{
			ID:   string(guildID), // @everyone ロールの ID は guildID と同じ
			Type: discordgo.PermissionOverwriteTypeRole,
			Deny: discordgo.PermissionViewChannel,
		},
	}

	// テキストチャンネル作成
	channel, err := a.session.GuildChannelCreateComplex(string(guildID), discordgo.GuildChannelCreateData{
		Name:                 textChannelName,
		Type:                 discordgo.ChannelTypeGuildText,
		ParentID:             voiceChannel.ParentID,
		PermissionOverwrites: permissionOverwrites,
	})
	if err != nil {
		return discordid.TextChannelID(""), fmt.Errorf("failed to create text channel: %w", err)
	}

	return discordid.TextChannelID(channel.ID), nil
}

func (a *DiscordAdapter) DeleteTextChannel(ctx context.Context, textChannelID discordid.TextChannelID) error {
	_, err := a.session.ChannelDelete(string(textChannelID))
	if err != nil {
		return fmt.Errorf("failed to delete text channel: %w", err)
	}
	return nil
}

func (a *DiscordAdapter) IsVoiceChannelExists(ctx context.Context, channelID discordid.VoiceChannelID) (bool, error) {
	_, err := a.session.Channel(string(channelID))
	if err != nil {
		if restErr, ok := err.(*discordgo.RESTError); ok {
			if restErr.Message != nil && restErr.Message.Code == discordgo.ErrCodeUnknownChannel {
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to check voice channel existence: %w", err)
	}
	return true, nil
}

func (a *DiscordAdapter) IsTextChannelExists(ctx context.Context, channelID discordid.TextChannelID) (bool, error) {
	_, err := a.session.Channel(string(channelID))
	if err != nil {
		if restErr, ok := err.(*discordgo.RESTError); ok {
			if restErr.Message != nil && restErr.Message.Code == discordgo.ErrCodeUnknownChannel {
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to check text channel existence: %w", err)
	}
	return true, nil
}

func (a *DiscordAdapter) AddMemberToTextChannel(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID, userID discordid.UserID) error {
	allow := int64(discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory)
	deny := int64(0)

	err := a.session.ChannelPermissionSet(
		string(textChannelID),
		string(userID),
		discordgo.PermissionOverwriteTypeMember,
		allow,
		deny,
	)
	if err != nil {
		return fmt.Errorf("failed to add member to text channel: %w", err)
	}
	return nil
}

func (a *DiscordAdapter) RemoveMemberFromTextChannel(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID, userID discordid.UserID) error {
	err := a.session.ChannelPermissionDelete(string(textChannelID), string(userID))
	if err != nil {
		return fmt.Errorf("failed to remove member from text channel: %w", err)
	}
	return nil
}

func (a *DiscordAdapter) GetVoiceChannelMemberCount(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (int, error) {
	guild, err := a.session.State.Guild(string(guildID))
	if err != nil {
		return 0, fmt.Errorf("failed to get guild: %w", err)
	}

	count := 0
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == string(voiceChannelID) {
			count++
		}
	}

	return count, nil
}
