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
	connected := client.Connect(conn, logger.NewServer())
	if !connected {
		fmt.Println("MQTT server did not allow")
		return
	}
}
