package radio

import (
	"context"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type session struct {
	vc     *discordgo.VoiceConnection
	cancel context.CancelFunc
}

// sessionManager はギルドごとのラジオ再生セッションを管理する
type sessionManager struct {
	mu       sync.Mutex
	sessions map[string]*session
}

func newSessionManager() *sessionManager {
	return &sessionManager{
		sessions: make(map[string]*session),
	}
}

func (m *sessionManager) set(guildID string, s *session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[guildID] = s
}

func (m *sessionManager) get(guildID string) (*session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[guildID]
	return s, ok
}

func (m *sessionManager) stop(guildID string) {
	m.mu.Lock()
	s, ok := m.sessions[guildID]
	if ok {
		delete(m.sessions, guildID)
	}
	m.mu.Unlock()

	if ok {
		s.cancel()
		s.vc.Disconnect()
	}
}
