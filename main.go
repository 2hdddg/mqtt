package main

import (
	"errors"
	"fmt"
	"net"

	"github.com/2hdddg/mqtt/conn"
	"github.com/2hdddg/mqtt/logger"
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
	scid := s.ClientId
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
	delete(sessions, s.ClientId)
	go func() {
		s.Dispose()
	}()
}

func makeSession(netc net.Conn) {
	c := conn.New(netc)
	au := &authorize{}
	pu := &publisher{}

	pack, err := server.Connect(c, au, logger.NewServer())
	if err != nil {
		return
	}
	clientId := pack.ClientIdentifier
	sess := server.NewSession(c, pack, pu, logger.NewSession(clientId))
	sessions[clientId] = sess
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

