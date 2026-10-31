package discord

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/bwmarrin/discordgo"
)

type errorTransport struct{ err error }

func (tr errorTransport) RoundTrip(*http.Request) (*http.Response, error) { return nil, tr.err }

func TestChannelExistsRecognizesWrappedRESTErrors(t *testing.T) {
	for _, code := range []int{discordgo.ErrCodeUnknownChannel, discordgo.ErrCodeMissingAccess} {
		for _, wrapped := range []bool{false, true} {
			t.Run(fmt.Sprintf("code=%d/wrapped=%v", code, wrapped), func(t *testing.T) {
				restErr := &discordgo.RESTError{
					Response: &http.Response{Status: "404 Not Found"},
					Message:  &discordgo.APIErrorMessage{Code: code},
				}
				var cause error = restErr
				if wrapped {
					cause = fmt.Errorf("request failed: %w", cause)
				}
				session, err := discordgo.New("Bot test")
				if err != nil {
					t.Fatal(err)
				}
				session.Client = &http.Client{Transport: errorTransport{err: cause}}
				adapter := NewDiscordAdapter(session)
				checks := []func() (bool, error){
					func() (bool, error) { return adapter.IsVoiceChannelExists(context.Background(), "voice") },
					func() (bool, error) { return adapter.IsTextChannelExists(context.Background(), "text") },
				}
				for _, check := range checks {
					exists, err := check()
					if exists {
						t.Fatal("channel unexpectedly exists")
					}
					if code == discordgo.ErrCodeUnknownChannel {
						if err != nil {
							t.Fatalf("unknown channel: %v", err)
						}
					} else if !errors.Is(err, restErr) {
						t.Fatalf("error = %v, want original REST error", err)
					}
				}
			})
		}
	}
}

func TestDeleteTextChannelTreatsOnlyUnknownChannelAsSuccess(t *testing.T) {
	for _, code := range []int{discordgo.ErrCodeUnknownChannel, discordgo.ErrCodeMissingAccess} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			var cause error = &discordgo.RESTError{Response: &http.Response{Status: "404 Not Found"}, Message: &discordgo.APIErrorMessage{Code: code}}
			session, err := discordgo.New("Bot test")
			if err != nil {
				t.Fatal(err)
			}
			session.Client = &http.Client{Transport: errorTransport{err: fmt.Errorf("wrapped: %w", cause)}}
			err = NewDiscordAdapter(session).DeleteTextChannel(context.Background(), "text")
			if code == discordgo.ErrCodeUnknownChannel {
				if err != nil {
					t.Fatal(err)
				}
			} else if !errors.Is(err, cause) {
				t.Fatalf("lost failure: %v", err)
			}
		})
	}
}
