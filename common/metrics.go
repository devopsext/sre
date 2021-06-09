package common

type MetricsCounter struct {
	counters map[Metricer]Counter
	metrics  *Metrics
}

type Metrics struct {
	metricers []Metricer
}

func (msc *MetricsCounter) Inc(values ...string) Counter {

	for _, m := range msc.counters {
		m.Inc(values...)
	}
	return msc
}

func (ms *Metrics) Counter(name, description string, labels []string, prefixes ...string) Counter {

	if len(ms.metricers) <= 0 {
		return nil
	}

	counter := MetricsCounter{
		metrics:  ms,
		counters: make(map[Metricer]Counter),
	}

	for _, m := range ms.metricers {

		c := m.Counter(name, description, labels, prefixes...)
		if c != nil {
			counter.counters[m] = c
		}
	}
	return &counter
}

func (ms *Metrics) Register(m Metricer) {
	if ms != nil {
		ms.metricers = append(ms.metricers, m)
	}
}

func NewMetrics() *Metrics {
	return &Metrics{}
}
