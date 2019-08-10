package server

import (
	"container/list"
)

type writeQueue struct {
	wr      Writer
	q       *list.List
	addChan chan interface{}
}

func newWriteQueue(wr Writer) *writeQueue {
	q := &writeQueue{
		wr:      wr,
		q:       list.New(),
		addChan: make(chan interface{}),
	}
	go q.monitor()
	return q
}

func (q *writeQueue) dispose() {
	close(q.addChan)
}

func (q *writeQueue) add(x interface{}) {
	q.addChan <- x
}

func (q *writeQueue) monitor() {
	writing := false
	wrChan := make(chan error)

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
		go func() {
			err := q.wr.WritePacket(e.Value)
			wrChan <- err
		}()
	}

	for {
		select {
		case x := <-q.addChan:
			q.q.PushBack(x)
			write()
		case <-wrChan:
			writing = false
			write()
		}
	}
}
