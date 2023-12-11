package provider

import (
	"errors"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/devopsext/sre/common"
)

func newrelicNewMeter(endpoint string) (*NewRelicMeter, *Stdout) {

	stdout := NewStdout(StdoutOptions{
		Format:          "template",
		Level:           "debug",
		Template:        "{{.msg}}",
		TimestampFormat: time.RFC3339Nano,
	})
	if stdout == nil {
		return nil, nil
	}
	stdout.SetCallerOffset(1)

	newrelic := NewNewRelicMeter(NewRelicMeterOptions{
		Endpoint: endpoint,
		Prefix:   "test",
		NewRelicOptions: NewRelicOptions{
			ApiKey:      "sdfsFFDfd",
			ServiceName: "sre-newrelic-meter-test",
			Attributes:  "tag1=value1,,tag3=${key3:value3}",
			Debug:       true,
		},
	}, nil, stdout)

	return newrelic, stdout
}

func newrelicNewLogger(agentHost, level string) (*NewRelicLogger, *Stdout, net.Listener) {

	agentPort := 51710
	listener, err := net.Listen("tcp", agentHost+":"+strconv.Itoa(agentPort))
	if err != nil {
		return nil, nil, nil
	}

	stdout := NewStdout(StdoutOptions{
		Format:          "template",
		Level:           "debug",
		Template:        "{{.msg}}",
		TimestampFormat: time.RFC3339Nano,
	})
	if stdout == nil {
		return nil, nil, listener
	}
	stdout.SetCallerOffset(1)

	newrelic := NewNewRelicLogger(NewRelicLoggerOptions{
		AgentHost: agentHost,
		AgentPort: agentPort,
		Level:     level,
		NewRelicOptions: NewRelicOptions{
			ServiceName: "sre-newrelic-tracer-test",
			Attributes:  "tag1=value1,,tag3=${key3:value3}",
			Debug:       true,
		},
	}, nil, stdout)

	return newrelic, stdout, listener
}

func TestNewRelicMeter(t *testing.T) {

	newrelic, _ := newrelicNewMeter("localhost")
	if newrelic == nil {
		t.Fatal("Invalid newrelic")
	}
	newrelic.SetCallerOffset(1)

	secondPrefix := "counter"
	metricName := "some"

	labels := make(common.Labels)
	labels["one"] = "value1"
	labels["two"] = "value2"
	labels["three"] = "value2"

	counter := newrelic.Counter(metricName, "description", labels, secondPrefix)
	if counter == nil {
		t.Fatal("Invalid newrelic counter")
	}

	maxCounter := 5
	for i := 0; i < maxCounter; i++ {
		counter.Inc()
	}

	newrelic.Stop()
}

func TestNewRelicMeterWrongAgentHost(t *testing.T) {

	newrelic, _ := newrelicNewMeter("")
	if newrelic != nil {
		t.Fatal("Valid newrelic")
	}
}

func testNewRelicLogger(t *testing.T, level string) (*NewRelicLogger, common.TracerSpan, net.Listener) {

	NewRelic, _, listener := newrelicNewLogger("localhost", level)
	if NewRelic == nil {
		t.Fatal("Invalid NewRelic")
	}
	NewRelic.Stack(-1).Stack(1)

	/*tracer, _ := NewRelicNewTracer("localhost")
	if tracer == nil {
		t.Fatal("Invalid tracer")
	}

	span := tracer.StartSpan()
	if span == nil {
		t.Fatal("Invalid span")
	}

	fields := NewRelic.addSpanFields(nil, nil)
	if fields != nil {
		t.Fatal("Invalid fields")
	}

	fields = logrus.Fields{}
	spanEmpty := &NewRelicTracerSpan{}
	fields = NewRelic.addSpanFields(spanEmpty, fields)
	if fields == nil {
		t.Fatal("Invalid fields")
	}*/

	return NewRelic, nil, listener
}

func TestNewRelicLoggerInfo(t *testing.T) {

	NewRelic, _, listener := testNewRelicLogger(t, "info")
	defer listener.Close()
	NewRelic.Info(nil)
	NewRelic.Info("info")
	//NewRelic.SpanInfo(span, "info")
}

func TestNewRelicLoggerWarn(t *testing.T) {

	NewRelic, _, listener := testNewRelicLogger(t, "warn")
	defer listener.Close()
	NewRelic.Warn(nil)
	NewRelic.Warn("warn")
	//NewRelic.SpanWarn(span, "warn")
}

func TestNewRelicLoggerDebug(t *testing.T) {

	NewRelic, _, listener := testNewRelicLogger(t, "debug")
	defer listener.Close()
	NewRelic.Debug(nil)
	NewRelic.Debug("debug")
	//NewRelic.SpanDebug(span, "debug")
}

func TestNewRelicLoggerError(t *testing.T) {

	NewRelic, _, listener := testNewRelicLogger(t, "error")
	defer listener.Close()
	NewRelic.Error(nil)
	NewRelic.Error("error")
	NewRelic.Error(errors.New("some error"))
	NewRelic.Error("error => %s", "message")
	//NewRelic.SpanError(span, "error")
}

func TestNewRelicLoggerPanic(t *testing.T) {

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("It should be paniced")
		}
	}()

	NewRelic, _, listener := testNewRelicLogger(t, "panic")
	defer listener.Close()
	NewRelic.Panic("panic")
}

/*func TestNewRelicLoggerSpanPanic(t *testing.T) {

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("It should be paniced")
		}
	}()

	NewRelic, span := testNewRelicLogger(t, "panic")
	NewRelic.SpanPanic(span, "panic")
}*/

func TestNewRelicLoggerEmptyAgentHost(t *testing.T) {

	NewRelic, _, listener := newrelicNewLogger("", "")
	defer listener.Close()
	if NewRelic != nil {
		t.Fatal("Valid NewRelic")
	}
}

func TestNewRelicLoggerWrongAgentHost(t *testing.T) {

	NewRelic, _, _ := newrelicNewLogger("ewqdWDEW1111ss", "how")
	if NewRelic != nil {
		t.Fatal("Valid NewRelic")
	}
}
