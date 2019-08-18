package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/2hdddg/mqtt/client"
	"github.com/2hdddg/mqtt/conn"
	"github.com/2hdddg/mqtt/logger"
	"github.com/2hdddg/mqtt/packet"
	t "github.com/2hdddg/mqtt/topic"
)

type connParams struct {
	network   string
	address   string
	clientid  string
	keepAlive uint
}

func connect(cp *connParams) *client.Session {
	nc, err := net.Dial(cp.network, cp.address)
	if err != nil {
		fmt.Println("Failed to connect")
		return nil
	}
	conn := conn.New(nc)

	opts := client.Options{
		ClientId:        cp.clientid,
		ProtocolVersion: 4,
		KeepAliveSecs:   uint16(cp.keepAlive),
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
	args := os.Args[1:]

	// Connection parameters
	cp := connParams{}
	flags := flag.NewFlagSet("Connection parameters", flag.ExitOnError)
	flags.StringVar(&cp.network, "network", "tcp", "Network")
	flags.StringVar(&cp.address, "address", "localhost:6666", "Address")
	flags.StringVar(&cp.clientid, "clientid", "cid1", "Client id")
	flags.UintVar(&cp.keepAlive, "keepalive", 10, "Keep alive in secs")
	flags.Parse(args)
	args = flags.Args()

	sess := connect(&cp)
	if sess == nil {
		return
	}

	for {
		args = flags.Args()
		if len(args) == 0 {
			break
		}
		cmd := args[0]
		args = args[1:]
		switch cmd {
		case "subscribe":
			var topic string
			var qos int

			flags = flag.NewFlagSet("subscribe", flag.ExitOnError)
			flags.StringVar(&topic, "topic", "", "Topic filter")
			flags.IntVar(&qos, "qos", 0, "Quality of service")
			flags.Parse(args)
			if topic == "" {
				fmt.Println("Missing topic")
				continue
			}
			sess.Subscribe([]client.Subscription{
				client.Subscription{
					Filter: t.NewFilter(topic),
					QoS:    packet.QoS(qos),
				}},
				func(acked []client.Subscription) {
					fmt.Println("Subscribed", acked)
				})
		case "sleep":
			var d time.Duration
			flags = flag.NewFlagSet("sleep", flag.ExitOnError)
			flags.DurationVar(&d, "duration", 7*time.Second, "Duration")
			flags.Parse(args)
			time.Sleep(d)
		default:
			fmt.Println("Unknown command:", cmd)
			break
		}
	}
}
