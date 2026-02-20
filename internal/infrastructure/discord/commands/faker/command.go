package faker

import (
	"context"
	"log"

	appfaker "github.com/aktnb/discord-bot-go/internal/application/faker"
	"github.com/bwmarrin/discordgo"
)

type FakerCommand struct {
	service *appfaker.Service
}

func NewFakerCommand(service *appfaker.Service) *FakerCommand {
	return &FakerCommand{
		service: service,
	}
}

func (c *FakerCommand) Name() string {
	return "faker"
}

func (c *FakerCommand) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "LOL プロプレイヤー Faker の伝説エピソードをランダムに紹介します",
	}
}

func (c *FakerCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	episode, err := c.service.GetRandomEpisode(ctx)
	if err != nil {
		log.Printf("Error getting faker episode: %v", err)
		return err
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: episode,
		},
	})
	if err != nil {
		log.Printf("Error responding to faker: %v", err)
		return err
	}

	return nil
}
