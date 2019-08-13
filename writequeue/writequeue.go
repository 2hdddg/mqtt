package writequeue

import (
	"container/list"

	"github.com/2hdddg/mqtt/logger"
	"github.com/2hdddg/mqtt/packet"
)

type Writer interface {
	WritePacket(packet packet.Packet, log logger.L) error
}

type Queue struct {
	wr       Writer
	q        *list.List
	addChan  chan *Item
	stopChan chan bool
	log      logger.L
}

type Item struct {
	Packet  packet.Packet
	Written func()
}

func New(wr Writer) *Queue {
	q := &Queue{
		wr:       wr,
		q:        list.New(),
		addChan:  make(chan *Item),
		stopChan: make(chan bool),
	}
	go q.monitor()
	return q
}

func (q *Queue) SetLogger(log logger.L) {
	q.log = log
}

func (q *Queue) Flush() {
	q.stopChan <- true
	<-q.stopChan
}

func (q *Queue) Add(i *Item) {
	q.addChan <- i
}

func (q *Queue) monitor() {
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
		i := e.Value.(*Item)
		go func() {
			err := q.wr.WritePacket(i.Packet, q.log)
			wrChan <- err
			if i.Written != nil {
				i.Written()
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
