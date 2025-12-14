package ping

import (
	"context"
	"log"

	"github.com/aktnb/discord-bot-go/internal/application/ping"
	"github.com/bwmarrin/discordgo"
)

type PingCommandDefinition struct{}

func NewPingCommandDefinition() *PingCommandDefinition {
	return &PingCommandDefinition{}
}

func (p *PingCommandDefinition) Name() string {
	return "ping"
}

func (p *PingCommandDefinition) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        p.Name(),
		Description: "Pong!と応答します",
	}
}

type PingCommandHandler struct {
	service *ping.Service
}

func NewPingCommandHandler(service *ping.Service) *PingCommandHandler {
	return &PingCommandHandler{
		service: service,
	}
}

func (h *PingCommandHandler) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	response, err := h.service.Ping(ctx)
	if err != nil {
		log.Printf("Error handling ping command: %v", err)
		return err
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
	if err != nil {
		log.Printf("Error responding to ping: %v", err)
		return err
	}

	return nil
}
