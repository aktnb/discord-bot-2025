package mahjong

import (
	"bytes"
	"context"
	"log"

	appmahjong "github.com/aktnb/discord-bot-go/internal/application/mahjong"
	"github.com/bwmarrin/discordgo"
)

type MahjongCommandDefinition struct{}

func NewMahjongCommandDefinition() *MahjongCommandDefinition {
	return &MahjongCommandDefinition{}
}

func (m *MahjongCommandDefinition) Name() string {
	return "mahjong"
}

func (m *MahjongCommandDefinition) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        m.Name(),
		Description: "ランダムな麻雀の配牌を表示します",
	}
}

type MahjongCommandHandler struct {
	service *appmahjong.Service
}

func NewMahjongCommandHandler(service *appmahjong.Service) *MahjongCommandHandler {
	return &MahjongCommandHandler{
		service: service,
	}
}

func (h *MahjongCommandHandler) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// API呼び出しに時間がかかる可能性があるため、応答を遅延させる
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("Error deferring response: %v", err)
		return err
	}

	hand, err := h.service.GetRandomStartingHand(ctx)
	if err != nil {
		log.Printf("Error fetching mahjong image: %v", err)
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "麻雀の配牌を取得できませんでした。もう一度お試しください。",
		})
		return err
	}

	// 画像バイナリデータをBytesReaderに変換してファイルとして添付
	imageReader := bytes.NewReader(hand.ImageData)

	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Files: []*discordgo.File{
			{
				Name:   "mahjong-starting-hand.png",
				Reader: imageReader,
			},
		},
	})
	if err != nil {
		log.Printf("Error sending mahjong image: %v", err)
		return err
	}

	return nil
}
