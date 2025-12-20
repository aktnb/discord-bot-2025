package ping

import (
	"context"
	"log"

	"github.com/aktnb/discord-bot-go/internal/application/ping"
	"github.com/bwmarrin/discordgo"
)

type PingCommand struct {
	service *ping.Service
}

func NewPingCommand(service *ping.Service) *PingCommand {
	return &PingCommand{
		service: service,
	}
}

func (c *PingCommand) Name() string {
	return "ping"
}

func (c *PingCommand) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "Pong!と応答します",
	}
}

func (c *PingCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	response, err := c.service.Ping(ctx)
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
