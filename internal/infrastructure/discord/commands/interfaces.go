package commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

type CommandDefinition interface {
	Name() string
	ToDiscordCommand() *discordgo.ApplicationCommand
}

type CommandHandler interface {
	Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error
}
