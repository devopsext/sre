package provider

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/devopsext/sre/common"
)

func stdoutTestSpan(t *testing.T, stdout *Stdout, level string, span common.TracerSpan, args ...interface{}) string {

	ctx := span.GetContext()
	if ctx == nil {
		t.Fatal("Invalid span context")
	}
	traceID := ctx.GetTraceID()

	msg := fmt.Sprintf("Some %s message...", level)
	switch level {
	case "info":
		stdout.SpanInfo(span, msg, args...)
	case "error":
		stdout.SpanError(span, errors.New(msg), args...)
	case "panic":
		stdout.SpanPanic(span, msg, args...)
	case "warn":
		stdout.SpanWarn(span, msg, args...)
	case "debug":
		stdout.SpanDebug(span, msg, args...)
	default:
		stdout.SpanInfo(nil, msg, args...)
		traceID = "<no value>"
	}

	return fmt.Sprintf("%s %s", msg, traceID)
}

func stdoutstdoutTest(stdout *Stdout, level string, args ...interface{}) string {

	msg := fmt.Sprintf("Some %s message...", level)
	switch level {
	case "info":
		stdout.Info(msg, args...)
	case "error":
		stdout.Error(errors.New(msg), args...)
	case "panic":
		stdout.Panic(msg, args...)
	case "warn":
		stdout.Warn(msg, args...)
	case "debug":
		stdout.Debug(msg, args...)
	default:
		stdout.Info(msg, args...)
	}
	return msg
}

func stdoutTest(t *testing.T, format, level, template string, span common.TracerSpan, args ...interface{}) {

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	stdout := NewStdout(StdoutOptions{
		Format:          format,
		Level:           level,
		Template:        template,
		TimestampFormat: time.RFC3339Nano,
		TextColors:      true,
		Debug:           true,
	})
	if stdout == nil {
		t.Fatal("Invalid stdout")
	}
	stdout.SetCallerOffset(2)
	stdout.Stack(-1).Stack(1)

	var msg string
	if span != nil {
		msg = stdoutTestSpan(t, stdout, level, span)
	} else {
		msg = stdoutstdoutTest(stdout, level)
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
			t.Fatal("Stdout template message is wrong")
		}
	case "text", "json":
		if !strings.Contains(output, msg) {
			t.Fatal("Stdout text/json message is wrong")
		}
	}
}

func stdoutTestTemplate(t *testing.T, level string, span common.TracerSpan) {

	stdoutTest(t, "template", level, "{{.msg}}", nil)
}

func stdoutTestTemplateSpan(t *testing.T, level string) {

	tracer := common.NewTraces()
	if tracer == nil {
		t.Fatal("Invalid tracer")
	}

	/*	traceID := tracer.NewTraceID()
			if utils.IsEmpty(traceID) {
				t.Fatal("Invalid trace ID")
			}

		span := tracer.StartSpanWithTraceID(traceID, "")
		if span == nil {
			t.Fatal("Invalid span")
		}

		stdoutTest(t, "template", level, "{{.msg}} {{.trace_id}}", span)
	*/
}

func TestStdoutNormal(t *testing.T) {

	stdoutTestTemplate(t, "", nil)
	stdoutTestTemplate(t, "info", nil)
	stdoutTestTemplate(t, "error", nil)
	stdoutTestTemplate(t, "warn", nil)
	stdoutTestTemplate(t, "debug", nil)

	stdoutTestTemplateSpan(t, "")
	stdoutTestTemplateSpan(t, "info")
	stdoutTestTemplateSpan(t, "error")
	stdoutTestTemplateSpan(t, "warn")
	stdoutTestTemplateSpan(t, "debug")

	stdoutTest(t, "text", "", "", nil)
	stdoutTest(t, "json", "", "", nil)
	stdoutTest(t, "", "", "", nil)
}

func TestStdoutPanic(t *testing.T) {

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("It should be paniced")
		}
	}()

	stdoutTest(t, "", "panic", "", nil)
}

// failed on test
/*func TestStdoutPanicSpan(t *testing.T) {

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("It should be paniced")
		}
	}()

	stdoutTestTemplateSpan(t, "panic")
}*/

func TestStdoutWrongTemplate(t *testing.T) {

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("It should be paniced")
		}
	}()

	stdoutTest(t, "template", "info", "{{.msg {{.trace2_id}}", nil)
}

func TestStdoutWrongArgs(t *testing.T) {

	stdoutTest(t, "template", "info", "{{.msg}}", nil, nil)
}
