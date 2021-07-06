package provider

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/devopsext/sre/common"
)

func outputSpan(t *testing.T, stdout *Stdout, level string, span common.TracerSpan) string {

	ctx := span.GetContext()
	if ctx == nil {
		t.Error("Invalid span context")
	}
	traceID := ctx.GetTraceID()

	msg := fmt.Sprintf("Some %s message...", level)
	switch level {
	case "info":
		stdout.SpanInfo(span, msg)
	case "error":
		stdout.SpanError(span, msg)
	case "panic":
		stdout.SpanPanic(span, msg)
	case "warn":
		stdout.SpanWarn(span, msg)
	case "debug":
		stdout.SpanDebug(span, msg)
	default:
		stdout.SpanInfo(span, msg)
	}

	return fmt.Sprintf("%s %d", msg, traceID)
}

func output(stdout *Stdout, level string) string {

	msg := fmt.Sprintf("Some %s message...", level)
	switch level {
	case "info":
		stdout.Info(msg)
	case "error":
		stdout.Error(msg)
	case "panic":
		stdout.Panic(msg)
	case "warn":
		stdout.Warn(msg)
	case "debug":
		stdout.Debug(msg)
	default:
		stdout.Info(msg)
	}
	return msg
}

func testTemplate(t *testing.T, level, template string, span common.TracerSpan) {

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	stdout := NewStdout(StdoutOptions{
		Format:          "template",
		Level:           level,
		Template:        template,
		TimestampFormat: time.RFC3339Nano,
		TextColors:      true,
	})
	if stdout == nil {
		t.Error("Invalid stdout")
	}
	stdout.SetCallerOffset(2)

	var msg string
	if span != nil {
		msg = outputSpan(t, stdout, level, span)
	} else {
		msg = output(stdout, level)
	}

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = oldStdout

	output := string(out)
	output = strings.TrimRight(output, "\n")

	t.Logf("Output is ... [%s]", output)
	t.Logf("Message is ... [%s]", msg)

	if strings.Compare(output, msg) != 0 {
		t.Error("Stdout message is wrong")
	}
}

func testSpan(t *testing.T, level string) {

	tracer := common.NewTraces()
	if tracer == nil {
		t.Error("Invalid tracer")
	}

	traceID := tracer.NewTraceID()
	if traceID == 0 {
		t.Error("Invalid trace ID")
	}

	span := tracer.StartSpanWithTraceID(traceID)
	if span == nil {
		t.Error("Invalid span")
	}

	testTemplate(t, "info", "{{.msg}} {{.trace_id}}", span)
}

func TestStdout(t *testing.T) {
	testTemplate(t, "info", "{{.msg}}", nil)
	testTemplate(t, "error", "{{.msg}}", nil)
	testTemplate(t, "warn", "{{.msg}}", nil)
	testTemplate(t, "debug", "{{.msg}}", nil)
	testSpan(t, "info")
	testSpan(t, "error")
	testSpan(t, "warn")
	testSpan(t, "debug")
}
