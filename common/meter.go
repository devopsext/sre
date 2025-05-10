package common

type Labels map[string]string

type Counter interface {
	Inc() Counter
	Add(value int) Counter
}

type Gauge interface {
	Set(value float64) Gauge
}

type Histogram interface {
	Observe(value float64) Histogram
}

type Group interface {
	Clear()
}

type Meter interface {
	Counter(group, name, description string, labels Labels, prefixes ...string) Counter
	Gauge(group, name, description string, labels Labels, prefixes ...string) Gauge
	Histogram(group, name, description string, labels Labels, prefixes ...string) Histogram
	Group(name string) Group
	Stop()
}
