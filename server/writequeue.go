package server

import (
	"container/list"

	"github.com/2hdddg/mqtt/packet"
)

type writeQueue struct {
	wr       Writer
	q        *list.List
	addChan  chan *writeQueueItem
	stopChan chan bool
}

type writeQueueItem struct {
	packet  packet.Packet
	written func()
}

func newWriteQueue(wr Writer) *writeQueue {
	q := &writeQueue{
		wr:       wr,
		q:        list.New(),
		addChan:  make(chan *writeQueueItem),
		stopChan: make(chan bool),
	}
	go q.monitor()
	return q
}

func (q *writeQueue) flush() {
	q.stopChan <- true
	<-q.stopChan
}

func (q *writeQueue) add(i *writeQueueItem) {
	q.addChan <- i
}

func (q *writeQueue) monitor() {
	writing := false
	wrChan := make(chan error)
	stopped := false

	write := func() {
		if writing {
			return
		}
		if q.q.Len() == 0 {
			return
		}
		writing = true
		e := q.q.Front()
		q.q.Remove(e)
		i := e.Value.(*writeQueueItem)
		go func() {
			err := q.wr.WritePacket(i.packet)
			wrChan <- err
			if i.written != nil {
				i.written()
			}
		}()
	}

	for {
		select {
		case x := <-q.addChan:
			q.q.PushBack(x)
			write()
		case <-wrChan:
			writing = false
			if stopped {
				q.stopChan <- false
				return
			}
			write()
		case <-q.stopChan:
			stopped = true
			if !writing {
				q.stopChan <- false
				return
			}
		}
	}
}
