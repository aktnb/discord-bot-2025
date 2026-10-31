package mahjongapi

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type countedBody struct {
	io.Reader
	read   int
	closed bool
}

func (b *countedBody) Read(p []byte) (int, error) {
	n, err := b.Reader.Read(p)
	b.read += n
	return n, err
}

func (b *countedBody) Close() error { b.closed = true; return nil }

func TestImageResponseSize(t *testing.T) {
	for _, size := range []int{0, 1, maxImageSize, maxImageSize + 10} {
		body := &countedBody{Reader: strings.NewReader(strings.Repeat("x", size))}
		client := &MahjongAPIClient{httpClient: &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: body}, nil
			}),
		}}
		hand, err := client.FetchRandomStartingHand(context.Background())
		wantErr := size == 0 || size > maxImageSize
		if (err != nil) != wantErr {
			t.Fatalf("size %d: error = %v", size, err)
		}
		if !wantErr && (len(hand.ImageData) != size || hand.ContentType != "image/png") {
			t.Fatalf("size %d: unexpected hand", size)
		}
		if body.read > maxImageSize+1 || !body.closed {
			t.Fatalf("size %d: read=%d closed=%v", size, body.read, body.closed)
		}
	}
}
