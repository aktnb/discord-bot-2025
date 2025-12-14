package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type CommandRegistrar struct {
	session *discordgo.Session
}

func NewCommandRegistrar(session *discordgo.Session) *CommandRegistrar {
	return &CommandRegistrar{
		session: session,
	}
}

// RegisterApplicationCommands registers all application commands globally
func (r *CommandRegistrar) RegisterApplicationCommands() error {
	commands := GetApplicationCommands()

	log.Printf("Registering %d application commands globally...", len(commands))

	_, err := r.session.ApplicationCommandBulkOverwrite(r.session.State.User.ID, "", commands)
	if err != nil {
		return err
	}

	log.Printf("Successfully registered %d application commands globally", len(commands))
	return nil
}

// RegisterGuildCommands registers all commands for a specific guild
func (r *CommandRegistrar) RegisterGuildCommands(guildID string) error {
	commands := GetApplicationCommands()

	log.Printf("Registering %d commands for guild %s...", len(commands), guildID)

	_, err := r.session.ApplicationCommandBulkOverwrite(r.session.State.User.ID, guildID, commands)
	if err != nil {
		return err
	}

	log.Printf("Successfully registered %d commands for guild %s", len(commands), guildID)
	return nil
}
