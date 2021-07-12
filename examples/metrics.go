package main

import (
	"sync"
	"time"

	"github.com/devopsext/sre/common"
	"github.com/devopsext/sre/provider"
)

var logs = common.NewLogs()
var metrics = common.NewMetrics()
var mainWG sync.WaitGroup

func test() {
	counter := metrics.Counter("calls", "Calls counter", []string{"time"})
	counter.Inc(time.Now().String())
}

func main() {

	// initialize Stdout logger
	stdout := provider.NewStdout(provider.StdoutOptions{
		Format:          "template",
		Level:           "debug",
		Template:        "{{.file}} {{.msg}}",
		TimestampFormat: time.RFC3339Nano,
		TextColors:      true,
	})
	// set caller offset for file:line proper usage
	stdout.SetCallerOffset(2)

	// add Stdout logger
	logs.Register(stdout)

	// initialize Prometheus meter
	prometheus := provider.NewPrometheusMeter(provider.PrometheusOptions{
		URL:    "/metrics",
		Listen: "127.0.0.1:8080",
		Prefix: "sre",
	}, logs, stdout)
	prometheus.StartInWaitGroup(&mainWG)

	// initialize DataDog meter
	datadog := provider.NewDataDogMeter(provider.DataDogMeterOptions{
		DataDogOptions: provider.DataDogOptions{
			ServiceName: "sre-datadog",
			Environment: "stage",
		},
		AgentHost: "localhost", // set DataDog agent UDP metrics host
		AgentPort: 10518,       // set DataDog agent UDP metrics port
	}, logs, stdout)

	// initialize Opentelemetry meter
	opentelemetry := provider.NewOpentelemetryMeter(provider.OpentelemetryMeterOptions{
		OpentelemetryOptions: provider.OpentelemetryOptions{
			ServiceName: "sre-opentelemetry",
			Environment: "stage",
		},
		AgentHost: "localhost", // set Opentelemetry agent metrics host
		AgentPort: 4317,        // set Opentelemetry agent metrics port
		Prefix:    "sre",
	}, logs, stdout)

	// add meters
	metrics.Register(prometheus)
	metrics.Register(datadog)
	metrics.Register(opentelemetry)

	test()

	mainWG.Wait()
	metrics.Stop() // finalize metrics delivery
}
