package main

import (
	"bufio"
	"fmt"
	"net"

	packet "github.com/2hdddg/mqtt/controlpacket"
	"github.com/2hdddg/mqtt/server"
)

type authorize struct{}
type publisher struct{}

func (a *authorize) CheckConnect(c *packet.Connect) packet.ConnRetCode {
	/* Mosquitto uses / and - in identifier
	if !packet.VerifyClientId(c.ClientIdentifier) {
		return packet.ConnRefusedIdentifier
	}
	*/
	return packet.ConnAccepted
}

func (_ *publisher) Publish(s *server.Session, p *packet.Publish) error {
	fmt.Println("About to publish", p, "from", s)
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

