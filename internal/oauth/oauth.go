package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"sync"
	"time"
)

type StateRecord struct {
	StateHash    string
	CodeVerifier string
	ProviderID   string
	ExpiresAt    time.Time
}

type Manager struct {
	mu         sync.RWMutex
	states     map[string]*StateRecord
	refreshMus sync.Map
}

func NewManager() *Manager {
	return &Manager{
		states: make(map[string]*StateRecord),
	}
}

func (m *Manager) GenerateState(providerID string, ttl time.Duration) (rawState string, codeChallenge string, err error) {
	stateBytes := make([]byte, 24)
	if _, err := io.ReadFull(rand.Reader, stateBytes); err != nil {
		return "", "", err
	}
	rawState = hex.EncodeToString(stateBytes)

	verifierBytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, verifierBytes); err != nil {
		return "", "", err
	}
	codeVerifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	s256 := sha256.Sum256([]byte(codeVerifier))
	codeChallenge = base64.RawURLEncoding.EncodeToString(s256[:])

	stateHash := HashState(rawState)

	m.mu.Lock()
	m.states[stateHash] = &StateRecord{
		StateHash:    stateHash,
		CodeVerifier: codeVerifier,
		ProviderID:   providerID,
		ExpiresAt:    time.Now().Add(ttl),
	}
	m.mu.Unlock()

	return rawState, codeChallenge, nil
}

func (m *Manager) ConsumeState(rawState string) (*StateRecord, error) {
	stateHash := HashState(rawState)

	m.mu.Lock()
	defer m.mu.Unlock()

	rec, ok := m.states[stateHash]
	if !ok {
		return nil, errors.New("invalid or expired oauth state")
	}

	delete(m.states, stateHash)

	if time.Now().After(rec.ExpiresAt) {
		return nil, errors.New("oauth state has expired")
	}

	return rec, nil
}

func (m *Manager) RefreshLock(accountID string) (unlock func()) {
	val, _ := m.refreshMus.LoadOrStore(accountID, &sync.Mutex{})
	mu := val.(*sync.Mutex)
	mu.Lock()
	return func() {
		mu.Unlock()
	}
}

func HashState(rawState string) string {
	sum := sha256.Sum256([]byte(rawState))
	return hex.EncodeToString(sum[:])
}
