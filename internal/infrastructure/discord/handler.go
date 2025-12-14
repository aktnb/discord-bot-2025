package discord

import (
	"context"
	"log"

	"github.com/aktnb/discord-bot-go/internal/application/voicetext"
	"github.com/aktnb/discord-bot-go/internal/shared/discordid"
	"github.com/bwmarrin/discordgo"
)

type ReadyHandler struct {
	service    *voicetext.Service
	registrar  *CommandRegistrar
}

func NewReadyHandler(service *voicetext.Service, registrar *CommandRegistrar) *ReadyHandler {
	return &ReadyHandler{
		service:   service,
		registrar: registrar,
	}
}

func (h *ReadyHandler) Handle() func(*discordgo.Session, *discordgo.Ready) {
	return func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is ready.")

		// Register application commands
		if err := h.registrar.RegisterApplicationCommands(); err != nil {
			log.Printf("Warning: command registration failed: %v", err)
			// コマンド登録失敗は警告のみで続行
		}

		// Sync voice-text links
		log.Println("Starting voice-text link synchronization...")
		if err := h.service.SyncVoiceTextLinks(context.Background()); err != nil {
			log.Printf("Warning: sync failed: %v", err)
			// 同期失敗は警告のみで続行（既存機能は動作）
		}
		log.Println("Voice-text link synchronization completed.")
	}
}

type InteractionCreateHandler struct {
	service *voicetext.Service
}

func NewInteractionCreateHandler(service *voicetext.Service) *InteractionCreateHandler {
	return &InteractionCreateHandler{service: service}
}

func (h *InteractionCreateHandler) Handle() func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Currently no commands are implemented
		// When commands are added, handle them here based on i.ApplicationCommandData().Name
		log.Printf("Received interaction: %s", i.ApplicationCommandData().Name)
	}
}

type VoiceStateUpdateHandler struct {
	service *voicetext.Service
}

func NewVoiceStateUpdateHandler(service *voicetext.Service) *VoiceStateUpdateHandler {
	return &VoiceStateUpdateHandler{service: service}
}

func (h *VoiceStateUpdateHandler) Handle() func(*discordgo.Session, *discordgo.VoiceStateUpdate) {
	return func(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
		if e.UserID == s.State.User.ID {
			return
		}

		var beforeChannelID *discordid.VoiceChannelID = nil
		var afterChannelID *discordid.VoiceChannelID = nil
		if e.BeforeUpdate != nil && e.BeforeUpdate.ChannelID != "" {
			id := discordid.VoiceChannelID(e.BeforeUpdate.ChannelID)
			beforeChannelID = &id
		}
		if e.ChannelID != "" {
			id := discordid.VoiceChannelID(e.ChannelID)
			afterChannelID = &id
		}

		log.Printf("VoiceStateUpdate: UserID=%s, GuildID=%s, BeforeVoiceChannelID=%v, AfterVoiceChannelID=%v",
			e.UserID,
			e.GuildID,
			beforeChannelID,
			afterChannelID,
		)

		cmd := voicetext.VoiceStateUpdateCommand{
			GuildID:              discordid.GuildID(e.GuildID),
			BeforeVoiceChannelID: beforeChannelID,
			AfterVoiceChannelID:  afterChannelID,
			UserID:               discordid.UserID(e.UserID),
		}

		if err := h.service.VoiceStateUpdate(context.Background(), cmd); err != nil {
			log.Printf("Error handling VoiceStateUpdate: %v", err)
		}
	}
}
