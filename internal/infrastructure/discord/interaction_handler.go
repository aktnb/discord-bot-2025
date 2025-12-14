package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type InteractionCreateHandler struct {
}

func NewInteractionCreateHandler() *InteractionCreateHandler {
	return &InteractionCreateHandler{}
}

func (h *InteractionCreateHandler) Handle() func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Currently no commands are implemented
		// When commands are added, handle them here based on i.ApplicationCommandData().Name
		log.Printf("Received interaction: %s", i.ApplicationCommandData().Name)
	}
}
