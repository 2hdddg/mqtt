package main

import (
	"fmt"
	"net"

	"github.com/2hdddg/mqtt/client"
	"github.com/2hdddg/mqtt/conn"
	"github.com/2hdddg/mqtt/logger"
)

func main() {
	nc, err := net.Dial("tcp", "localhost:6666")
	if err != nil {
		fmt.Println("Failed to connect")
		return
	}
	defer nc.Close()
	conn := conn.New(nc)

	opts := client.Options{
		ClientId:        "a client",
		ProtocolVersion: 4,
		KeepAliveSecs:   3,
	}

	log := logger.NewServer()
	err = client.Connect(conn, &opts, log)
	if err != nil {
		fmt.Println("MQTT server did not allow", err)
		return
	}

	_ = client.NewSession(conn, log, opts.KeepAliveSecs)
	for {
	}
}
