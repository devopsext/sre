package provider

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/devopsext/utils"
	"github.com/opentracing/opentracing-go"
)

func jaegerNew(agentHost string) (*JaegerTracer, *Stdout) {

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

	jaeger := NewJaegerTracer(JaegerOptions{
		AgentHost:   agentHost,
		AgentPort:   6831,
		ServiceName: "sre-jaeger-test",
		Tags:        "tag1=value1,,tag3=${key3:value3}",
		Debug:       true,
	}, nil, stdout)

	return jaeger, stdout
}

func TestJaeger(t *testing.T) {

	jaeger, _ := jaegerNew("localhost")
	if jaeger == nil {
		t.Fatal("Invalid jaeger")
	}
	jaeger.SetCallerOffset(1)

	span := jaeger.StartSpan()
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

	traceSpan := jaeger.StartSpanWithTraceID(traceID, "")
	if traceSpan == nil {
		t.Fatal("Invalid trace span")
	}
	defer traceSpan.Finish()
	traceSpan.SetName("some-trace-span")
	traceSpan.SetBaggageItem("key", "value")
	traceSpan.SetTag("parent-span-ID", spanID)

	childSpan := jaeger.StartChildSpan(ctx)
	if childSpan == nil {
		t.Fatal("Invalid child span")
	}
	defer childSpan.Finish()
	childSpan.SetName("some-child-span")

	followSpan := jaeger.StartFollowSpan(ctx)
	if followSpan == nil {
		t.Fatal("Invalid child span")
	}
	defer followSpan.Finish()
	followSpan.SetName("some-follow-span")

	nilChildSpan := jaeger.StartChildSpan(t)
	if nilChildSpan != nil {
		t.Fatal("Invalid nil child span")
	}

	nilFollowSpan := jaeger.StartFollowSpan(t)
	if nilFollowSpan != nil {
		t.Fatal("Invalid nil follow span")
	}

	headers := make(http.Header)

	span.SetCarrier(t)
	nilHeaderSpan := jaeger.StartFollowSpan(headers)
	if nilHeaderSpan != nil {
		t.Fatal("Valid nil header span")
	}

	span.SetCarrier(headers)
	headerSpan := jaeger.StartChildSpan(headers)
	if headerSpan == nil {
		t.Fatal("Invalid nil header span")
	}

	nilSpan := jaeger.StartSpanWithTraceID("", "")
	if nilSpan != nil {
		t.Fatal("Valid nil span")
	}
}

func TestJaegerWrongAgentHost(t *testing.T) {

	jaeger, _ := jaegerNew("")
	if jaeger != nil {
		t.Fatal("Valid jaeger")
	}
}

func TestJaegerInternalLogger(t *testing.T) {

	_, stdout := jaegerNew("localhost")
	if stdout == nil {
		t.Fatal("Valid stdout")
	}

	internalLogger := JaegerInternalLogger{
		logger: stdout,
	}

	internalLogger.Error("Some internal message")
	internalLogger.Infof("")
}

func TestJaegerWrongSpan(t *testing.T) {

	span := JaegerSpan{}

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

	var fields map[string]interface{}

	s = span.LogFields(fields)
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

	tracer := opentracing.NoopTracer{}
	span.span = tracer.StartSpan("some-noop-span")

	span.spanContext = &JaegerSpanContext{}

	ctx = span.GetContext()
	if ctx == nil {
		t.Fatal("Invalid span context")
	}

	span.span = nil
	span.Finish()
}

func TestJaegerWrongSpanContext(t *testing.T) {

	ctx := JaegerSpanContext{}

	traceID := ctx.GetTraceID()
	if !utils.IsEmpty(traceID) {
		t.Fatal("Valid trace ID")
	}

	spanID := ctx.GetSpanID()
	if !utils.IsEmpty(spanID) {
		t.Fatal("Valid span ID")
	}

	tracer := opentracing.NoopTracer{}
	span := tracer.StartSpan("some-noop-span")

	ctx.context = span.Context()

	traceID = ctx.GetTraceID()
	if !utils.IsEmpty(traceID) {
		t.Fatal("Valid trace ID")
	}

	spanID = ctx.GetSpanID()
	if !utils.IsEmpty(spanID) {
		t.Fatal("Valid span ID")
	}
}
