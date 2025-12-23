package collatz

import (
	"context"
	"strings"
	"testing"
)

func TestCalculate(t *testing.T) {
	service := NewCollatzService()

	tests := []struct {
		name         string
		input        int64
		expectError  bool
		expectSteps  int // expected number of steps to reach 1
	}{
		{
			name:        "small number 7",
			input:       7,
			expectError: false,
			expectSteps: 17, // 7 → 22 → 11 → 34 → 17 → 52 → 26 → 13 → 40 → 20 → 10 → 5 → 16 → 8 → 4 → 2 → 1
		},
		{
			name:        "number 1 (already at target)",
			input:       1,
			expectError: false,
			expectSteps: 1,
		},
		{
			name:        "even number 8",
			input:       8,
			expectError: false,
			expectSteps: 4, // 8 → 4 → 2 → 1
		},
		{
			name:        "invalid number 0",
			input:       0,
			expectError: true,
		},
		{
			name:        "negative number",
			input:       -5,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := service.Calculate(context.Background(), tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(messages) == 0 {
				t.Errorf("expected at least one message but got none")
				return
			}

			// Check that messages are within Discord's character limit
			for i, msg := range messages {
				if len(msg) > 2000 {
					t.Errorf("message %d exceeds Discord's 2000 character limit: %d characters", i, len(msg))
				}
			}
		})
	}
}

func TestFormatSequence(t *testing.T) {
	service := NewCollatzService()

	// Test with a number that will generate a moderately long sequence
	messages, err := service.Calculate(context.Background(), 27)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(messages) == 0 {
		t.Fatal("expected at least one message but got none")
	}

	// All messages should be within Discord's limit
	for i, msg := range messages {
		if len(msg) > 2000 {
			t.Errorf("message %d exceeds 2000 character limit: %d", i, len(msg))
		}
	}
}

func TestMessageSplitting(t *testing.T) {
	service := NewCollatzService()

	// Test with a number that generates a very long sequence (97 is known to have 118 steps)
	messages, err := service.Calculate(context.Background(), 97)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("Number 97 generated %d message(s)", len(messages))

	// Verify each message is within limits
	for i, msg := range messages {
		if len(msg) > 2000 {
			t.Errorf("message %d exceeds 2000 character limit: got %d characters", i, len(msg))
		}
		t.Logf("Message %d: %d characters", i+1, len(msg))
	}

	// If multiple messages, verify continuation messages are properly formatted
	if len(messages) > 1 {
		for i := 1; i < len(messages); i++ {
			expectedStart := "**（続き）**"
			if !strings.HasPrefix(messages[i], expectedStart) {
				t.Errorf("continuation message %d should start with '%s'", i, expectedStart)
			}
		}
	}
}
