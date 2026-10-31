package bootstrap

import (
	"context"
	"fmt"
	"log"

	"github.com/aktnb/discord-bot-go/internal/config"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/discord"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Run は設定を読み込み、Botを起動し、Contextのキャンセルまで実行します。
func Run(ctx context.Context, version string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	return run(ctx, cfg, version)
}

func run(ctx context.Context, cfg config.Config, version string) error {
	// Initialize Discord session
	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return fmt.Errorf("failed to create Discord session: %w", err)
	}

	// Add necessary intents
	session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildVoiceStates

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to create database connection pool: %w", err)
	}
	defer pool.Close()
	// Open失敗時も途中まで確保されたGatewayリソースを解放します。
	// deferの逆順実行により、Discordを先に閉じてからDBプールを解放します。
	defer func() {
		if err := session.Close(); err != nil {
			log.Printf("failed to close Discord session: %v", err)
		}
	}()
	vtlService := newVoiceTextService(pool, session)
	registry := newCommandRegistry(version)

	// Register handlers before opening session
	commandRegistrar := commands.NewRegistrar(session, registry)
	readyHandler := discord.NewReadyHandler(vtlService, commandRegistrar)
	interactionHandler := discord.NewInteractionCreateHandler(registry)
	voiceStateHandler := discord.NewVoiceStateUpdateHandler(vtlService)

	session.AddHandlerOnce(readyHandler.Handle())
	session.AddHandler(interactionHandler.Handle())
	session.AddHandler(voiceStateHandler.Handle())

	if err := session.Open(); err != nil {
		return fmt.Errorf("cannot open Discord session: %w", err)
	}
	if err := session.UpdateCustomStatus(version); err != nil {
		log.Printf("failed to update bot status: %v", err)
	}

	log.Println("Bot is now running. Press CTRL+C to exit.")

	// Wait for interrupt signal
	<-ctx.Done()

	log.Println("Bot is shutting down...")
	return nil
}
