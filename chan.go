package chan2

import "sync"

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

type chanImpl[T any] struct {
	buf      []T
	capacity uint
	size     uint
	recvIdx  uint
	sendIdx  uint

	mu       *sync.Mutex
	recvCond sync.Cond
	sendCond sync.Cond
}

func New[T any](capacity uint) Chan[T] {
	var mu sync.Mutex
	return &chanImpl[T]{
		mu:       &mu,
		buf:      make([]T, capacity),
		capacity: capacity,
		recvCond: *sync.NewCond(&mu),
		sendCond: *sync.NewCond(&mu),
	}
}

func (ch *chanImpl[T]) Read() T {
	defer ch.sendCond.Signal()

	ch.mu.Lock()
	defer ch.mu.Unlock()

	for ch.size == 0 {
		ch.recvCond.Wait()
	}

	ret := ch.buf[ch.recvIdx]
	ch.recvIdx++
	ch.recvIdx %= ch.capacity
	ch.size--
	return ret
}

func (ch *chanImpl[T]) Write(val T) {
	defer ch.recvCond.Signal()

	ch.mu.Lock()
	defer ch.mu.Unlock()

	for ch.size == ch.capacity {
		ch.sendCond.Wait()
	}

	ch.buf[ch.sendIdx] = val
	ch.sendIdx++
	ch.sendIdx %= ch.capacity
	ch.size++
}
