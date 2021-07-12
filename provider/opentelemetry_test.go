package provider

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/devopsext/utils"
	"go.opentelemetry.io/otel/trace"
)

func opentelemetryNewTracer(agentHost string) (*OpentelemetryTracer, *Stdout) {

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

	opentelemetry := NewOpentelemetryTracer(OpentelemetryTracerOptions{
		AgentHost: agentHost,
		AgentPort: 8126,
		OpentelemetryOptions: OpentelemetryOptions{
			ServiceName: "sre-opentelemetry-tracer-test",
			Attributes:  "tag1=value1,,tag3=${key3:value3}",
		},
	}, nil, stdout)

	return opentelemetry, stdout
}

func opentelemetryNewMeter(agentHost string) (*OpentelemetryMeter, *Stdout) {

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

	opentelemetry := NewOpentelemetryMeter(OpentelemetryMeterOptions{
		AgentHost:     agentHost,
		AgentPort:     8126,
		Prefix:        "test",
		CollectPeriod: 0,
		OpentelemetryOptions: OpentelemetryOptions{
			ServiceName: "sre-opentelemetry-meter-test",
			Attributes:  "tag1=value1,,tag3=${key3:value3}",
		},
	}, nil, stdout)

	return opentelemetry, stdout
}

func TestOpentelemetryTracer(t *testing.T) {

	opentelemetry, _ := opentelemetryNewTracer("localhost")
	if opentelemetry == nil {
		t.Fatal("Invalid opentelemetry")
	}
	opentelemetry.SetCallerOffset(1)

	span := opentelemetry.StartSpan()
	if span == nil {
		t.Fatal("Invalid span")
	}
	defer span.Finish()
	span.SetName("some-span")
	span.SetTag("string", "some-string")
	span.SetTag("int", 21412)
	span.SetTag("int64", 2131242354364645434)
	span.SetTag("flota64", 2131242354364645434.99)

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

	traceSpan := opentelemetry.StartSpanWithTraceID(traceID, "")
	if traceSpan == nil {
		t.Fatal("Invalid trace span")
	}
	defer traceSpan.Finish()
	traceSpan.SetName("some-trace-span")
	traceSpan.SetBaggageItem("key", "value")
	traceSpan.SetTag("parent-span-ID", spanID)

	childSpan := opentelemetry.StartChildSpan(ctx)
	if childSpan == nil {
		t.Fatal("Invalid child span")
	}
	defer childSpan.Finish()
	childSpan.SetName("some-child-span")

	followSpan := opentelemetry.StartFollowSpan(ctx)
	if followSpan == nil {
		t.Fatal("Invalid child span")
	}
	defer followSpan.Finish()
	followSpan.SetName("some-follow-span")

	nilChildSpan := opentelemetry.StartChildSpan(t)
	if nilChildSpan != nil {
		t.Fatal("Invalid nil child span")
	}

	nilFollowSpan := opentelemetry.StartFollowSpan(t)
	if nilFollowSpan != nil {
		t.Fatal("Invalid nil follow span")
	}

	headers := make(http.Header)

	span.SetCarrier(t)
	nilHeaderSpan := opentelemetry.StartFollowSpan(headers)
	if nilHeaderSpan != nil {
		t.Fatal("Valid nil header span")
	}

	span.SetCarrier(headers)
	headerSpan := opentelemetry.StartChildSpan(headers)
	if headerSpan == nil {
		t.Fatal("Invalid nil header span")
	}

	nilSpan := opentelemetry.StartSpanWithTraceID("", "")
	if nilSpan != nil {
		t.Fatal("Valid nil span")
	}

	opentelemetry.Stop()
}

func TestOpentelemetryTracerWrongAgentHost(t *testing.T) {

	opentelemetry, _ := opentelemetryNewTracer("")
	if opentelemetry != nil {
		t.Fatal("Valid opentelemetry")
	}
}

func TestOpentelemetryTracerWrongSpan(t *testing.T) {

	span := OpentelemetryTracerSpan{}

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

func TestOpentelemetryTracerWrongSpanContext(t *testing.T) {

	ctx := OpentelemetryTracerSpanContext{}

	traceID := ctx.GetTraceID()
	if !utils.IsEmpty(traceID) {
		t.Fatal("Valid trace ID")
	}

	spanID := ctx.GetSpanID()
	if !utils.IsEmpty(spanID) {
		t.Fatal("Valid span ID")
	}

	ctx.context = &trace.SpanContext{}

	traceID = ctx.GetTraceID()
	if !utils.IsEmpty(traceID) {
		t.Fatal("Valid trace ID")
	}

	spanID = ctx.GetSpanID()
	if !utils.IsEmpty(spanID) {
		t.Fatal("Valid span ID")
	}
}

func TestOpentelemetryMeter(t *testing.T) {

	opentelemetry, _ := opentelemetryNewMeter("localhost")
	if opentelemetry == nil {
		t.Fatal("Invalid opentelemetry")
	}
	opentelemetry.SetCallerOffset(1)

	secondPrefix := "counter"
	metricName := "some"

	counter := opentelemetry.Counter(metricName, "description", []string{"one", "two", "three"}, secondPrefix)
	if counter == nil {
		t.Fatal("Invalid opentelemetry")
	}

	maxCounter := 5
	for i := 0; i < maxCounter; i++ {
		counter.Inc("value1", "value2", "value3")
	}

	opentelemetry.Stop()
}

func TestOpentelemetryMeterWrongAgentHost(t *testing.T) {

	opentelemetry, _ := opentelemetryNewMeter("")
	if opentelemetry != nil {
		t.Fatal("Valid opentelemetry")
	}
}
