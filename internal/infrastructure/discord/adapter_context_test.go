package discord

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

type contextTransport func(*http.Request) (*http.Response, error)

func (f contextTransport) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestAdapterPropagatesRequestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cause := errors.New("transport stopped")
	session, err := discordgo.New("Bot test")
	if err != nil {
		t.Fatal(err)
	}
	adapter := NewDiscordAdapter(session)
	tests := []struct {
		name string
		call func() error
	}{
		{"create", func() error { _, err := adapter.CreateTextChannelForVoice(ctx, "guild", "voice"); return err }},
		{"delete", func() error { return adapter.DeleteTextChannel(ctx, "text") }},
		{"voice exists", func() error { _, err := adapter.IsVoiceChannelExists(ctx, "voice"); return err }},
		{"text exists", func() error { _, err := adapter.IsTextChannelExists(ctx, "text"); return err }},
		{"add member", func() error { return adapter.AddMemberToTextChannel(ctx, "guild", "text", "user") }},
		{"remove member", func() error { return adapter.RemoveMemberFromTextChannel(ctx, "guild", "text", "user") }},
		{"list members", func() error { _, err := adapter.GetTextChannelMembers(ctx, "text"); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			session.Client = &http.Client{Transport: contextTransport(func(req *http.Request) (*http.Response, error) {
				calls++
				if req.Context() != ctx {
					t.Error("request did not receive caller context")
				}
				if tt.name == "create" && calls == 1 {
					return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"id":"voice","name":"test"}`))}, nil
				}
				return nil, cause
			})}
			if err := tt.call(); !errors.Is(err, cause) {
				t.Fatalf("error = %v, want transport cause", err)
			}
			if calls == 0 {
				t.Fatal("transport was not called")
			}
			if tt.name == "create" && calls != 2 {
				t.Fatalf("create requests = %d, want 2", calls)
			}
		})
	}
}
