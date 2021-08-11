package provider

import (
	"errors"
	"testing"
	"time"

	"github.com/devopsext/sre/common"
)

func newrelicNewLogger(agentHost, level string) (*NewRelicLogger, *Stdout) {

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

	newrelic := NewNewRelicLogger(NewRelicLoggerOptions{
		AgentHost: agentHost,
		AgentPort: 5171,
		Level:     level,
		NewRelicOptions: NewRelicOptions{
			ServiceName: "sre-newrelic-tracer-test",
			Labels:      "tag1=value1,,tag3=${key3:value3}",
			Debug:       true,
		},
	}, nil, stdout)

	return newrelic, stdout
}

func testNewRelicLogger(t *testing.T, level string) (*NewRelicLogger, common.TracerSpan) {

	NewRelic, _ := newrelicNewLogger("localhost", level)
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

	return NewRelic, nil
}

func TestNewRelicLoggerInfo(t *testing.T) {

	NewRelic, _ := testNewRelicLogger(t, "info")
	NewRelic.Info(nil)
	NewRelic.Info("info")
	//NewRelic.SpanInfo(span, "info")
}

func TestNewRelicLoggerWarn(t *testing.T) {

	NewRelic, _ := testNewRelicLogger(t, "warn")
	NewRelic.Warn(nil)
	NewRelic.Warn("warn")
	//NewRelic.SpanWarn(span, "warn")
}

func TestNewRelicLoggerDebug(t *testing.T) {

	NewRelic, _ := testNewRelicLogger(t, "debug")
	NewRelic.Debug(nil)
	NewRelic.Debug("debug")
	//NewRelic.SpanDebug(span, "debug")
}

func TestNewRelicLoggerError(t *testing.T) {

	NewRelic, _ := testNewRelicLogger(t, "error")
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

	NewRelic, _ := testNewRelicLogger(t, "panic")
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

	NewRelic, _ := newrelicNewLogger("", "")
	if NewRelic != nil {
		t.Fatal("Valid NewRelic")
	}
}

func TestNewRelicLoggerWrongAgentHost(t *testing.T) {

	NewRelic, _ := newrelicNewLogger("ewqdWDEW1111ss", "how")
	if NewRelic != nil {
		t.Fatal("Valid NewRelic")
	}
}
