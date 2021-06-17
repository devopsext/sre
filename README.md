# SRE framework 

Framework for golang applications which helps to send metrics, logs and traces into different monitoring tools or vendors. 

[![GoDoc](https://godoc.org/github.com/devopsext/sre?status.svg)](https://godoc.org/github.com/devopsext/sre)
[![build status](https://img.shields.io/travis/devopsext/sre/master.svg?style=flat-square)](https://travis-ci.org/devopsext/sre)

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

## Build

```sh
git clone https://github.com/devopsext/sre.git
cd sre/
go build

## Example