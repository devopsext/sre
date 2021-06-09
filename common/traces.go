package common

import (
	"math/rand"
	"sync"
	"time"

	"github.com/uber/jaeger-client-go/utils"
)

type TracesSpanContext struct {
	contexts map[Tracer]TracerSpanContext
	span     *TracesSpan
}

type TracesSpan struct {
	spans       map[Tracer]TracerSpan
	traceID     uint64
	spanContext *TracesSpanContext
	traces      *Traces
}

type Traces struct {
	randomNumber func() uint64
	tracers      []Tracer
}

func (tssc TracesSpanContext) GetTraceID() uint64 {

	if tssc.span == nil {
		return 0
	}
	return tssc.span.traceID
}

func (tss *TracesSpan) GetContext() TracerSpanContext {

	if len(tss.spans) <= 0 {
		return nil
	}

	if tss.spanContext != nil {
		return tss.spanContext
	}

	tss.spanContext = &TracesSpanContext{
		contexts: make(map[Tracer]TracerSpanContext),
		span:     tss,
	}

	for t, s := range tss.spans {

		ctx := s.GetContext()
		if ctx != nil {
			tss.spanContext.contexts[t] = ctx
		}
	}

	return tss.spanContext
}

func (tss *TracesSpan) SetCarrier(object interface{}) TracerSpan {

	for _, s := range tss.spans {
		s.SetCarrier(object)
	}
	return tss
}

func (tss *TracesSpan) SetTag(key string, value interface{}) TracerSpan {

	for _, s := range tss.spans {
		s.SetTag(key, value)
	}
	return tss
}

func (tss *TracesSpan) SetBaggageItem(restrictedKey, value string) TracerSpan {

	for _, s := range tss.spans {
		s.SetBaggageItem(restrictedKey, value)
	}
	return tss
}

func (tss *TracesSpan) Error(err error) TracerSpan {

	for _, s := range tss.spans {
		s.Error(err)
	}
	return tss
}

func (tss *TracesSpan) Finish() {
	for _, s := range tss.spans {
		s.Finish()
	}
}

func (ts *Traces) Register(t Tracer) {
	if t != nil {
		ts.tracers = append(ts.tracers, t)
	}
}

func (ts *Traces) StartSpan() TracerSpan {

	if len(ts.tracers) <= 0 {
		return nil
	}

	span := TracesSpan{
		traces:  ts,
		spans:   make(map[Tracer]TracerSpan),
		traceID: ts.randomNumber(),
	}

	for _, t := range ts.tracers {

		if !t.Enabled() {
			continue
		}

		s := t.StartSpanWithTraceID(span.traceID)
		if s != nil {
			span.spans[t] = s
		}
	}
	return &span
}

func (ts *Traces) StartSpanWithTraceID(traceID uint64) TracerSpan {

	if len(ts.tracers) <= 0 {
		return nil
	}

	span := TracesSpan{
		traces:  ts,
		spans:   make(map[Tracer]TracerSpan),
		traceID: traceID,
	}

	for _, t := range ts.tracers {

		if !t.Enabled() {
			continue
		}

		s := t.StartSpanWithTraceID(span.traceID)
		if s != nil {
			span.spans[t] = s
		}
	}
	return &span
}

func (ts *Traces) StartChildSpan(object interface{}) TracerSpan {

	if len(ts.tracers) <= 0 {
		return nil
	}

	var traceID uint64
	spanCtx, spanCtxOk := object.(*TracesSpanContext)
	if spanCtxOk {
		traceID = spanCtx.GetTraceID()
	}

	span := TracesSpan{
		traces:  ts,
		spans:   make(map[Tracer]TracerSpan),
		traceID: traceID,
	}

	for _, t := range ts.tracers {

		if !t.Enabled() {
			continue
		}

		var s TracerSpan
		if spanCtxOk {
			s = t.StartChildSpan(spanCtx.contexts[t])
		} else {
			s = t.StartChildSpan(object)
		}
		if s != nil {
			span.spans[t] = s

			// find first traceID if there is no on
			sCtx := s.GetContext()
			if sCtx != nil && span.traceID == 0 {
				span.traceID = sCtx.GetTraceID()
			}
		}
	}
	return &span
}

func (ts *Traces) StartFollowSpan(object interface{}) TracerSpan {

	if len(ts.tracers) <= 0 {
		return nil
	}

	var traceID uint64
	spanCtx, spanCtxOk := object.(*TracesSpanContext)
	if spanCtxOk {
		traceID = spanCtx.GetTraceID()
	}

	span := TracesSpan{
		traces:  ts,
		spans:   make(map[Tracer]TracerSpan),
		traceID: traceID,
	}

	for _, t := range ts.tracers {

		if !t.Enabled() {
			continue
		}

		var s TracerSpan
		if spanCtxOk {
			s = t.StartFollowSpan(spanCtx.contexts[t])
		} else {
			s = t.StartFollowSpan(object)
		}
		if s != nil {
			span.spans[t] = s

			// find first traceID if there is no on
			sCtx := s.GetContext()
			if sCtx != nil && span.traceID == 0 {
				span.traceID = sCtx.GetTraceID()
			}
		}
	}
	return &span
}

func (ts *Traces) Enabled() bool {
	return true
}

func NewTraces() *Traces {

	ts := Traces{}

	seedGenerator := utils.NewRand(time.Now().UnixNano())
	pool := sync.Pool{
		New: func() interface{} {
			return rand.NewSource(seedGenerator.Int63())
		},
	}

	ts.randomNumber = func() uint64 {
		generator := pool.Get().(rand.Source)
		number := uint64(generator.Int63())
		pool.Put(generator)
		return number
	}

	return &ts
}
