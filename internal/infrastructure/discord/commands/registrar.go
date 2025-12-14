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

// RegisterApplicationCommands registers all application commands globally
func (r *CommandRegistrar) RegisterApplicationCommands() error {
	commandDefs := r.registry.GetAllDefinitions()

	log.Printf("Registering %d application commands globally...", len(commandDefs))

	_, err := r.session.ApplicationCommandBulkOverwrite(r.session.State.User.ID, "", commandDefs)
	if err != nil {
		return err
	}

	log.Printf("Successfully registered %d application commands globally", len(commandDefs))
	return nil
}
