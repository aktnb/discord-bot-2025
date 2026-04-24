package legend

import "testing"

func TestRandom(t *testing.T) {
	episodes := []string{"aaa", "bbb", "ccc"}
	ep := Random(episodes)

	if ep.Number < 1 || ep.Number > len(episodes) {
		t.Errorf("expected number between 1 and %d, got %d", len(episodes), ep.Number)
	}
	if ep.Text == "" {
		t.Error("expected non-empty text")
	}
	if episodes[ep.Number-1] != ep.Text {
		t.Errorf("expected %q at index %d, got %q", episodes[ep.Number-1], ep.Number-1, ep.Text)
	}
}
