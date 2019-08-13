package logger

import "fmt"

type Session struct {
	id string
}

func NewSession(id string) *Session {
	return &Session{id: id}
}

func (l *Session) Info(s string) {
	fmt.Printf("[%s] INF: %s\n", l.id, s)
}

func (l *Session) Error(s string) {
	fmt.Printf("[%s] ERR: %s\n", l.id, s)
}

func (l *Session) Debug(s string) {
	fmt.Printf("[%s] DBG: %s\n", l.id, s)
}

