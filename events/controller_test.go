package events

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestControllerAllowsReentrantPublish(t *testing.T) {
	c := New()

	done := make(chan struct{})
	var handled atomic.Int32

	c.Subscribe(func(e Event) {
		switch e.Key {
		case "root":
			for i := range 256 {
				c.New(MakeEvent("child", i))
			}
		case "child":
			if handled.Add(1) == 256 {
				close(done)
			}
		}
	}, "root", "child")

	c.New(MakeEvent("root", nil))

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for reentrant events to drain")
	}
}

func TestControllerAllowsSubscribeDuringNotify(t *testing.T) {
	c := New()

	done := make(chan struct{})

	c.Subscribe(func(e Event) {
		c.Subscribe(func(inner Event) {
			if inner.Key == "second" {
				close(done)
			}
		}, "second")
	}, "first")

	c.New(MakeEvent("first", nil))
	c.New(MakeEvent("second", nil))

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for dynamically subscribed handler")
	}
}
