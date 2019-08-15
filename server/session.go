package server

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

type maybePublish struct {
	topic   topic.Name
	publish packet.Publish
	qoS     packet.QoS
}

type connState byte

const (
	connStateUp           connState = 0
	connStateDisconnected connState = 1
	connStateError        connState = 2
)

type Session struct {
	ClientId         string
	conn             conn.C
	connState        connState
	publisher        Publisher
	connPacket       *packet.Connect
	stopChan         chan bool
	subs             *subscriptions
	wrQueue          *writequeue.Queue
	maybePublishChan chan *maybePublish
	qos              *qos.QoS
	log              logger.L
}

type Authorize interface {
	CheckConnect(c *packet.Connect) packet.ConnRetCode
}

type Publisher interface {
	Publish(s *Session, p *packet.Publish) error
	//TakeOwnership()
	Stopped(s *Session)
}

func (s *Session) EvalPublish(tn *topic.Name, p *packet.Publish) error {
	// Make a copy of the packet to avoid sharing it with other sessions
	m := &maybePublish{
		topic:   *tn,
		publish: *p,
	}
	s.maybePublishChan <- m
	return nil
}

func (s *Session) write(p packet.Packet) {
	i := &writequeue.Item{Packet: p}
	s.wrQueue.Add(i)
}

func (s *Session) receivedSubscribe(sub *packet.Subscribe) {
	// The SUBACK Packet sent by the Server to the Client MUST contain a
	// return code for each Topic Filter/QoS pair. This return code
	// MUST either show the maximum QoS that was granted for that
	// Subscription or indicate that the subscription failed
	// [MQTT-3.8.4-5]. The Server might grant a lower maximum QoS than
	// the subscriber requested. The QoS of Payload Messages sent in
	// response to a Subscription MUST be the minimum of the QoS of the
	// originally published message and the maximum QoS granted by the
	// Server. The server is permitted to send duplicate copies of a
	// message to a subscriber in the case where the original message
	// was published with QoS 1 and the maximum QoS granted was QoS 0
	// [MQTT-3.8.4-6].
	retCodes := make([]packet.QoS, len(sub.Subscriptions))
	for i, _ := range sub.Subscriptions {
		retCodes[i] = s.subs.subscribe(&sub.Subscriptions[i])
	}

	// When the Server receives a SUBSCRIBE Packet from a Client,
	// the Server MUST respond with a SUBACK Packet [MQTT-3.8.4-1].
	// The SUBACK Packet MUST have the same Packet Identifier as the
	// SUBSCRIBE Packet that it is acknowledging [MQTT-3.8.4-2].
	s.write(&packet.SubscribeAck{
		PacketId:    sub.PacketId,
		ReturnCodes: retCodes,
	})
}

func (s *Session) receivedPublish(p *packet.Publish) {
	// Retain:
	// The Server MUST store the Application Message and its QoS, so
	// that it can be delivered to future subscribers whose subscriptions
	// match its topic name [MQTT-3.3.1-5]. When a new subscription is
	// established, the last retained message, if any, on each matching
	// topic name MUST be sent to the subscriber [MQTT-3.3.1-6]. If the
	// Server receives a QoS 0 message with the RETAIN flag set to 1 it
	// MUST discard any message previously retained for that topic. It
	// SHOULD store the new QoS 0 message as the new retained message for
	// that topic, but MAY choose to discard it at any time - if this
	// happens there will be no retained message for that topic.
	if p.Retain {
		s.log.Error("Retain for received PUBLISH not implemented")
		s.enterConnState(connStateError)
		return
	}

	// Delegate to state handler
	err := s.qos.ReceivedPublish(p)
	if err != nil {
		s.enterConnState(connStateError)
		return
	}
}

