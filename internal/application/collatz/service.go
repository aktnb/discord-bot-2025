package collatz

import (
	"context"
	"fmt"
	"strings"

	"github.com/aktnb/discord-bot-go/internal/domain/collatz"
)

const (
	// Discord の文字列制限は2000文字
	maxMessageLength = 2000
)

type Service struct{}

func NewCollatzService() *Service {
	return &Service{}
}

// Calculate はコラッツ予想の計算を実行し、結果を文字列のスライスとして返す
// Discord の文字列制限に対応するため、必要に応じてメッセージを分割する
func (s *Service) Calculate(ctx context.Context, start int64) ([]string, error) {
	if start <= 0 {
		return nil, fmt.Errorf("開始値は正の整数である必要があります")
	}

	// コラッツ予想の計算
	sequence := collatz.NewSequence(start)
	sequence.Calculate()

	// 結果を文字列化
	messages := s.formatSequence(sequence)

	return messages, nil
}

// formatSequence は計算結果を Discord 用にフォーマットする
// メッセージが長すぎる場合は自動的に分割する
func (s *Service) formatSequence(sequence *collatz.Sequence) []string {
	var messages []string
	var currentMessage strings.Builder

	// ヘッダー
	header := fmt.Sprintf("🔢 **コラッツ予想シミュレーション**\n開始値: %d\nステップ数: %d\n\n",
		sequence.Steps[0].Value,
		sequence.Length()-1)

	currentMessage.WriteString(header)
	currentMessage.WriteString("**計算過程:**\n")

	// 各ステップを追加
	for i, step := range sequence.Steps {
		var line string
		if i == 0 {
			line = fmt.Sprintf("%d", step.Value)
		} else {
			line = fmt.Sprintf(" → %d", step.Value)
		}

		// メッセージが制限を超える場合は分割
		if currentMessage.Len()+len(line) > maxMessageLength {
			messages = append(messages, currentMessage.String())
			currentMessage.Reset()
			currentMessage.WriteString("**（続き）**\n")
			// 前の値を含めて続きを書く（連続性を保つため）
			if i > 0 {
				line = fmt.Sprintf("%d → %d", sequence.Steps[i-1].Value, step.Value)
			}
		}

		currentMessage.WriteString(line)
	}

	// 最後のメッセージを追加
	if currentMessage.Len() > 0 {
		messages = append(messages, currentMessage.String())
	}

	return messages
}
