package common

type Labels = map[string]string

type Counter interface {
	Inc() Counter
	Add(value int) Counter
}

type Gauge interface {
	Set(value float64) Gauge
}

type Group interface {
	Clear()
}

type Meter interface {
	Group(name string) Group
	Counter(group, name, description string, labels Labels, prefixes ...string) Counter
	Gauge(group, name, description string, labels Labels, prefixes ...string) Gauge
	Stop()
}
