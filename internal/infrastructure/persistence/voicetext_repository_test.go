package persistence

import (
	"errors"
	"testing"
	"time"

	"github.com/aktnb/discord-bot-go/internal/domain/voicetext"
	"github.com/jackc/pgx/v5"
)

type rowFunc func(...any) error

func (f rowFunc) Scan(dest ...any) error { return f(dest...) }

func TestScanVoiceTextLink(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := created.Add(time.Hour)
	row := rowFunc(func(dest ...any) error {
		if len(dest) != 6 {
			t.Fatalf("Scan destinations = %d, want 6", len(dest))
		}
		for i, value := range []string{"id", "guild", "voice", "text"} {
			*dest[i].(*string) = value
		}
		*dest[4].(*time.Time) = created
		*dest[5].(*time.Time) = updated
		return nil
	})
	link, err := scanVoiceTextLink(row)
	if err != nil {
		t.Fatal(err)
	}
	if link.ID() != "id" || link.GuildID() != "guild" || link.VoiceChannelID() != "voice" || link.TextChannelID() != "text" || !link.CreatedAt().Equal(created) || !link.UpdatedAt().Equal(updated) {
		t.Fatalf("unexpected link: %+v", link)
	}
}

func TestScanVoiceTextLinkPreservesErrors(t *testing.T) {
	for _, cause := range []error{pgx.ErrNoRows, errors.New("scan failed")} {
		link, err := scanVoiceTextLink(rowFunc(func(...any) error { return cause }))
		if link != nil || !errors.Is(err, cause) {
			t.Fatalf("got (%v, %v), want (nil, %v)", link, err, cause)
		}
	}
	link, err := scanVoiceTextLink(rowFunc(func(...any) error { return nil }))
	if link != nil || !errors.Is(err, voicetext.ErrInvalidID) {
		t.Fatalf("got (%v, %v), want invalid ID", link, err)
	}
}
