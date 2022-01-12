package common

type Events struct {
	eventers []Eventer
}

func (es *Events) Trigger(message string) {

	for _, e := range es.eventers {
		e.Trigger(message)
	}
}

func (es *Events) Register(e Eventer) {
	if es != nil {
		es.eventers = append(es.eventers, e)
	}
}

func NewEvents() *Events {
	return &Events{}
}
