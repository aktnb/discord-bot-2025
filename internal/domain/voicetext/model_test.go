package voicetext

import (
	"testing"
	"time"

	"github.com/aktnb/discord-bot-go/internal/shared/discordid"
)

func TestNewVoiceTextLink(t *testing.T) {
	tests := []struct {
		name           string
		guildID        discordid.GuildID
		voiceChannelID discordid.VoiceChannelID
		textChannelID  discordid.TextChannelID
		expectError    error
	}{
		{
			name:           "valid input with text channel",
			guildID:        "guild123",
			voiceChannelID: "voice456",
			textChannelID:  "text789",
			expectError:    nil,
		},
		{
			name:           "valid input without text channel",
			guildID:        "guild123",
			voiceChannelID: "voice456",
			textChannelID:  "",
			expectError:    nil,
		},
		{
			name:           "empty guild ID",
			guildID:        "",
			voiceChannelID: "voice456",
			textChannelID:  "text789",
			expectError:    ErrInvalidGuildID,
		},
		{
			name:           "empty voice channel ID",
			guildID:        "guild123",
			voiceChannelID: "",
			textChannelID:  "text789",
			expectError:    ErrInvalidVoiceChannelID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link, err := NewVoiceTextLink(tt.guildID, tt.voiceChannelID, tt.textChannelID)

			if tt.expectError != nil {
				if err != tt.expectError {
					t.Errorf("expected error %v, got %v", tt.expectError, err)
				}
				if link != nil {
					t.Error("expected nil link when error occurs")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if link == nil {
				t.Fatal("expected non-nil link")
			}

			// Verify fields
			if link.ID() == "" {
				t.Error("expected non-empty ID")
			}
			if link.GuildID() != tt.guildID {
				t.Errorf("expected GuildID %s, got %s", tt.guildID, link.GuildID())
			}
			if link.VoiceChannelID() != tt.voiceChannelID {
				t.Errorf("expected VoiceChannelID %s, got %s", tt.voiceChannelID, link.VoiceChannelID())
			}
			if link.TextChannelID() != tt.textChannelID {
				t.Errorf("expected TextChannelID %s, got %s", tt.textChannelID, link.TextChannelID())
			}
			if link.CreatedAt().IsZero() {
				t.Error("expected non-zero CreatedAt")
			}
			if link.UpdatedAt().IsZero() {
				t.Error("expected non-zero UpdatedAt")
			}
			// CreatedAt and UpdatedAt should be very close in time (within 1 second)
			timeDiff := link.UpdatedAt().Sub(link.CreatedAt())
			if timeDiff < 0 {
				timeDiff = -timeDiff
			}
			if timeDiff > time.Second {
				t.Errorf("expected CreatedAt and UpdatedAt to be close in time, difference was %v", timeDiff)
			}
		})
	}
}

func TestRebuildVoiceTextLink(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-time.Hour)

	tests := []struct {
		name           string
		id             VoiceTextID
		guildID        discordid.GuildID
		voiceChannelID discordid.VoiceChannelID
		textChannelID  discordid.TextChannelID
		createdAt      time.Time
		updatedAt      time.Time
		expectError    error
	}{
		{
			name:           "valid input",
			id:             "id123",
			guildID:        "guild123",
			voiceChannelID: "voice456",
			textChannelID:  "text789",
			createdAt:      earlier,
			updatedAt:      now,
			expectError:    nil,
		},
		{
			name:           "empty ID",
			id:             "",
			guildID:        "guild123",
			voiceChannelID: "voice456",
			textChannelID:  "text789",
			createdAt:      earlier,
			updatedAt:      now,
			expectError:    ErrInvalidID,
		},
		{
			name:           "empty guild ID",
			id:             "id123",
			guildID:        "",
			voiceChannelID: "voice456",
			textChannelID:  "text789",
			createdAt:      earlier,
			updatedAt:      now,
			expectError:    ErrInvalidGuildID,
		},
		{
			name:           "empty voice channel ID",
			id:             "id123",
			guildID:        "guild123",
			voiceChannelID: "",
			textChannelID:  "text789",
			createdAt:      earlier,
			updatedAt:      now,
			expectError:    ErrInvalidVoiceChannelID,
		},
		{
			name:           "empty text channel ID",
			id:             "id123",
			guildID:        "guild123",
			voiceChannelID: "voice456",
			textChannelID:  "",
			createdAt:      earlier,
			updatedAt:      now,
			expectError:    ErrInvalidTextChannelID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link, err := RebuildVoiceTextLink(
				tt.id,
				tt.guildID,
				tt.voiceChannelID,
				tt.textChannelID,
				tt.createdAt,
				tt.updatedAt,
			)

			if tt.expectError != nil {
				if err != tt.expectError {
					t.Errorf("expected error %v, got %v", tt.expectError, err)
				}
				if link != nil {
					t.Error("expected nil link when error occurs")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if link == nil {
				t.Fatal("expected non-nil link")
			}

			// Verify all fields are properly set
			if link.ID() != tt.id {
				t.Errorf("expected ID %s, got %s", tt.id, link.ID())
			}
			if link.GuildID() != tt.guildID {
				t.Errorf("expected GuildID %s, got %s", tt.guildID, link.GuildID())
			}
			if link.VoiceChannelID() != tt.voiceChannelID {
				t.Errorf("expected VoiceChannelID %s, got %s", tt.voiceChannelID, link.VoiceChannelID())
			}
			if link.TextChannelID() != tt.textChannelID {
				t.Errorf("expected TextChannelID %s, got %s", tt.textChannelID, link.TextChannelID())
			}
			if !link.CreatedAt().Equal(tt.createdAt) {
				t.Errorf("expected CreatedAt %v, got %v", tt.createdAt, link.CreatedAt())
			}
			if !link.UpdatedAt().Equal(tt.updatedAt) {
				t.Errorf("expected UpdatedAt %v, got %v", tt.updatedAt, link.UpdatedAt())
			}
		})
	}
}

func TestChangeTextChannel(t *testing.T) {
	// Create a link
	link, err := NewVoiceTextLink("guild123", "voice456", "text789")
	if err != nil {
		t.Fatalf("failed to create link: %v", err)
	}

	originalUpdatedAt := link.UpdatedAt()
	newTextChannelID := discordid.TextChannelID("text999")

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Change text channel
	err = link.ChangeTextChannel(newTextChannelID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify text channel ID was changed
	if link.TextChannelID() != newTextChannelID {
		t.Errorf("expected TextChannelID %s, got %s", newTextChannelID, link.TextChannelID())
	}

	// Verify UpdatedAt was changed
	if !link.UpdatedAt().After(originalUpdatedAt) {
		t.Error("expected UpdatedAt to be updated after change")
	}

	// Verify other fields remain unchanged
	if link.GuildID() != "guild123" {
		t.Error("GuildID should not change")
	}
	if link.VoiceChannelID() != "voice456" {
		t.Error("VoiceChannelID should not change")
	}
}

func TestGetterMethods(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-time.Hour)

	link, err := RebuildVoiceTextLink(
		"id123",
		"guild123",
		"voice456",
		"text789",
		earlier,
		now,
	)
	if err != nil {
		t.Fatalf("failed to rebuild link: %v", err)
	}

	// Test all getter methods
	t.Run("ID", func(t *testing.T) {
		if link.ID() != "id123" {
			t.Errorf("expected ID id123, got %s", link.ID())
		}
	})

	t.Run("GuildID", func(t *testing.T) {
		if link.GuildID() != "guild123" {
			t.Errorf("expected GuildID guild123, got %s", link.GuildID())
		}
	})

	t.Run("VoiceChannelID", func(t *testing.T) {
		if link.VoiceChannelID() != "voice456" {
			t.Errorf("expected VoiceChannelID voice456, got %s", link.VoiceChannelID())
		}
	})

	t.Run("TextChannelID", func(t *testing.T) {
		if link.TextChannelID() != "text789" {
			t.Errorf("expected TextChannelID text789, got %s", link.TextChannelID())
		}
	})

	t.Run("CreatedAt", func(t *testing.T) {
		if !link.CreatedAt().Equal(earlier) {
			t.Errorf("expected CreatedAt %v, got %v", earlier, link.CreatedAt())
		}
	})

	t.Run("UpdatedAt", func(t *testing.T) {
		if !link.UpdatedAt().Equal(now) {
			t.Errorf("expected UpdatedAt %v, got %v", now, link.UpdatedAt())
		}
	})
}
