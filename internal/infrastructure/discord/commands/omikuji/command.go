package omikuji

import (
	"context"
	"fmt"
	"log"

	appomikuji "github.com/aktnb/discord-bot-go/internal/application/omikuji"
	"github.com/bwmarrin/discordgo"
)

type OmikujiCommandDefinition struct{}

func NewOmikujiCommandDefinition() *OmikujiCommandDefinition {
	return &OmikujiCommandDefinition{}
}

func (o *OmikujiCommandDefinition) Name() string {
	return "omikuji"
}

func (o *OmikujiCommandDefinition) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        o.Name(),
		Description: "今日の運勢を占います（同じ日は同じ結果になります）",
	}
}

type OmikujiCommandHandler struct {
	service *appomikuji.Service
}

func NewOmikujiCommandHandler(service *appomikuji.Service) *OmikujiCommandHandler {
	return &OmikujiCommandHandler{
		service: service,
	}
}

func (h *OmikujiCommandHandler) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// ユーザーID取得（nilチェック）
	var userID string
	if i.Member != nil && i.Member.User != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	} else {
		log.Printf("Error: unable to get user ID from interaction")
		return fmt.Errorf("unable to get user ID")
	}

	// おみくじを引く
	fortune, err := h.service.DrawFortune(ctx, userID)
	if err != nil {
		log.Printf("Error drawing fortune: %v", err)
		// ユーザーにエラーメッセージを返す
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "おみくじを引けませんでした。もう一度お試しください。",
			},
		})
		return err
	}

	// レスポンスメッセージを作成
	responseContent := fmt.Sprintf("🎴 **今日のおみくじ結果** 🎴\n\n**%s**\n\n%s",
		fortune.Level.String(),
		fortune.Message,
	)

	// 即座に応答
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: responseContent,
		},
	})
	if err != nil {
		log.Printf("Error responding to omikuji: %v", err)
		return err
	}

	return nil
}
