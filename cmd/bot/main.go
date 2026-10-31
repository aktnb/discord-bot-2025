package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aktnb/discord-bot-go/internal/application/cat"
	"github.com/aktnb/discord-bot-go/internal/application/collatz"
	"github.com/aktnb/discord-bot-go/internal/application/dog"
	"github.com/aktnb/discord-bot-go/internal/application/faker"
	"github.com/aktnb/discord-bot-go/internal/application/ichiro"
	"github.com/aktnb/discord-bot-go/internal/application/jeffdean"
	"github.com/aktnb/discord-bot-go/internal/application/mahjong"
	"github.com/aktnb/discord-bot-go/internal/application/omikuji"
	"github.com/aktnb/discord-bot-go/internal/application/ping"
	versionapp "github.com/aktnb/discord-bot-go/internal/application/version"
	"github.com/aktnb/discord-bot-go/internal/application/voicetext"
	"github.com/aktnb/discord-bot-go/internal/application/yamada"
	"github.com/aktnb/discord-bot-go/internal/config"
	voicetextdomain "github.com/aktnb/discord-bot-go/internal/domain/voicetext"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/catapi"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/discord"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands"
	catcmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/cat"
	collatzcmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/collatz"
	dogcmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/dog"
	fakercmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/faker"
	ichirocmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/ichiro"
	jeffdeancmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/jeffdean"
	mahjongcmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/mahjong"
	omikujicmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/omikuji"
	pingcmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/ping"
	versioncmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/version"
	yamadacmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/yamada"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/dogapi"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/mahjongapi"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/persistence"
	"github.com/aktnb/discord-bot-go/internal/interfaces/db"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
)

// version はボットのバージョン情報を保持します。
// ビルド時に ldflags で上書き可能です: go build -ldflags "-X main.version=v1.0.0"
// デフォルトは "develop" です。
var version = "develop"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := runWithConfig(ctx); err != nil {
		log.Printf("bot stopped: %v", err)
		os.Exit(1)
	}
}

func runWithConfig(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	return run(ctx, cfg)
}

func run(ctx context.Context, cfg config.Config) error {

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
	txm := persistence.NewTxManager(pool)

	newRepository := func(tx db.Tx) voicetextdomain.Repository {
		return persistence.NewVoiceTextLinkRepository(tx)
	}
	discordAdapter := discord.NewDiscordAdapter(session)
	vtlService := voicetext.NewVoiceTextService(newRepository, txm, discordAdapter)

	// Command registry
	registry := commands.NewCommandRegistry()

	// Version command
	versionService := versionapp.NewVersionService(version)
	versionCmd := versioncmd.NewVersionCommand(versionService)
	registry.Register(versionCmd)

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

	// Collatz command
	collatzService := collatz.NewCollatzService()
	collatzCmd := collatzcmd.NewCollatzCommand(collatzService)
	registry.Register(collatzCmd)

	// Faker command
	fakerService := faker.NewFakerService()
	fakerCmd := fakercmd.NewFakerCommand(fakerService)
	registry.Register(fakerCmd)

	// Ichiro command
	ichiroService := ichiro.NewIchiroService()
	ichiroCmd := ichirocmd.NewIchiroCommand(ichiroService)
	registry.Register(ichiroCmd)

	// Jeff Dean command
	jeffDeanService := jeffdean.NewJeffDeanService()
	jeffDeanCmd := jeffdeancmd.NewJeffDeanCommand(jeffDeanService)
	registry.Register(jeffDeanCmd)

	// Yamada command (guild-specific)
	yamadaService := yamada.NewYamadaService()
	yamadaCmd := yamadacmd.NewYamadaCommand(yamadaService)
	registry.Register(yamadaCmd)

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
	defer session.Close()

	log.Println("Bot is now running. Press CTRL+C to exit.")

	// Wait for interrupt signal
	<-ctx.Done()

	log.Println("Bot is shutting down...")
	return nil
}
