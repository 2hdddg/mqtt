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
	fmt.Println("About to publish", p.Topic, "from", scid)
	for cid, sess := range sessions {
		if cid == scid {
			continue
		}

		fmt.Println("Sending publish to session", cid)
		sess.Publish(tn, []byte{})
	}
	return nil
}

func makeSession(conn net.Conn) {
	rd := &packet.Reader{bufio.NewReader(conn)}
	wr := &packet.Writer{conn}
	au := &authorize{}
	pu := &publisher{}

	sess, err := server.Connect(conn, rd, wr, au)
	if err != nil {
		return
	}
	sessions[sess.ClientId()] = sess

	sess.Start(pu)
}

func main() {
	l, err := net.Listen("tcp", "localhost:6666")
	if err != nil {
		fmt.Println("Failed to listen")
		return
	}
	defer l.Close()

	for {
		fmt.Println("Accepting connections")
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Failed to accept")
			return
		}
		makeSession(conn)
	}
}

