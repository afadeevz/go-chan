package chan2_test

import (
	"sync"
	"testing"

	chan2 "github.com/afadeevz/go-chan"
	"github.com/stretchr/testify/assert"
)

func TestChanBufNoBlock(t *testing.T) {
	ch := chan2.New[int](1)
	ch.Write(1)
	assert.Equal(t, 1, ch.Read())
}

func TestChanBufWriteBlock(t *testing.T) {
	ch := chan2.New[int](1)

	ch.Write(1)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.Equal(t, 1, ch.Read())
		assert.Equal(t, 2, ch.Read())
	}()

	ch.Write(2)
	wg.Wait()
}

func TestChanBufReadBlock(t *testing.T) {
	ch := chan2.New[int](1)

	go func() {
		ch.Write(1)
	}()

	assert.Equal(t, 1, ch.Read())
}
