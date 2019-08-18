package client

import (
	"fmt"
	"time"

	"github.com/2hdddg/mqtt/conn"
	"github.com/2hdddg/mqtt/logger"
	"github.com/2hdddg/mqtt/packet"
	"github.com/2hdddg/mqtt/qos"
	"github.com/2hdddg/mqtt/topic"
	"github.com/2hdddg/mqtt/writequeue"
)

type Session struct {
	ClientId   string
	keepAlive  time.Duration
	conn       conn.C
	stopChan   chan bool
	log        logger.L
	wrQueue    *writequeue.Queue
	subscrChan chan *subscriptions
	qos        *qos.QoS
}

func (s *Session) received(px packet.Packet) {
	switch p := px.(type) {
	case *packet.PingResp:
	case *packet.SubscribeAck:
		err := s.qos.ReceivedSubscribeAck(p)
		if err != nil {
			s.log.Error(fmt.Sprintf("Failed to ack subscribe %v", err))
			s.conn.Close()
		}

	default:
		s.log.Error(fmt.Sprintf("Received unhandled packet %t", p))
		s.conn.Close()
	}
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
			if !s.conn.IsClosed() {
				readAsync()
			}

		// Received subscribe request
		case sub := <-s.subscrChan:
			s.subscribe(sub)

		// Read failure
		case err := <-readErrChan:
			s.log.Error(fmt.Sprintf("Receive error %s", err))
			s.conn.Close()

		// Stop requested
		case <-s.stopChan:
			s.stopChan <- true
			s.wrQueue.Flush()
			s.log.Info("Stopped")
			return
		}
	}
}

func (s *Session) Dispose() {
	s.stopChan <- true
	<-s.stopChan
}

type Subscription struct {
	Filter *topic.Filter
	QoS    packet.QoS
}

type SubscribeAck func(subs []Subscription)

type subscriptions struct {
	subs []Subscription
	ack  SubscribeAck
}

func (s *Session) Subscribe(subs []Subscription, ack SubscribeAck) {
	s.subscrChan <- &subscriptions{
		subs: subs,
		ack:  ack,
	}
}

func (s *Session) subscribe(sub *subscriptions) {
	// Repackage to packet format
	subs := []packet.Subscription{}
	for _, x := range sub.subs {
		subs = append(subs, packet.Subscription{
			Topic: x.Filter.String(),
			QoS:   x.QoS,
		})
	}
	s.qos.SendSubscribe(
		&packet.Subscribe{
			Subscriptions: subs,
		},
		// Called when SUBACK received
		func(s *packet.Subscribe, a *packet.SubscribeAck) {
			// Package to external callback format
			// Order of return codes in SUBACK matches above
			for i, _ := range sub.subs {
				sub.subs[i].QoS = a.ReturnCodes[i]
			}
			sub.ack(sub.subs)
		})
}

func NewSession(
	conn conn.C, log logger.L, keepAliveSecs uint16) *Session {

	wrQueue := writequeue.New(conn, log)
	s := &Session{
		ClientId:   "x",
		conn:       conn,
		keepAlive:  time.Duration(keepAliveSecs) * time.Second,
		log:        log,
		stopChan:   make(chan bool),
		wrQueue:    wrQueue,
		subscrChan: make(chan *subscriptions),
		qos:        qos.New(wrQueue, log),
	}

	conn.SetLog(log)

	go s.pump()
	return s
}
