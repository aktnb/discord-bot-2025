package commands

import "github.com/bwmarrin/discordgo"

type CommandRegistry struct {
	commands map[string]SlashCommand
}

func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]SlashCommand),
	}
}

// Register は SlashCommand を登録する
func (r *CommandRegistry) Register(cmd SlashCommand) {
	r.commands[cmd.Name()] = cmd
}

// GetCommand は指定された名前のコマンドを取得する
func (r *CommandRegistry) GetCommand(name string) (SlashCommand, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

// GetAllDefinitions は全てのコマンド定義をDiscord API用に変換して返す
func (r *CommandRegistry) GetAllDefinitions() []*discordgo.ApplicationCommand {
	var definitions []*discordgo.ApplicationCommand
	for _, cmd := range r.commands {
		definitions = append(definitions, cmd.ToDiscordCommand())
	}
	return definitions
}
