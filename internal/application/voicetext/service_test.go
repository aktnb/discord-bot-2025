package voicetext

import (
	"context"
	"errors"
	"testing"

	"github.com/aktnb/discord-bot-go/internal/domain/voicetext"
	"github.com/aktnb/discord-bot-go/internal/interfaces/db"
	"github.com/aktnb/discord-bot-go/internal/shared/discordid"
)

// Mock implementations

type mockTx struct{}

func (m *mockTx) Exec(ctx context.Context, sql string, arguments ...any) (db.CommandTag, error) {
	return nil, nil
}

func (m *mockTx) Query(ctx context.Context, sql string, args ...any) (db.Rows, error) {
	return nil, nil
}

func (m *mockTx) QueryRow(ctx context.Context, sql string, args ...any) db.Row {
	return nil
}

type mockTxManager struct {
	withTxFunc      func(ctx context.Context, fn func(ctx context.Context, tx db.Tx) error) error
	withKeyLockFunc func(ctx context.Context, key db.LockKey, fn func(ctx context.Context, tx db.Tx) error) error
}

func (m *mockTxManager) WithTx(ctx context.Context, fn func(ctx context.Context, tx db.Tx) error) error {
	if m.withTxFunc != nil {
		return m.withTxFunc(ctx, fn)
	}
	return fn(ctx, &mockTx{})
}

func (m *mockTxManager) WithKeyLock(ctx context.Context, key db.LockKey, fn func(ctx context.Context, tx db.Tx) error) error {
	if m.withKeyLockFunc != nil {
		return m.withKeyLockFunc(ctx, key, fn)
	}
	return fn(ctx, &mockTx{})
}

type mockRepository struct {
	findByVoiceChannelFunc func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error)
	findByTextChannelFunc  func(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID) (*voicetext.VoiceTextLink, error)
	findAllFunc            func(ctx context.Context) ([]*voicetext.VoiceTextLink, error)
	saveFunc               func(ctx context.Context, vtl *voicetext.VoiceTextLink) error
	deleteFunc             func(ctx context.Context, id voicetext.VoiceTextID) error
}

func (m *mockRepository) FindByVoiceChannel(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
	if m.findByVoiceChannelFunc != nil {
		return m.findByVoiceChannelFunc(ctx, guildID, voiceChannelID)
	}
	return nil, voicetext.ErrVoiceTextLinkNotFound
}

func (m *mockRepository) FindByTextChannel(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID) (*voicetext.VoiceTextLink, error) {
	if m.findByTextChannelFunc != nil {
		return m.findByTextChannelFunc(ctx, guildID, textChannelID)
	}
	return nil, voicetext.ErrVoiceTextLinkNotFound
}

func (m *mockRepository) FindAll(ctx context.Context) ([]*voicetext.VoiceTextLink, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx)
	}
	return []*voicetext.VoiceTextLink{}, nil
}

func (m *mockRepository) Save(ctx context.Context, vtl *voicetext.VoiceTextLink) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, vtl)
	}
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id voicetext.VoiceTextID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

type mockRepositories struct {
	repo *mockRepository
}

func (m *mockRepositories) VoiceTextLink(tx db.Tx) voicetext.Repository {
	return m.repo
}

type mockDiscordPort struct {
	createTextChannelForVoiceFunc  func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (discordid.TextChannelID, error)
	deleteTextChannelFunc          func(ctx context.Context, textChannelID discordid.TextChannelID) error
	isVoiceChannelExistsFunc       func(ctx context.Context, channelID discordid.VoiceChannelID) (bool, error)
	isTextChannelExistsFunc        func(ctx context.Context, channelID discordid.TextChannelID) (bool, error)
	addMemberToTextChannelFunc     func(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID, userID discordid.UserID) error
	removeMemberFromTextChannelFunc func(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID, userID discordid.UserID) error
	getVoiceChannelMemberCountFunc func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (int, error)
	getGuildsFunc                  func(ctx context.Context) ([]discordid.GuildID, error)
	getGuildVoiceStatesFunc        func(ctx context.Context, guildID discordid.GuildID) (map[discordid.VoiceChannelID][]discordid.UserID, error)
	getTextChannelMembersFunc      func(ctx context.Context, textChannelID discordid.TextChannelID) ([]discordid.UserID, error)
}

