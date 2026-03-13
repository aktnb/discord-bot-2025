package radiko

import (
	"context"
	"fmt"
	"log"

	appradiko "github.com/aktnb/discord-bot-go/internal/application/radiko"
	domain "github.com/aktnb/discord-bot-go/internal/domain/radiko"
	"github.com/bwmarrin/discordgo"
)

// RadikoCommand はradiko選局・再生コマンド
type RadikoCommand struct {
	service *appradiko.Service
}

// NewRadikoCommand はRadikoCommandを生成する
func NewRadikoCommand(service *appradiko.Service) *RadikoCommand {
	return &RadikoCommand{service: service}
}

func (c *RadikoCommand) Name() string {
	return "radiko"
}

func (c *RadikoCommand) ToDiscordCommand() *discordgo.ApplicationCommand {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(domain.PredefinedStations))
	for i, s := range domain.PredefinedStations {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  s.Name,
			Value: s.ID,
		}
	}

	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "radikoでラジオを聴く",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "play",
				Description: "ラジオ局を選んで再生する（ボイスチャンネルに参加している必要があります）",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "station",
						Description: "ラジオ局",
						Required:    true,
						Choices:     choices,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "stop",
				Description: "ラジオの再生を停止してボイスチャンネルから退出する",
			},
		},
	}
}

func (c *RadikoCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		return nil
	}

	switch options[0].Name {
	case "play":
		return c.handlePlay(ctx, s, i, options[0].Options)
	case "stop":
		return c.handleStop(ctx, s, i)
	}
	return nil
}

func (c *RadikoCommand) handlePlay(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	options []*discordgo.ApplicationCommandInteractionDataOption,
) error {
	// 認証やストリーム開始に時間がかかるため、先にDeferredレスポンスを返す
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return fmt.Errorf("DeferredResponseに失敗しました: %w", err)
	}

	guildID := i.GuildID
	userID := i.Member.User.ID

	// ユーザーのボイスチャンネルを取得する
	vs, err := s.State.VoiceState(guildID, userID)
	if err != nil || vs.ChannelID == "" {
		_, followErr := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: "ボイスチャンネルに参加してからコマンドを実行してください。",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		return followErr
	}

	// ラジオ局IDを取得する
	if len(options) == 0 {
		return nil
	}
	stationID := options[0].StringValue()
	station, _ := domain.FindStation(stationID)

	// ラジオ再生を開始する
	if err := c.service.Play(ctx, guildID, vs.ChannelID, stationID); err != nil {
		log.Printf("[RadikoCommand] Play error: %v", err)
		_, followErr := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: fmt.Sprintf("再生に失敗しました: %v", err),
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		return followErr
	}

	_, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Content: fmt.Sprintf("%s の再生を開始しました。", station.Name),
		Flags:   discordgo.MessageFlagsEphemeral,
	})
	return err
}

func (c *RadikoCommand) handleStop(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	guildID := i.GuildID

	if !c.service.IsPlaying(guildID) {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "現在ラジオは再生していません。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	if err := c.service.Stop(guildID); err != nil {
		log.Printf("[RadikoCommand] Stop error: %v", err)
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("停止に失敗しました: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "ラジオを停止しました。",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
