package events

import "sync"

func New() *Controller {
	var c Controller
	c.in = make(chan Event, 128)
	c.s = make(subscribers, 32)
	go c.scan()
	return &c
}

type Controller struct {
	mux sync.RWMutex

	in chan Event
	s  subscribers
}

func (c *Controller) New(e Event) {
	c.in <- e
}

func (c *Controller) Subscribe(fn func(Event), subscribingTo ...string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.s.subscribe(fn, subscribingTo)
}

func (c *Controller) scan() {
	for event := range c.in {
		c.notify(event)
	}
}

func (c *Controller) notify(event Event) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	c.s.notify(event)
}