func (s *Session) maybeSendPublish(m *maybePublish) {
	matched, maxQoS := s.subs.match(&m.topic)
	if !matched {
		return
	}

	s.log.Info(
		fmt.Sprintf("PUBLISH of %d matched, publishing",
			m.publish.PacketId))

	// The publish packet now reflects the publish info received on
	// another session, prepare the publish packet for being sent to
	// the client on this session. Acting on a copy.
	p := &m.publish

	// DUP:
	// The value of the DUP flag from an incoming PUBLISH packet is not
	// propagated when the PUBLISH Packet is sent to subscribers by the
	// Server. The DUP flag in the outgoing PUBLISH packet is set
	// independently to the incoming PUBLISH packet, its value MUST be
	// determined solely by whether the outgoing PUBLISH packet is a
	// retransmission [MQTT-3.3.1-3].
	p.Duplicate = false

	// Retain:
	// When sending a PUBLISH Packet to a Client the Server MUST set the
	// RETAIN flag to 1 if a message is sent as a result of a new
	// subscription being made by a Client [MQTT-3.3.1-8].
	// It MUST set the RETAIN flag to 0 when a PUBLISH Packet is sent to
	// a Client because it matches an established subscription regardless
	// of how the flag was set in the message it received [MQTT-3.3.1-9].
	p.Retain = false // TODO: [MQTT-3.3.1-8]

	// QoS:
	// the Server MUST deliver the message to the Client respecting the
	// maximum QoS of all the matching subscriptions [MQTT-3.3.5-1].
	p.QoS = maxQoS

	// Publish
	err := s.qos.SendPublish(p)
	if err != nil {
		s.enterConnState(connStateError)
		return
	}
}

func (s *Session) received(px packet.Packet) {
	switch p := px.(type) {
	case *packet.Connect:
		// CONNECT not allowed here.
		s.enterConnState(connStateError)

	case *packet.Subscribe:
		s.receivedSubscribe(p)

	case *packet.Publish:
		s.receivedPublish(p)

	case *packet.PingReq:
		s.write(&packet.PingResp{})

	case *packet.Disconnect:
		s.enterConnState(connStateDisconnected)

	case *packet.PublishAck:
		s.qos.ReceivedPublishAck(p)

	default:
		s.log.Error(fmt.Sprintf("Received unhandled packet %t", p))
	}
}

func (s *Session) pump() {
	readPackChan := make(chan packet.Packet)
	readErrChan := make(chan error)

	s.log.Info("Started")

	readAsync := func() {
		keepAlive := time.Duration(s.connPacket.KeepAliveSecs)
		keepAlive *= time.Second
		if keepAlive > 0 {
			s.conn.SetReadDeadline(time.Now().Add(keepAlive))
		}

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
		// Received packet
		case px := <-readPackChan:
			s.received(px)
			if s.connState == connStateUp {
				readAsync()
			}

		// Read failure
		case err := <-readErrChan:
			s.log.Error(fmt.Sprintf("Receive error %s", err))
			s.enterConnState(connStateError)

		// Server received a publish in another session, evaluate
		// the topic and see if it should be sent to this client.
		case maybe := <-s.maybePublishChan:
			if s.connState == connStateUp {
				s.maybeSendPublish(maybe)
			}

		// Stop requested
		case <-s.stopChan:
			s.stopChan <- true
			s.wrQueue.Flush()
			s.log.Info("Stopped")
			return
		}
	}
}

func NewSession(
	conn conn.C, pack *packet.Connect,
	pub Publisher, log logger.L) *Session {

	s := &Session{
		ClientId:         pack.ClientIdentifier,
		conn:             conn,
		connPacket:       pack,
		subs:             newSubscriptions(),
		wrQueue:          writequeue.New(conn, log),
		maybePublishChan: make(chan *maybePublish),
		log:              log,
		publisher:        pub,
		stopChan:         make(chan bool),
	}
	s.qos = qos.New(
		func(p *packet.Publish) {
			log.Info(fmt.Sprintf("Accepted publish of %d", p.PacketId))
			pub.Publish(s, p)
		}, s.wrQueue, s.log)

	go s.pump()
	return s
}

func (s *Session) enterConnState(x connState) {
	s.connState = x
	s.conn.Close()
	s.publisher.Stopped(s)
	s.log.Info("Entered stopped state")
}

func (s *Session) Dispose() {
	s.stopChan <- true
	<-s.stopChan
}
