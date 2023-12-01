package common

type Counter interface {
	Inc(labelValues ...string) Counter
}

type Gauge interface {
	Set(value float64, labelValues ...string) Gauge
}

type Meter interface {
	Counter(name, description string, labels []string, prefixes ...string) Counter
	Gauge(name, description string, labels []string, prefixes ...string) Gauge
	Stop()
}
