package voicetext

import (
	"time"

	"github.com/aktnb/discord-bot-go/internal/shared/discordid"

	"github.com/google/uuid"
)

type VoiceTextID string

type VoiceTextLink struct {
	id             VoiceTextID
	guildID        discordid.GuildID
	voiceChannelID discordid.VoiceChannelID
	textChannelID  discordid.TextChannelID
	createdAt      time.Time
	updatedAt      time.Time
}

func (v *VoiceTextLink) ID() VoiceTextID {
	return v.id
}

func (v *VoiceTextLink) GuildID() discordid.GuildID {
	return v.guildID
}

func (v *VoiceTextLink) VoiceChannelID() discordid.VoiceChannelID {
	return v.voiceChannelID
}

func (v *VoiceTextLink) TextChannelID() discordid.TextChannelID {
	return v.textChannelID
}

func (v *VoiceTextLink) CreatedAt() time.Time {
	return v.createdAt
}

func (v *VoiceTextLink) UpdatedAt() time.Time {
	return v.updatedAt
}

func (v *VoiceTextLink) ChangeTextChannel(textChannelID discordid.TextChannelID) error {
	v.textChannelID = textChannelID
	v.updatedAt = time.Now()
	return nil
}

func NewVoiceTextLink(
	guildId discordid.GuildID,
	voiceChannelId discordid.VoiceChannelID,
	textChannelId discordid.TextChannelID,
) (*VoiceTextLink, error) {
	if guildId == "" {
		return nil, ErrInvalidGuildID
	}
	if voiceChannelId == "" {
		return nil, ErrInvalidVoiceChannelID
	}
	return &VoiceTextLink{
		id:             VoiceTextID(uuid.New().String()),
		guildID:        guildId,
		voiceChannelID: voiceChannelId,
		textChannelID:  textChannelId,
		createdAt:      time.Now(),
		updatedAt:      time.Now(),
	}, nil
}

func RebuildVoiceTextLink(
	id VoiceTextID,
	guildId discordid.GuildID,
	voiceChannelId discordid.VoiceChannelID,
	textChannelId discordid.TextChannelID,
	createdAt, updatedAt time.Time,
) (*VoiceTextLink, error) {
	if id == "" {
		return nil, ErrInvalidID
	}
	if guildId == "" {
		return nil, ErrInvalidGuildID
	}
	if voiceChannelId == "" {
		return nil, ErrInvalidVoiceChannelID
	}
	if textChannelId == "" {
		return nil, ErrInvalidTextChannelID
	}

	return &VoiceTextLink{
		id:             id,
		guildID:        guildId,
		voiceChannelID: voiceChannelId,
		textChannelID:  textChannelId,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}, nil
}
