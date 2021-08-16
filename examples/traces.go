package main

import (
	"fmt"
	"time"

	"github.com/devopsext/sre/common"
	"github.com/devopsext/sre/provider"
)

var logs = common.NewLogs()
var traces = common.NewTraces()

func test() {

	rootSpan := traces.StartSpan()
	spanCtx := rootSpan.GetContext()
	for i := 0; i < 10; i++ {

		span := traces.StartChildSpan(spanCtx)
		span.SetName(fmt.Sprintf("name-%d", i))

		time.Sleep(time.Duration(100*i) * time.Millisecond)
		logs.SpanDebug(span, "Counter increment %d", i)

		spanCtx = span.GetContext()
		span.Finish()
	}

	// emulate delay of 100 msecs
	time.Sleep(time.Duration(200) * time.Millisecond)

	rootSpan.Finish()

	// wait for a while to delivery all spans to provider
	time.Sleep(time.Duration(3000) * time.Millisecond)
}

func main() {

	defer logs.Stop()   // finalize logs delivery
	defer traces.Stop() // finalize traces delivery

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

	// initialize Jaeger tracer
	jaeger := provider.NewJaegerTracer(provider.JaegerOptions{
		ServiceName:         "sre-jaeger",
		AgentHost:           "localhost", // set Jaeger agent host
		AgentPort:           6831,        // set Jaeger agent port
		BufferFlushInterval: 0,
		QueueSize:           0,
		Tags:                "key1=value1",
	}, logs, stdout)

	// initialize DataDog tracer
	datadog := provider.NewDataDogTracer(provider.DataDogTracerOptions{
		DataDogOptions: provider.DataDogOptions{
			ServiceName: "sre-datadog",
			Environment: "stage",
		},
		AgentHost: "",   // set DataDog agent traces host
		AgentPort: 8126, // set DataDog agent traces port
	}, logs, stdout)

	// initialize Opentelemetry tracer
	opentelemetry := provider.NewOpentelemetryTracer(provider.OpentelemetryTracerOptions{
		OpentelemetryOptions: provider.OpentelemetryOptions{
			ServiceName: "sre-opentelemetry",
			Environment: "stage",
		},
		AgentHost: "localhost", // set Opentelemetry agent traces host
		AgentPort: 4317,        // set Opentelemetry agnet traces port
	}, logs, stdout)

	// add traces
	traces.Register(jaeger)
	if datadog != nil {
		traces.Register(datadog)
	}
	traces.Register(opentelemetry)

	test()
}
