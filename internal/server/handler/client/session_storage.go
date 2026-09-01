package client

import (
	"sync"

	"github.com/olahol/melody"
)

type SessionStorage interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
}

type melodySessionAdapter struct {
	session *melody.Session
}

func (m *melodySessionAdapter) Get(key string) (interface{}, bool) {
	return m.session.Get(key)
}

func (m *melodySessionAdapter) Set(key string, value interface{}) {
	m.session.Set(key, value)
}

func NewSessionStorageFromMelody(session *melody.Session) SessionStorage {
	return &melodySessionAdapter{session: session}
}

type mapSessionStorage struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

func NewMapSessionStorage() SessionStorage {
	return &mapSessionStorage{
		data: make(map[string]interface{}),
	}
}

func (m *mapSessionStorage) Get(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.data[key]
	return value, ok
}

func (m *mapSessionStorage) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}