func (m *mockDiscordPort) CreateTextChannelForVoice(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (discordid.TextChannelID, error) {
	if m.createTextChannelForVoiceFunc != nil {
		return m.createTextChannelForVoiceFunc(ctx, guildID, voiceChannelID)
	}
	return "text123", nil
}

func (m *mockDiscordPort) DeleteTextChannel(ctx context.Context, textChannelID discordid.TextChannelID) error {
	if m.deleteTextChannelFunc != nil {
		return m.deleteTextChannelFunc(ctx, textChannelID)
	}
	return nil
}

func (m *mockDiscordPort) IsVoiceChannelExists(ctx context.Context, channelID discordid.VoiceChannelID) (bool, error) {
	if m.isVoiceChannelExistsFunc != nil {
		return m.isVoiceChannelExistsFunc(ctx, channelID)
	}
	return true, nil
}

func (m *mockDiscordPort) IsTextChannelExists(ctx context.Context, channelID discordid.TextChannelID) (bool, error) {
	if m.isTextChannelExistsFunc != nil {
		return m.isTextChannelExistsFunc(ctx, channelID)
	}
	return true, nil
}

func (m *mockDiscordPort) AddMemberToTextChannel(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID, userID discordid.UserID) error {
	if m.addMemberToTextChannelFunc != nil {
		return m.addMemberToTextChannelFunc(ctx, guildID, textChannelID, userID)
	}
	return nil
}

func (m *mockDiscordPort) RemoveMemberFromTextChannel(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID, userID discordid.UserID) error {
	if m.removeMemberFromTextChannelFunc != nil {
		return m.removeMemberFromTextChannelFunc(ctx, guildID, textChannelID, userID)
	}
	return nil
}

func (m *mockDiscordPort) GetVoiceChannelMemberCount(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (int, error) {
	if m.getVoiceChannelMemberCountFunc != nil {
		return m.getVoiceChannelMemberCountFunc(ctx, guildID, voiceChannelID)
	}
	return 1, nil
}

func (m *mockDiscordPort) GetGuilds(ctx context.Context) ([]discordid.GuildID, error) {
	if m.getGuildsFunc != nil {
		return m.getGuildsFunc(ctx)
	}
	return []discordid.GuildID{}, nil
}

func (m *mockDiscordPort) GetGuildVoiceStates(ctx context.Context, guildID discordid.GuildID) (map[discordid.VoiceChannelID][]discordid.UserID, error) {
	if m.getGuildVoiceStatesFunc != nil {
		return m.getGuildVoiceStatesFunc(ctx, guildID)
	}
	return map[discordid.VoiceChannelID][]discordid.UserID{}, nil
}

func (m *mockDiscordPort) GetTextChannelMembers(ctx context.Context, textChannelID discordid.TextChannelID) ([]discordid.UserID, error) {
	if m.getTextChannelMembersFunc != nil {
		return m.getTextChannelMembersFunc(ctx, textChannelID)
	}
	return []discordid.UserID{}, nil
}

// Test cases

func TestJoinVoice_NewLink(t *testing.T) {
	// Setup mocks
	repo := &mockRepository{}
	repos := &mockRepositories{repo: repo}
	txm := &mockTxManager{}
	discord := &mockDiscordPort{}

	linkCreated := false
	textChannelCreated := false
	memberAdded := false

	repo.findByVoiceChannelFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
		return nil, voicetext.ErrVoiceTextLinkNotFound
	}

	repo.saveFunc = func(ctx context.Context, vtl *voicetext.VoiceTextLink) error {
		linkCreated = true
		if vtl.GuildID() != "guild123" {
			t.Errorf("expected GuildID guild123, got %s", vtl.GuildID())
		}
		if vtl.VoiceChannelID() != "voice456" {
			t.Errorf("expected VoiceChannelID voice456, got %s", vtl.VoiceChannelID())
		}
		if vtl.TextChannelID() != "text789" {
			t.Errorf("expected TextChannelID text789, got %s", vtl.TextChannelID())
		}
		return nil
	}

	discord.createTextChannelForVoiceFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (discordid.TextChannelID, error) {
		textChannelCreated = true
		return "text789", nil
	}

	discord.addMemberToTextChannelFunc = func(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID, userID discordid.UserID) error {
		memberAdded = true
		return nil
	}

	// Create service
	service := NewVoiceTextService(repos, txm, discord)

	// Test
	cmd := JoinVoiceCommand{
		GuildID:        "guild123",
		VoiceChannelID: "voice456",
		UserID:         "user789",
	}

	err := service.JoinVoice(context.Background(), cmd)

	// Assertions
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !textChannelCreated {
		t.Error("expected text channel to be created")
	}
	if !linkCreated {
		t.Error("expected link to be saved")
	}
	if !memberAdded {
		t.Error("expected member to be added")
	}
}

func TestJoinVoice_ExistingLink(t *testing.T) {
	// Setup mocks
	repo := &mockRepository{}
	repos := &mockRepositories{repo: repo}
	txm := &mockTxManager{}
	discord := &mockDiscordPort{}

	existingLink, _ := voicetext.NewVoiceTextLink("guild123", "voice456", "text789")
	memberAdded := false

	repo.findByVoiceChannelFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
		return existingLink, nil
	}

	discord.isTextChannelExistsFunc = func(ctx context.Context, channelID discordid.TextChannelID) (bool, error) {
		return true, nil
	}

	discord.addMemberToTextChannelFunc = func(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID, userID discordid.UserID) error {
		memberAdded = true
		if textChannelID != "text789" {
			t.Errorf("expected TextChannelID text789, got %s", textChannelID)
		}
		return nil
	}

	// Create service
	service := NewVoiceTextService(repos, txm, discord)

	// Test
	cmd := JoinVoiceCommand{
		GuildID:        "guild123",
		VoiceChannelID: "voice456",
		UserID:         "user789",
	}

	err := service.JoinVoice(context.Background(), cmd)

	// Assertions
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !memberAdded {
		t.Error("expected member to be added")
	}
}

