package voicetext

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aktnb/discord-bot-go/internal/domain/voicetext"
	"github.com/aktnb/discord-bot-go/internal/interfaces/db"
	"github.com/aktnb/discord-bot-go/internal/interfaces/discord"
	"github.com/aktnb/discord-bot-go/internal/shared/discordid"
	"github.com/aktnb/discord-bot-go/internal/shared/keylock"
)

// RepositoryFactory はトランザクションに紐づくリポジトリを生成します。
type RepositoryFactory func(db.Tx) voicetext.Repository

// Service owns process-local channel locks. Share one instance across event and sync handlers.
// Multiple Bot processes require distributed coordination instead.
type Service struct {
	newRepository RepositoryFactory
	txm           db.TxManager
	discord       discord.DiscordPort
	locks         *keylock.Locker
}

func NewVoiceTextService(
	newRepository RepositoryFactory,
	txm db.TxManager,
	discordPort discord.DiscordPort,
) *Service {
	return &Service{
		newRepository: newRepository,
		txm:           txm,
		discord:       discordPort,
		locks:         keylock.New(),
	}
}

func (s *Service) VoiceStateUpdate(ctx context.Context, cmd VoiceStateUpdateCommand) error {
	if cmd.BeforeVoiceChannelID == nil && cmd.AfterVoiceChannelID == nil {
		return nil
	}
	if cmd.BeforeVoiceChannelID != nil && cmd.AfterVoiceChannelID != nil &&
		*cmd.BeforeVoiceChannelID == *cmd.AfterVoiceChannelID {
		return nil
	}

	if cmd.BeforeVoiceChannelID != nil {
		leaveCmd := LeaveVoiceCommand{
			GuildID:        cmd.GuildID,
			VoiceChannelID: *cmd.BeforeVoiceChannelID,
			UserID:         cmd.UserID,
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

func (s *Service) withChannelLock(ctx context.Context, guild discordid.GuildID, voice discordid.VoiceChannelID, fn func() error) error {
	return s.locks.WithLock(ctx, string(guild)+":"+string(voice), fn)
}

func (s *Service) findLink(ctx context.Context, guild discordid.GuildID, voice discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
	var link *voicetext.VoiceTextLink
	err := s.txm.WithTx(ctx, func(ctx context.Context, tx db.Tx) error {
		var err error
		link, err = s.newRepository(tx).FindByVoiceChannel(ctx, guild, voice)
		return err
	})
	return link, err
}

// ensureLink runs under the channel lock. No Discord call holds a transaction.
func (s *Service) ensureLink(ctx context.Context, guild discordid.GuildID, voice discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
	link, err := s.findLink(ctx, guild, voice)
	if err != nil && !errors.Is(err, voicetext.ErrVoiceTextLinkNotFound) {
		return nil, err
	}
	if link != nil {
		exists, err := s.discord.IsTextChannelExists(ctx, link.TextChannelID())
		if err != nil {
			return nil, err
		}
		if exists {
			return link, nil
		}
	}
	text, err := s.discord.CreateTextChannelForVoice(ctx, guild, voice)
	if err != nil {
		return nil, err
	}
	if link == nil {
		link, err = voicetext.NewVoiceTextLink(guild, voice, text)
	} else {
		err = link.ChangeTextChannel(text)
	}
	if err != nil {
		return nil, s.compensate(ctx, text, err)
	}
	// A successful callback followed by an error may be an ambiguous Commit.
	saved := false
	err = s.txm.WithTx(ctx, func(ctx context.Context, tx db.Tx) error {
		if err := s.newRepository(tx).Save(ctx, link); err != nil {
			return err
		}
		saved = true
		return nil
	})
	if err == nil {
		return link, nil
	}
	if saved {
		checkCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		actual, checkErr := s.findLink(checkCtx, guild, voice)
		if checkErr == nil && actual != nil && actual.TextChannelID() == text {
			return actual, nil
		}
		if checkErr != nil && !errors.Is(checkErr, voicetext.ErrVoiceTextLinkNotFound) {
			verificationErr := fmt.Errorf("cannot verify channel save: guild=%s voice=%s text=%s: %w", guild, voice, text, checkErr)
			log.Printf("[ERROR] %v; original error: %v", verificationErr, err)
			return nil, errors.Join(err, verificationErr)
		}
	}
	return nil, s.compensate(ctx, text, err)
}

func (s *Service) compensate(ctx context.Context, text discordid.TextChannelID, cause error) error {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := s.discord.DeleteTextChannel(cleanupCtx, text); err != nil {
		failure := fmt.Errorf("compensation failed: text=%s: %w", text, err)
		log.Printf("[ERROR] %v; original error: %v", failure, cause)
		return errors.Join(cause, failure)
	}
	return cause
}

func (s *Service) JoinVoice(ctx context.Context, cmd JoinVoiceCommand) error {
	return s.withChannelLock(ctx, cmd.GuildID, cmd.VoiceChannelID, func() error {
		link, err := s.ensureLink(ctx, cmd.GuildID, cmd.VoiceChannelID)
		if err != nil {
			return err
		}
		return s.discord.AddMemberToTextChannel(ctx, cmd.GuildID, link.TextChannelID(), cmd.UserID)
	})
}

func (s *Service) LeaveVoice(ctx context.Context, cmd LeaveVoiceCommand) error {
	return s.withChannelLock(ctx, cmd.GuildID, cmd.VoiceChannelID, func() error {
		link, err := s.findLink(ctx, cmd.GuildID, cmd.VoiceChannelID)
		if errors.Is(err, voicetext.ErrVoiceTextLinkNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		count, err := s.discord.GetVoiceChannelMemberCount(ctx, cmd.GuildID, cmd.VoiceChannelID)
		if err != nil {
			return err
		}
		if count == 0 {
			return s.deleteLink(ctx, link)
		}
		// A queued leave event can be overtaken by a rejoin. Preserve its permission.
		states, err := s.discord.GetGuildVoiceStates(ctx, cmd.GuildID)
		if err != nil {
			return err
		}
		for _, user := range states[cmd.VoiceChannelID] {
			if user == cmd.UserID {
				return nil
			}
		}
		return s.discord.RemoveMemberFromTextChannel(ctx, cmd.GuildID, link.TextChannelID(), cmd.UserID)
	})
}

func (s *Service) deleteLink(ctx context.Context, link *voicetext.VoiceTextLink) error {
	if err := s.discord.DeleteTextChannel(ctx, link.TextChannelID()); err != nil {
		return err
	}
	return s.txm.WithTx(ctx, func(ctx context.Context, tx db.Tx) error {
		err := s.newRepository(tx).Delete(ctx, link.ID())
		if errors.Is(err, voicetext.ErrVoiceTextLinkNotFound) {
			return nil
		}
		return err
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
		repo := s.newRepository(tx)
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

func (s *Service) cleanupLink(ctx context.Context, candidate *voicetext.VoiceTextLink) error {
	return s.withChannelLock(ctx, candidate.GuildID(), candidate.VoiceChannelID(), func() error {
		link, err := s.findLink(ctx, candidate.GuildID(), candidate.VoiceChannelID())
		if errors.Is(err, voicetext.ErrVoiceTextLinkNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		guilds, err := s.discord.GetGuilds(ctx)
		if err != nil {
			return err
		}
		present := false
		for _, guild := range guilds {
			if guild == link.GuildID() {
				present = true
				break
			}
		}
		if present {
			exists, err := s.discord.IsVoiceChannelExists(ctx, link.VoiceChannelID())
			if err != nil {
				return err
			}
			if exists {
				count, err := s.discord.GetVoiceChannelMemberCount(ctx, link.GuildID(), link.VoiceChannelID())
				if err != nil {
					return err
				}
				if count != 0 {
					return nil
				}
			}
		}
		return s.deleteLink(ctx, link)
	})
}

func (s *Service) syncLinkPermissions(ctx context.Context, candidate *voicetext.VoiceTextLink, _ []discordid.UserID) error {
	return s.withChannelLock(ctx, candidate.GuildID(), candidate.VoiceChannelID(), func() error {
		states, err := s.discord.GetGuildVoiceStates(ctx, candidate.GuildID())
		if err != nil {
			return err
		}
		users := states[candidate.VoiceChannelID()]
		if len(users) == 0 {
			return nil
		}
		link, err := s.ensureLink(ctx, candidate.GuildID(), candidate.VoiceChannelID())
		if err != nil {
			return err
		}
		return s.syncPermissions(ctx, link, users)
	})
}

func (s *Service) syncPermissions(ctx context.Context, link *voicetext.VoiceTextLink, voiceChannelUsers []discordid.UserID) error {
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

func (s *Service) createLinkWithUsers(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID, _ []discordid.UserID) error {
	return s.withChannelLock(ctx, guildID, voiceChannelID, func() error {
		states, err := s.discord.GetGuildVoiceStates(ctx, guildID)
		if err != nil {
			return err
		}
		users := states[voiceChannelID]
		if len(users) == 0 {
			return nil
		}
		link, err := s.ensureLink(ctx, guildID, voiceChannelID)
		if err != nil {
			return err
		}
		return s.syncPermissions(ctx, link, users)
	})
}
