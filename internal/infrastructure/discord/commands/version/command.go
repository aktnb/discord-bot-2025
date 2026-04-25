package version

import (
	"context"
	"log"

	"github.com/aktnb/discord-bot-go/internal/application/version"
	"github.com/bwmarrin/discordgo"
)

type VersionCommand struct {
	service *version.Service
}

func NewVersionCommand(service *version.Service) *VersionCommand {
	return &VersionCommand{service: service}
}

func (c *VersionCommand) Name() string {
	return "version"
}

func (c *VersionCommand) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "ボットのバージョンを表示します",
	}
}

func (c *VersionCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	v, err := c.service.Version(ctx)
	if err != nil {
		log.Printf("Error handling version command: %v", err)
		return err
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: v,
		},
	})
	if err != nil {
		log.Printf("Error responding to version: %v", err)
		return err
	}

	return nil
}
