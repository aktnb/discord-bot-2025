package collatz

import (
	"context"
	"fmt"
	"log"

	appcollatz "github.com/aktnb/discord-bot-go/internal/application/collatz"
	"github.com/bwmarrin/discordgo"
)

type CollatzCommand struct {
	service *appcollatz.Service
}

func NewCollatzCommand(service *appcollatz.Service) *CollatzCommand {
	return &CollatzCommand{
		service: service,
	}
}

func (c *CollatzCommand) Name() string {
	return "collatz"
}

func (c *CollatzCommand) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "コラッツ予想をシミュレーションします",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "number",
				Description: "開始する正の整数",
				Required:    true,
				MinValue:    func() *float64 { v := 1.0; return &v }(),
			},
		},
	}
}

func (c *CollatzCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// パラメータ取得
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		log.Printf("Error: no options provided")
		return fmt.Errorf("number parameter is required")
	}

	number := options[0].IntValue()

	// 計算処理に時間がかかる可能性があるため、応答を遅延させる
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("Error deferring response: %v", err)
		return err
	}

	// コラッツ予想の計算
	messages, err := c.service.Calculate(ctx, number)
	if err != nil {
		log.Printf("Error calculating collatz sequence: %v", err)
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("エラーが発生しました: %v", err),
		})
		return err
	}

	// 最初のメッセージを送信（Followup）
	if len(messages) > 0 {
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: messages[0],
		})
		if err != nil {
			log.Printf("Error sending first message: %v", err)
			return err
		}
	}

	// 残りのメッセージを通常のチャンネルメッセージとして送信
	for idx, message := range messages[1:] {
		_, err = s.ChannelMessageSend(i.ChannelID, message)
		if err != nil {
			log.Printf("Error sending message %d: %v", idx+2, err)
			return err
		}
	}

	return nil
}
