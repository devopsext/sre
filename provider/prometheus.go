package provider

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/devopsext/sre/common"
	"github.com/devopsext/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusOptions struct {
	URL     string
	Listen  string
	Version string
	Prefix  string
}

type PrometheusCounter struct {
	meter      *PrometheusMeter
	counterVec *prometheus.CounterVec
}

type PrometheusGauge struct {
	meter    *PrometheusMeter
	gaugeVec *prometheus.GaugeVec
}

type PrometheusMeter struct {
	options      PrometheusOptions
	logger       common.Logger
	callerOffset int
	listener     *net.Listener
}

func (pc *PrometheusCounter) Inc(labelValues ...string) common.Counter {

	_, file, line := utils.CallerGetInfo(pc.meter.callerOffset + 3)
	newValues := append(labelValues, fmt.Sprintf("%s:%d", file, line))

	pc.counterVec.WithLabelValues(newValues...).Inc()
	return pc
}

func (p *PrometheusMeter) Counter(name, description string, labels []string, prefixes ...string) common.Counter {

	var names []string

	if !utils.IsEmpty(p.options.Prefix) {
		names = append(names, p.options.Prefix)
	}

	names = append(names, prefixes...)
	names = append(names, name)
	newName := strings.Join(names, "_")

	config := prometheus.CounterOpts{
		Name: newName,
		Help: description,
	}

	labels = append(labels, "file")

	counterVec := prometheus.NewCounterVec(config, labels)
	prometheus.Register(counterVec)

	return &PrometheusCounter{
		meter:      p,
		counterVec: counterVec,
	}
}

func (pc *PrometheusGauge) Set(value float64, labelValues ...string) common.Gauge {

	_, file, line := utils.CallerGetInfo(pc.meter.callerOffset + 3)
	newValues := append(labelValues, fmt.Sprintf("%s:%d", file, line))

	pc.gaugeVec.WithLabelValues(newValues...).Set(value)
	return pc
}

func (p *PrometheusMeter) Gauge(name, description string, labels []string, prefixes ...string) common.Gauge {

	var names []string

	if !utils.IsEmpty(p.options.Prefix) {
		names = append(names, p.options.Prefix)
	}

	names = append(names, prefixes...)
	names = append(names, name)
	newName := strings.Join(names, "_")

	config := prometheus.GaugeOpts{
		Name: newName,
		Help: description,
	}

	labels = append(labels, "file")

	gaugeVec := prometheus.NewGaugeVec(config, labels)
	prometheus.Register(gaugeVec)

	return &PrometheusGauge{
		meter:    p,
		gaugeVec: gaugeVec,
	}
}

func (p *PrometheusMeter) SetCallerOffset(offset int) {
	p.callerOffset = offset
}

func (p *PrometheusMeter) Start() bool {

	p.logger.Info("Start prometheus endpoint...")

	http.Handle(p.options.URL, promhttp.Handler())

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
		options:      options,
		logger:       logger,
		callerOffset: 1,
	}
}
