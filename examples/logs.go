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
		AgentHost: "",    // set DataDog agent UDP logs host
		AgentPort: 10518, // set DataDog agent UDP logs port
		Level:     "info",
	}, logs, stdout)

	// initialize NewRelic logger
	newrelic := provider.NewNewRelicLogger(provider.NewRelicLoggerOptions{
		NewRelicOptions: provider.NewRelicOptions{
			ServiceName: "sre-newrelic",
			Environment: "stage",
		},
		AgentHost: "localhost", // set NewRelic agent TCP logs host
		AgentPort: 5171,        // set NewRelic agent TCP logs port
		Level:     "info",
	}, logs, stdout)

	// add loggers
	logs.Register(stdout)
	if datadog != nil {
		logs.Register(datadog)
	}
	logs.Register(newrelic)

	test()
}
