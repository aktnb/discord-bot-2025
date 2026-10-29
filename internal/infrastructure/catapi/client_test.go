package catapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/aktnb/discord-bot-go/internal/domain/cat"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestFetchRandomImagePreservesTransportErrors(t *testing.T) {
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded, errors.New("connection failed")} {
		t.Run(cause.Error(), func(t *testing.T) {
			client := &CatAPIClient{httpClient: &http.Client{
				Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
					return nil, cause
				}),
			}}

			_, err := client.FetchRandomImage(context.Background())
			if !errors.Is(err, cat.ErrAPIUnavailable) {
				t.Errorf("expected ErrAPIUnavailable, got %v", err)
			}
			if !errors.Is(err, cause) {
				t.Errorf("expected cause %v, got %v", cause, err)
			}
		})
	}
}

func TestFetchRandomImagePreservesDecodeErrors(t *testing.T) {
	tests := []struct {
		name  string
		body  string
		check func(*testing.T, error)
	}{
		{
			name: "empty response",
			body: "",
			check: func(t *testing.T, err error) {
				t.Helper()
				if !errors.Is(err, io.EOF) {
					t.Errorf("expected EOF, got %v", err)
				}
			},
		},
		{
			name: "invalid JSON",
			body: "invalid",
			check: func(t *testing.T, err error) {
				t.Helper()
				var syntaxErr *json.SyntaxError
				if !errors.As(err, &syntaxErr) {
					t.Errorf("expected json.SyntaxError, got %v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &CatAPIClient{httpClient: &http.Client{
				Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(tt.body)),
						Header:     make(http.Header),
					}, nil
				}),
			}}

			_, err := client.FetchRandomImage(context.Background())
			if !errors.Is(err, cat.ErrInvalidResponse) {
				t.Errorf("expected ErrInvalidResponse, got %v", err)
			}
			tt.check(t, err)
		})
	}
}
