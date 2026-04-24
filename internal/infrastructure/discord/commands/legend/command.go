package legend

import (
	"context"
	"fmt"
	"log"

	"github.com/aktnb/discord-bot-go/internal/domain/legend"
	"github.com/bwmarrin/discordgo"
)

type EpisodeGetter func(ctx context.Context) (legend.Episode, error)

// Command はエピソードをランダムに返す伝説コマンドの共通実装
type Command struct {
	name        string
	description string
	prefix      string
	getEpisode  EpisodeGetter
}

func New(name, description, prefix string, getter EpisodeGetter) *Command {
	return &Command{
		name:        name,
		description: description,
		prefix:      prefix,
		getEpisode:  getter,
	}
}

func (c *Command) Name() string {
	return c.name
}

func (c *Command) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.name,
		Description: c.description,
	}
}

func (c *Command) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	episode, err := c.getEpisode(ctx)
	if err != nil {
		log.Printf("Error getting %s episode: %v", c.name, err)
		return err
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s伝説 その%d\n> %s", c.prefix, episode.Number, episode.Text),
		},
	})
	if err != nil {
		log.Printf("Error responding to %s: %v", c.name, err)
		return err
	}

	return nil
}
