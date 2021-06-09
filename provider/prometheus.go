package provider

import (
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
	counterVec *prometheus.CounterVec
}

type Prometheus struct {
	options      PrometheusOptions
	logger       common.Logger
	callerOffset int
}

func (pc *PrometheusCounter) Inc(labelValues ...string) common.Counter {

	pc.counterVec.WithLabelValues(labelValues...).Inc()
	return pc
}

func (p *Prometheus) SetCallerOffset(offset int) {
	p.callerOffset = offset
}

func (p *Prometheus) Counter(name, description string, labels []string, prefixes ...string) common.Counter {

	var names []string

	if !utils.IsEmpty(p.options.Prefix) {
		names = append(names, p.options.Prefix)
	}

	for _, v := range prefixes {
		names = append(names, v)
	}

	names = append(names, name)
	newName := strings.Join(names, "_")

	config := prometheus.CounterOpts{
		Name: newName,
		Help: description,
	}

	counterVec := prometheus.NewCounterVec(config, labels)
	prometheus.Register(counterVec)

	return &PrometheusCounter{
		counterVec: counterVec,
	}
}

func (p *Prometheus) Start(wg *sync.WaitGroup) {

	wg.Add(1)

	go func(wg *sync.WaitGroup) {

		defer wg.Done()

		p.logger.Info("Start prometheus endpoint...")

		http.Handle(p.options.URL, promhttp.Handler())

		listener, err := net.Listen("tcp", p.options.Listen)
		if err != nil {
			p.logger.Error(err)
			return
		}

		p.logger.Info("Prometheus is up. Listening...")
		err = http.Serve(listener, nil)
		if err != nil {
			p.logger.Error(err)
			return
		}

	}(wg)
}

func NewPrometheus(options PrometheusOptions, logger common.Logger, stdout *Stdout) *Prometheus {

	return &Prometheus{
		options:      options,
		logger:       logger,
		callerOffset: 0,
	}
}
