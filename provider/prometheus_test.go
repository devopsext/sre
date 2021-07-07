package provider

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/devopsext/sre/common"
)

func TestPrometheus(t *testing.T) {

	stdout := NewStdout(StdoutOptions{
		Format:          "template",
		Level:           "debug",
		Template:        "{{.msg}}",
		TimestampFormat: time.RFC3339Nano,
	})
	if stdout == nil {
		t.Error("Invalid stdout")
	}
	stdout.SetCallerOffset(1)

	URL := "/metrics"
	port := 9999

	// for prometheus it should be:
	// firstPrefix_secondPrefix_metricName => test_counter_some

	firstPrefix := "test"
	secondPrefix := "counter"
	metricName := "some"

	prometheus := NewPrometheusMeter(PrometheusOptions{
		URL:    URL,
		Listen: fmt.Sprintf(":%d", port),
		Prefix: firstPrefix,
	}, nil, stdout)
	if prometheus == nil {
		t.Error("Invalid prometheus")
	}
	prometheus.SetCallerOffset(1)

	var wg sync.WaitGroup
	prometheus.StartInWaitGroup(&wg)
	defer prometheus.Stop()

	counter := prometheus.Counter(metricName, "description", []string{"one", "two", "three"}, secondPrefix)
	if counter == nil {
		t.Error("Invalid prometheus")
	}

	maxCounter := 5
	for i := 0; i < maxCounter; i++ {
		counter.Inc("value1", "value2", "value3")
	}

	time.Sleep(time.Duration(1) * time.Second)

	r, err := http.Get(fmt.Sprintf("http://localhost:%d%s", port, URL))
	if err != nil {
		t.Error(err)
	}

	if r.StatusCode != 200 {
		t.Errorf("None 200 response: %d", r.StatusCode)
	}

	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Error(err)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 {
		t.Error("No lines in output")
	}

	m := make(map[string]string)
	for _, line := range lines {
		parts := strings.Split(line, " ")
		if len(parts) > 1 {

			value := parts[1]
			names := strings.Split(parts[0], "{")
			if len(names) > 0 {
				m[names[0]] = value
			}
		}
	}

	value := m[fmt.Sprintf("%s_%s_%s", firstPrefix, secondPrefix, metricName)]
	if value == "" {
		t.Error("No metric or value in output")
	}

	if value != strconv.Itoa(maxCounter) {
		t.Errorf("Invalid metric value %s, expected %d", value, maxCounter)
	}
}

func TestPrometheusWrongListen(t *testing.T) {

	stdout := NewStdout(StdoutOptions{
		Format:          "template",
		Level:           "debug",
		Template:        "{{.msg}}",
		TimestampFormat: time.RFC3339Nano,
	})
	if stdout == nil {
		t.Error("Invalid stdout")
	}
	stdout.SetCallerOffset(1)

	URL := "/wrong"
	port := 10000
	host := common.GetGuid()

	prometheus := NewPrometheusMeter(PrometheusOptions{
		URL:    URL,
		Listen: fmt.Sprintf("%s:%d", host, port),
		Prefix: "test",
	}, nil, stdout)
	if prometheus == nil {
		t.Error("Invalid prometheus")
	}
	prometheus.SetCallerOffset(1)

	if prometheus.Start() {
		t.Error("Invalid startup option")
	}
}