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
