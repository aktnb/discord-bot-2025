package commands

import (
	"cmp"
	"slices"

	"github.com/bwmarrin/discordgo"
)

type CommandRegistry struct {
	commands map[string]SlashCommand
}

func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{}
}

// Register は SlashCommand を登録する
func (r *CommandRegistry) Register(cmd SlashCommand) {
	if r.commands == nil {
		r.commands = make(map[string]SlashCommand)
	}
	r.commands[cmd.Name()] = cmd
}

// GetCommand は指定された名前のコマンドを取得する
func (r *CommandRegistry) GetCommand(name string) (SlashCommand, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

// GetAllCommands は全てのコマンドを名前順に返す
func (r *CommandRegistry) GetAllCommands() []SlashCommand {
	cmds := make([]SlashCommand, 0, len(r.commands))
	for _, cmd := range r.commands {
		cmds = append(cmds, cmd)
	}
	slices.SortFunc(cmds, func(a, b SlashCommand) int {
		return cmp.Compare(a.Name(), b.Name())
	})
	return cmds
}

// GetAllDefinitions は全てのコマンド定義を名前順にDiscord API用に変換して返す
func (r *CommandRegistry) GetAllDefinitions() []*discordgo.ApplicationCommand {
	var definitions []*discordgo.ApplicationCommand
	for _, cmd := range r.GetAllCommands() {
		definitions = append(definitions, cmd.ToDiscordCommand())
	}
	return definitions
}
