package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aktnb/discord-bot-go/internal/application/cat"
	"github.com/aktnb/discord-bot-go/internal/application/dog"
	"github.com/aktnb/discord-bot-go/internal/application/mahjong"
	"github.com/aktnb/discord-bot-go/internal/application/omikuji"
	"github.com/aktnb/discord-bot-go/internal/application/ping"
	"github.com/aktnb/discord-bot-go/internal/application/voicetext"
	"github.com/aktnb/discord-bot-go/internal/config"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/catapi"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/discord"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands"
	catcmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/cat"
	dogcmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/dog"
	mahjongcmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/mahjong"
	omikujicmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/omikuji"
	pingcmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/ping"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/dogapi"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/mahjongapi"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/persistence"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
)

// version はボットのバージョン情報を保持します。
// ビルド時に ldflags で上書き可能です: go build -ldflags "-X main.version=v1.0.0"
// デフォルトは "develop" です。
var version = "develop"

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
	pingCmd := pingcmd.NewPingCommand(pingService)
	registry.Register(pingCmd)

	// Cat command
	catAPIClient := catapi.NewCatAPIClient()
	catService := cat.NewCatService(catAPIClient)
	catCmd := catcmd.NewCatCommand(catService)
	registry.Register(catCmd)

	// Dog command
	dogAPIClient := dogapi.NewDogAPIClient()
	dogService := dog.NewDogService(dogAPIClient)
	dogCmd := dogcmd.NewDogCommand(dogService)
	registry.Register(dogCmd)

	// Mahjong command
	mahjongAPIClient := mahjongapi.NewMahjongAPIClient()
	mahjongService := mahjong.NewMahjongService(mahjongAPIClient)
	mahjongCmd := mahjongcmd.NewMahjongCommand(mahjongService)
	registry.Register(mahjongCmd)

	// Omikuji command
	omikujiService := omikuji.NewOmikujiService()
	omikujiCmd := omikujicmd.NewOmikujiCommand(omikujiService)
	registry.Register(omikujiCmd)

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
	if err := session.UpdateCustomStatus(version); err != nil {
		log.Printf("failed to update bot status: %v", err)
	}
	defer session.Close()

	log.Println("Bot is now running. Press CTRL+C to exit.")

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Bot is shutting down...")
}
