package common

type MetricsCounter struct {
	counters map[Meter]Counter
	metrics  *Metrics
}

type MetricsGauge struct {
	gauges  map[Meter]Gauge
	metrics *Metrics
}

type MetricsHistogram struct {
	histograms map[Meter]Histogram
	metrics    *Metrics
}

type MetricsGroup struct {
	groups  map[Meter]Group
	metrics *Metrics
}

type Metrics struct {
	meters []Meter
}

func (mg *MetricsGroup) Clear() {

	for _, m := range mg.groups {
		m.Clear()
	}
}

func (msc *MetricsCounter) Inc() Counter {

	for _, m := range msc.counters {
		m.Inc()
	}
	return msc
}

func (msc *MetricsCounter) Add(value int) Counter {

	for _, m := range msc.counters {
		m.Add(value)
	}
	return msc
}

func (ms *Metrics) Counter(group, name, description string, labels Labels, prefixes ...string) Counter {

	counter := MetricsCounter{
		metrics:  ms,
		counters: make(map[Meter]Counter),
	}

	for _, m := range ms.meters {

		c := m.Counter(group, name, description, labels, prefixes...)
		if c != nil {
			counter.counters[m] = c
		}
	}
	return &counter
}

func (msg *MetricsGauge) Set(value float64) Gauge {

	for _, m := range msg.gauges {
		m.Set(value)
	}
	return msg
}

func (ms *Metrics) Gauge(group, name, description string, labels Labels, prefixes ...string) Gauge {

	gauge := MetricsGauge{
		metrics: ms,
		gauges:  make(map[Meter]Gauge),
	}

	for _, m := range ms.meters {

		g := m.Gauge(group, name, description, labels, prefixes...)
		if g != nil {
			gauge.gauges[m] = g
		}
	}
	return &gauge
}

func (msg *MetricsHistogram) Observe(value float64) Histogram {
	for _, m := range msg.histograms {
		m.Observe(value)
	}
	return msg
}

func (ms *Metrics) Histogram(group, name, description string, labels Labels, prefixes ...string) Histogram {
	histogram := MetricsHistogram{
		metrics:    ms,
		histograms: make(map[Meter]Histogram),
	}

	for _, m := range ms.meters {
		h := m.Histogram(group, name, description, labels, prefixes...)
		if h != nil {
			histogram.histograms[m] = h
		}
	}
	return &histogram
}

func (ms *Metrics) Group(group string) Group {

	gr := MetricsGroup{
		metrics: ms,
		groups:  make(map[Meter]Group),
	}

	for _, m := range ms.meters {

		g := m.Group(group)
		if g != nil {
			gr.groups[m] = g
		}
	}
	return &gr
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
