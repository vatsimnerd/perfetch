package perfetch

import (
	"fmt"
	"testing"
	"time"
)

func TestFetchDuration(t *testing.T) {
	c := 0
	fetcher := func() (string, error) {
		c++
		return fmt.Sprintf("%d", c), nil
	}

	pf := New(100*time.Millisecond, fetcher)
	sub := pf.Subscribe(100)

	select {
	case data := <-sub:
		t.Errorf("unexpected string '%s' on channel", data)
		return
	default:
	}

	err := pf.Start()
	if err != nil {
		t.Errorf("error starting server: %v", err)
	}
	defer pf.Stop()

	time.Sleep(50 * time.Millisecond)
	select {
	case data := <-sub:
		if data != "1" {
			t.Errorf("expected '1' to be on channel, got '%s'", data)
			return
		}
	default:
		t.Error("expected a new string on channel")
		return
	}

	select {
	case data := <-sub:
		t.Errorf("unexpected string '%s' on channel", data)
		return
	default:
	}

	time.Sleep(150 * time.Millisecond)

	select {
	case data := <-sub:
		if data != "2" {
			t.Errorf("expected '2' to be on channel, got '%s'", data)
			return
		}
	default:
		t.Error("expected a new string on channel")
		return
	}
}

func TestSubscribeDataReady(t *testing.T) {
	c := 0
	fetcher := func() (string, error) {
		c++
		return fmt.Sprintf("%d", c), nil
	}

	pf := New(100*time.Millisecond, fetcher)

	err := pf.Start()
	if err != nil {
		t.Errorf("error starting server: %v", err)
	}
	defer pf.Stop()

	time.Sleep(50 * time.Millisecond)

	sub := pf.Subscribe(100)
	select {
	case data := <-sub:
		if data != "1" {
			t.Errorf("expected '1' to be on channel, got '%s'", data)
			return
		}
	default:
		t.Error("expected a new string on channel")
		return
	}

	select {
	case data := <-sub:
		t.Errorf("unexpected string '%s' on channel", data)
		return
	default:
	}

	time.Sleep(150 * time.Millisecond)

	select {
	case data := <-sub:
		if data != "2" {
			t.Errorf("expected '2' to be on channel, got '%s'", data)
			return
		}
	default:
		t.Error("expected a new string on channel")
		return
	}
}

func TestFetchAbort(t *testing.T) {
	fetcher := func() (string, error) {
		return "", ErrAbort
	}

	pf := New(100*time.Millisecond, fetcher)

	err := pf.Start()
	if err != nil {
		t.Errorf("error starting server: %v", err)
	}
	defer pf.Stop()

	time.Sleep(100 * time.Millisecond)

	sub := pf.Subscribe(100)

	for data := range sub {
		fmt.Println(data)
	}
}
