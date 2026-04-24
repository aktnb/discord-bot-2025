package jeffdean

import (
	appjeffdean "github.com/aktnb/discord-bot-go/internal/application/jeffdean"
	legendcmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/legend"
)

func NewJeffDeanCommand(service *appjeffdean.Service) *legendcmd.Command {
	return legendcmd.New(
		"jeff-dean",
		"Googleのエンジニア Jeff Dean の伝説をランダムに紹介します",
		"Jeff Dean",
		service.GetRandomFact,
	)
}
