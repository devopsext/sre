package common

type Labels = map[string]string

type Counter interface {
	Inc() Counter
	Add(value int) Counter
}

type Gauge interface {
	Set(value float64) Gauge
}

type Meter interface {
	Counter(name, description string, labels Labels, prefixes ...string) Counter
	Gauge(name, description string, labels Labels, prefixes ...string) Gauge
	Stop()
}
