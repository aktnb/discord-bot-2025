package voicetext

import (
	"context"
	"errors"
	"sync"
	"testing"

	domain "github.com/aktnb/discord-bot-go/internal/domain/voicetext"
	"github.com/aktnb/discord-bot-go/internal/interfaces/db"
	"github.com/aktnb/discord-bot-go/internal/interfaces/discord"
	"github.com/aktnb/discord-bot-go/internal/shared/discordid"
)

// The fakes assert both boundaries: repositories require a transaction and
// every Discord method must run after that transaction has finished.
type boundaryFixture struct {
	repoDeleteErr                                                   error
	t                                                               *testing.T
	active                                                          bool
	link                                                            *domain.VoiceTextLink
	saveErr, commitErr, lookupErr, deleteErr, addErr, compensateErr error
	commitPersists                                                  bool
	cancel                                                          context.CancelFunc
	creates, deletes, adds, removes                                 int
	users                                                           []discordid.UserID
	members                                                         []discordid.UserID
	exists                                                          bool
	createEntered                                                   chan struct{}
	createRelease                                                   chan struct{}
}
type boundaryTx struct{ f *boundaryFixture }

func (m boundaryTx) WithTx(ctx context.Context, fn func(context.Context, db.Tx) error) error {
	f := m.f
	if f.active {
		f.t.Fatal("nested transaction")
	}
	f.active = true
	defer func() { f.active = false }()
	old := f.link
	if err := fn(ctx, nil); err != nil {
		f.link = old
		return err
	}
	if f.commitErr != nil && f.link != old {
		err := f.commitErr
		f.commitErr = nil
		if !f.commitPersists {
			f.link = old
		}
		return err
	}
	return nil
}

type boundaryRepo struct {
	domain.Repository
	f *boundaryFixture
}

func (r boundaryRepo) check() {
	if !r.f.active {
		r.f.t.Fatal("repository outside transaction")
	}
}
func (r boundaryRepo) FindByVoiceChannel(context.Context, discordid.GuildID, discordid.VoiceChannelID) (*domain.VoiceTextLink, error) {
	r.check()
	if r.f.lookupErr != nil {
		return nil, r.f.lookupErr
	}
	if r.f.link == nil {
		return nil, domain.ErrVoiceTextLinkNotFound
	}
	l := r.f.link
	return domain.RebuildVoiceTextLink(l.ID(), l.GuildID(), l.VoiceChannelID(), l.TextChannelID(), l.CreatedAt(), l.UpdatedAt())
}
func (r boundaryRepo) FindAll(context.Context) ([]*domain.VoiceTextLink, error) {
	r.check()
	if r.f.link == nil {
		return nil, nil
	}
	return []*domain.VoiceTextLink{r.f.link}, nil
}
func (r boundaryRepo) Save(_ context.Context, l *domain.VoiceTextLink) error {
	r.check()
	if r.f.cancel != nil {
		r.f.cancel()
	}
	if r.f.saveErr != nil {
		return r.f.saveErr
	}
	r.f.link = l
	return nil
}
func (r boundaryRepo) Delete(context.Context, domain.VoiceTextID) error {
	r.check()
	if r.f.repoDeleteErr != nil {
		return r.f.repoDeleteErr
	}
	r.f.link = nil
	return nil
}

type boundaryDiscord struct {
	discord.DiscordPort
	f *boundaryFixture
}

