package yamada

import (
	"context"
	"fmt"
	"log"

	appyamada "github.com/aktnb/discord-bot-go/internal/application/yamada"
	"github.com/bwmarrin/discordgo"
)

const guildID = "1128971828644294666"

// Command は山田嘘ニュースコマンド
type Command struct {
	service *appyamada.Service
}

func NewYamadaCommand(service *appyamada.Service) *Command {
	return &Command{service: service}
}

func (c *Command) Name() string {
	return "yamada"
}

// GuildIDs はコマンドを登録するギルド ID 一覧を返す
func (c *Command) GuildIDs() []string {
	return []string{guildID}
}

func (c *Command) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "山田に関する最新ニュースをお届けします",
	}
}

func (c *Command) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	episode, err := c.service.GetRandomEpisode(ctx)
	if err != nil {
		log.Printf("Error getting yamada episode: %v", err)
		return err
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("【山田速報】\n> %s", episode.Text),
		},
	})
	if err != nil {
		log.Printf("Error responding to yamada: %v", err)
		return err
	}

	return nil
}
