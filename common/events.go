package common

import "time"

type Events struct {
	eventers []Eventer
}

func (es *Events) Now(name string, attributes map[string]string) {
	for _, e := range es.eventers {
		e.Now(name, attributes)
	}
}

func (es *Events) At(name string, attributes map[string]string, when time.Time) {
	for _, e := range es.eventers {
		e.At(name, attributes, when)
	}
}

func (es *Events) Interval(name string, attributes map[string]string, begin, end time.Time) {
	for _, e := range es.eventers {
		e.Interval(name, attributes, begin, end)
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
