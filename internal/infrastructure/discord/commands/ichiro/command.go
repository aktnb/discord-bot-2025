package ichiro

import (
	appichiro "github.com/aktnb/discord-bot-go/internal/application/ichiro"
	legendcmd "github.com/aktnb/discord-bot-go/internal/infrastructure/discord/commands/legend"
)

func NewIchiroCommand(service *appichiro.Service) *legendcmd.Command {
	return legendcmd.New(
		"ichiro",
		"全盛期のイチロー伝説をランダムに紹介します",
		"イチロー",
		service.GetRandomEpisode,
	)
}