func (d boundaryDiscord) check() {
	if d.f.active {
		d.f.t.Fatal("Discord called inside transaction")
	}
}
func (d boundaryDiscord) CreateTextChannelForVoice(context.Context, discordid.GuildID, discordid.VoiceChannelID) (discordid.TextChannelID, error) {
	d.check()
	d.f.creates++
	if d.f.createEntered != nil {
		close(d.f.createEntered)
		<-d.f.createRelease
	}
	return "new-text", nil
}
func (d boundaryDiscord) IsTextChannelExists(context.Context, discordid.TextChannelID) (bool, error) {
	d.check()
	return d.f.exists, nil
}
func (d boundaryDiscord) IsVoiceChannelExists(context.Context, discordid.VoiceChannelID) (bool, error) {
	d.check()
	return true, nil
}
func (d boundaryDiscord) DeleteTextChannel(ctx context.Context, _ discordid.TextChannelID) error {
	d.check()
	d.f.deletes++
	if err := ctx.Err(); err != nil {
		return err
	}
	if d.f.deleteErr != nil {
		return d.f.deleteErr
	}
	return d.f.compensateErr
}
func (d boundaryDiscord) AddMemberToTextChannel(context.Context, discordid.GuildID, discordid.TextChannelID, discordid.UserID) error {
	d.check()
	d.f.adds++
	return d.f.addErr
}
func (d boundaryDiscord) RemoveMemberFromTextChannel(context.Context, discordid.GuildID, discordid.TextChannelID, discordid.UserID) error {
	d.check()
	d.f.removes++
	return nil
}
func (d boundaryDiscord) GetVoiceChannelMemberCount(context.Context, discordid.GuildID, discordid.VoiceChannelID) (int, error) {
	d.check()
	return len(d.f.users), nil
}
func (d boundaryDiscord) GetGuildVoiceStates(context.Context, discordid.GuildID) (map[discordid.VoiceChannelID][]discordid.UserID, error) {
	d.check()
	return map[discordid.VoiceChannelID][]discordid.UserID{"voice": d.f.users}, nil
}
func (d boundaryDiscord) GetGuilds(context.Context) ([]discordid.GuildID, error) {
	d.check()
	return []discordid.GuildID{"guild"}, nil
}
func (d boundaryDiscord) GetTextChannelMembers(context.Context, discordid.TextChannelID) ([]discordid.UserID, error) {
	d.check()
	return d.f.members, nil
}
func (f *boundaryFixture) service() *Service {
	return NewVoiceTextService(func(db.Tx) domain.Repository { return boundaryRepo{f: f} }, boundaryTx{f}, boundaryDiscord{f: f})
}
func (f *boundaryFixture) seed() {
	f.t.Helper()
	var err error
	f.link, err = domain.NewVoiceTextLink("guild", "voice", "old-text")
	if err != nil {
		f.t.Fatal(err)
	}
}
func joinCmd() JoinVoiceCommand {
	return JoinVoiceCommand{GuildID: "guild", VoiceChannelID: "voice", UserID: "user"}
}
func TestJoinTransactionBoundaryAndCompensation(t *testing.T) {
	cause := errors.New("save failed")
	compensation := errors.New("delete failed")
	for _, name := range []string{"create", "existing", "recreate", "save failure", "cancelled save", "compensation failure", "commit persisted", "commit rolled back", "commit unverifiable", "permission failure"} {
		t.Run(name, func(t *testing.T) {
			f := &boundaryFixture{t: t, exists: true}
			ctx := context.Background()
			wantErr := false
			switch name {
			case "existing":
				f.seed()
			case "recreate":
				f.seed()
				f.exists = false
			case "save failure":
				f.saveErr = cause
				wantErr = true
			case "cancelled save":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				defer cancel()
				f.cancel = cancel
				f.saveErr = cause
				wantErr = true
			case "compensation failure":
				f.saveErr = cause
				f.compensateErr = compensation
				wantErr = true
			case "commit persisted":
				f.commitErr = cause
				f.commitPersists = true
			case "commit rolled back":
				f.commitErr = cause
				wantErr = true
			case "permission failure":
				f.addErr = cause
				wantErr = true
			}
			// Inject a lookup failure only after saving, simulating an unavailable DB
			// during verification of an ambiguous commit.
			svc := f.service()
			if name == "commit unverifiable" {
				f.commitErr = cause
				f.commitPersists = true
				wantErr = true
				svc.txm = unverifiableTx{base: boundaryTx{f}, f: f}
			}
			err := svc.JoinVoice(ctx, joinCmd())
			if (err != nil) != wantErr {
				t.Fatalf("error=%v, wantErr=%v", err, wantErr)
			}
			if wantErr && !errors.Is(err, cause) {
				t.Fatalf("lost original cause: %v", err)
			}
			wantDeletes := 0
			switch name {
			case "save failure", "cancelled save", "compensation failure", "commit rolled back":
				wantDeletes = 1
			}
			if f.deletes != wantDeletes {
				t.Fatalf("deletes=%d, want %d", f.deletes, wantDeletes)
			}
			if name == "compensation failure" && !errors.Is(err, compensation) {
				t.Fatal("lost compensation error")
			}
			if name == "permission failure" && f.link == nil {
				t.Fatal("permission failure discarded saved link")
			}
		})
	}
}

type unverifiableTx struct {
	base boundaryTx
	f    *boundaryFixture
}

