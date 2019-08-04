package server

import (
	"errors"
	"fmt"
	"net"

	packet "github.com/2hdddg/mqtt/controlpacket"
)

type Session struct {
	conn          net.Conn
	rd            Reader
	wr            Writer
	connPacket    *packet.Connect
	alive         bool
	stopChan      chan bool
	writeWaiting  bool
	writeErrChan  chan error
	writeDoneChan chan bool
	pingReq       bool
}

type Reader interface {
	ReadPacket(version uint8) (interface{}, error)
}

type Writer interface {
	WriteAckConnection(ack *packet.AckConnection) error
	WritePingResp(resp *packet.PingResp) error
}

func read(
	rd Reader, packetChan chan interface{}, errorChan chan error) {

	// Set read deadline
	p, err := rd.ReadPacket(4)
	if err != nil {
		errorChan <- err
	} else {
		packetChan <- p
	}
}

func (s *Session) eval() {
	if s.writeWaiting {
		return
	}

	if s.pingReq {
		s.writeWaiting = true
		s.pingReq = false
		go func() {
			err := s.wr.WritePingResp(&packet.PingResp{})
			if err != nil {
				s.writeErrChan <- err
			} else {
				s.writeDoneChan <- true
			}
		}()
		return
	}
}

func (s *Session) pump() {
	s.writeErrChan = make(chan error)
	s.writeDoneChan = make(chan bool)
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
			case *packet.Publish:
				if p.QoS == 0 && !p.Retain {
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
		}
	}
}

func (s *Session) Start() error {
	if s.stopChan != nil {
		return errors.New("Already started")
	}

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
