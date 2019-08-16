package client

import (
	"fmt"
	"time"

	"github.com/2hdddg/mqtt/conn"
	"github.com/2hdddg/mqtt/logger"
	"github.com/2hdddg/mqtt/packet"
	"github.com/2hdddg/mqtt/writequeue"
)

type Session struct {
	ClientId  string
	keepAlive time.Duration
	conn      conn.C
	stopChan  chan bool
	log       logger.L
	wrQueue   *writequeue.Queue
}

func (s *Session) received(px packet.Packet) {
}

func (s *Session) pump() {
	readPackChan := make(chan packet.Packet)
	readErrChan := make(chan error)
	// TODO: Handle zero!
	// TODO: Pings too often..
	ticker := time.NewTicker(s.keepAlive)

	s.log.Info("Started")

	readAsync := func() {
		go func() {
			p, err := s.conn.ReadPacket(4, s.log)
			if err != nil {
				readErrChan <- err
			} else {
				readPackChan <- p
			}
		}()
	}

	readAsync()
	for {
		select {
		case <-ticker.C:
			pingReq := &packet.PingReq{}
			i := &writequeue.Item{Packet: pingReq}
			s.wrQueue.Add(i)

		// Received packet
		case px := <-readPackChan:
			s.received(px)
			//if s.connState == connStateUp {
			readAsync()
			//}

		// Read failure
		case err := <-readErrChan:
			s.log.Error(fmt.Sprintf("Receive error %s", err))
			//s.enterConnState(connStateError)

		// Stop requested
		case <-s.stopChan:
			s.stopChan <- true
			//s.wrQueue.Flush()
			s.log.Info("Stopped")
			return
		}
	}
}

func (s *Session) Dispose() {
	s.stopChan <- true
	<-s.stopChan
}

func NewSession(conn conn.C, log logger.L, keepAliveSecs uint16) *Session {
	s := &Session{
		ClientId:  "x",
		conn:      conn,
		keepAlive: time.Duration(keepAliveSecs) * time.Second,
		log:       log,
		stopChan:  make(chan bool),
		wrQueue:   writequeue.New(conn, log),
	}

	go s.pump()
	return s
}
