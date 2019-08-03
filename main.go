package main

import (
	"bufio"
	"fmt"
	"net"

	"github.com/2hdddg/mqtt/controlpacket"
	"github.com/2hdddg/mqtt/server"
)

func makeSession(conn net.Conn) {
	rd := &controlpacket.Reader{bufio.NewReader(conn)}
	wr := &controlpacket.Writer{conn}

	sess, err := server.Connect(conn, rd, wr)
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

