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
