package server

import (
	"errors"
	"fmt"
	"net"

	"github.com/2hdddg/mqtt/packet"
	"github.com/2hdddg/mqtt/topic"
)

type maybePublish struct {
	topic   topic.Name
	publish packet.Publish
	qoS     packet.QoS
}

type Session struct {
	conn             net.Conn
	rd               Reader
	publisher        Publisher
	connPacket       *packet.Connect
	id               string
	stopChan         chan bool
	subs             *subscriptions
	wrQueue          *writeQueue
	maybePublishChan chan *maybePublish
}

type Reader interface {
	ReadPacket(version uint8) (packet.Packet, error)
}

type Writer interface {
	WritePacket(packet packet.Packet) error
}

type Authorize interface {
	CheckConnect(c *packet.Connect) packet.ConnRetCode
}

type Publisher interface {
	Publish(s *Session, p *packet.Publish) error
}

func read(
	rd Reader, packetChan chan interface{}, errorChan chan error) {

	// TODO: Set read deadline
	p, err := rd.ReadPacket(4)
	if err != nil {
		errorChan <- err
	} else {
		packetChan <- p
	}
}

func (s *Session) ClientId() string {
	return s.id
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

func (s *Session) receivedSubscribe(sub *packet.Subscribe) {
	retCodes := make([]packet.QoS, len(sub.Subscriptions))
	for i, _ := range sub.Subscriptions {
		retCodes[i] = s.subs.subscribe(&sub.Subscriptions[i])
	}

	s.wrQueue.add(&packet.SubscribeAck{
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
		fmt.Println("Retain NOT implemented")
		return
	}

	if p.QoS > packet.QoS0 {
		fmt.Println("QoS above 0 is not implemented")
		return
	}

	// Notify framework that we received a PUBLISH packet, it's up
	// to the framework to distribute this to other sessions.
	s.publisher.Publish(s, p)
}

func (s *Session) maybeSendPublish(m *maybePublish) {
	matched, maxQoS := s.subs.match(&m.topic)
	if !matched {
		return
	}

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
	s.wrQueue.add(p)
}

func (s *Session) received(px packet.Packet) {
	switch p := px.(type) {
	case *packet.Connect:
		// TODO: Close connection, CONNECT not allowed here.

	case *packet.Subscribe:
		s.receivedSubscribe(p)

	case *packet.Publish:
		s.receivedPublish(p)

	case *packet.PingReq:
		s.wrQueue.add(&packet.PingResp{})

	case *packet.Disconnect:
		return

	default:
		fmt.Printf("Received unhandled: %t\n", p)
	}
}

func (s *Session) pump() {
	readPackChan := make(chan interface{})
	readErrChan := make(chan error)

	// Start reading immediately
	go read(s.rd, readPackChan, readErrChan)

	for {
		select {
		// Received packet
		case px := <-readPackChan:
			// Start reader immediately again
			go read(s.rd, readPackChan, readErrChan)
			s.received(px)

		// Read failure
		case <-readErrChan:
			fmt.Println("Receive error")
			return

		// Server received a publish in another session, evaluate
		// the topic and see if it should be sent to this client.
		case maybe := <-s.maybePublishChan:
			s.maybeSendPublish(maybe)

		// Stop requested
		case <-s.stopChan:
			s.stopChan <- true
			s.wrQueue.flush()
			return
		}
	}
}

func newSession(conn net.Conn, rd Reader, wr Writer,
	connect *packet.Connect) *Session {

	return &Session{
		conn:             conn,
		rd:               rd,
		connPacket:       connect,
		id:               connect.ClientIdentifier,
		subs:             newSubscriptions(),
		wrQueue:          newWriteQueue(wr),
		maybePublishChan: make(chan *maybePublish),
	}
}

func (s *Session) Start(p Publisher) error {
	if s.stopChan != nil {
		return errors.New("Already started")
	}

	s.publisher = p
	s.stopChan = make(chan bool)
	go s.pump()
	fmt.Println("Started session")

	return nil
}

func (s *Session) Stop() {
	if s.stopChan == nil {
		return
	}

	s.stopChan <- true
	fmt.Println("Requested session stop")
	<-s.stopChan
	fmt.Println("Stopped session")
}
