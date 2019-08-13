package server

import (
	"errors"
	"fmt"
	"time"

	"github.com/2hdddg/mqtt/conn"
	"github.com/2hdddg/mqtt/logger"
	"github.com/2hdddg/mqtt/packet"
)

func Connect(conn conn.C, au Authorize, log logger.L) (*Session, error) {
	// If server does not receive CONNECT in a reasonable amount of time,
	// the server should close the network connection.
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Wait for CONNECT
	log.Info("Waiting for CONNECT")
	p, err := conn.ReadPacket(5, log)
	if err != nil {
		return nil, err
	}

	c, ok := p.(*packet.Connect)
	if !ok {
		conn.Close()
		log.Error("Unexpected packet")
		return nil, errors.New("Wrong package")
	}

	// If this fails, the server should close the connection without
	// sending a CONNACK
	if c.ProtocolName != "MQTT" && c.ProtocolName != "MQIsdp" {
		conn.Close()
		log.Info("Wrong protocol")
		return nil, errors.New("Invalid protocol")
	}

	// If version check fails, notify client about wrong version
	if c.ProtocolVersion < 4 || c.ProtocolVersion > 5 {
		conn.WritePacket(
			packet.RefuseConnection(packet.ConnRefusedVersion), log)
		conn.Close()
		log.Info("Version not supported")
		return nil, errors.New("Protocol version not supported")
	}

	// Authorization hook
	ret := au.CheckConnect(c)
	if ret != packet.ConnAccepted {
		conn.WritePacket(packet.RefuseConnection(ret), log)
		conn.Close()
		log.Info("Unauthorized")
		return nil, errors.New("External auth refused")
	}

	// Accept connection by acking it
	ack := &packet.AckConnection{
		SessionPresent: false, // TODO:
		RetCode:        packet.ConnAccepted,
	}
	err = conn.WritePacket(ack, log)
	if err != nil {
		conn.Close()
		return nil, errors.New("Failed to send CONNACK")
	}
	log.Info(fmt.Sprintf("Client %s accepted", c.ClientIdentifier))

	return newSession(conn, c), nil
}
