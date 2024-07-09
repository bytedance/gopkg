package syncx

import (
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var errFoo = errors.New("foo")

func TestPromiseAndFuture(t *testing.T) {
	p := NewPromise()
	f := p.Future()
	p.Set(1, errFoo)
	val, err := f.Get()
	assert.Equal(t, val, 1)
	assert.Equal(t, err, errFoo)
}

func TestPromiseAndFutureConcurrency(t *testing.T) {
	n := runtime.NumCPU() - 1

	ch := make(chan struct{}, n)
	p := NewPromise()
	go func() {
		for i := 0; i < n; i++ {
			ch <- struct{}{}
		}
		time.Sleep(1 * time.Second)
		p.Set(1, errFoo)
	}()

	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ch
			f := p.Future()
			val, err := f.Get()
			assert.Equal(t, val, 1)
			assert.Equal(t, err, errFoo)
		}()
	}
	wg.Wait()
}

func TestPromiseSetTwice(t *testing.T) {
	p := NewPromise()
	p.Set(1, nil)
	assert.Panics(t, func() {
		p.Set(1, nil)
	})
}
