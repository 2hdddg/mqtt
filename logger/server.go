package logger

import "fmt"

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

func (l *Server) Info(s string) {
	fmt.Printf("[SRV] INF: %s\n", s)
}

func (l *Server) Error(s string) {
	fmt.Printf("[SRV] ERR: %s\n", s)
}

func (l *Server) Debug(s string) {
	fmt.Printf("[SRV] DBG: %s\n", s)
}

