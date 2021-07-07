package provider

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
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
	sTraceID := strconv.Itoa(int(traceID))

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
		stdout.SpanInfo(nil, msg)
		sTraceID = "<no value>"
	}

	return fmt.Sprintf("%s %s", msg, sTraceID)
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

func test(t *testing.T, format, level, template string, span common.TracerSpan) {

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	stdout := NewStdout(StdoutOptions{
		Format:          format,
		Level:           level,
		Template:        template,
		TimestampFormat: time.RFC3339Nano,
		TextColors:      true,
	})
	if stdout == nil {
		t.Error("Invalid stdout")
	}
	stdout.SetCallerOffset(2)
	stdout.Stack(-1).Stack(1)

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

	switch format {
	case "template":
		if strings.Compare(output, msg) != 0 {
			t.Error("Stdout template message is wrong")
		}
	case "text", "json":
		if !strings.Contains(output, msg) {
			t.Error("Stdout text/json message is wrong")
		}
	}
}

func testTemplate(t *testing.T, level string, span common.TracerSpan) {

	test(t, "template", level, "{{.msg}}", nil)
}

func testTemplateSpan(t *testing.T, level string) {

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

	test(t, "template", level, "{{.msg}} {{.trace_id}}", span)
}

func TestStdoutNormal(t *testing.T) {

	testTemplate(t, "", nil)
	testTemplate(t, "info", nil)
	testTemplate(t, "error", nil)
	testTemplate(t, "warn", nil)
	testTemplate(t, "debug", nil)

	testTemplateSpan(t, "")
	testTemplateSpan(t, "info")
	testTemplateSpan(t, "error")
	testTemplateSpan(t, "warn")
	testTemplateSpan(t, "debug")

	test(t, "text", "", "", nil)
	test(t, "json", "", "", nil)
	test(t, "", "", "", nil)
}

func TestStdoutPanic(t *testing.T) {

	defer func() {
		if r := recover(); r == nil {
			t.Error("It should be paniced")
		}
	}()

	test(t, "", "panic", "", nil)
}

func TestStdoutPanicSpan(t *testing.T) {

	defer func() {
		if r := recover(); r == nil {
			t.Error("It should be paniced")
		}
	}()

	testTemplateSpan(t, "panic")
}
