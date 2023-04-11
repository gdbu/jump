package events

type subscribers map[string][]func(Event)

func (s subscribers) notify(e Event) {
	fns, ok := s[e.Key]
	if !ok {
		return
	}

	for _, fn := range fns {
		fn(e)
	}
}

func (s subscribers) subscribe(fn func(Event), subscribingTo []string) {
	for _, key := range subscribingTo {
		s[key] = append(s[key], fn)
	}
}
