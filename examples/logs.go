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
		AgentHost: "localhost", // set DataDog agent UDP logs host
		AgentPort: 10518,       // set DataDog agent UDP logs port
		Level:     "info",
	}, logs, stdout)

	// add loggers
	logs.Register(stdout)
	logs.Register(datadog)

	test()
}
