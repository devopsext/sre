package provider

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/devopsext/utils"
)

func datadogNewTracer(agentHost string) (*DataDogTracer, *Stdout) {

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

	datadog := NewDataDogTracer(DataDogTracerOptions{
		AgentHost: agentHost,
		AgentPort: 8126,
		DataDogOptions: DataDogOptions{
			ServiceName: "sre-datadog-tracer-test",
			Tags:        "tag1=value1,,tag3=${key3:value3}",
			Debug:       true,
		},
	}, nil, stdout)

	return datadog, stdout
}

func datadogNewMeter(agentHost string) (*DataDogMeter, *Stdout) {

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

	datadog := NewDataDogMeter(DataDogMeterOptions{
		AgentHost: agentHost,
		AgentPort: 8126,
		Prefix:    "test",
		DataDogOptions: DataDogOptions{
			ServiceName: "sre-datadog-meter-test",
			Tags:        "tag1=value1,,tag3=${key3:value3}",
			Debug:       true,
		},
	}, nil, stdout)

	return datadog, stdout
}

func datadogNewLogger(agentHost string) (*DataDogLogger, *Stdout) {

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

	datadog := NewDataDogLogger(DataDogLoggerOptions{
		AgentHost: agentHost,
		AgentPort: 8126,
		Level:     "debug",
		DataDogOptions: DataDogOptions{
			ServiceName: "sre-datadog-tracer-test",
			Tags:        "tag1=value1,,tag3=${key3:value3}",
			Debug:       true,
		},
	}, nil, stdout)

	return datadog, stdout
}

func TestDataDogTracer(t *testing.T) {

	datadog, _ := datadogNewTracer("localhost")
	if datadog == nil {
		t.Fatal("Invalid datadog")
	}
	datadog.SetCallerOffset(1)

	span := datadog.StartSpan()
	if span == nil {
		t.Fatal("Invalid span")
	}
	defer span.Finish()
	span.SetName("some-span")
	span.SetTag("key1", "Value1")
	span.Error(errors.New("some-span-error"))

	ctx := span.GetContext()
	if ctx == nil {
		t.Fatal("Invalid span context")
	}

	traceID := ctx.GetTraceID()
	if utils.IsEmpty(traceID) {
		t.Fatal("Invalid trace ID")
	}
	t.Logf("Trace ID is %s", traceID)

	spanID := ctx.GetSpanID()
	if utils.IsEmpty(spanID) {
		t.Fatal("Invalid span ID")
	}
	t.Logf("Span ID is %s", spanID)

	traceSpan := datadog.StartSpanWithTraceID(traceID, "")
	if traceSpan == nil {
		t.Fatal("Invalid trace span")
	}
	defer traceSpan.Finish()
	traceSpan.SetName("some-trace-span")
	traceSpan.SetBaggageItem("key", "value")
	traceSpan.SetTag("parent-span-ID", spanID)

	childSpan := datadog.StartChildSpan(ctx)
	if childSpan == nil {
		t.Fatal("Invalid child span")
	}
	defer childSpan.Finish()
	childSpan.SetName("some-child-span")

	followSpan := datadog.StartFollowSpan(ctx)
	if followSpan == nil {
		t.Fatal("Invalid child span")
	}
	defer followSpan.Finish()
	followSpan.SetName("some-follow-span")

	nilChildSpan := datadog.StartChildSpan(t)
	if nilChildSpan != nil {
		t.Fatal("Invalid nil child span")
	}

	nilFollowSpan := datadog.StartFollowSpan(t)
	if nilFollowSpan != nil {
		t.Fatal("Invalid nil follow span")
	}

	headers := make(http.Header)

	span.SetCarrier(t)
	nilHeaderSpan := datadog.StartFollowSpan(headers)
	if nilHeaderSpan != nil {
		t.Fatal("Valid nil header span")
	}

	span.SetCarrier(headers)
	headerSpan := datadog.StartChildSpan(headers)
	if headerSpan == nil {
		t.Fatal("Invalid nil header span")
	}

	nilSpan := datadog.StartSpanWithTraceID("", "")
	if nilSpan != nil {
		t.Fatal("Valid nil span")
	}

	datadog.Stop()
}

func TestDataDogTracerWrongAgentHost(t *testing.T) {

	datadog, _ := datadogNewTracer("")
	if datadog != nil {
		t.Fatal("Valid datadog")
	}
}

func TestDataDogTracerInternalLogger(t *testing.T) {

	_, stdout := datadogNewTracer("localhost")
	if stdout == nil {
		t.Fatal("Valid stdout")
	}

	internalLogger := DataDogInternalLogger{
		logger: stdout,
	}

	internalLogger.Log("Some message")
}

func TestDataDogWrongSpan(t *testing.T) {

	span := DataDogTracerSpan{}

	ctx := span.GetContext()
	if ctx != nil {
		t.Fatal("Valid span context")
	}

	s := span.SetCarrier(t)
	if s != nil {
		t.Fatal("Valid span")
	}

	s = span.SetName("some-name")
	if s != nil {
		t.Fatal("Valid span")
	}

	s = span.SetTag("some-tag", "some-value")
	if s != nil {
		t.Fatal("Valid span")
	}

	s = span.Error(errors.New("some-error"))
	if s != nil {
		t.Fatal("Valid span")
	}

	s = span.SetBaggageItem("key", "value")
	if s != nil {
		t.Fatal("Valid span")
	}

	span.span = nil
	span.Finish()
}

func TestDataDogWrongSpanContext(t *testing.T) {

	ctx := DataDogTracerSpanContext{}

	traceID := ctx.GetTraceID()
	if !utils.IsEmpty(traceID) {
		t.Fatal("Valid trace ID")
	}

	spanID := ctx.GetSpanID()
	if !utils.IsEmpty(spanID) {
		t.Fatal("Valid span ID")
	}

}

func TestDataDogMeter(t *testing.T) {

	datadog, _ := datadogNewMeter("localhost")
	if datadog == nil {
		t.Fatal("Invalid datadog")
	}
	datadog.SetCallerOffset(1)

	secondPrefix := "counter"
	metricName := "some"

	counter := datadog.Counter(metricName, "description", []string{"one", "two", "three"}, secondPrefix)
	if counter == nil {
		t.Fatal("Invalid datadog")
	}

	maxCounter := 5
	for i := 0; i < maxCounter; i++ {
		counter.Inc("value1", "value2", "value3")
	}

	datadog.Stop()
}

func TestDataDogMeterWrongAgentHost(t *testing.T) {

	datadog, _ := datadogNewMeter("")
	if datadog != nil {
		t.Fatal("Valid datadog")
	}
}

func TestDataDogLogger(t *testing.T) {

	datadog, _ := datadogNewLogger("localhost")
	if datadog == nil {
		t.Fatal("Invalid datadog")
	}
	datadog.Info(nil)
	datadog.Info("info")
	datadog.Warn("warn")
	datadog.Debug("debug")

	datadog.Error("error")
	datadog.Error(errors.New("some error"))
	datadog.Error("error => %s", "message")

}
