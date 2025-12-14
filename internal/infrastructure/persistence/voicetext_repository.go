package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/aktnb/discord-bot-go/internal/domain/voicetext"
	"github.com/aktnb/discord-bot-go/internal/interfaces/db"
	"github.com/aktnb/discord-bot-go/internal/shared/discordid"
	"github.com/jackc/pgx/v5"
)

type VoiceTextLinkRepositoryFactory struct{}

func NewVoiceTextLinkRepositoryFactory() *VoiceTextLinkRepositoryFactory {
	return &VoiceTextLinkRepositoryFactory{}
}

func (f *VoiceTextLinkRepositoryFactory) VoiceTextLink(tx db.Tx) voicetext.Repository {
	return NewVoiceTextLinkRepository(&tx)
}

type VoiceTextLinkRepository struct {
	tx db.Tx
}

func NewVoiceTextLinkRepository(tx *db.Tx) *VoiceTextLinkRepository {
	return &VoiceTextLinkRepository{
		tx: *tx,
	}
}

func (r *VoiceTextLinkRepository) FindByVoiceChannel(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
	query := `
		SELECT id, guild_id, voice_channel_id, text_channel_id, created_at, updated_at
		FROM voice_text_links
		WHERE guild_id = $1 AND voice_channel_id = $2
	`

	var (
		dbID             string
		dbGuildID        string
		dbVoiceChannelID string
		dbTextChannelID  string
		dbCreatedAt      time.Time
		dbUpdatedAt      time.Time
	)

	err := r.tx.QueryRow(ctx, query, string(guildID), string(voiceChannelID)).Scan(
		&dbID,
		&dbGuildID,
		&dbVoiceChannelID,
		&dbTextChannelID,
		&dbCreatedAt,
		&dbUpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, voicetext.ErrVoiceTextLinkNotFound
		}
		return nil, err
	}

	return voicetext.RebuildVoiceTextLink(
		voicetext.VoiceTextID(dbID),
		discordid.GuildID(dbGuildID),
		discordid.VoiceChannelID(dbVoiceChannelID),
		discordid.TextChannelID(dbTextChannelID),
		dbCreatedAt,
		dbUpdatedAt,
	)
}

func (r *VoiceTextLinkRepository) FindByTextChannel(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID) (*voicetext.VoiceTextLink, error) {
	query := `
		SELECT id, guild_id, voice_channel_id, text_channel_id, created_at, updated_at
		FROM voice_text_links
		WHERE guild_id = $1 AND text_channel_id = $2
	`

	var (
		dbID             string
		dbGuildID        string
		dbVoiceChannelID string
		dbTextChannelID  string
		dbCreatedAt      time.Time
		dbUpdatedAt      time.Time
	)

	err := r.tx.QueryRow(ctx, query, string(guildID), string(textChannelID)).Scan(
		&dbID,
		&dbGuildID,
		&dbVoiceChannelID,
		&dbTextChannelID,
		&dbCreatedAt,
		&dbUpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, voicetext.ErrVoiceTextLinkNotFound
		}
		return nil, err
	}

	return voicetext.RebuildVoiceTextLink(
		voicetext.VoiceTextID(dbID),
		discordid.GuildID(dbGuildID),
		discordid.VoiceChannelID(dbVoiceChannelID),
		discordid.TextChannelID(dbTextChannelID),
		dbCreatedAt,
		dbUpdatedAt,
	)
}

func (r *VoiceTextLinkRepository) Save(ctx context.Context, vtl *voicetext.VoiceTextLink) error {
	query := `
		INSERT INTO voice_text_links (id, guild_id, voice_channel_id, text_channel_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			text_channel_id = EXCLUDED.text_channel_id,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.tx.Exec(ctx, query,
		string(vtl.ID()),
		string(vtl.GuildID()),
		string(vtl.VoiceChannelID()),
		string(vtl.TextChannelID()),
		vtl.CreatedAt(),
		vtl.UpdatedAt(),
	)

	return err
}

func (r *VoiceTextLinkRepository) Delete(ctx context.Context, id voicetext.VoiceTextID) error {
	query := `DELETE FROM voice_text_links WHERE id = $1`

	result, err := r.tx.Exec(ctx, query, string(id))
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return voicetext.ErrVoiceTextLinkNotFound
	}

	return nil
}

func (r *VoiceTextLinkRepository) FindAll(ctx context.Context) ([]*voicetext.VoiceTextLink, error) {
	query := `
		SELECT id, guild_id, voice_channel_id, text_channel_id, created_at, updated_at
		FROM voice_text_links
		ORDER BY created_at
	`

	rows, err := r.tx.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*voicetext.VoiceTextLink
	for rows.Next() {
		var (
			dbID             string
			dbGuildID        string
			dbVoiceChannelID string
			dbTextChannelID  string
			dbCreatedAt      time.Time
			dbUpdatedAt      time.Time
		)

		if err := rows.Scan(&dbID, &dbGuildID, &dbVoiceChannelID, &dbTextChannelID, &dbCreatedAt, &dbUpdatedAt); err != nil {
			return nil, err
		}

		link, err := voicetext.RebuildVoiceTextLink(
			voicetext.VoiceTextID(dbID),
			discordid.GuildID(dbGuildID),
			discordid.VoiceChannelID(dbVoiceChannelID),
			discordid.TextChannelID(dbTextChannelID),
			dbCreatedAt,
			dbUpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		links = append(links, link)
	}

	return links, rows.Err()
}
