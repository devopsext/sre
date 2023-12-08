package common

type MetricsCounter struct {
	counters map[Meter]Counter
	metrics  *Metrics
}

type MetricsGauge struct {
	gauges  map[Meter]Gauge
	metrics *Metrics
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

func (msg *MetricsGauge) WithLabels(labels Labels) Gauge {
	for _, m := range msg.gauges {
		m.WithLabels(labels)
	}
	return msg
}

func (msg *MetricsGauge) Set(value float64, values ...string) Gauge {

	for _, m := range msg.gauges {
		m.Set(value, values...)
	}
	return msg
}

func (ms *Metrics) Gauge(name, description string, labels []string, prefixes ...string) Gauge {

	gauge := MetricsGauge{
		metrics: ms,
		gauges:  make(map[Meter]Gauge),
	}

	for _, m := range ms.meters {

		g := m.Gauge(name, description, labels, prefixes...)
		if g != nil {
			gauge.gauges[m] = g
		}
	}
	return &gauge
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
