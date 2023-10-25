package chan2

import (
	"sync"

	"github.com/zyedidia/generic/queue"
)

type RChan[T any] interface {
	Read() T
}

type WChan[T any] interface {
	Write(T)
}

type Chan[T any] interface {
	RChan[T]
	WChan[T]
}

type queueElem[T any] struct {
	wg  *sync.WaitGroup
	val *T
}
type chanImpl[T any] struct {
	buf      []T
	capacity uint
	size     uint
	recvIdx  uint
	sendIdx  uint

	mu *sync.Mutex

	recvq *queue.Queue[queueElem[T]]
	sendq *queue.Queue[queueElem[T]]
}

func New[T any](capacity uint) Chan[T] {
	var mu sync.Mutex
	return &chanImpl[T]{
		mu:       &mu,
		buf:      make([]T, capacity),
		capacity: capacity,
		recvq:    queue.New[queueElem[T]](),
		sendq:    queue.New[queueElem[T]](),
	}
}

func (ch *chanImpl[T]) Read() T {
	ch.mu.Lock()

	if ch.size > 0 {
		ret := ch.buf[ch.recvIdx]
		ch.recvIdx++
		ch.recvIdx %= ch.capacity
		ch.size--
		ch.mu.Unlock()
		return ret
	}

	if !ch.sendq.Empty() {
		elem := ch.sendq.Dequeue()
		ch.mu.Unlock()

		ret := elem.val
		elem.wg.Done()
		return *ret
	}

	var ret T
	var wg sync.WaitGroup
	wg.Add(1)
	ch.recvq.Enqueue(queueElem[T]{
		wg:  &wg,
		val: &ret,
	})
	ch.mu.Unlock()

	wg.Wait()
	return ret
}

func (ch *chanImpl[T]) Write(val T) {
	ch.mu.Lock()

	if !ch.recvq.Empty() {
		elem := ch.recvq.Dequeue()
		ch.mu.Unlock()

		*elem.val = val
		elem.wg.Done()
		return
	}

	if ch.size < ch.capacity {
		ch.buf[ch.sendIdx] = val
		ch.sendIdx++
		ch.sendIdx %= ch.capacity
		ch.size++
		ch.mu.Unlock()
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	ch.sendq.Enqueue(queueElem[T]{
		wg:  &wg,
		val: &val,
	})
	ch.mu.Unlock()

	wg.Wait()
}
