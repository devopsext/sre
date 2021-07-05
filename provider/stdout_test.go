package provider

import (
	"testing"
	"time"
)

func TestStdout(t *testing.T) {

	stdout := NewStdout(StdoutOptions{
		Format:          "template",
		Level:           "info",
		Template:        "{{.msg}}",
		TimestampFormat: time.RFC3339Nano,
		TextColors:      true,
	})
	if stdout == nil {
		t.Error("Stdout is not defined")
	}
	//stdout.SetCallerOffset(2)
	stdout.Info("Some info message...")
}
