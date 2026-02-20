package faker

import "testing"

func TestRandomEpisode(t *testing.T) {
	episode := RandomEpisode()

	if episode.Number < 1 || episode.Number > len(episodes) {
		t.Errorf("expected episode number between 1 and %d, got %d", len(episodes), episode.Number)
	}

	if episode.Text == "" {
		t.Error("expected non-empty episode text")
	}

	if episodes[episode.Number-1] != episode.Text {
		t.Errorf("expected episode text %q at index %d, got %q", episodes[episode.Number-1], episode.Number-1, episode.Text)
	}
}
