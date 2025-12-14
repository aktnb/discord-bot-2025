package commands

import "github.com/bwmarrin/discordgo"

type CommandRegistry struct {
	commands map[string]struct {
		definition CommandDefinition
		handler    CommandHandler
	}
}

func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]struct {
			definition CommandDefinition
			handler    CommandHandler
		}),
	}
}

func (r *CommandRegistry) Register(def CommandDefinition, handler CommandHandler) {
	r.commands[def.Name()] = struct {
		definition CommandDefinition
		handler    CommandHandler
	}{definition: def, handler: handler}
}

func (r *CommandRegistry) GetHandler(name string) (CommandHandler, bool) {
	cmd, ok := r.commands[name]
	if !ok {
		return nil, false
	}
	return cmd.handler, true
}

// GetAllDefinitions は全てのコマンド定義をDiscord API用に変換して返す
func (r *CommandRegistry) GetAllDefinitions() []*discordgo.ApplicationCommand {
	var definitions []*discordgo.ApplicationCommand
	for _, cmd := range r.commands {
		definitions = append(definitions, cmd.definition.ToDiscordCommand())
	}
	return definitions
}
