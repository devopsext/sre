package common

import "time"

type Events struct {
	eventers []Eventer
}

func (es *Events) Now(name string, attributes map[string]string) error {
	for _, e := range es.eventers {
		e.Now(name, attributes)
	}
	return nil
}

func (es *Events) At(name string, attributes map[string]string, when time.Time) error {
	for _, e := range es.eventers {
		e.At(name, attributes, when)
	}
	return nil
}

func (es *Events) Interval(name string, attributes map[string]string, begin, end time.Time) error {
	for _, e := range es.eventers {
		e.Interval(name, attributes, begin, end)
	}
	return nil
}

func (es *Events) Stop() {
	for _, e := range es.eventers {
		e.Stop()
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
