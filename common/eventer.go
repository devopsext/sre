package common

import "time"

type Eventer interface {
	Now(name string, message string, attributes map[string]string) error
	At(name string, message string, attributes map[string]string, when time.Time) error
	Interval(name string, message string, attributes map[string]string, begin, end time.Time) error
	Stop()
}
