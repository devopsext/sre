package main

import (
	_ "embed"
	"time"

	"github.com/devopsext/sre/common"
	"github.com/devopsext/sre/provider"
)

//go:embed slack.json
var payload []byte

var logs = common.NewLogs()
var events = common.NewEvents()

func test() {

	events.Now(":exclamation: First", nil)

	m := make(map[string]string)
	m["payload"] = string(payload)

	events.Now(":grey_exclamation:Second", m)
	events.Now(":grey_question: *Third*", nil)
}

func main() {

	defer logs.Stop()   // finalize logs delivery
	defer events.Stop() // finalize events delivery

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

	slack := provider.NewSlackEventer(provider.SlackOptions{
		WebHook: "",
		Tags:    "",
		Timeout: 2,
	}, logs, stdout)

	// add events
	events.Register(slack)

	test()
}
