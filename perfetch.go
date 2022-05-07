package perfetch

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type (
	Fetcher[T any] func() (T, error)

	Subscription[T any] struct {
		id string
		ch chan T
	}

	Server[T any] struct {
		fetcher Fetcher[T]
		period  time.Duration
		subs    map[string]Subscription[T]
		data    T

		stop    chan bool
		stopped bool
		ticker  *time.Ticker
		hasData bool

		lock sync.RWMutex
	}
)

var (
	ErrAbort   = fmt.Errorf("abort")
	ErrStopped = fmt.Errorf("stopped fetcher")
)

func (s Subscription[T]) Updates() <-chan T {
	return s.ch
}

func New[T any](period time.Duration, fetcher Fetcher[T]) *Server[T] {
	return &Server[T]{
		fetcher: fetcher,
		period:  period,
		subs:    make(map[string]Subscription[T]),
		stop:    make(chan bool),
		stopped: false,
		hasData: false,
	}
}

func (s *Server[T]) Start() error {
	if s.stopped {
		return ErrStopped
	}
	go s.loop()
	return nil
}

func (s *Server[T]) Stop() {
	if s.stopped {
		return
	}
	s.stop <- true
}

func (s *Server[T]) loop() {
	// initial fetch
	if s.fetch() {
		s.ticker = time.NewTicker(s.period)
	loop:
		for {
			select {
			case <-s.ticker.C:
				if !s.fetch() {
					break loop
				}
			case <-s.stop:
				break loop
			}
		}
		s.ticker.Stop()
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	s.stopped = true

	for _, sub := range s.subs {
		close(sub.ch)
	}
	close(s.stop)
}

func (s *Server[T]) fetch() (cont bool) {
	cont = true

	data, err := s.fetcher()

	if err == nil {
		s.hasData = true
		s.data = data
		s.notify()
	} else if err == ErrAbort {
		cont = false
	}

	return
}

func (s *Server[T]) notify() {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, sub := range s.subs {
		sub.ch <- s.data
	}
}

func (s *Server[T]) Subscribe(chSize int) Subscription[T] {
	sub := Subscription[T]{
		id: uuid.New().String(),
		ch: make(chan T, chSize),
	}

	if s.stopped {
		close(sub.ch)
		return sub
	}

	s.lock.Lock()
	s.subs[sub.id] = sub
	s.lock.Unlock()

	if s.hasData {
		sub.ch <- s.data
	}

	return sub
}

func (s *Server[T]) Unsubscribe(sub Subscription[T]) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.subs, sub.id)
	close(sub.ch)
}
