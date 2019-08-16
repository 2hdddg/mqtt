package client

import (
	"errors"

	"github.com/2hdddg/mqtt/conn"
	"github.com/2hdddg/mqtt/logger"
	"github.com/2hdddg/mqtt/packet"
)

type Options struct {
	ClientId        string
	ProtocolVersion byte
	KeepAliveSecs   uint16
}

func Connect(conn conn.C, opts *Options, log logger.L) error {
	c := &packet.Connect{
		ProtocolName:     "MQTT",
		ProtocolVersion:  opts.ProtocolVersion,
		CleanStart:       true,
		KeepAliveSecs:    opts.KeepAliveSecs,
		ClientIdentifier: opts.ClientId,
	}
	err := conn.WritePacket(c, log)
	if err != nil {
		conn.Close()
		return errors.New("Failed to write CONNECT")
	}

	// Clients are allowed to send further Control Packets immediately
	// after sending a CONNECT Packet; Clients need not wait for a CONNACK
	// Packet to arrive from the Server. If the Server rejects the
	// CONNECT, it MUST NOT process any data sent by the Client after the
	// CONNECT Packet [MQTT-3.1.4-5].
	//
	// Keep Connect simple and make it sync, so wait for CONNACK.
	p, err := conn.ReadPacket(c.ProtocolVersion, log)
	if err != nil {
		conn.Close()
		return errors.New("Failed to read packet, waiting for CONNACK")
	}
	a, ok := p.(*packet.AckConnection)
	if !ok {
		conn.Close()
		return errors.New("Expected CONNACK")
	}
	if a.RetCode != packet.ConnAccepted {
		conn.Close()
		return errors.New("Connection not accepted")
	}

	return nil
}
