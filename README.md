# SRE framework 

Framework for golang applications which helps to send metrics, logs and traces into different monitoring tools or vendors. 

[![godoc](https://godoc.org/github.com/devopsext/sre?status.svg)](https://godoc.org/github.com/devopsext/sre)
[![go report](	https://goreportcard.com/badge/github.com/devopsext/sre)](https://goreportcard.com/report/github.com/devopsext/sre)
[![codecov](https://codecov.io/gh/devopsext/sre/branch/main/graph/badge.svg?token=M78C7PVMDV)](https://codecov.io/gh/devopsext/sre)
[![build status](https://travis-ci.com/devopsext/sre.svg?branch=main)](https://travis-ci.com/devopsext/sre)

## Features

- Provide plain text, json logs with trace ID (if log entry is based on a span) and source line info
- Provide additional labels and tags for metrics, like: source line, service name and it's version
- Support logging tools (aka logs):
  - Stdout (text, json, template) based on [Logrus](github.com/sirupsen/logrus)
  - DataDog based on [Logrus](github.com/sirupsen/logrus) over UDP
- Support monitoring tools (aka metrics)
  - [Prometheus](github.com/prometheus/client_golang)
  - [DataDog](https://github.com/DataDog/datadog-go)
  - [Opentelemetry](https://github.com/open-telemetry/opentelemetry-go)
- Support tracing tools (aka traces)
  - [Jaeger](https://github.com/jaegertracing/jaeger-client-go)
  - [DataDog](https://github.com/DataDog/dd-trace-go)
  - [Opentelemetry](https://github.com/open-telemetry/opentelemetry-go)


## Usage

### Requirements

- Jaeger works with its [Jaeger agent](https://www.jaegertracing.io/docs/latest/getting-started/)
- DataDog uses [DataDog agent](https://docs.datadoghq.com/agent/) for logs, metrics and traces
- Opentelemetry communicates with its [Opentelemetry agent](https://github.com/open-telemetry/opentelemetry-collector)

### Envs

Set proper GOROOT and PATH variables
```sh
export GOROOT="$HOME/go/root/1.16.4"
export PATH="$PATH:$GOROOT/bin"
```

### Go modules

Set go.mod manually
```plain
module sre

go 1.16

require github.com/devopsext/sre v0.0.6
```

Collect go modules
```sh
go get
```
```log
go: finding module for package github.com/devopsext/sre/provider
go: finding module for package github.com/devopsext/sre/common
go: found github.com/devopsext/sre/common in github.com/devopsext/sre v0.0.6
go: found github.com/devopsext/sre/provider in github.com/devopsext/sre v0.0.6
```


### Logs usage

Create logs.go file to test logging functionality
```golang
package main

import (
  "time"

  "github.com/devopsext/sre/common"
  "github.com/devopsext/sre/provider"
)

var logs = common.NewLogs()

func test() {
  logs.Info("Info message to every log provider...")
  logs.Debug("Debug message to every log provider...")
  logs.Warn("Warn message to every log provider...")
}

func main() {

  // initialize Stdout logger
  stdout := provider.NewStdout(provider.StdoutOptions{
    Format:          "template",
    Level:           "info",
    Template:        "{{.file}} {{.msg}}",
    TimestampFormat: time.RFC3339Nano,
    TextColors:      true,
  })
  // set caller offset for file:line proper usage 
  stdout.SetCallerOffset(2)

  // initialize DataDog logger
  datadog := provider.NewDataDogLogger(provider.DataDogLoggerOptions{
    DataDogOptions: provider.DataDogOptions{
      ServiceName: "sre-datadog",
      Environment: "stage",
    },
    AgentHost:  "localhost", // set DataDog agent UDP logs host
    AgentPort:  10518, // set DataDog agent UDP logs port
    Level: "info",
  }, logs, stdout)

  // add loggers
  logs.Register(stdout)
  logs.Register(datadog)

  test()
}
```

Run logs example
```sh
go run logs.go
```
```log
go/sre/logs.go:13 Info message to every log provider...
go/sre/logs.go:14 Debug message to every log provider...
go/sre/logs.go:15 Warn message to every log provider...
...
```

### Metrics usage

Create metrics.go file to test metrics functionality
```golang
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
    AgentHost:  "localhost", // set DataDog agent UDP metrics host
    AgentPort:  10518, // set DataDog agent UDP metrics port
  }, logs, stdout)

  // initialize Opentelemetry meter
  opentelemetry := provider.NewOpentelemetryMeter(provider.OpentelemetryMeterOptions{
    OpentelemetryOptions: provider.OpentelemetryOptions{
      ServiceName: "sre-opentelemetry",
      Environment: "stage",
    },
    AgentHost:   "localhost", // set Opentelemetry agent metrics host
    AgentPort:   4317,        // set Opentelemetry agent metrics port
    Prefix: "sre",
  }, logs, stdout)

  // add meters
  metrics.Register(prometheus)
  metrics.Register(datadog)
  metrics.Register(opentelemetry)

  test()

  mainWG.Wait()
  metrics.Stop() // finalize metrics delivery
}
```

Run metrcis example
```sh
go run metrics.go
```
```log
sre@v0.0.6/provider/prometheus.go:83 Start prometheus endpoint...
sre@v0.0.6/provider/prometheus.go:93 Prometheus is up. Listening...
sre@v0.0.6/provider/datadog.go:648 DataDog meter is up...
```

Check Prometheus metrics
```sh
curl -sk http://127.0.0.1:8080/metrics | grep sre_
```
```prometheus
# HELP sre_calls Calls counter
# TYPE sre_calls counter
sre_calls{time="2021-06-17 18:00:30.248990729 +0300 EEST m=+0.002878298"} 1
```

### Traces usage

Create traces.go file to test traces functionality
```golang
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
    AgentPort:           6831, // set Jaeger agent port
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
    AgentHost: "", // set DataDog agent traces host
    AgentPort: 8126,        // set DataDog agent traces port
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

  traces.Stop() // finalize traces delivery
}
```

Run traces example
```sh
go run traces.go
```
```log
...
sre@v0.0.6/provider/jaeger.go:444 Jaeger tracer is up...
go/sre/traces.go:66 DataDog tracer is disabled.
sre@v0.0.6/provider/opentelemetry.go:493 Opentelemetry tracer is up...
go/sre/traces.go:24 Counter increment 0
go/sre/traces.go:24 Counter increment 1
go/sre/traces.go:24 Counter increment 2
...
```

Go to Jaeger UI and there should be seen

![Jaeger](/jaeger.png)

For a case of Opentelemetry (Lightstep) screenshoot will be

![Lightstep](/lightstep.png)

## Framework in other projects

- [devopsext/events](https://github.com/devopsext/events) Kubernetes & Alertmanager events to Telegram, Slack, Workchat and other messengers.
