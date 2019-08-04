package main

import (
	"bufio"
	"fmt"
	"net"

	packet "github.com/2hdddg/mqtt/controlpacket"
	"github.com/2hdddg/mqtt/server"
)

type authorize struct{}

func (a *authorize) CheckConnect(c *packet.Connect) packet.ConnRetCode {
	/* Mosquitto uses / and - in identifier
	if !packet.VerifyClientId(c.ClientIdentifier) {
		return packet.ConnRefusedIdentifier
	}
	*/
	return packet.ConnAccepted
}

func makeSession(conn net.Conn) {
	rd := &packet.Reader{bufio.NewReader(conn)}
	wr := &packet.Writer{conn}
	au := &authorize{}

	sess, err := server.Connect(conn, rd, wr, au)
	if err != nil {
		return
	}

	sess.Start()
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

