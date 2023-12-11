package provider

import (
	"fmt"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/VictoriaMetrics/metrics"
	"github.com/devopsext/sre/common"
	"github.com/devopsext/utils"
)

type PrometheusOptions struct {
	URL     string
	Listen  string
	Version string
	Prefix  string
}

type PrometheusCounter struct {
	meter   *PrometheusMeter
	counter *metrics.Counter
}

type PrometheusGauge struct {
	meter *PrometheusMeter
	value float64
	gauge *metrics.Gauge
}

type PrometheusMeter struct {
	options  PrometheusOptions
	logger   common.Logger
	listener *net.Listener
	counters *sync.Map
	gauges   *sync.Map
}

func (p *PrometheusMeter) buildIdent(name string, labels common.Labels, prefixes ...string) string {

	var names []string

	if !utils.IsEmpty(p.options.Prefix) {
		names = append(names, p.options.Prefix)
	}

	names = append(names, prefixes...)
	names = append(names, name)
	name = strings.Join(names, "_")

	lbs := ""
	if len(labels) > 0 {
		arr := []string{}
		for k, v := range labels {
			arr = append(arr, fmt.Sprintf(`%s="%s"`, k, v))
		}
		sort.Strings(arr)
		lbs = fmt.Sprintf("{%s}", strings.Join(arr, ","))
	}
	return fmt.Sprintf(`%s%s`, name, lbs)
}

func (pc *PrometheusCounter) Inc() common.Counter {

	pc.counter.Inc()
	return pc
}

func (pc *PrometheusCounter) Add(value int) common.Counter {

	pc.counter.Add(value)
	return pc
}

func (p *PrometheusMeter) Counter(name, description string, labels common.Labels, prefixes ...string) common.Counter {

	ident := p.buildIdent(name, labels, prefixes...)
	co, ok := p.counters.Load(ident)
	if ok && co != nil {
		return co.(*PrometheusCounter)
	}

	counter := &PrometheusCounter{
		meter:   p,
		counter: metrics.GetOrCreateCounter(ident),
	}
	p.counters.Store(ident, counter)
	return counter
}

func (pg *PrometheusGauge) Set(value float64) common.Gauge {

	pg.value = value
	return pg
}

func (p *PrometheusMeter) Gauge(name, description string, labels common.Labels, prefixes ...string) common.Gauge {

	ident := p.buildIdent(name, labels, prefixes...)
	gg, ok := p.gauges.Load(ident)
	if ok && gg != nil {
		return gg.(*PrometheusGauge)
	}
	var gauge *PrometheusGauge

	gauge = &PrometheusGauge{
		meter: p,
		gauge: metrics.GetOrCreateGauge(ident, func() float64 {
			return gauge.value
		}),
	}
	p.gauges.Store(ident, gauge)
	return gauge
}

func (p *PrometheusMeter) Start() bool {

	p.logger.Info("Start prometheus endpoint...")

	http.HandleFunc(p.options.URL, func(w http.ResponseWriter, req *http.Request) {
		metrics.WritePrometheus(w, false)
	})

	listener, err := net.Listen("tcp", p.options.Listen)
	if err != nil {
		p.logger.Error(err)
		return false
	}
	p.listener = &listener
	p.logger.Info("Prometheus is up. Listening...")
	err = http.Serve(listener, nil)
	if err != nil {
		p.logger.Error(err)
		return false
	}
	return true
}

func (p *PrometheusMeter) StartInWaitGroup(wg *sync.WaitGroup) {

	wg.Add(1)

	go func(wg *sync.WaitGroup) {

		defer wg.Done()
		p.Start()
	}(wg)
}

func (p *PrometheusMeter) Stop() {
	if p.listener != nil {
		l := *p.listener
		l.Close()
	}
}

func NewPrometheusMeter(options PrometheusOptions, logger common.Logger, stdout *Stdout) *PrometheusMeter {

	if logger == nil {
		logger = stdout
	}

	return &PrometheusMeter{
		options:  options,
		logger:   logger,
		counters: &sync.Map{},
		gauges:   &sync.Map{},
	}
}
