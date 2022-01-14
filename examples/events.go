package main

import (
	"time"

	"github.com/devopsext/sre/common"
	"github.com/devopsext/sre/provider"
)

var logs = common.NewLogs()
var events = common.NewEvents()

func test() {

	events.Now("First", nil)

	m := make(map[string]string)
	m["alert_type"] = "warning"
	m["device_name"] = "test machine"
	m["host"] = "xciber's macbook"
	m["priority"] = "low"
	m["related_event_id"] = "123"
	m["source_type_name"] = "sre framework"

	events.At("Second", m, time.Now().Add(time.Second*5))
	m["priority"] = "normal"
	events.Interval("Third", m, time.Now().Add(-time.Second*2), time.Now().Add(-time.Second*1))
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

	// initialize Grafana eventer
	grafana := provider.NewGrafanaEventer(provider.GrafanaEventerOptions{
		GrafanaOptions: provider.GrafanaOptions{
			URL:    "localhost",
			ApiKey: "admin:admin", // set API key
		},
		Endpoint: "/api/annotations",
	}, logs, stdout)

	// initialize Newrelic eventer
	newrelic := provider.NewNewRelicEventer(provider.NewRelicEventerOptions{
		NewRelicOptions: provider.NewRelicOptions{
			ApiKey: "", // set API key
		},
		Endpoint: "https://insights-collector.eu01.nr-data.net/v1/accounts/$ACCOUNT_ID/events",
	}, logs, stdout)

	// initialize DataDog Eventer
	datadog := provider.NewDataDogEventer(provider.DataDogEventerOptions{
		DataDogOptions: provider.DataDogOptions{
			ApiKey:      "", // set API key
			ServiceName: "sre-datadog",
			Environment: "stage",
			Tags:        "tag1=asd,tag2=kjh",
			Debug:       true,
		},
		Site: "datadoghq.eu", //
	},
		logs, stdout)

	// add events
	events.Register(grafana)
	events.Register(newrelic)
	events.Register(datadog)

	test()
}
