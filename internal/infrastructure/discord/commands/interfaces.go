package commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// SlashCommand は Discord スラッシュコマンドの定義と処理を統合したインターフェース
type SlashCommand interface {
	// Name はコマンド名を返す
	Name() string
	// ToDiscordCommand は Discord API 用のコマンド定義を返す
	ToDiscordCommand() *discordgo.ApplicationCommand
	// Handle はコマンドの実行処理を行う
	Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error
}

// GuildSlashCommand はギルド固有のスラッシュコマンド
// このインターフェースを実装するコマンドは、指定されたギルドにのみ登録される
type GuildSlashCommand interface {
	SlashCommand
	// GuildIDs はコマンドを登録するギルド ID の一覧を返す
	GuildIDs() []string
}

// AutocompleteHandler はオートコンプリートに対応するコマンドが実装するオプショナルインターフェース
type AutocompleteHandler interface {
	HandleAutocomplete(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error
}
