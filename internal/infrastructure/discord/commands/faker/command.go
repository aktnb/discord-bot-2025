package faker

import (
	appfaker "github.com/aktnb/discord-bot-go/internal/application/faker"
	legendcmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/legend"
)

func NewFakerCommand(service *appfaker.Service) *legendcmd.Command {
	return legendcmd.New(
		"faker",
		"LOL プロプレイヤー Faker の伝説エピソードをランダムに紹介します",
		"Faker",
		service.GetRandomEpisode,
	)
}
