package common

import (
	"math"
	"testing"
)

func TestUtilsTraceID(t *testing.T) {

	s := TraceIDUint64ToHex(0)
	if len(s) != 32 {
		t.Fatal("Wrong trace ID lenght")
	}

	i := TraceIDHexToUint64(s)
	if i != 0 {
		t.Fatal("Wrong trace ID hex")
	}

	s = TraceIDUint64ToHex(1)
	if s != "00000000000000000000000000000001" {
		t.Fatal("Wrong trace ID num")
	}

	i = TraceIDHexToUint64(s)
	if i != 1 {
		t.Fatal("Wrong trace ID hex")
	}

	s = TraceIDUint64ToHex(uint64(math.Pow(2, 32)))
	if s != "00000000000000000000000100000000" {
		t.Fatal("Wrong trace ID num")
	}

	i = TraceIDHexToUint64(s)
	if i != uint64(math.Pow(2, 32)) {
		t.Fatal("Wrong trace ID hex")
	}

	s = TraceIDUint64ToHex(uint64(math.Pow(2, 64)))
	if s != "00000000000000008000000000000000" {
		t.Fatal("Wrong trace ID num")
	}

	i = TraceIDHexToUint64(s)
	if i != uint64(math.Pow(2, 64)) {
		t.Fatal("Wrong trace ID hex")
	}
}

func TestUtilsSpanID(t *testing.T) {

	s := SpanIDUint64ToHex(0)
	if len(s) != 16 {
		t.Fatal("Wrong span ID lenght")
	}

	i := SpanIDHexToUint64(s)
	if i != 0 {
		t.Fatal("Wrong trace ID hex")
	}

	s = SpanIDUint64ToHex(1)
	if s != "0000000000000001" {
		t.Fatal("Wrong span ID num")
	}

	i = SpanIDHexToUint64(s)
	if i != 1 {
		t.Fatal("Wrong trace ID hex")
	}

	s = SpanIDUint64ToHex(uint64(math.Pow(2, 32)))
	if s != "0000000100000000" {
		t.Fatal("Wrong span ID num")
	}

	i = SpanIDHexToUint64(s)
	if i != uint64(math.Pow(2, 32)) {
		t.Fatal("Wrong trace ID hex")
	}

	s = SpanIDUint64ToHex(uint64(math.Pow(2, 64)))
	if s != "8000000000000000" {
		t.Fatal("Wrong span ID num")
	}

	i = SpanIDHexToUint64(s)
	if i != uint64(math.Pow(2, 64)) {
		t.Fatal("Wrong trace ID hex")
	}
}