func TestJoinVoice_ExistingLinkTextChannelDeleted(t *testing.T) {
	// Setup mocks
	repo := &mockRepository{}
	repos := &mockRepositories{repo: repo}
	txm := &mockTxManager{}
	discord := &mockDiscordPort{}

	existingLink, _ := voicetext.NewVoiceTextLink("guild123", "voice456", "text789")
	textChannelCreated := false
	linkUpdated := false

	repo.findByVoiceChannelFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
		return existingLink, nil
	}

	discord.isTextChannelExistsFunc = func(ctx context.Context, channelID discordid.TextChannelID) (bool, error) {
		return false, nil
	}

	discord.createTextChannelForVoiceFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (discordid.TextChannelID, error) {
		textChannelCreated = true
		return "text999", nil
	}

	repo.saveFunc = func(ctx context.Context, vtl *voicetext.VoiceTextLink) error {
		linkUpdated = true
		if vtl.TextChannelID() != "text999" {
			t.Errorf("expected TextChannelID to be updated to text999, got %s", vtl.TextChannelID())
		}
		return nil
	}

	// Create service
	service := NewVoiceTextService(repos, txm, discord)

	// Test
	cmd := JoinVoiceCommand{
		GuildID:        "guild123",
		VoiceChannelID: "voice456",
		UserID:         "user789",
	}

	err := service.JoinVoice(context.Background(), cmd)

	// Assertions
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !textChannelCreated {
		t.Error("expected new text channel to be created")
	}
	if !linkUpdated {
		t.Error("expected link to be updated with new text channel")
	}
}

