package publish

import (
	"errors"
	"fmt"
	"testing"
	"time"

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

func (w *tWriter) WritePacket(p packet.Packet) error {
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

func TestReceivedPublishQoS0(t *testing.T) {
	acc := tNewAccepter(t)
	wr := &tWriter{}
	wrQueue := writequeue.New(wr)
	rec := NewReceiver(acc.accept, wrQueue, &tLogger{})

	p := &packet.Publish{
		QoS: 0,
	}
	rec.Received(p)
	wrQueue.Flush()

	if err := acc.tWaitForAccept(); err != nil {
		t.Errorf("Should have accepted")
	}
	if len(wr.written) > 0 {
		t.Errorf("Should NOT have written")
	}
}

func TestReceivedPublishQoS1(t *testing.T) {
	acc := tNewAccepter(t)
	wr := &tWriter{}
	wrQueue := writequeue.New(wr)
	rec := NewReceiver(acc.accept, wrQueue, &tLogger{})

	p := &packet.Publish{
		QoS: 1,
	}
	rec.Received(p)
	wrQueue.Flush()

	// Ok to accept before or after PUBACK written
	if err := acc.tWaitForAccept(); err != nil {
		t.Errorf("Should have accepted")
	}
	if len(wr.written) != 1 {
		t.Errorf("Should have written once")
	}
}
