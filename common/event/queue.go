package event

import (
	"errors"
	"io"
	"sync"
	"time"
)

var EventReadTimeout = errors.New("EventQueue read timeout")
var EventWriteTimeout = errors.New("EventQueue write timeout")

type EventQueue struct {
	closed bool
	mutex  sync.Mutex
	peek   Event
	queue  chan Event
}

func (q *EventQueue) Publish(ev Event, timeout time.Duration) error {
	for {
		select {
		case q.queue <- ev:
			return nil
		case <-time.After(timeout):
			return EventWriteTimeout
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}
}
func (q *EventQueue) Close() {
	if !q.closed {
		q.closed = true
		close(q.queue)
	}
}

func (q *EventQueue) Peek(timeout time.Duration) (Event, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if nil != q.peek {
		return q.peek, nil
	}
	select {
	case ev := <-q.queue:
		if nil == ev {
			return nil, io.EOF
		}
		q.peek = ev
		return ev, nil
	case <-time.After(timeout):
		return nil, EventReadTimeout
	}
}
func (q *EventQueue) ReadPeek() Event {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	ev := q.peek
	q.peek = nil
	return ev
}

func (q *EventQueue) Read(timeout time.Duration) (Event, error) {
	select {
	case ev := <-q.queue:
		if nil == ev {
			return nil, io.EOF
		}
		return ev, nil
	case <-time.After(timeout):
		return nil, EventReadTimeout
	}
}

func NewEventQueue() *EventQueue {
	q := new(EventQueue)
	q.queue = make(chan Event, 10)
	return q
}