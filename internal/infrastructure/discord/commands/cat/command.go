package cat

import (
	"context"
	"log"

	appcat "github.com/aktnb/discord-bot-go/internal/application/cat"
	"github.com/bwmarrin/discordgo"
)

type CatCommandDefinition struct{}

func NewCatCommandDefinition() *CatCommandDefinition {
	return &CatCommandDefinition{}
}

func (c *CatCommandDefinition) Name() string {
	return "cat"
}

func (c *CatCommandDefinition) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "ランダムな猫の画像を表示します",
	}
}

type CatCommandHandler struct {
	service *appcat.Service
}

func NewCatCommandHandler(service *appcat.Service) *CatCommandHandler {
	return &CatCommandHandler{
		service: service,
	}
}

func (h *CatCommandHandler) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// まず応答を遅延させる（API呼び出しに時間がかかる可能性があるため）
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("Error deferring response: %v", err)
		return err
	}

	image, err := h.service.GetRandomCatImage(ctx)
	if err != nil {
		log.Printf("Error fetching cat image: %v", err)
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "猫の画像を取得できませんでした。もう一度お試しください。",
		})
		return err
	}

	// 画像URLを直接返す
	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: image.URL,
	})
	if err != nil {
		log.Printf("Error sending cat image: %v", err)
		return err
	}

	return nil
}
