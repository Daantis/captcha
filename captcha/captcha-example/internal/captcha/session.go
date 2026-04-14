package captcha

import (
	"math/rand"
	"sync"
	"time"
)

type EntityToken struct {
	ID   uint16
	Kind string
	Key  string
}

type ScheduledPush struct {
	At    time.Time
	Kind  string
	Step  int
	Extra int
}

type SessionState struct {
	mu sync.Mutex

	ID         string
	ModeID     string
	Difficulty DifficultyProfile
	Seed       int64
	RNG        *rand.Rand

	CreatedAt  time.Time
	LastSeenAt time.Time
	TTL        time.Duration

	Started bool
	Done    bool

	Phase        uint8
	ClientSeq    uint16
	NextFrameSeq uint16
	LastAckFrame uint16
	NextToken    uint16

	Errors    int
	Anomalies int
	UserEvents int

	Completed int
	Total     int

	ActiveEntities map[uint16]EntityToken
	PendingPushes  []ScheduledPush
	CurrentView    ViewModel
	ModeState      any
}

func NewSessionState(id, modeID string, profile DifficultyProfile, now time.Time) *SessionState {
	seed := now.UnixNano() ^ int64(len(id)<<8)
	return &SessionState{
		ID:             id,
		ModeID:         modeID,
		Difficulty:     profile,
		Seed:           seed,
		RNG:            rand.New(rand.NewSource(seed)),
		CreatedAt:      now,
		LastSeenAt:     now,
		TTL:            defaultSessionTTL,
		ActiveEntities: make(map[uint16]EntityToken),
	}
}

func (s *SessionState) Touch(now time.Time) {
	s.LastSeenAt = now
}

func (s *SessionState) IsExpired(now time.Time) bool {
	return now.Sub(s.LastSeenAt) > s.TTL
}

func (s *SessionState) BeginFrame() {
	s.Phase++
	s.NextFrameSeq++
	s.ActiveEntities = make(map[uint16]EntityToken)
}

func (s *SessionState) NextEntity(kind, key string) uint16 {
	s.NextToken++
	id := s.NextToken
	s.ActiveEntities[id] = EntityToken{ID: id, Kind: kind, Key: key}
	return id
}

func (s *SessionState) AddPush(at time.Time, kind string, step int, extra int) {
	s.PendingPushes = append(s.PendingPushes, ScheduledPush{
		At:    at,
		Kind:  kind,
		Step:  step,
		Extra: extra,
	})
}

func (s *SessionState) PopDuePush(now time.Time, kind string) (ScheduledPush, bool) {
	for i, push := range s.PendingPushes {
		if push.Kind != kind || push.At.After(now) {
			continue
		}

		s.PendingPushes = append(s.PendingPushes[:i], s.PendingPushes[i+1:]...)
		return push, true
	}

	return ScheduledPush{}, false
}

func (s *SessionState) ProgressPercent() uint8 {
	if s.Total <= 0 {
		return 0
	}

	value := (s.Completed * 100) / s.Total
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}
	return uint8(value)
}

func (s *SessionState) Confidence() int32 {
	confidence := 98
	confidence -= s.Errors * 10
	confidence -= s.Anomalies * 14

	if s.UserEvents < 5 {
		confidence -= 20
	}

	if duration := int(time.Since(s.CreatedAt) / time.Second); duration > 45 {
		confidence -= (duration - 45) / 5
	}

	if confidence < 1 {
		confidence = 1
	}
	if confidence > 100 {
		confidence = 100
	}
	return int32(confidence)
}

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*SessionState
	ttl      time.Duration
}

func NewSessionStore(ttl time.Duration) *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*SessionState),
		ttl:      ttl,
	}
}

func (s *SessionStore) Put(session *SessionState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session.TTL = s.ttl
	s.sessions[session.ID] = session
}

func (s *SessionStore) Get(id string) (*SessionState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[id]
	return session, ok
}

func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, id)
}

func (s *SessionStore) Snapshot() []*SessionState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*SessionState, 0, len(s.sessions))
	for _, session := range s.sessions {
		out = append(out, session)
	}
	return out
}
