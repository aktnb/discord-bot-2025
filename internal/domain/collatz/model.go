package collatz

// Step はコラッツ予想の1ステップを表す
type Step struct {
	Value int64
}

// Sequence はコラッツ予想の計算過程全体を表す
type Sequence struct {
	Steps []Step
}

// NewSequence は新しいSequenceを生成する
func NewSequence(start int64) *Sequence {
	return &Sequence{
		Steps: []Step{{Value: start}},
	}
}

// Calculate はコラッツ予想の計算を実行する
// 偶数の場合: n / 2
// 奇数の場合: 3n + 1
// 1に到達するまで繰り返す
func (s *Sequence) Calculate() {
	current := s.Steps[0].Value

	for current != 1 {
		if current%2 == 0 {
			current = current / 2
		} else {
			current = current*3 + 1
		}
		s.Steps = append(s.Steps, Step{Value: current})
	}
}

// Length は計算ステップの長さを返す
func (s *Sequence) Length() int {
	return len(s.Steps)
}
