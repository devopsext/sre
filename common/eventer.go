package common

import "time"

type Eventer interface {
	Now(name string, attributes map[string]string)
	At(name string, attributes map[string]string, when time.Time)
	Interval(name string, attributes map[string]string, begin, end time.Time)
	Stop()
}
