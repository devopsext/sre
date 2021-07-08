package provider

import (
	"errors"
	"testing"
	"time"
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
		Tags:        "tag1=value1,tag2=value2",
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

	traceSpan := jaeger.StartSpanWithTraceID(ctx.GetTraceID())
	if traceSpan == nil {
		t.Fatal("Invalid trace span")
	}
	defer traceSpan.Finish()
	traceSpan.SetName("some-trace-span")
	traceSpan.SetBaggageItem("key", "value")
	traceSpan.SetTag("parent-span-ID", ctx.GetSpanID())

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
}

func TestJaegerWrongAgentHost(t *testing.T) {

	jaeger, _ := jaegerNew("")
	if jaeger != nil {
		t.Fatal("Valid jaeger")
	}
}