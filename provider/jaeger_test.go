package provider

import (
	"errors"
	"testing"
	"time"
)

func TestJaeger(t *testing.T) {

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

	jaeger := NewJaegerTracer(JaegerOptions{
		AgentHost:   "localhost",
		AgentPort:   6831,
		ServiceName: "sre-jaeger-test",
	}, nil, stdout)
	if jaeger == nil {
		t.Error("Invalid jaeger")
	}
	jaeger.SetCallerOffset(1)

	span := jaeger.StartSpan()
	if span == nil {
		t.Error("Invalid span")
	}
	defer span.Finish()
	span.SetName("some-span")
	span.SetTag("key1", "Value1")
	span.Error(errors.New("some-span-error"))

	ctx := span.GetContext()
	if ctx == nil {
		t.Error("Invalid span context")
	}

	traceSpan := jaeger.StartSpanWithTraceID(ctx.GetTraceID())
	if traceSpan == nil {
		t.Error("Invalid trace span")
	}
	defer traceSpan.Finish()
	traceSpan.SetName("some-trace-span")
	traceSpan.SetBaggageItem("key", "value")
	traceSpan.SetTag("parent-span-ID", ctx.GetSpanID())

	childSpan := jaeger.StartChildSpan(ctx)
	if childSpan == nil {
		t.Error("Invalid child span")
	}
	defer childSpan.Finish()
	childSpan.SetName("some-child-span")

	followSpan := jaeger.StartFollowSpan(ctx)
	if followSpan == nil {
		t.Error("Invalid child span")
	}
	defer followSpan.Finish()
	followSpan.SetName("some-follow-span")

}
