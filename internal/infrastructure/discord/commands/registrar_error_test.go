package commands

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

type registrationTransport func(*http.Request) (*http.Response, error)

func (f registrationTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type guildCommand struct{ namedCommand }

func (guildCommand) GuildIDs() []string { return []string{"guild-id"} }

func TestRegistrationErrorContext(t *testing.T) {
	for _, guild := range []bool{false, true} {
		registry := NewCommandRegistry()
		cmd := namedCommand{name: "test"}
		want := "global"
		if guild {
			registry.Register(guildCommand{cmd})
			want = "guild-id"
		} else {
			registry.Register(cmd)
		}
		session, err := discordgo.New("Bot test")
		if err != nil {
			t.Fatal(err)
		}
		session.State.User = &discordgo.User{ID: "app"}
		cause := errors.New("request failed")
		session.Client = &http.Client{Transport: registrationTransport(func(*http.Request) (*http.Response, error) {
			return nil, cause
		})}
		err = NewRegistrar(session, registry).RegisterApplicationCommands()
		if !errors.Is(err, cause) || !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %v, want cause and %q", err, want)
		}
	}
}
