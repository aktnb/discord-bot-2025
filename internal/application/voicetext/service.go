package voicetext

import (
	"context"
	"log"

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

func (s *Service) SyncVoiceTextLinks(ctx context.Context) error {
	log.Println("[INFO] Sync started")

	// 1. 準備フェーズ: Guild一覧とDB全リンクを取得
	guilds, err := s.discord.GetGuilds(ctx)
	if err != nil {
		log.Printf("[ERROR] Failed to get guilds: %v", err)
		return err
	}

	var dbLinks []*voicetext.VoiceTextLink
	err = s.txm.WithTx(ctx, func(ctx context.Context, tx db.Tx) error {
		repo := s.repositories.VoiceTextLink(tx)
		links, err := repo.FindAll(ctx)
		if err != nil {
			return err
		}
		dbLinks = links
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Failed to get DB links: %v", err)
		return err
	}

	log.Printf("[INFO] Sync preparation: guilds=%d, db_links=%d", len(guilds), len(dbLinks))

	// Guild一覧をマップ化
	guildMap := make(map[discordid.GuildID]bool)
	for _, guildID := range guilds {
		guildMap[guildID] = true
	}

	// Guild毎のVoiceStatesを取得してマップ化
	guildVoiceStates := make(map[discordid.GuildID]map[discordid.VoiceChannelID][]discordid.UserID)
	for _, guildID := range guilds {
		voiceStates, err := s.discord.GetGuildVoiceStates(ctx, guildID)
		if err != nil {
			log.Printf("[ERROR] Failed to get guild voice states: guild=%s err=%v", guildID, err)
			continue
		}
		guildVoiceStates[guildID] = voiceStates
	}

	cleanedCount := 0
	syncedCount := 0
	createdCount := 0
	errorCount := 0

	// DBリンクをマップ化（作成フェーズで使用）
	dbLinkMap := make(map[string]*voicetext.VoiceTextLink)
	for _, link := range dbLinks {
		key := string(link.GuildID()) + ":" + string(link.VoiceChannelID())
		dbLinkMap[key] = link
	}

	// 2. クリーンアップフェーズ: DBにあるが不要なリンクを削除
	for _, link := range dbLinks {
		// Guildが存在しない場合
		if !guildMap[link.GuildID()] {
			log.Printf("[WARN] Guild not found, deleting link: guild=%s voice=%s", link.GuildID(), link.VoiceChannelID())
			if err := s.cleanupLink(ctx, link); err != nil {
				log.Printf("[ERROR] Failed to cleanup link for missing guild: %v", err)
				errorCount++
			} else {
				cleanedCount++
			}
			delete(dbLinkMap, string(link.GuildID())+":"+string(link.VoiceChannelID()))
			continue
		}

		// ボイスチャンネルが存在しない場合
		exists, err := s.discord.IsVoiceChannelExists(ctx, link.VoiceChannelID())
		if err != nil {
			log.Printf("[ERROR] Failed to check voice channel existence: guild=%s voice=%s err=%v", link.GuildID(), link.VoiceChannelID(), err)
			errorCount++
			continue
		}
		if !exists {
			log.Printf("[WARN] Voice channel not found, deleting link: guild=%s voice=%s", link.GuildID(), link.VoiceChannelID())
			if err := s.cleanupLink(ctx, link); err != nil {
				log.Printf("[ERROR] Failed to cleanup link for missing voice channel: %v", err)
				errorCount++
			} else {
				cleanedCount++
			}
			delete(dbLinkMap, string(link.GuildID())+":"+string(link.VoiceChannelID()))
			continue
		}

		// VoiceStatesを取得
		voiceStates, ok := guildVoiceStates[link.GuildID()]
		if !ok {
			continue
		}
		userIDs, ok := voiceStates[link.VoiceChannelID()]

		// メンバーが0人の場合
		if !ok || len(userIDs) == 0 {
			log.Printf("[WARN] Voice channel is empty, deleting link: guild=%s voice=%s", link.GuildID(), link.VoiceChannelID())
			if err := s.cleanupLink(ctx, link); err != nil {
				log.Printf("[ERROR] Failed to cleanup link for empty voice channel: %v", err)
				errorCount++
			} else {
				cleanedCount++
			}
			delete(dbLinkMap, string(link.GuildID())+":"+string(link.VoiceChannelID()))
			continue
		}

		// 3. 同期フェーズ: 既存のリンクについて、ユーザー権限を完全に同期
		if err := s.syncLinkPermissions(ctx, link, userIDs); err != nil {
			log.Printf("[ERROR] Failed to sync link permissions: guild=%s voice=%s err=%v", link.GuildID(), link.VoiceChannelID(), err)
			errorCount++
		} else {
			syncedCount++
		}
	}

	// 4. 作成フェーズ: Discordにあるが未作成のリンクを作成
	for guildID, voiceStates := range guildVoiceStates {
		for channelID, userIDs := range voiceStates {
			key := string(guildID) + ":" + string(channelID)
			if _, exists := dbLinkMap[key]; exists {
				// すでにDBに存在する場合はスキップ（同期フェーズで処理済み）
				continue
			}

			// 新規リンクを作成
			log.Printf("[INFO] Creating new link: guild=%s voice=%s users=%d", guildID, channelID, len(userIDs))
			if err := s.createLinkWithUsers(ctx, guildID, channelID, userIDs); err != nil {
				log.Printf("[ERROR] Failed to create link: guild=%s voice=%s err=%v", guildID, channelID, err)
				errorCount++
			} else {
				createdCount++
			}
		}
	}

	log.Printf("[INFO] Sync completed: cleaned=%d, synced=%d, created=%d, errors=%d", cleanedCount, syncedCount, createdCount, errorCount)

	return nil
}

func (s *Service) cleanupLink(ctx context.Context, link *voicetext.VoiceTextLink) error {
	return s.txm.WithKeyLock(ctx, db.LockKey(string(link.GuildID())+string(link.VoiceChannelID())), func(ctx context.Context, tx db.Tx) error {
		repo := s.repositories.VoiceTextLink(tx)

		// テキストチャンネル削除
		if err := s.discord.DeleteTextChannel(ctx, link.TextChannelID()); err != nil {
			log.Printf("[WARN] Failed to delete text channel (may already be deleted): text=%s err=%v", link.TextChannelID(), err)
			// テキストチャンネル削除失敗は警告のみで続行
		}

		// DB削除
		if err := repo.Delete(ctx, link.ID()); err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) syncLinkPermissions(ctx context.Context, link *voicetext.VoiceTextLink, voiceChannelUsers []discordid.UserID) error {
	// VoiceChannelにいるユーザーをマップ化
	voiceUserMap := make(map[discordid.UserID]bool)
	for _, userID := range voiceChannelUsers {
		voiceUserMap[userID] = true
	}

	// テキストチャンネルに権限を持つユーザー一覧を取得
	textChannelUsers, err := s.discord.GetTextChannelMembers(ctx, link.TextChannelID())
	if err != nil {
		return err
	}

	log.Printf("[INFO] Syncing permissions: guild=%s voice=%s text_members=%d voice_members=%d", link.GuildID(), link.VoiceChannelID(), len(textChannelUsers), len(voiceChannelUsers))

	// VoiceChannelにいる全ユーザーに権限を付与
	for _, userID := range voiceChannelUsers {
		if err := s.discord.AddMemberToTextChannel(ctx, link.GuildID(), link.TextChannelID(), userID); err != nil {
			log.Printf("[ERROR] Failed to add member permission: guild=%s text=%s user=%s err=%v", link.GuildID(), link.TextChannelID(), userID, err)
		}
	}

	// テキストチャンネルに権限があるが、VoiceChannelにいないユーザーの権限を剥奪
	for _, userID := range textChannelUsers {
		if !voiceUserMap[userID] {
			log.Printf("[INFO] Removing permission from user not in voice: guild=%s text=%s user=%s", link.GuildID(), link.TextChannelID(), userID)
			if err := s.discord.RemoveMemberFromTextChannel(ctx, link.GuildID(), link.TextChannelID(), userID); err != nil {
				log.Printf("[ERROR] Failed to remove member permission: guild=%s text=%s user=%s err=%v", link.GuildID(), link.TextChannelID(), userID, err)
			}
		}
	}

	return nil
}

func (s *Service) createLinkWithUsers(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID, userIDs []discordid.UserID) error {
	return s.txm.WithKeyLock(ctx, db.LockKey(string(guildID)+string(voiceChannelID)), func(ctx context.Context, tx db.Tx) error {
		repo := s.repositories.VoiceTextLink(tx)

		// テキストチャンネル作成
		textChannelID, err := s.discord.CreateTextChannelForVoice(ctx, guildID, voiceChannelID)
		if err != nil {
			return err
		}

		// VoiceTextLinkエンティティ作成
		vtl, err := voicetext.NewVoiceTextLink(guildID, voiceChannelID, textChannelID)
		if err != nil {
			return err
		}

		// DB保存
		if err := repo.Save(ctx, vtl); err != nil {
			return err
		}

		// 全ユーザーに権限付与
		for _, userID := range userIDs {
			if err := s.discord.AddMemberToTextChannel(ctx, guildID, textChannelID, userID); err != nil {
				log.Printf("[ERROR] Failed to add member to text channel: guild=%s text=%s user=%s err=%v", guildID, textChannelID, userID, err)
				// ユーザー権限付与失敗は警告のみで続行
			}
		}

		return nil
	})
}
