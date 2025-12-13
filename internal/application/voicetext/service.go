package voicetext

import (
	"context"

	"github.com/aktnb/discord-bot-go/internal/domain/voicetext"
	"github.com/aktnb/discord-bot-go/internal/interfaces/db"
	"github.com/aktnb/discord-bot-go/internal/interfaces/discord"
	"github.com/aktnb/discord-bot-go/internal/shared/discordid"
)

type Service struct {
	repositories voicetext.Repositories
	txm          db.TxManager
	discord      discord.DiscordPort
}

func NewVoiceTextService(
	repositories voicetext.Repositories,
	txm db.TxManager,
	discordPort discord.DiscordPort,
) *Service {
	return &Service{
		repositories: repositories,
		txm:          txm,
		discord:      discordPort,
	}
}

func (s *Service) VoiceStateUpdate(ctx context.Context, cmd VoiceStateUpdateCommand) error {
	if cmd.BeforeVoiceChannelID == cmd.AfterVoiceChannelID {
		return nil
	}

	if cmd.BeforeVoiceChannelID != nil {
		count, err := s.discord.GetVoiceChannelMemberCount(ctx, cmd.GuildID, *cmd.BeforeVoiceChannelID)
		if err != nil {
			return err
		}

		leaveCmd := LeaveVoiceCommand{
			GuildID:        cmd.GuildID,
			VoiceChannelID: *cmd.BeforeVoiceChannelID,
			UserID:         cmd.UserID,
			IsLastMember:   count == 0,
		}
		if err := s.LeaveVoice(ctx, leaveCmd); err != nil {
			return err
		}
	}

	if cmd.AfterVoiceChannelID != nil {
		joinCmd := JoinVoiceCommand{
			GuildID:        cmd.GuildID,
			VoiceChannelID: *cmd.AfterVoiceChannelID,
			UserID:         cmd.UserID,
		}
		if err := s.JoinVoice(ctx, joinCmd); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) JoinVoice(ctx context.Context, cmd JoinVoiceCommand) error {
	return s.txm.WithKeyLock(ctx, db.LockKey(string(cmd.GuildID)+string(cmd.VoiceChannelID)), func(ctx context.Context, tx db.Tx) error {
		var repo = s.repositories.VoiceTextLink(tx)
		var textChannelID discordid.TextChannelID

		vtl, err := repo.FindByVoiceChannel(ctx, cmd.GuildID, cmd.VoiceChannelID)
		if err != nil && err != voicetext.ErrVoiceTextLinkNotFound {
			return err
		}

		if vtl == nil {
			textChannelID, err = s.discord.CreateTextChannelForVoice(ctx, cmd.GuildID, cmd.VoiceChannelID)
			if err != nil {
				return err
			}
			vtl, err = voicetext.NewVoiceTextLink(
				cmd.GuildID,
				cmd.VoiceChannelID,
				textChannelID,
			)
			if err != nil {
				return err
			}

			if err := repo.Save(ctx, vtl); err != nil {
				return err
			}
		} else {
			exists, err := s.discord.IsTextChannelExists(ctx, vtl.TextChannelID())
			if err != nil {
				return err
			}
			if !exists {
				textChannelID, err = s.discord.CreateTextChannelForVoice(ctx, cmd.GuildID, cmd.VoiceChannelID)
				if err != nil {
					return err
				}
				if err := vtl.ChangeTextChannel(textChannelID); err != nil {
					return err
				}
				if err := repo.Save(ctx, vtl); err != nil {
					return err
				}
			}
		}

		if err := s.discord.AddMemberToTextChannel(ctx, cmd.GuildID, vtl.TextChannelID(), cmd.UserID); err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) LeaveVoice(ctx context.Context, cmd LeaveVoiceCommand) error {
	return s.txm.WithKeyLock(ctx, db.LockKey(string(cmd.GuildID)+string(cmd.VoiceChannelID)), func(ctx context.Context, tx db.Tx) error {
		var repo = s.repositories.VoiceTextLink(tx)
		vtl, err := repo.FindByVoiceChannel(ctx, cmd.GuildID, cmd.VoiceChannelID)
		if err != nil {
			return err
		}

		if cmd.IsLastMember {
			if err := s.discord.DeleteTextChannel(ctx, vtl.TextChannelID()); err != nil {
				return err
			}

			if err := repo.Delete(ctx, vtl.ID()); err != nil {
				return err
			}

			return nil
		}

		if err := s.discord.RemoveMemberFromTextChannel(ctx, cmd.GuildID, vtl.TextChannelID(), cmd.UserID); err != nil {
			return err
		}

		return nil
	})
}
