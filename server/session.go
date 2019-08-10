package server

import (
	"errors"
	"fmt"
	"net"

	"github.com/2hdddg/mqtt/packet"
	"github.com/2hdddg/mqtt/topic"
)

type subscription struct {
	filter topic.Filter
	qoS    packet.QoS
}

type publish struct {
	topic   topic.Name
	publish packet.Publish
	qoS     packet.QoS
}

type Session struct {
	conn       net.Conn
	rd         Reader
	wr         Writer
	publisher  Publisher
	connPacket *packet.Connect
	id         string
	alive      bool
	stopChan   chan bool
	subs       []subscription

	writeWaiting  bool
	writeErrChan  chan error
	writeDoneChan chan bool
	pingReq       bool
	subAcks       []packet.SubscribeAck
	publishChan   chan *publish
	toPublish     []*publish
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
	for i, x := range sub.Subscriptions {
		filter := topic.NewFilter(x.Topic)
		if filter != nil {
			retCodes[i] = x.QoS
			s.subs = append(s.subs, subscription{
				filter: *filter,
				qoS:    x.QoS,
			})
		} else {
			retCodes[i] = packet.QoSFailure
		}
	}

	s.subAcks = append(s.subAcks, packet.SubscribeAck{
		PacketId:    sub.PacketId,
		ReturnCodes: retCodes,
	})
}

func (s *Session) eval() {
	if s.writeWaiting {
		return
	}

	write := func(x interface{}) {
		s.writeWaiting = true
		go func() {
			err := s.wr.WritePacket(x)
			if err != nil {
				s.writeErrChan <- err
			} else {
				s.writeDoneChan <- true
			}
		}()
	}

	if s.pingReq {
		s.writeWaiting = true
		s.pingReq = false
		go write(&packet.PingResp{})
		return
	}

	if len(s.subAcks) > 0 {
		s.writeWaiting = true
		ack := s.subAcks[0]
		s.subAcks = s.subAcks[1:]
		go write(&ack)
		return
	}

	if len(s.toPublish) > 0 {
		s.writeWaiting = true
		pub := s.toPublish[0]
		s.toPublish = s.toPublish[1:]
		fmt.Println("Writing publish")
		go write(&pub.publish)
		return
	}
}

func (s *Session) pump() {
	s.writeErrChan = make(chan error)
	s.writeDoneChan = make(chan bool)
	s.publishChan = make(chan *publish)
	readPackChan := make(chan interface{})
	readErrChan := make(chan error)
	go read(s.rd, readPackChan, readErrChan)

	for {
		select {
		case <-s.stopChan:
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
				s.eval()
			case *packet.Publish:
				if p.QoS == 0 && !p.Retain {
					s.publisher.Publish(s, p)
				} else {
					fmt.Println("Not supported")
				}
			case *packet.PingReq:
				s.pingReq = true
			case *packet.Disconnect:
				return
			default:
				fmt.Printf("Received unhandled: %t\n", p)
			}
			s.eval()
		case <-readErrChan:
			fmt.Println("Receive error")
			return
		case <-s.writeErrChan:
			fmt.Println("Write error")
		case <-s.writeDoneChan:
			s.writeWaiting = false
			s.eval()
		case pub := <-s.publishChan:
			qoS := -1
			for _, sub := range s.subs {
				fmt.Println("Evaluating subscription", sub)
				if sub.filter.Match(&pub.topic) {
					if int(sub.qoS) > qoS {
						qoS = int(sub.qoS)
					}
				}
			}

			if qoS != -1 {
				fmt.Println("Found matching subscription")
				pub.qoS = packet.QoS(qoS)
				// Put the publish in internal queue
				s.toPublish = append(s.toPublish, pub)
				s.eval()
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
	fmt.Println("Requested session stop")
	<-s.stopChan
	fmt.Println("Stopped session")
}
