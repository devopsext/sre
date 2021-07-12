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

### Set envs

Set proper GOROOT and PATH variables
```sh
export GOROOT="$HOME/go/root/1.16.4"
export PATH="$PATH:$GOROOT/bin"
```

### Get Go modules

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

### Examples

Run one of example below: 
- [Logs](examples/logs.md)
- [Metrics](examples/metrics.md)
- [Traces](examples/traces.md)


## Framework in other projects

- [devopsext/events](https://github.com/devopsext/events) Kubernetes & Alertmanager events to Telegram, Slack, Workchat and other messengers.
