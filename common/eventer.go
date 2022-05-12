package common

import "time"

type Eventer interface {
	Now(name string, attributes map[string]string) error
	At(name string, attributes map[string]string, when time.Time) error
	Interval(name string, attributes map[string]string, begin, end time.Time) error
	Stop()
}
