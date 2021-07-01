package common

type Counter interface {
	Inc(labelValues ...string) Counter
}

type Meter interface {
	Counter(name, description string, labels []string, prefixes ...string) Counter
}
