package discord

import (
	"context"
	"log"

	"github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands"
	"github.com/bwmarrin/discordgo"
)

type InteractionCreateHandler struct {
	registry *commands.CommandRegistry
}

func NewInteractionCreateHandler(registry *commands.CommandRegistry) *InteractionCreateHandler {
	return &InteractionCreateHandler{
		registry: registry,
	}
}

func (h *InteractionCreateHandler) Handle() func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			h.routeApplicationCommand(s, i)
		default:
			log.Printf("Unsupported interaction type: %v", i.Type)
		}
	}
}

func (h *InteractionCreateHandler) routeApplicationCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	commandName := i.ApplicationCommandData().Name

	cmd, ok := h.registry.GetCommand(commandName)
	if !ok {
		log.Printf("Unknown command: %s", commandName)
		return
	}

	if err := cmd.Handle(context.Background(), s, i); err != nil {
		log.Printf("Error handling command %s: %v", commandName, err)
	}
}
