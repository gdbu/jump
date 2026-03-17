package events

type subscribers map[string][]func(Event)

func (s subscribers) snapshot(key string) []func(Event) {
	fns, ok := s[key]
	if !ok {
		return nil
	}

	cp := make([]func(Event), len(fns))
	copy(cp, fns)
	return cp
}

func (s subscribers) subscribe(fn func(Event), subscribingTo []string) {
	for _, key := range subscribingTo {
		s[key] = append(s[key], fn)
	}
}
