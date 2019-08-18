package qos

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/2hdddg/mqtt/logger"
	"github.com/2hdddg/mqtt/packet"
	"github.com/2hdddg/mqtt/writequeue"
)

type tLogger struct {
}

func (l *tLogger) Info(s string) {
	fmt.Println(s)
}

func (l *tLogger) Error(s string) {
	fmt.Println(s)
}

func (l *tLogger) Debug(s string) {
	fmt.Println(s)
}

type tWriter struct {
	written []packet.Packet
}

func (w *tWriter) WritePacket(p packet.Packet, l logger.L) error {
	w.written = append(w.written, p)
	return nil
}

type tAccepter struct {
	ch chan bool
}

func tNewAccepter(t *testing.T) *tAccepter {
	return &tAccepter{
		ch: make(chan bool, 1),
	}
}

func (a *tAccepter) tWaitForAccept() error {
	t := time.NewTimer(1 * time.Second)
	defer t.Stop()
	select {
	case <-a.ch:
		return nil
	case <-t.C:
		return errors.New("Timeout")
	}
	return nil
}

func (a *tAccepter) accept(p *packet.Publish) {
	a.ch <- true
}

func tNew(t *testing.T) (*QoS, *tAccepter, *tWriter, *writequeue.Queue) {
	acc := tNewAccepter(t)
	wr := &tWriter{}
	log := logger.NewServer()
	wrQueue := writequeue.New(wr, log)
	qos := New( /*acc.accept,*/ wrQueue, log)

	return qos, acc, wr, wrQueue
}

func TestReceivedPublishQoS0(t *testing.T) {
	qos, acc, wr, wrQueue := tNew(t)
	p := &packet.Publish{
		QoS: 0,
	}
	qos.ReceivedPublish(p, acc.accept)
	wrQueue.Flush()

	if err := acc.tWaitForAccept(); err != nil {
		t.Errorf("Should have accepted")
	}
	if len(wr.written) > 0 {
		t.Errorf("Should NOT have written")
	}
}

func TestReceivedPublishQoS1(t *testing.T) {
	qos, acc, wr, wrQueue := tNew(t)
	p := &packet.Publish{
		QoS: 1,
	}
	qos.ReceivedPublish(p, acc.accept)
	wrQueue.Flush()

	// Ok to accept before or after PUBACK written
	if err := acc.tWaitForAccept(); err != nil {
		t.Errorf("Should have accepted")
	}
	if len(wr.written) != 1 {
		t.Errorf("Should have written once")
	}
	// Ensure that it is a PUBACK
	_ = wr.written[0].(*packet.PublishAck)
}

func TestSendPublishQoS0(t *testing.T) {
	qos, _, wr, _ := tNew(t)
	pub := &packet.Publish{
		QoS: packet.QoS0,
	}
	err := qos.SendPublish(pub)

	if err != nil {
		t.Errorf("Failed %s", err)
	}
	if len(wr.written) != 1 {
		t.Errorf("Should have written once")
	}
	pubw := wr.written[0].(*packet.Publish)
	if pubw.PacketId != 0 {
		t.Errorf("Packet id should be zero for QoS 0")
	}
	if pubw.Duplicate {
		t.Errorf("DUP should be false for QoS 0")
	}
}

func TestSendPublishQoS1(t *testing.T) {
	qos, _, wr, _ := tNew(t)
	pub := &packet.Publish{
		QoS: packet.QoS1,
	}
	err := qos.SendPublish(pub)

	if err != nil {
		t.Errorf("Failed %s", err)
	}
	if len(wr.written) != 1 {
		t.Errorf("Should have written once")
	}
	pubw := wr.written[0].(*packet.Publish)
	if pubw.PacketId == 0 {
		t.Errorf("Packet id should be NON zero for QoS 0")
	}
	if pubw.Duplicate {
		t.Errorf("DUP should be false for QoS 1")
	}

	// White-box check
	_, exists := qos.sent[pubw.PacketId]
	if !exists {
		t.Errorf("Should keep PUBLISH packet until acked")
	}

	// Acknowledge PUBLISH
	ack := &packet.PublishAck{
		PacketId: pub.PacketId,
	}
	err = qos.ReceivedPublishAck(ack)
	if err != nil {
		t.Errorf("Failed %s", err)
	}

	// White-box check
	_, exists = qos.sent[pubw.PacketId]
	if exists {
		t.Errorf("Should remove PUBLISH packet after acked")
	}
}
