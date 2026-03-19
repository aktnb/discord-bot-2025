package ichiro

import (
	"context"
	"fmt"
	"log"

	appichiro "github.com/aktnb/discord-bot-go/internal/application/ichiro"
	"github.com/bwmarrin/discordgo"
)

type IchiroCommand struct {
	service *appichiro.Service
}

func NewIchiroCommand(service *appichiro.Service) *IchiroCommand {
	return &IchiroCommand{
		service: service,
	}
}

func (c *IchiroCommand) Name() string {
	return "ichiro"
}

func (c *IchiroCommand) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "全盛期のイチロー伝説をランダムに紹介します",
	}
}

func (c *IchiroCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	episode, err := c.service.GetRandomEpisode(ctx)
	if err != nil {
		log.Printf("Error getting ichiro episode: %v", err)
		return err
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("イチロー伝説 その%d\n> %s", episode.Number, episode.Text),
		},
	})
	if err != nil {
		log.Printf("Error responding to ichiro: %v", err)
		return err
	}

	return nil
}
