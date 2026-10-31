package bootstrap

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands"
	"github.com/bwmarrin/discordgo"
)

func TestCommandRegistry(t *testing.T) {
	registry := newCommandRegistry("test-version")
	want := []string{"cat", "collatz", "dog", "faker", "ichiro", "jeff-dean", "mahjong", "omikuji", "ping", "version", "yamada"}
	var got []string
	for _, cmd := range registry.GetAllCommands() {
		got = append(got, cmd.Name())
		if definition := cmd.ToDiscordCommand(); definition.Name != cmd.Name() {
			t.Errorf("command %q has definition name %q", cmd.Name(), definition.Name)
		}
		guildCmd, guildSpecific := cmd.(commands.GuildSlashCommand)
		if cmd.Name() == "yamada" {
			if !guildSpecific {
				t.Fatal("yamada must remain guild-specific")
			}
			if ids := guildCmd.GuildIDs(); !reflect.DeepEqual(ids, []string{"1128971828644294666"}) {
				t.Errorf("yamada guild IDs = %v", ids)
			}
		} else if guildSpecific {
			t.Errorf("%s must remain global", cmd.Name())
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("registered commands = %v, want %v", got, want)
	}
}

type responseTransport func(*http.Request) (*http.Response, error)

func (f responseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestVersionCommandUsesBootstrapVersion(t *testing.T) {
	const version = "v-bootstrap-test"
	cmd, ok := newCommandRegistry(version).GetCommand("version")
	if !ok {
		t.Fatal("version command is missing")
	}
	session, err := discordgo.New("Bot test")
	if err != nil {
		t.Fatal(err)
	}
	called := false
	session.Client = &http.Client{Transport: responseTransport(func(req *http.Request) (*http.Response, error) {
		called = true
		var response discordgo.InteractionResponse
		if err := json.NewDecoder(req.Body).Decode(&response); err != nil {
			t.Fatal(err)
		}
		if response.Data == nil || response.Data.Content != version {
			t.Errorf("version response = %+v, want %q", response.Data, version)
		}
		return &http.Response{StatusCode: http.StatusNoContent, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
	interaction := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{ID: "id", Token: "token"}}
	if err := cmd.Handle(context.Background(), session, interaction); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("version command did not respond")
	}
}
