package voicetext

import "github.com/aktnb/discord-bot-go/internal/shared/discordid"

type VoiceStateUpdateCommand struct {
	GuildID              discordid.GuildID
	BeforeVoiceChannelID *discordid.VoiceChannelID
	AfterVoiceChannelID  *discordid.VoiceChannelID
	UserID               discordid.UserID
}

type JoinVoiceCommand struct {
	GuildID        discordid.GuildID
	VoiceChannelID discordid.VoiceChannelID
	UserID         discordid.UserID
}

type LeaveVoiceCommand struct {
	GuildID        discordid.GuildID
	VoiceChannelID discordid.VoiceChannelID
	UserID         discordid.UserID
	IsLastMember   bool
}
