package engine

import (
	"sync"
	"time"

	"mhkyle/my-harness/internal/schema"
)

type Session struct {
	ID        string
	WorkDir   string
	CreateAt  time.Time
	UpdatedAt time.Time

	history []schema.Message
	mu      sync.RWMutex
}

func NewSession(id string, workDir string) *Session {
	return &Session{
		ID:        id,
		WorkDir:   workDir,
		CreateAt:  time.Now(),
		UpdatedAt: time.Now(),
		history:   []schema.Message{},
	}
}

func (s *Session) Append(msgs ...schema.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.history = append(s.history, msgs...)
	s.UpdatedAt = time.Now()
	// only append in the mem
}

func (s *Session) GetWorkingMemory(limit int) []schema.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totol := len(s.history)
	if totol <= limit {
		res := make([]schema.Message, totol)
		copy(res, s.history)
		return res
	}

	res := make([]schema.Message, limit)
	copy(res, s.history[totol-limit:])

	for len(res) > 0 {
		if res[0].Role == schema.RoleUser && res[0].ToolCallID != "" {
			res = res[1:]
		} else {
			// do nothing
			break
		}
	}

	return res
}

type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

var GlobalSessionMgr = &SessionManager{
	sessions: make(map[string]*Session),
}

func (sm *SessionManager) GetOrCreate(id string, workDir string) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[id]; exists {
		return session
	}

	newSession := NewSession(id, workDir)
	sm.sessions[id] = newSession
	return newSession
}
