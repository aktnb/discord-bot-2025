package commands

import "testing"

func TestCommandRegistryZeroValue(t *testing.T) {
	var registry CommandRegistry
	if _, ok := registry.GetCommand("ping"); ok {
		t.Fatal("empty registry contains ping")
	}
	if len(registry.GetAllCommands()) != 0 || len(registry.GetAllDefinitions()) != 0 {
		t.Fatal("empty registry returned commands")
	}
	registry.Register(namedCommand{name: "ping"})
	registry.Register(namedCommand{name: "cat"})
	registry.Register(namedCommand{name: "ping"})
	if cmd, ok := registry.GetCommand("ping"); !ok || cmd.Name() != "ping" {
		t.Fatal("registered command not found")
	}
	defs := registry.GetAllDefinitions()
	if len(defs) != 2 || defs[0].Name != "cat" || defs[1].Name != "ping" {
		t.Fatalf("unexpected definitions: %v", defs)
	}
}