func TestLeaveVoice_LastMember(t *testing.T) {
	// Setup mocks
	repo := &mockRepository{}
	repos := &mockRepositories{repo: repo}
	txm := &mockTxManager{}
	discord := &mockDiscordPort{}

	existingLink, _ := voicetext.NewVoiceTextLink("guild123", "voice456", "text789")
	textChannelDeleted := false
	linkDeleted := false

	repo.findByVoiceChannelFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
		return existingLink, nil
	}

	discord.deleteTextChannelFunc = func(ctx context.Context, textChannelID discordid.TextChannelID) error {
		textChannelDeleted = true
		if textChannelID != "text789" {
			t.Errorf("expected TextChannelID text789, got %s", textChannelID)
		}
		return nil
	}

	repo.deleteFunc = func(ctx context.Context, id voicetext.VoiceTextID) error {
		linkDeleted = true
		if id != existingLink.ID() {
			t.Errorf("expected ID %s, got %s", existingLink.ID(), id)
		}
		return nil
	}

	// Create service
	service := NewVoiceTextService(repos, txm, discord)

	// Test
	cmd := LeaveVoiceCommand{
		GuildID:        "guild123",
		VoiceChannelID: "voice456",
		UserID:         "user789",
		IsLastMember:   true,
	}

	err := service.LeaveVoice(context.Background(), cmd)

	// Assertions
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !textChannelDeleted {
		t.Error("expected text channel to be deleted")
	}
	if !linkDeleted {
		t.Error("expected link to be deleted")
	}
}

func TestLeaveVoice_NotLastMember(t *testing.T) {
	// Setup mocks
	repo := &mockRepository{}
	repos := &mockRepositories{repo: repo}
	txm := &mockTxManager{}
	discord := &mockDiscordPort{}

	existingLink, _ := voicetext.NewVoiceTextLink("guild123", "voice456", "text789")
	memberRemoved := false

	repo.findByVoiceChannelFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
		return existingLink, nil
	}

	discord.removeMemberFromTextChannelFunc = func(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID, userID discordid.UserID) error {
		memberRemoved = true
		if textChannelID != "text789" {
			t.Errorf("expected TextChannelID text789, got %s", textChannelID)
		}
		if userID != "user789" {
			t.Errorf("expected UserID user789, got %s", userID)
		}
		return nil
	}

	// Create service
	service := NewVoiceTextService(repos, txm, discord)

	// Test
	cmd := LeaveVoiceCommand{
		GuildID:        "guild123",
		VoiceChannelID: "voice456",
		UserID:         "user789",
		IsLastMember:   false,
	}

	err := service.LeaveVoice(context.Background(), cmd)

	// Assertions
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !memberRemoved {
		t.Error("expected member to be removed")
	}
}

func TestVoiceStateUpdate_NoChange(t *testing.T) {
	// Setup mocks
	repo := &mockRepository{}
	repos := &mockRepositories{repo: repo}
	txm := &mockTxManager{}
	discord := &mockDiscordPort{}

	// Create service
	service := NewVoiceTextService(repos, txm, discord)

	// Test - same channel before and after
	beforeChannel := discordid.VoiceChannelID("voice456")
	afterChannel := discordid.VoiceChannelID("voice456")
	cmd := VoiceStateUpdateCommand{
		GuildID:              "guild123",
		BeforeVoiceChannelID: &beforeChannel,
		AfterVoiceChannelID:  &afterChannel,
		UserID:               "user789",
	}

	err := service.VoiceStateUpdate(context.Background(), cmd)

	// Assertions
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVoiceStateUpdate_JoinChannel(t *testing.T) {
	// Setup mocks
	repo := &mockRepository{}
	repos := &mockRepositories{repo: repo}
	txm := &mockTxManager{}
	discord := &mockDiscordPort{}

	joinCalled := false
	existingLink, _ := voicetext.NewVoiceTextLink("guild123", "voice111", "text111")

	repo.findByVoiceChannelFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
		if voiceChannelID == "voice111" {
			return existingLink, nil
		}
		return nil, voicetext.ErrVoiceTextLinkNotFound
	}

	discord.createTextChannelForVoiceFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (discordid.TextChannelID, error) {
		return "text789", nil
	}

	discord.addMemberToTextChannelFunc = func(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID, userID discordid.UserID) error {
		joinCalled = true
		return nil
	}

	// Mock for leave operation
	discord.getVoiceChannelMemberCountFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (int, error) {
		return 1, nil // Not last member
	}

	discord.removeMemberFromTextChannelFunc = func(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID, userID discordid.UserID) error {
		return nil
	}

	// Create service
	service := NewVoiceTextService(repos, txm, discord)

	// Test - joining a channel from another channel
	beforeChannel := discordid.VoiceChannelID("voice111")
	afterChannel := discordid.VoiceChannelID("voice456")
	cmd := VoiceStateUpdateCommand{
		GuildID:              "guild123",
		BeforeVoiceChannelID: &beforeChannel,
		AfterVoiceChannelID:  &afterChannel,
		UserID:               "user789",
	}

	err := service.VoiceStateUpdate(context.Background(), cmd)

	// Assertions
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !joinCalled {
		t.Error("expected join to be called")
	}
}

