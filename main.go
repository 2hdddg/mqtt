package main

import (
	"bufio"
	"fmt"
	"net"

	"github.com/2hdddg/mqtt/controlpacket"
	"github.com/2hdddg/mqtt/server"
)

func connection(conn net.Conn) {
	//cont := true
	r := &controlpacket.Reader{bufio.NewReader(conn)}
	w := &controlpacket.Writer{conn}

	sess, err := server.Connect(conn, r, w)
	if err != nil {
		fmt.Println("Connection refused:", err)
		return
	}

	fmt.Println("Connected")
	sess.Start()

	/*
		handler := &controlpacket.Handler{
			OnConnect: func(c *controlpacket.Connect) {
				// Protocol error if CONNECT is sent again!
				cont = false
			},
			OnDisconnect: func(d *controlpacket.Disconnect) {
				cont = false
			},
			OnPublish: func(p *controlpacket.Publish) {
				fmt.Println(p)
			},
		}

		for cont {
			err := r.ReadPacket(4, handler)
			if err != nil {
				fmt.Println(err)
				cont = false
			}
		}
		conn.Close()
	*/
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
		go connection(conn)
	}
}

