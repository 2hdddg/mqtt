package server

import (
	"errors"
	"fmt"
	"net"

	"github.com/2hdddg/mqtt/packet"
	"github.com/2hdddg/mqtt/topic"
)

type publish struct {
	topic   topic.Name
	publish packet.Publish
	qoS     packet.QoS
}

type Session struct {
	conn        net.Conn
	rd          Reader
	wr          Writer
	publisher   Publisher
	connPacket  *packet.Connect
	id          string
	stopChan    chan bool
	subs        *subscriptions
	wrQueue     *writeQueue
	publishChan chan *publish
}

type Reader interface {
	ReadPacket(version uint8) (interface{}, error)
}

type Writer interface {
	WritePacket(packet interface{}) error
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
	s.publishChan <- &publish{topic: *tn, publish: *p}
	return nil
}

func (s *Session) subscribe(sub *packet.Subscribe) {
	retCodes := make([]packet.QoS, len(sub.Subscriptions))
	for i, _ := range sub.Subscriptions {
		retCodes[i] = s.subs.subscribe(&sub.Subscriptions[i])
	}

	s.wrQueue.add(packet.SubscribeAck{
		PacketId:    sub.PacketId,
		ReturnCodes: retCodes,
	})
}

func (s *Session) pump() {
	s.subs = newSubscriptions()
	s.wrQueue = newWriteQueue(s.wr)
	s.publishChan = make(chan *publish)
	readPackChan := make(chan interface{})
	readErrChan := make(chan error)
	go read(s.rd, readPackChan, readErrChan)

	for {
		select {
		case <-s.stopChan:
			s.wrQueue.dispose()
			s.stopChan <- true
			return
		case px := <-readPackChan:
			// Start reader immediately again
			go read(s.rd, readPackChan, readErrChan)

			switch p := px.(type) {
			case *packet.Connect:
				// TODO: Close connection, CONNECT not allowed here.
			case *packet.Subscribe:
				s.subscribe(p)
			case *packet.Publish:
				if p.QoS == 0 && !p.Retain {
					s.publisher.Publish(s, p)
				} else {
					fmt.Println("Not supported")
				}
			case *packet.PingReq:
				s.wrQueue.add(&packet.PingResp{})
			case *packet.Disconnect:
				return
			default:
				fmt.Printf("Received unhandled: %t\n", p)
			}
		case <-readErrChan:
			fmt.Println("Receive error")
			return
		// Server received a publish in another session, evaluate
		// the topic and see if it should be sent to this client.
		case pub := <-s.publishChan:
			matched, qoS := s.subs.match(&pub.topic)
			if matched {
				fmt.Println("Found matching subscription")
				pub.qoS = packet.QoS(qoS)
				// Put the publish in internal queue
				s.wrQueue.add(pub)
			}
		}
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
	s.wrQueue.dispose()
	fmt.Println("Requested session stop")
	<-s.stopChan
	fmt.Println("Stopped session")
}
