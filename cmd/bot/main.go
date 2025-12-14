package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aktnb/discord-bot-go/internal/application/ping"
	"github.com/aktnb/discord-bot-go/internal/application/voicetext"
	"github.com/aktnb/discord-bot-go/internal/config"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/discord"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands"
	pingcmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/ping"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/persistence"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	// Initialize Discord session
	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatalf("failed to create Discord session: %v", err)
	}

	// Add necessary intents
	session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildVoiceStates

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to create database connection pool: %v", err)
	}
	defer pool.Close()
	txm := persistence.NewTxManager(pool)

	vtlRepositories := persistence.NewVoiceTextLinkRepositoryFactory()
	discordAdapter := discord.NewDiscordAdapter(session)
	vtlService := voicetext.NewVoiceTextService(vtlRepositories, txm, discordAdapter)

	// Command registry
	registry := commands.NewCommandRegistry()

	// Ping command
	pingService := ping.NewPingService()
	pingDef := pingcmd.NewPingCommandDefinition()
	pingHandler := pingcmd.NewPingCommandHandler(pingService)

	registry.Register(pingDef, pingHandler)

	// Register handlers before opening session
	commandRegistrar := commands.NewRegistrar(session, registry)
	readyHandler := discord.NewReadyHandler(vtlService, commandRegistrar)
	interactionHandler := discord.NewInteractionCreateHandler(registry)
	voiceStateHandler := discord.NewVoiceStateUpdateHandler(vtlService)

	session.AddHandlerOnce(readyHandler.Handle())
	session.AddHandler(interactionHandler.Handle())
	session.AddHandler(voiceStateHandler.Handle())

	if err := session.Open(); err != nil {
		log.Fatalf("cannot open Discord session: %v", err)
	}
	defer session.Close()

	log.Println("Bot is now running. Press CTRL+C to exit.")

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Bot is shutting down...")
}
