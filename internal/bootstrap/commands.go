package bootstrap

import (
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
	"github.com/aktnb/discord-bot-go/internal/application/yamada"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/catapi"
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
)

// newCommandRegistry は全スラッシュコマンドを手動DIで構築・登録します。
func newCommandRegistry(version string) *commands.CommandRegistry {
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

	return registry
}
