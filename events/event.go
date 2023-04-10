package events

func MakeEvent(key string, value interface{}) (e Event) {
	e.Key = key
	e.Value = value
	return
}

type Event struct {
	Key   string
	Value interface{}
}