func TestVoiceStateUpdate_LeaveChannel(t *testing.T) {
	// Setup mocks
	repo := &mockRepository{}
	repos := &mockRepositories{repo: repo}
	txm := &mockTxManager{}
	discord := &mockDiscordPort{}

	existingLink, _ := voicetext.NewVoiceTextLink("guild123", "voice456", "text789")
	leaveCalled := false

	repo.findByVoiceChannelFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
		return existingLink, nil
	}

	discord.getVoiceChannelMemberCountFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (int, error) {
		return 0, nil // Last member
	}

	discord.deleteTextChannelFunc = func(ctx context.Context, textChannelID discordid.TextChannelID) error {
		leaveCalled = true
		return nil
	}

	// Create service
	service := NewVoiceTextService(repos, txm, discord)

	// Test - leaving a channel to another channel
	beforeChannel := discordid.VoiceChannelID("voice456")
	afterChannel := discordid.VoiceChannelID("voice789")
	cmd := VoiceStateUpdateCommand{
		GuildID:              "guild123",
		BeforeVoiceChannelID: &beforeChannel,
		AfterVoiceChannelID:  &afterChannel,
		UserID:               "user789",
	}

	// Also mock join to new channel
	discord.createTextChannelForVoiceFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (discordid.TextChannelID, error) {
		return "text999", nil
	}

	discord.addMemberToTextChannelFunc = func(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID, userID discordid.UserID) error {
		return nil
	}

	err := service.VoiceStateUpdate(context.Background(), cmd)

	// Assertions
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !leaveCalled {
		t.Error("expected leave to be called")
	}
}

func TestVoiceStateUpdate_SwitchChannels(t *testing.T) {
	// Setup mocks
	repo := &mockRepository{}
	repos := &mockRepositories{repo: repo}
	txm := &mockTxManager{}
	discord := &mockDiscordPort{}

	existingLink1, _ := voicetext.NewVoiceTextLink("guild123", "voice456", "text789")
	leaveCalled := false
	joinCalled := false
	callOrder := []string{}

	repo.findByVoiceChannelFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
		if voiceChannelID == "voice456" {
			return existingLink1, nil
		}
		return nil, voicetext.ErrVoiceTextLinkNotFound
	}

	discord.getVoiceChannelMemberCountFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (int, error) {
		return 1, nil // Not last member
	}

	discord.removeMemberFromTextChannelFunc = func(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID, userID discordid.UserID) error {
		leaveCalled = true
		callOrder = append(callOrder, "leave")
		return nil
	}

	discord.createTextChannelForVoiceFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (discordid.TextChannelID, error) {
		return "text999", nil
	}

	discord.addMemberToTextChannelFunc = func(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID, userID discordid.UserID) error {
		joinCalled = true
		callOrder = append(callOrder, "join")
		return nil
	}

	// Create service
	service := NewVoiceTextService(repos, txm, discord)

	// Test - switching channels
	beforeChannel := discordid.VoiceChannelID("voice456")
	afterChannel := discordid.VoiceChannelID("voice789")
	cmd := VoiceStateUpdateCommand{
		GuildID:              "guild123",
		BeforeVoiceChannelID: &beforeChannel,
		AfterVoiceChannelID:  &afterChannel,
		UserID:               "user789",
	}

	err := service.VoiceStateUpdate(context.Background(), cmd)

	// Assertions
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !leaveCalled {
		t.Error("expected leave to be called")
	}
	if !joinCalled {
		t.Error("expected join to be called")
	}
	if len(callOrder) != 2 || callOrder[0] != "leave" || callOrder[1] != "join" {
		t.Errorf("expected leave to be called before join, got %v", callOrder)
	}
}

