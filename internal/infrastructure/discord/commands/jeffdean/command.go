package jeffdean

import (
	"context"
	"fmt"
	"log"

	appjeffdean "github.com/aktnb/discord-bot-go/internal/application/jeffdean"
	"github.com/bwmarrin/discordgo"
)

type JeffDeanCommand struct {
	service *appjeffdean.Service
}

func NewJeffDeanCommand(service *appjeffdean.Service) *JeffDeanCommand {
	return &JeffDeanCommand{
		service: service,
	}
}

func (c *JeffDeanCommand) Name() string {
	return "jeff-dean"
}

func (c *JeffDeanCommand) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "Googleのエンジニア Jeff Dean の伝説をランダムに紹介します",
	}
}

func (c *JeffDeanCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	fact, err := c.service.GetRandomFact(ctx)
	if err != nil {
		log.Printf("Error getting jeff-dean fact: %v", err)
		return err
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Jeff Dean伝説 その%d\n> %s", fact.Number, fact.Text),
		},
	})
	if err != nil {
		log.Printf("Error responding to jeff-dean: %v", err)
		return err
	}

	return nil
}
