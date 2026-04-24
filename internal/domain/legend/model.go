package legend

import "math/rand/v2"

// Episode は伝説エピソード
type Episode struct {
	Number int
	Text   string
}

// Random はエピソード一覧からランダムに1つ返す
func Random(episodes []string) Episode {
	i := rand.IntN(len(episodes))
	return Episode{Number: i + 1, Text: episodes[i]}
}
