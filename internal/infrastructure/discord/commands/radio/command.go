package radio

import (
	"context"
	"fmt"
	"log"
	"strings"

	appradio "github.com/aktnb/discord-bot-go/internal/application/radio"
	"github.com/bwmarrin/discordgo"
)

// RadioCommand は /radio スラッシュコマンドを実装する
type RadioCommand struct {
	service  *appradio.Service
	sessions *sessionManager
}

func NewRadioCommand(service *appradio.Service) *RadioCommand {
	return &RadioCommand{
		service:  service,
		sessions: newSessionManager(),
	}
}

func (c *RadioCommand) Name() string { return "radio" }

func (c *RadioCommand) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "radio",
		Description: "Radiko のラジオを再生します",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "play",
				Description: "ラジオを再生します（ボイスチャンネルに参加してから実行してください）",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:         "station",
						Description:  "ラジオ局を選択してください",
						Type:         discordgo.ApplicationCommandOptionString,
						Required:     true,
						Autocomplete: true,
					},
				},
			},
			{
				Name:        "stop",
				Description: "ラジオを停止します",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
	}
}

func (c *RadioCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		return c.respond(s, i, "サブコマンドを指定してください。")
	}

	switch options[0].Name {
	case "play":
		return c.handlePlay(ctx, s, i, options[0].Options)
	case "stop":
		return c.handleStop(ctx, s, i)
	default:
		return c.respond(s, i, "不明なサブコマンドです。")
	}
}

func (c *RadioCommand) handlePlay(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) error {
	// ユーザーのボイスチャンネルを検索
	guildID := i.GuildID
	voiceChannelID, err := findUserVoiceChannel(s, guildID, i.Member.User.ID)
	if err != nil || voiceChannelID == "" {
		return c.respond(s, i, "ボイスチャンネルに参加してから実行してください。")
	}

	// station オプションを取得
	var stationID string
	for _, opt := range opts {
		if opt.Name == "station" {
			stationID = opt.StringValue()
		}
	}
	if stationID == "" {
		return c.respond(s, i, "ラジオ局を指定してください。")
	}

	// 既に再生中の場合は停止
	if _, ok := c.sessions.get(guildID); ok {
		c.sessions.stop(guildID)
	}

	// 即時応答（ボイス接続などに時間がかかるため Defer）
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		return fmt.Errorf("defer response: %w", err)
	}

	// ストリーム情報取得
	streamURL, authToken, err := c.service.GetStreamInfo(ctx, stationID)
	if err != nil {
		log.Printf("radio: GetStreamInfo: %v", err)
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "ラジオ局の情報を取得できませんでした。",
		})
		return err
	}

	// ボイスチャンネルに接続
	vc, err := s.ChannelVoiceJoin(ctx, guildID, voiceChannelID, false, true)
	if err != nil {
		log.Printf("radio: voice join: %v", err)
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "ボイスチャンネルに接続できませんでした。",
		})
		return err
	}

	// セッション開始
	playCtx, cancel := context.WithCancel(context.Background())
	c.sessions.set(guildID, &session{vc: vc, cancel: cancel})

	// ラジオ局名を取得してメッセージ表示
	stationName := stationID
	if stations, err := c.service.GetStations(ctx); err == nil {
		for _, st := range stations {
			if st.ID == stationID {
				stationName = st.Name
				break
			}
		}
	}

	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf("📻 **%s** を再生中...", stationName),
	})
	if err != nil {
		log.Printf("radio: followup message: %v", err)
	}

	// バックグラウンドでストリーミング開始
	go func() {
		defer c.sessions.stop(guildID)
		playStream(playCtx, vc, streamURL, authToken)
	}()

	return nil
}

func (c *RadioCommand) handleStop(_ context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guildID := i.GuildID
	if _, ok := c.sessions.get(guildID); !ok {
		return c.respond(s, i, "現在ラジオを再生していません。")
	}
	c.sessions.stop(guildID)
	return c.respond(s, i, "⏹ ラジオを停止しました。")
}

// HandleAutocomplete は station オプションのオートコンプリートを処理する
func (c *RadioCommand) HandleAutocomplete(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		return nil
	}

	// play サブコマンドの station オプションのみ対応
	if options[0].Name != "play" {
		return nil
	}

	subOpts := options[0].Options
	var input string
	for _, opt := range subOpts {
		if opt.Name == "station" && opt.Focused {
			input = strings.ToLower(opt.StringValue())
		}
	}

	stations, err := c.service.GetStations(ctx)
	if err != nil {
		log.Printf("radio: autocomplete GetStations: %v", err)
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionApplicationCommandAutocompleteResult,
			Data: &discordgo.InteractionResponseData{
				Choices: []*discordgo.ApplicationCommandOptionChoice{},
			},
		})
	}

	var choices []*discordgo.ApplicationCommandOptionChoice
	for _, st := range stations {
		if input == "" ||
			strings.Contains(strings.ToLower(st.Name), input) ||
			strings.Contains(strings.ToLower(st.ID), input) {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  st.Name,
				Value: st.ID,
			})
		}
		// Discord の選択肢は最大 25 件
		if len(choices) >= 25 {
			break
		}
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
}

func (c *RadioCommand) respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

// findUserVoiceChannel はユーザーが参加しているボイスチャンネル ID を返す
func findUserVoiceChannel(s *discordgo.Session, guildID, userID string) (string, error) {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return "", fmt.Errorf("get guild: %w", err)
	}
	for _, vs := range guild.VoiceStates {
		if vs.UserID == userID {
			return vs.ChannelID, nil
		}
	}
	return "", nil
}
