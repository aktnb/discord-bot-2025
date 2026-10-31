package omikuji

import (
	"context"
	"fmt"
	"log"

	appomikuji "github.com/aktnb/discord-bot-go/internal/application/omikuji"
	"github.com/bwmarrin/discordgo"
)

type OmikujiCommand struct {
	service *appomikuji.Service
}

func NewOmikujiCommand(service *appomikuji.Service) *OmikujiCommand {
	return &OmikujiCommand{
		service: service,
	}
}

func (c *OmikujiCommand) Name() string {
	return "omikuji"
}

func (c *OmikujiCommand) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "今日の運勢を占います（同じ日は同じ結果になります）",
	}
}

func (c *OmikujiCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
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
	fortune, err := c.service.DrawFortune(ctx, userID)
	if err != nil {
		log.Printf("Error drawing fortune: %v", err)
		// ユーザーにエラーメッセージを返す
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "おみくじを引けませんでした。もう一度お試しください。",
			},
		}, discordgo.WithContext(ctx))
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
	}, discordgo.WithContext(ctx))
	if err != nil {
		log.Printf("Error responding to omikuji: %v", err)
		return err
	}

	return nil
}
