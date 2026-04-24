package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type CommandRegistrar struct {
	session  *discordgo.Session
	registry *CommandRegistry
}

func NewRegistrar(session *discordgo.Session, registry *CommandRegistry) *CommandRegistrar {
	return &CommandRegistrar{
		session:  session,
		registry: registry,
	}
}

// RegisterApplicationCommands はグローバルコマンドとギルド固有コマンドを登録する
func (r *CommandRegistrar) RegisterApplicationCommands() error {
	var globalDefs []*discordgo.ApplicationCommand
	guildDefs := make(map[string][]*discordgo.ApplicationCommand)

	for _, cmd := range r.registry.GetAllCommands() {
		if guildCmd, ok := cmd.(GuildSlashCommand); ok {
			for _, guildID := range guildCmd.GuildIDs() {
				guildDefs[guildID] = append(guildDefs[guildID], cmd.ToDiscordCommand())
			}
		} else {
			globalDefs = append(globalDefs, cmd.ToDiscordCommand())
		}
	}

	if len(globalDefs) > 0 {
		log.Printf("Registering %d application commands globally...", len(globalDefs))
		if _, err := r.session.ApplicationCommandBulkOverwrite(r.session.State.User.ID, "", globalDefs); err != nil {
			return err
		}
		log.Printf("Successfully registered %d application commands globally", len(globalDefs))
	}

	for guildID, defs := range guildDefs {
		log.Printf("Registering %d application commands for guild %s...", len(defs), guildID)
		if _, err := r.session.ApplicationCommandBulkOverwrite(r.session.State.User.ID, guildID, defs); err != nil {
			return err
		}
		log.Printf("Successfully registered %d application commands for guild %s", len(defs), guildID)
	}

	return nil
}
