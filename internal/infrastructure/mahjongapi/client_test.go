package mahjongapi

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"

	"github.com/aktnb/discord-bot-go/internal/domain/mahjong"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestFetchRandomStartingHandPreservesTransportErrors(t *testing.T) {
	transportErr := &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection failed")}
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded, transportErr} {
		t.Run(cause.Error(), func(t *testing.T) {
			client := &MahjongAPIClient{httpClient: &http.Client{
				Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
					return nil, cause
				}),
			}}
			hand, err := client.FetchRandomStartingHand(context.Background())
			if hand != nil {
				t.Fatalf("expected no hand, got %v", hand)
			}
			if !errors.Is(err, mahjong.ErrAPIUnavailable) || !errors.Is(err, cause) {
				t.Fatalf("expected APIUnavailable and cause %v, got %v", cause, err)
			}
			if cause == transportErr {
				var opErr *net.OpError
				if !errors.As(err, &opErr) || opErr != transportErr {
					t.Fatalf("expected original net.OpError, got %v", err)
				}
			}
		})
	}
}
