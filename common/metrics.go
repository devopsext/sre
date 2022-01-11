package common

type MetricsCounter struct {
	counters map[Meter]Counter
	metrics  *Metrics
}

type Metrics struct {
	meters []Meter
}

func (msc *MetricsCounter) Inc(values ...string) Counter {

	for _, m := range msc.counters {
		m.Inc(values...)
	}
	return msc
}

func (ms *Metrics) Counter(name, description string, labels []string, prefixes ...string) Counter {

	counter := MetricsCounter{
		metrics:  ms,
		counters: make(map[Meter]Counter),
	}

	for _, m := range ms.meters {

		c := m.Counter(name, description, labels, prefixes...)
		if c != nil {
			counter.counters[m] = c
		}
	}
	return &counter
}

func (ms *Metrics) Stop() {

	for _, m := range ms.meters {
		m.Stop()
	}
}

func (ms *Metrics) Register(m Meter) {
	if ms != nil {
		ms.meters = append(ms.meters, m)
	}
}

func NewMetrics() *Metrics {
	return &Metrics{}
}
