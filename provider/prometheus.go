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

type PrometheusGroup struct {
	meter *PrometheusMeter
	name  string
	set   *metrics.Set
}

type PrometheusMeter struct {
	options  PrometheusOptions
	logger   common.Logger
	listener *net.Listener
	groups   *sync.Map
}

func (p *PrometheusGroup) Clear() {
	p.set.UnregisterAllMetrics()
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

func (p *PrometheusMeter) Counter(group, name, description string, labels common.Labels, prefixes ...string) common.Counter {

	ident := p.buildIdent(name, labels, prefixes...)

	set := metrics.GetDefaultSet()
	gr := p.findGroup(group)
	if gr != nil {
		set = gr.set
	}

	counter := &PrometheusCounter{
		meter:   p,
		counter: set.GetOrCreateCounter(ident),
	}
	return counter
}

func (pg *PrometheusGauge) Set(value float64) common.Gauge {

	pg.value = value
	return pg
}

func (p *PrometheusMeter) Gauge(group, name, description string, labels common.Labels, prefixes ...string) common.Gauge {

	ident := p.buildIdent(name, labels, prefixes...)

	set := metrics.GetDefaultSet()
	gr := p.findGroup(group)
	if gr != nil {
		set = gr.set
	}

	var gauge *PrometheusGauge
	gauge = &PrometheusGauge{
		meter: p,
		gauge: set.GetOrCreateGauge(ident, func() float64 {
			return gauge.value
		}),
	}
	return gauge
}

func (p *PrometheusMeter) findGroup(name string) *PrometheusGroup {

	gr, ok := p.groups.Load(name)
	if ok && gr != nil {
		return gr.(*PrometheusGroup)
	}
	return nil
}

func (p *PrometheusMeter) Group(name string) common.Group {

	gr := p.findGroup(name)
	if gr != nil {
		return gr
	}

	set := metrics.NewSet()

	group := &PrometheusGroup{
		meter: p,
		name:  name,
		set:   set,
	}
	metrics.RegisterSet(set)
	p.groups.Store(name, group)
	return group
}

func (p *PrometheusMeter) Start() bool {

	p.logger.Info("Start prometheus endpoint...")

	http.HandleFunc(p.options.URL, func(w http.ResponseWriter, req *http.Request) {
		metrics.ExposeMetadata(true)
		defer metrics.ExposeMetadata(false)

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
		options: options,
		logger:  logger,
		groups:  &sync.Map{},
	}
}
