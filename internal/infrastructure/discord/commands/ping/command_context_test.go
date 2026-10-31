package ping

import (
	"context"
	"errors"
	"net/http"
	"testing"

	appping "github.com/aktnb/discord-bot-go/internal/application/ping"
	"github.com/bwmarrin/discordgo"
)

type contextTransport func(*http.Request) (*http.Response, error)

func (f contextTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestHandlePassesContextToDiscord(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cause := errors.New("transport failed")
	session, err := discordgo.New("Bot test")
	if err != nil {
		t.Fatal(err)
	}
	called := false
	session.Client = &http.Client{Transport: contextTransport(func(req *http.Request) (*http.Response, error) {
		called = true
		if req.Context() != ctx {
			t.Error("request did not receive caller context")
		}
		return nil, cause
	})}
	cmd := NewPingCommand(appping.NewPingService())
	interaction := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{ID: "id", Token: "token"}}
	if err := cmd.Handle(ctx, session, interaction); !errors.Is(err, cause) {
		t.Fatalf("error = %v, want transport cause", err)
	}
	if !called {
		t.Fatal("transport was not called")
	}
}
