package client

import (
	"github.com/olahol/melody"
)

type Context interface {
	GetUnderlyingSession() *melody.Session
}

type melodyClientContext struct {
	storage SessionStorage
	session *melody.Session
}

func NewMelodyClientContext(session *melody.Session) Context {
	return &melodyClientContext{
		storage: NewSessionStorageFromMelody(session),
		session: session,
	}
}

func (m *melodyClientContext) GetUnderlyingSession() *melody.Session {
	return m.session
}