func (m unverifiableTx) WithTx(ctx context.Context, fn func(context.Context, db.Tx) error) error {
	err := m.base.WithTx(ctx, fn)
	if err != nil {
		m.f.lookupErr = errors.New("verification unavailable")
	}
	return err
}
func TestLeaveAndSyncTransactionBoundaries(t *testing.T) {
	for _, name := range []string{"leave last", "leave member", "rejoined", "delete failure", "cleanup empty", "cleanup now occupied", "sync permissions", "sync create", "sync empty", "full sync"} {
		t.Run(name, func(t *testing.T) {
			f := &boundaryFixture{t: t, exists: true}
			f.seed()
			candidate := f.link
			s := f.service()
			var err error
			switch name {
			case "leave last":
				err = s.LeaveVoice(context.Background(), LeaveVoiceCommand{GuildID: "guild", VoiceChannelID: "voice", UserID: "user"})
			case "leave member":
				f.users = []discordid.UserID{"other"}
				err = s.LeaveVoice(context.Background(), LeaveVoiceCommand{GuildID: "guild", VoiceChannelID: "voice", UserID: "user", IsLastMember: true})
			case "rejoined":
				f.users = []discordid.UserID{"user"}
				err = s.LeaveVoice(context.Background(), LeaveVoiceCommand{GuildID: "guild", VoiceChannelID: "voice", UserID: "user"})
			case "delete failure":
				f.deleteErr = errors.New("forbidden")
				err = s.LeaveVoice(context.Background(), LeaveVoiceCommand{GuildID: "guild", VoiceChannelID: "voice"})
				if !errors.Is(err, f.deleteErr) || f.link == nil {
					t.Fatal("delete error lost link")
				}
				return
			case "cleanup empty":
				err = s.cleanupLink(context.Background(), candidate)
			case "cleanup now occupied":
				f.users = []discordid.UserID{"user"}
				err = s.cleanupLink(context.Background(), candidate)
			case "sync permissions":
				f.users = []discordid.UserID{"user"}
				f.members = []discordid.UserID{"stale"}
				err = s.syncLinkPermissions(context.Background(), candidate, nil)
			case "sync create":
				f.link = nil
				f.users = []discordid.UserID{"user"}
				err = s.createLinkWithUsers(context.Background(), "guild", "voice", nil)
			case "sync empty":
				f.link = nil
				err = s.createLinkWithUsers(context.Background(), "guild", "voice", []discordid.UserID{"stale"})
			case "full sync":
				f.users = []discordid.UserID{"user"}
				err = s.SyncVoiceTextLinks(context.Background())
			}
			if err != nil {
				t.Fatal(err)
			}
			switch name {
			case "leave last", "cleanup empty":
				if f.link != nil || f.deletes != 1 {
					t.Fatal("link not deleted")
				}
			case "leave member":
				if f.removes != 1 || f.deletes != 0 {
					t.Fatal("stale last-member flag used")
				}
			case "rejoined", "cleanup now occupied":
				if f.deletes != 0 || f.removes != 0 {
					t.Fatal("stale state caused deletion")
				}
			case "sync permissions":
				if f.adds != 1 || f.removes != 1 {
					t.Fatal("permissions not synchronized")
				}
			case "sync create":
				if f.creates != 1 || f.adds != 1 {
					t.Fatal("link not created")
				}
			case "sync empty":
				if f.creates != 0 {
					t.Fatal("created from stale snapshot")
				}
			}
		})
	}
}
func TestConcurrentJoinAndSyncCreateOnlyOneChannel(t *testing.T) {
	f := &boundaryFixture{t: t, exists: true, users: []discordid.UserID{"user"}, createEntered: make(chan struct{}), createRelease: make(chan struct{})}
	s := f.service()
	var wg sync.WaitGroup
	errs := make(chan error, 3)
	wg.Add(1)
	go func() { defer wg.Done(); errs <- s.JoinVoice(context.Background(), joinCmd()) }()
	<-f.createEntered
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); errs <- s.createLinkWithUsers(context.Background(), "guild", "voice", nil) }()
	}
	close(f.createRelease)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if f.creates != 1 {
		t.Fatalf("created %d channels", f.creates)
	}
}

func TestQueuedLeaveUsesCurrentStateAfterJoin(t *testing.T) {
	f := &boundaryFixture{t: t, exists: true, users: []discordid.UserID{"user"}, createEntered: make(chan struct{}), createRelease: make(chan struct{})}
	s := f.service()
	errs := make(chan error, 2)
	go func() { errs <- s.JoinVoice(context.Background(), joinCmd()) }()
	<-f.createEntered
	go func() {
		errs <- s.LeaveVoice(context.Background(), LeaveVoiceCommand{GuildID: "guild", VoiceChannelID: "voice", UserID: "user", IsLastMember: true})
	}()
	close(f.createRelease)
	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
	if f.creates != 1 || f.deletes != 0 || f.removes != 0 || f.link == nil {
		t.Fatal("queued leave used stale last-member flag")
	}
}

func TestDeletionCanRetryAfterDatabaseFailure(t *testing.T) {
	f := &boundaryFixture{t: t, exists: true, repoDeleteErr: errors.New("database unavailable")}
	f.seed()
	s := f.service()
	cmd := LeaveVoiceCommand{GuildID: "guild", VoiceChannelID: "voice"}
	if err := s.LeaveVoice(context.Background(), cmd); !errors.Is(err, f.repoDeleteErr) {
		t.Fatalf("error=%v", err)
	}
	if f.link == nil {
		t.Fatal("failed DB deletion lost link")
	}
	f.repoDeleteErr = nil
	if err := s.LeaveVoice(context.Background(), cmd); err != nil {
		t.Fatal(err)
	}
	if f.link != nil || f.deletes != 2 {
		t.Fatal("deletion was not retried")
	}
}
