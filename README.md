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
  // Set caller offset for file:line proper usage 
  stdout.SetCallerOffset(2)

  // initialize DataDog logger
  datadog := provider.NewDataDogLogger(provider.DataDogLoggerOptions{
    DataDogOptions: provider.DataDogOptions{
      ServiceName: "some-service",
      Environment: "stage",
    },
    Host:  "localhost", // set DataDog agent UDP log host
    Port:  10518, // set DataDog agent UDP log port
    Level: "info",
  }, logs, stdout)

  // Add loggers
  logs.Register(stdout)
  logs.Register(datadog)

  test()
}
```

Collect go modules
```sh
go mod init logs
go mod tidy
```

Run logging example
```sh
go run logs.go
```

## Example