package common

type Events struct {
	eventers []Eventer
}

func (es *Events) Trigger() {

	for _, e := range es.eventers {
		e.Trigger()
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