func TestJoinVoice_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mockRepository, *mockDiscordPort)
		expectError   bool
	}{
		{
			name: "repository find error",
			setupMocks: func(repo *mockRepository, discord *mockDiscordPort) {
				repo.findByVoiceChannelFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
					return nil, errors.New("database error")
				}
			},
			expectError: true,
		},
		{
			name: "create text channel error",
			setupMocks: func(repo *mockRepository, discord *mockDiscordPort) {
				repo.findByVoiceChannelFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
					return nil, voicetext.ErrVoiceTextLinkNotFound
				}
				discord.createTextChannelForVoiceFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (discordid.TextChannelID, error) {
					return "", errors.New("discord error")
				}
			},
			expectError: true,
		},
		{
			name: "save link error",
			setupMocks: func(repo *mockRepository, discord *mockDiscordPort) {
				repo.findByVoiceChannelFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
					return nil, voicetext.ErrVoiceTextLinkNotFound
				}
				discord.createTextChannelForVoiceFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (discordid.TextChannelID, error) {
					return "text789", nil
				}
				repo.saveFunc = func(ctx context.Context, vtl *voicetext.VoiceTextLink) error {
					return errors.New("save error")
				}
			},
			expectError: true,
		},
		{
			name: "add member error",
			setupMocks: func(repo *mockRepository, discord *mockDiscordPort) {
				repo.findByVoiceChannelFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
					return nil, voicetext.ErrVoiceTextLinkNotFound
				}
				discord.createTextChannelForVoiceFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (discordid.TextChannelID, error) {
					return "text789", nil
				}
				discord.addMemberToTextChannelFunc = func(ctx context.Context, guildID discordid.GuildID, textChannelID discordid.TextChannelID, userID discordid.UserID) error {
					return errors.New("add member error")
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			repos := &mockRepositories{repo: repo}
			txm := &mockTxManager{}
			discord := &mockDiscordPort{}

			tt.setupMocks(repo, discord)

			service := NewVoiceTextService(repos, txm, discord)

			cmd := JoinVoiceCommand{
				GuildID:        "guild123",
				VoiceChannelID: "voice456",
				UserID:         "user789",
			}

			err := service.JoinVoice(context.Background(), cmd)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLeaveVoice_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mockRepository, *mockDiscordPort)
		expectError   bool
	}{
		{
			name: "repository find error",
			setupMocks: func(repo *mockRepository, discord *mockDiscordPort) {
				repo.findByVoiceChannelFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
					return nil, errors.New("database error")
				}
			},
			expectError: true,
		},
		{
			name: "delete text channel error",
			setupMocks: func(repo *mockRepository, discord *mockDiscordPort) {
				existingLink, _ := voicetext.NewVoiceTextLink("guild123", "voice456", "text789")
				repo.findByVoiceChannelFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
					return existingLink, nil
				}
				discord.deleteTextChannelFunc = func(ctx context.Context, textChannelID discordid.TextChannelID) error {
					return errors.New("discord error")
				}
			},
			expectError: true,
		},
		{
			name: "delete link error",
			setupMocks: func(repo *mockRepository, discord *mockDiscordPort) {
				existingLink, _ := voicetext.NewVoiceTextLink("guild123", "voice456", "text789")
				repo.findByVoiceChannelFunc = func(ctx context.Context, guildID discordid.GuildID, voiceChannelID discordid.VoiceChannelID) (*voicetext.VoiceTextLink, error) {
					return existingLink, nil
				}
				repo.deleteFunc = func(ctx context.Context, id voicetext.VoiceTextID) error {
					return errors.New("delete error")
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			repos := &mockRepositories{repo: repo}
			txm := &mockTxManager{}
			discord := &mockDiscordPort{}

			tt.setupMocks(repo, discord)

			service := NewVoiceTextService(repos, txm, discord)

			cmd := LeaveVoiceCommand{
				GuildID:        "guild123",
				VoiceChannelID: "voice456",
				UserID:         "user789",
				IsLastMember:   true,
			}

			err := service.LeaveVoice(context.Background(), cmd)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
