package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"

	"github.com/2hdddg/mqtt/packet"
	"github.com/2hdddg/mqtt/server"
	"github.com/2hdddg/mqtt/topic"
)

type authorize struct{}
type publisher struct{}

var sessions map[string]*server.Session = map[string]*server.Session{}

func (a *authorize) CheckConnect(c *packet.Connect) packet.ConnRetCode {
	/* Mosquitto uses / and - in identifier
	if !packet.VerifyClientId(c.ClientIdentifier) {
		return packet.ConnRefusedIdentifier
	}
	*/

	// Check if session already exists
	// No support for clean start yet.
	_, exists := sessions[c.ClientIdentifier]
	if exists {
		return packet.ConnRefusedNotAuthorized
	}

	return packet.ConnAccepted
}

func (_ *publisher) Publish(s *server.Session, p *packet.Publish) error {
	scid := s.ClientId()
	tn := topic.NewName(p.Topic)
	if tn == nil {
		return errors.New("Illegal topic")
	}
	for cid, sess := range sessions {
		if cid == scid {
			continue
		}

		sess.EvalPublish(tn, p)
	}
	return nil
}

func (_ *publisher) Stopped(s *server.Session) {
	delete(sessions, s.ClientId())
	go func() {
		s.Stop()
	}()
}

type sessionLogger struct {
	id string
}

func (l *sessionLogger) Info(s string) {
	fmt.Printf("[%s] INF: %s\n", l.id, s)
}

func (l *sessionLogger) Error(s string) {
	fmt.Printf("[%s] ERR: %s\n", l.id, s)
}

func (l *sessionLogger) Debug(s string) {
	fmt.Printf("[%s] DBG: %s\n", l.id, s)
}

type Conn struct {
	net.Conn
	*packet.Reader
	*packet.Writer
}

func makeSession(conn net.Conn) {
	rd := &packet.Reader{bufio.NewReader(conn)}
	wr := &packet.Writer{conn}
	c := &Conn{conn, rd, wr}
	au := &authorize{}
	pu := &publisher{}

	sess, err := server.Connect(c, au, &sessionLogger{id: "CONN"})
	if err != nil {
		return
	}
	clientId := sess.ClientId()
	sessions[clientId] = sess

	sess.Start(pu, &sessionLogger{id: clientId})
}

func main() {
	l, err := net.Listen("tcp", "localhost:6666")
	if err != nil {
		fmt.Println("Failed to listen")
		return
	}
	defer l.Close()

	for {
		//fmt.Println("Accepting connections")
		conn, err := l.Accept()
		if err != nil {
			//fmt.Println("Failed to accept")
			return
		}
		makeSession(conn)
	}
}

