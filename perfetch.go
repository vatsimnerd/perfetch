package perfetch

import (
	"fmt"
	"sync"
	"time"
)

type Fetcher[T any] func() (T, error)

var (
	ErrAbort   = fmt.Errorf("abort")
	ErrStopped = fmt.Errorf("stopped fetcher")
)

type Server[T any] struct {
	fetcher Fetcher[T]
	period  time.Duration
	subs    []chan T
	data    T

	stop    chan bool
	stopped bool
	ticker  *time.Ticker
	hasData bool

	lock sync.RWMutex
}

func New[T any](period time.Duration, fetcher Fetcher[T]) *Server[T] {
	return &Server[T]{
		fetcher: fetcher,
		period:  period,
		subs:    make([]chan T, 0),
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
		close(sub)
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
		sub <- s.data
	}
}

func (s *Server[T]) Subscribe(chSize int) <-chan T {
	sub := make(chan T, chSize)
	if s.stopped {
		close(sub)
		return sub
	}

	s.lock.Lock()
	s.subs = append(s.subs, sub)
	s.lock.Unlock()

	if s.hasData {
		sub <- s.data
	}

	return sub
}

func (s *Server[T]) Unsubscribe(ch chan T) {
	m := -1
	for i := 0; i < len(s.subs); i++ {
		if ch == s.subs[i] {
			m = i
			break
		}
	}

	if m >= 0 {
		s.lock.Lock()
		defer s.lock.Unlock()
		close(ch)
		s.subs = append(s.subs[:m], s.subs[m+1:]...)
	}
}
