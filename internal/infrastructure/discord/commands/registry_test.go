package commands

import (
	"slices"
	"testing"

	"github.com/bwmarrin/discordgo"
)

type namedCommand struct {
	SlashCommand
	name string
}

func (c namedCommand) Name() string { return c.name }
func (c namedCommand) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{Name: c.name}
}

func TestCommandRegistryListsCommandsInNameOrder(t *testing.T) {
	for _, names := range [][]string{nil, {"ping"}, {"version", "cat", "ping"}, {"ping", "cat", "version"}} {
		registry := NewCommandRegistry()
		for _, name := range names {
			registry.Register(namedCommand{name: name})
		}
		want := slices.Clone(names)
		slices.Sort(want)
		// Repeat to catch dependence on map iteration order.
		for range 20 {
			var commandNames, definitionNames []string
			for _, cmd := range registry.GetAllCommands() {
				commandNames = append(commandNames, cmd.Name())
			}
			for _, def := range registry.GetAllDefinitions() {
				definitionNames = append(definitionNames, def.Name)
			}
			if !slices.Equal(commandNames, want) || !slices.Equal(definitionNames, want) {
				t.Fatalf("commands = %v, definitions = %v, want %v", commandNames, definitionNames, want)
			}
		}
		if len(names) > 0 {
			commands := registry.GetAllCommands()
			commands[0] = nil
			if cmd, ok := registry.GetCommand(want[0]); !ok || cmd == nil {
				t.Fatal("modifying returned slice changed the registry")
			}
		}
	}
}
