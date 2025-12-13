package voicetext

import "errors"

var (
	ErrVoiceTextLinkNotFound = errors.New("voice text link not found")
	ErrInvalidID             = errors.New("invalid ID")
	ErrInvalidGuildID        = errors.New("invalid Guild ID")
	ErrInvalidVoiceChannelID = errors.New("invalid Voice Channel ID")
	ErrInvalidTextChannelID  = errors.New("invalid Text Channel ID")
)
