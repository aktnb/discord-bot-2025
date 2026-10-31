package discord

import (
	"context"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
)

// Run with -race to verify that state reads synchronize with Gateway updates.
func TestAdapterStateReadsDuringUpdates(t *testing.T) {
	session, err := discordgo.New("Bot test")
	if err != nil {
		t.Fatal(err)
	}
	guild := &discordgo.Guild{ID: "guild", VoiceStates: []*discordgo.VoiceState{
		{GuildID: "guild", ChannelID: "voice", UserID: "user"},
	}}
	if err := session.State.GuildAdd(guild); err != nil {
		t.Fatal(err)
	}
	adapter := NewDiscordAdapter(session)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range 1000 {
			session.State.Lock()
			session.State.Guilds = []*discordgo.Guild{guild}
			guild.VoiceStates = []*discordgo.VoiceState{
				{GuildID: "guild", ChannelID: "voice", UserID: "user"},
			}
			session.State.Unlock()
		}
	}()
	defer wg.Wait()
	for range 1000 {
		ctx := context.Background()
		guilds, err := adapter.GetGuilds(ctx)
		if err != nil || len(guilds) != 1 || guilds[0] != "guild" {
			t.Fatalf("guilds=%v error=%v", guilds, err)
		}
		count, err := adapter.GetVoiceChannelMemberCount(ctx, "guild", "voice")
		if err != nil || count != 1 {
			t.Fatalf("count=%d error=%v", count, err)
		}
		states, err := adapter.GetGuildVoiceStates(ctx, "guild")
		if err != nil || len(states["voice"]) != 1 || states["voice"][0] != "user" {
			t.Fatalf("states=%v error=%v", states, err)
		}
	}
}
