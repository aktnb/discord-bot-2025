package dog

import (
	"context"
	"log"

	appdog "github.com/aktnb/discord-bot-go/internal/application/dog"
	"github.com/bwmarrin/discordgo"
)

type DogCommand struct {
	service *appdog.Service
}

func NewDogCommand(service *appdog.Service) *DogCommand {
	return &DogCommand{
		service: service,
	}
}

func (c *DogCommand) Name() string {
	return "dog"
}

func (c *DogCommand) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "ランダムな犬の画像を表示します",
	}
}

func (c *DogCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// まず応答を遅延させる（API呼び出しに時間がかかる可能性があるため）
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("Error deferring response: %v", err)
		return err
	}

	image, err := c.service.GetRandomDogImage(ctx)
	if err != nil {
		log.Printf("Error fetching dog image: %v", err)
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "犬の画像を取得できませんでした。もう一度お試しください。",
		})
		return err
	}

	// 画像URLを直接返す
	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: image.URL,
	})
	if err != nil {
		log.Printf("Error sending dog image: %v", err)
		return err
	}

	return nil
}
