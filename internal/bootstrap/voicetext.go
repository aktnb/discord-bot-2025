package bootstrap

import (
	"github.com/aktnb/discord-bot-go/internal/application/voicetext"
	voicetextdomain "github.com/aktnb/discord-bot-go/internal/domain/voicetext"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/discord"
	"github.com/aktnb/discord-bot-go/internal/infrastructure/persistence"
	"github.com/aktnb/discord-bot-go/internal/interfaces/db"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
)

// newVoiceTextService はトランザクションごとのRepositoryと既存Adapterを接続します。
func newVoiceTextService(pool *pgxpool.Pool, session *discordgo.Session) *voicetext.Service {
	txm := persistence.NewTxManager(pool)

	newRepository := func(tx db.Tx) voicetextdomain.Repository {
		return persistence.NewVoiceTextLinkRepository(tx)
	}
	discordAdapter := discord.NewDiscordAdapter(session)
	vtlService := voicetext.NewVoiceTextService(newRepository, txm, discordAdapter)

	return vtlService
}
