package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/2hdddg/mqtt/client"
	"github.com/2hdddg/mqtt/conn"
	"github.com/2hdddg/mqtt/logger"
	"github.com/2hdddg/mqtt/packet"
	t "github.com/2hdddg/mqtt/topic"
)

func connect() *client.Session {
	nc, err := net.Dial("tcp", "localhost:6666")
	if err != nil {
		fmt.Println("Failed to connect")
		return nil
	}
	//defer nc.Close()
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
		return nil
	}
	return client.NewSession(conn, log, opts.KeepAliveSecs)
}

func main() {
	if len(os.Args) <= 1 {
		return
	}

	var topic string
	var qos int

	flags := flag.NewFlagSet("x", flag.ExitOnError)
	cmd := os.Args[1]
	switch cmd {
	case "subscribe":
		flags.StringVar(&topic, "topic", "", "Topic filter")
		flags.IntVar(&qos, "qos", 0, "Quality of service")
	default:
		fmt.Println("Unknown command:", cmd)
		return
	}
	flags.Parse(os.Args[2:])
	//fmt.Println(flags.Args())

	switch cmd {
	case "subscribe":
		if topic == "" {
			fmt.Println("Missing topic")
			return
		}
	}

	sess := connect()
	if sess == nil {
		return
	}

	switch cmd {
	case "subscribe":
		sess.Subscribe([]client.Subscription{
			client.Subscription{
				Filter: t.NewFilter(topic),
				QoS:    packet.QoS(qos),
			}},
			func(acked []client.Subscription) {
				fmt.Println("Subscribed")
			})
	}

	for {
	}
}
