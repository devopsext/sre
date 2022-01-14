# SRE framework 

Framework for golang applications which helps to send metrics, logs, traces and events into different monitoring tools or vendors. 

[![godoc](https://godoc.org/github.com/devopsext/sre?status.svg)](https://godoc.org/github.com/devopsext/sre)
[![go report](	https://goreportcard.com/badge/github.com/devopsext/sre)](https://goreportcard.com/report/github.com/devopsext/sre)
[![codecov](https://codecov.io/gh/devopsext/sre/branch/main/graph/badge.svg?token=M78C7PVMDV)](https://codecov.io/gh/devopsext/sre)
[![build status](https://travis-ci.com/devopsext/sre.svg?branch=main)](https://app.travis-ci.com/github/devopsext/sre)

## Features

- Provide plain text, json logs with trace ID (if log entry is based on a span) and source line info
- Provide additional labels and tags for metrics, like: source line, service name and it's version
- Support logging tools (aka logs):
  - Stdout (text, json, template) based on [Logrus](github.com/sirupsen/logrus)
  - DataDog based on [Logrus](github.com/sirupsen/logrus) over UDP
  - NewRelic based on [Logrus](github.com/sirupsen/logrus) over TCP, as well as via [LogAPI](https://docs.newrelic.com/docs/logs/log-management/log-api/) by using [Telemetry](https://github.com/newrelic/newrelic-telemetry-sdk-go) 
- Support monitoring tools (aka metrics)
  - [Prometheus](github.com/prometheus/client_golang)
  - [DataDog](https://github.com/DataDog/datadog-go)
  - [NewRelic](https://github.com/newrelic/newrelic-telemetry-sdk-go)
  - [Opentelemetry](https://github.com/open-telemetry/opentelemetry-go)
- Support tracing tools (aka traces)
  - [Jaeger](https://github.com/jaegertracing/jaeger-client-go)
  - [DataDog](https://github.com/DataDog/dd-trace-go)
  - [Opentelemetry](https://github.com/open-telemetry/opentelemetry-go)
- Support eventing tools (aka events)
  - [NewRelic](https://github.com/newrelic/newrelic-telemetry-sdk-go)
  - [Grafana](https://github.com/grafana/grafana)


## Usage

### Requirements

- Jaeger works with its [Jaeger agent](https://www.jaegertracing.io/docs/latest/getting-started/)
- DataDog uses [DataDog agent](https://docs.datadoghq.com/agent/) for logs, metrics and traces
- NewRelic uses [NewRelic standalone infrastructure agent](https://docs.newrelic.com/docs/infrastructure/install-infrastructure-agent/) for logs
- NewRelic uses [NewRelic Telemetry SDK](https://docs.newrelic.com/docs/telemetry-data-platform/ingest-apis/telemetry-sdks-report-custom-telemetry-data/) for logs, metrics, traces, events
- Opentelemetry communicates with its [Opentelemetry agent](https://github.com/open-telemetry/opentelemetry-collector)
- Grafana uses [Grafana Annotations API](https://grafana.com/docs/grafana/latest/http_api/annotations/) 

### Set envs

Set proper GOROOT and PATH variables
```sh
export GOROOT="$HOME/go/root/1.17.4"
export PATH="$PATH:$GOROOT/bin"
```

### Get Go modules

Set go.mod manually
```plain
module sre

go 1.17

require github.com/devopsext/sre vX.Y.Z
```

### Collect go modules
```sh
go get
```
```log
go: finding module for package github.com/devopsext/sre/provider
go: finding module for package github.com/devopsext/sre/common
go: found github.com/devopsext/sre/common in github.com/devopsext/sre vX.Y.Z
go: found github.com/devopsext/sre/provider in github.com/devopsext/sre vX.Y.Z
```
*vX.Y.Z - tag (version) of the framework, for instance => v0.0.7

### Run one of example
 
- [Logs](examples/logs.md)
- [Metrics](examples/metrics.md)
- [Traces](examples/traces.md)
- [Events](examples/events.md)

## Framework in other projects

- [devopsext/events](https://github.com/devopsext/events) Kubernetes & Alertmanager events to Telegram, Slack, Workchat and other messengers.
