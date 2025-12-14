package discord

import (
	"context"
	"log"

	"github.com/aktnb/discord-bot-go/internal/application/voicetext"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands"
	"github.com/bwmarrin/discordgo"
)

type ReadyHandler struct {
	service   *voicetext.Service
	registrar *commands.CommandRegistrar
}

func NewReadyHandler(service *voicetext.Service, registrar *commands.CommandRegistrar) *ReadyHandler {
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
