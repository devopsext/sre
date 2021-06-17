# SRE framework 

Framework for golang applications which helps to send metrics, logs and traces into different monitoring tools or vendors. 

[![GoDoc](https://godoc.org/github.com/devopsext/sre?status.svg)](https://godoc.org/github.com/devopsext/sre)
[![build status](https://img.shields.io/travis/devopsext/sre/master.svg?style=flat-square)](https://travis-ci.com/devopsext/sre)

## Features

- Logging tools (aka logs):
  - Stdout (plain, json, patterns)
  - DataDog
- Monitoring tools (aka metrics)
  - Prometheus
  - DataDog
- Tracing tools (aka traces)
  - Jaeger
  - DataDog

## Usage

Set proper GOROOT and PATH variables
```sh
export GOROOT="$HOME/go/root/1.16.4"
export PATH="$PATH:$GOROOT/bin"
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
    Format:          "text",
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
      ServiceName: "some-service",
      Environment: "stage",
    },
    Host:  "localhost", // set DataDog agent UDP logs host
    Port:  10518, // set DataDog agent UDP logs port
    Level: "info",
  }, logs, stdout)

  // add loggers
  logs.Register(stdout)
  logs.Register(datadog)

  test()
}
```

Collect go modules
```sh
go mod init sre
go mod tidy
```
```log
go: finding module for package github.com/devopsext/sre/provider
go: finding module for package github.com/devopsext/sre/common
go: found github.com/devopsext/sre/common in github.com/devopsext/sre v0.0.2
go: found github.com/devopsext/sre/provider in github.com/devopsext/sre v0.0.2
```

Run logging example
```sh
go run logs.go
```
```log
INFO[2021-06-17T17:32:30.585651118+03:00] Info message to every log provider...         file="go/sre/logs.go:13" func=main.test
WARN[2021-06-17T17:32:30.585798024+03:00] Warn message to every log provider...         file="go/sre/logs.go:15" func=main.test
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
    Format:          "text",
    Level:           "info",
    Template:        "{{.file}} {{.msg}}",
    TimestampFormat: time.RFC3339Nano,
    TextColors:      true,
  })
  // set caller offset for file:line proper usage
  stdout.SetCallerOffset(2)

  // add Stdout logger
  logs.Register(stdout)

  // initialize Prometheus metricer
  prometheus := provider.NewPrometheus(provider.PrometheusOptions{
    URL:    "/metrics",
    Listen: "127.0.0.1:8080",
    Prefix: "sre",
  }, logs, stdout)
  // set caller offset for file:line proper usage
  prometheus.SetCallerOffset(1)
  prometheus.Start(&mainWG)

  // initialize DataDog metricer
  datadog := provider.NewDataDogMetricer(provider.DataDogMetricerOptions{
    DataDogOptions: provider.DataDogOptions{
      ServiceName: "some-service",
      Environment: "stage",
    },
    Host:  "localhost", // set DataDog agent UDP metrics host
    Port:  10518, // set DataDog agent UDP metrics port
  }, logs, stdout)
  datadog.SetCallerOffset(1)

  // add metricers
  metrics.Register(prometheus)
  metrics.Register(datadog)

  test()

  mainWG.Wait()
}
```

Collect go modules
```sh
go mod init sre
go mod tidy
```
```log
go: finding module for package github.com/devopsext/sre/provider
go: finding module for package github.com/devopsext/sre/common
go: found github.com/devopsext/sre/common in github.com/devopsext/sre v0.0.2
go: found github.com/devopsext/sre/provider in github.com/devopsext/sre v0.0.2
```

Run logging example
```sh
go run metrics.go
```
```log
INFO[2021-06-17T18:00:30.247316085+03:00] Start prometheus endpoint...                  file="sre@v0.0.1/provider/prometheus.go:78" func="provider.(*Prometheus).Start.func1"
INFO[2021-06-17T18:00:30.247526413+03:00] Prometheus is up. Listening...                file="sre@v0.0.1/provider/prometheus.go:88" func="provider.(*Prometheus).Start.func1"
INFO[2021-06-17T18:00:30.248965919+03:00] Datadog metrics are up...                     file="sre@v0.0.1/provider/datadog.go:614" func=provider.NewDataDogMetricer
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
  "time"

  "github.com/devopsext/sre/common"
  "github.com/devopsext/sre/provider"
)

var logs = common.NewLogs()
var traces = common.NewTraces()

func test() {
  rootSpan := traces.StartSpan()

  // emulate delay of 100 msecs
  time.Sleep(time.Duration(100) * time.Millisecond)

  span := traces.StartChildSpan(rootSpan.GetContext())

  // emulate delay of 100 msecs
  time.Sleep(time.Duration(100) * time.Millisecond)
  logs.SpanInfo(span, "Something happened")
  span.Finish()

  rootSpan.Finish()
}

func main() {

  // initialize Stdout logger
  stdout := provider.NewStdout(provider.StdoutOptions{
    Format:          "text",
    Level:           "info",
    Template:        "{{.file}} {{.msg}}",
    TimestampFormat: time.RFC3339Nano,
    TextColors:      true,
  })
  // set caller offset for file:line proper usage
  stdout.SetCallerOffset(2)

  // add Stdout logger
  logs.Register(stdout)

  // initialize Jaeger tracer
  jaeger := provider.NewJaeger(provider.JaegerOptions{
    ServiceName: "sre",
    AgentHost: "localhost", // set Jaeger agent host
    AgentPort: 6831, // set Jaeger agent port
  }, logs, stdout)
  // set caller offset for file:line proper usage
  jaeger.SetCallerOffset(1)

  // initialize DataDog metricer
  datadog := provider.NewDataDogTracer(provider.DataDogTracerOptions{
    DataDogOptions: provider.DataDogOptions{
      ServiceName: "some-service",
      Environment: "stage",
    },
    Host:  "localhost", // set DataDog agent traces host
    Port:  8126, // set DataDog agent traces port
  }, logs, stdout)
  datadog.SetCallerOffset(1)

  // add traces
  traces.Register(jaeger)
  traces.Register(datadog)

  test()
}
```

Collect go modules
```sh
go mod init sre
go mod tidy
```
```log
go: finding module for package github.com/devopsext/sre/provider
go: finding module for package github.com/devopsext/sre/common
go: found github.com/devopsext/sre/common in github.com/devopsext/sre v0.0.2
go: found github.com/devopsext/sre/provider in github.com/devopsext/sre v0.0.2
```

Run logging example
```sh
go run traces.go
```
```log
...
INFO[2021-06-17T18:28:45.178707109+03:00] Something happened                            file="go/sre/traces.go:23" func=main.test trace_id=1198730091955389962
INFO[2021-06-17T18:28:45.178840198+03:00] Reporting span 10a2beaae092860a:486b0277d5e7ae83:10a2beaae092860a:1  file="uber/jaeger-client-go@v2.29.1+incompatible/reporter.go:151" func="jaeger-client-go.(*compositeReporter).Report"
INFO[2021-06-17T18:28:45.178940724+03:00] Reporting span 10a2beaae092860a:10a2beaae092860a:0000000000000000:1  file="uber/jaeger-client-go@v2.29.1+incompatible/reporter.go:151" func="jaeger-client-go.(*compositeReporter).Report"
```
