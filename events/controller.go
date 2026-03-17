package events

import "sync"

func New() *Controller {
	var c Controller
	c.s = make(subscribers, 32)
	c.cond = sync.NewCond(&c.queueMux)
	go c.scan()
	return &c
}

type Controller struct {
	mux sync.RWMutex

	queueMux sync.Mutex
	cond     *sync.Cond
	queue    []Event

	s subscribers
}

func (c *Controller) New(e Event) {
	c.appendEvent(e)
	c.cond.Signal()
}

func (c *Controller) Subscribe(fn func(Event), subscribingTo ...string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.s.subscribe(fn, subscribingTo)
}

func (c *Controller) scan() {
	for {
		e := c.getNextEvent()
		c.notify(e)
	}
}

func (c *Controller) appendEvent(e Event) {
	c.queueMux.Lock()
	defer c.queueMux.Unlock()
	c.queue = append(c.queue, e)
}

func (c *Controller) getNextEvent() (e Event) {
	c.queueMux.Lock()
	defer c.queueMux.Unlock()
	for len(c.queue) == 0 {
		c.cond.Wait()
	}

	e = c.queue[0]
	c.queue[0] = Event{}
	c.queue = c.queue[1:]
	return e
}

func (c *Controller) getFuncs(key string) (out []func(Event)) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.s.snapshot(key)
}

func (c *Controller) notify(event Event) {
	fns := c.getFuncs(event.Key)

	for _, fn := range fns {
		fn(event)
	}
}
