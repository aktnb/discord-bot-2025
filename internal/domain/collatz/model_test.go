package collatz

import "testing"

func TestCollatzSequence(t *testing.T) {
	tests := []struct {
		name          string
		input         int64
		expectedSteps int
		finalValue    int64
	}{
		{
			name:          "number 1",
			input:         1,
			expectedSteps: 1,
			finalValue:    1,
		},
		{
			name:          "number 2",
			input:         2,
			expectedSteps: 2, // 2 → 1
			finalValue:    1,
		},
		{
			name:          "number 3",
			input:         3,
			expectedSteps: 8, // 3 → 10 → 5 → 16 → 8 → 4 → 2 → 1
			finalValue:    1,
		},
		{
			name:          "number 7",
			input:         7,
			expectedSteps: 17,
			finalValue:    1,
		},
		{
			name:          "number 8",
			input:         8,
			expectedSteps: 4, // 8 → 4 → 2 → 1
			finalValue:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sequence := NewSequence(tt.input)
			sequence.Calculate()

			if sequence.Length() != tt.expectedSteps {
				t.Errorf("expected %d steps, got %d", tt.expectedSteps, sequence.Length())
			}

			lastStep := sequence.Steps[len(sequence.Steps)-1]
			if lastStep.Value != tt.finalValue {
				t.Errorf("expected final value %d, got %d", tt.finalValue, lastStep.Value)
			}

			// Verify first step
			if sequence.Steps[0].Value != tt.input {
				t.Errorf("expected first step to be %d, got %d", tt.input, sequence.Steps[0].Value)
			}
		})
	}
}

func TestCollatzRules(t *testing.T) {
	// Test even number rule: n/2
	sequence := NewSequence(4)
	sequence.Calculate()

	// 4 → 2 → 1
	if len(sequence.Steps) != 3 {
		t.Errorf("expected 3 steps for input 4, got %d", len(sequence.Steps))
	}
	if sequence.Steps[1].Value != 2 {
		t.Errorf("expected second step to be 2, got %d", sequence.Steps[1].Value)
	}

	// Test odd number rule: 3n+1
	sequence = NewSequence(3)
	sequence.Calculate()

	// 3 → 10 → ... (3*3 + 1 = 10)
	if sequence.Steps[1].Value != 10 {
		t.Errorf("expected second step to be 10 (3*3+1), got %d", sequence.Steps[1].Value)
	}
}
