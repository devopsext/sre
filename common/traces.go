package common

import (
	utils "github.com/devopsext/utils"
)

type TracesSpanContext struct {
	contexts map[Tracer]TracerSpanContext
	span     *TracesSpan
}

type TracesSpan struct {
	spans       map[Tracer]TracerSpan
	traceID     string
	spanID      string
	spanContext *TracesSpanContext
	traces      *Traces
}

type Traces struct {
	tracers []Tracer
}

func (tssc TracesSpanContext) GetTraceID() string {

	if tssc.span == nil {
		return ""
	}
	return tssc.span.traceID
}

func (tssc TracesSpanContext) GetSpanID() string {

	if tssc.span == nil {
		return ""
	}
	return tssc.span.spanID
}

func (tss *TracesSpan) GetContext() TracerSpanContext {

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

func (tss *TracesSpan) SetName(name string) TracerSpan {

	for _, s := range tss.spans {
		s.SetName(name)
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

	traceID := NewTraceID()
	spanID := NewSpanID()

	span := TracesSpan{
		traces:  ts,
		spans:   make(map[Tracer]TracerSpan),
		traceID: traceID,
		spanID:  spanID,
	}

	for _, t := range ts.tracers {

		s := t.StartSpanWithTraceID(span.traceID, span.spanID)
		if s != nil {
			span.spans[t] = s
		}
	}
	return &span
}

func (ts *Traces) StartSpanWithTraceID(traceID, spanID string) TracerSpan {

	if utils.IsEmpty(traceID) {
		traceID = NewTraceID()
	}

	if utils.IsEmpty(spanID) {
		spanID = NewSpanID()
	}

	span := TracesSpan{
		traces:  ts,
		spans:   make(map[Tracer]TracerSpan),
		traceID: traceID,
		spanID:  spanID,
	}

	for _, t := range ts.tracers {

		s := t.StartSpanWithTraceID(span.traceID, span.spanID)
		if s != nil {
			span.spans[t] = s
		}
	}
	return &span
}

func (ts *Traces) StartChildSpan(object interface{}) TracerSpan {

	var traceID string
	var spanID string

	spanCtx, spanCtxOk := object.(*TracesSpanContext)
	if spanCtxOk {
		traceID = spanCtx.GetTraceID()
		spanID = spanCtx.GetSpanID()
	}

	span := TracesSpan{
		traces:  ts,
		spans:   make(map[Tracer]TracerSpan),
		traceID: traceID,
		spanID:  spanID,
	}

	for _, t := range ts.tracers {

		var s TracerSpan
		if spanCtxOk {
			s = t.StartChildSpan(spanCtx.contexts[t])
		} else {
			s = t.StartChildSpan(object)
		}
		if s != nil {
			span.spans[t] = s

			sCtx := s.GetContext()
			if sCtx != nil {

				// find first traceID if there is no one
				if utils.IsEmpty(span.traceID) {
					span.traceID = sCtx.GetTraceID()
				}
				// find first spanID if there is no one
				if utils.IsEmpty(span.spanID) {
					span.spanID = sCtx.GetSpanID()
				}
			}
		}
	}
	return &span
}

func (ts *Traces) StartFollowSpan(object interface{}) TracerSpan {

	var traceID string
	var spanID string

	spanCtx, spanCtxOk := object.(*TracesSpanContext)
	if spanCtxOk {
		traceID = spanCtx.GetTraceID()
		spanID = spanCtx.GetSpanID()
	}

	span := TracesSpan{
		traces:  ts,
		spans:   make(map[Tracer]TracerSpan),
		traceID: traceID,
		spanID:  spanID,
	}

	for _, t := range ts.tracers {

		var s TracerSpan
		if spanCtxOk {
			s = t.StartFollowSpan(spanCtx.contexts[t])
		} else {
			s = t.StartFollowSpan(object)
		}
		if s != nil {
			span.spans[t] = s

			sCtx := s.GetContext()
			if sCtx != nil {

				// find first traceID if there is no one
				if utils.IsEmpty(span.traceID) {
					span.traceID = sCtx.GetTraceID()
				}
				// find first spanID if there is no one
				if utils.IsEmpty(span.spanID) {
					span.spanID = sCtx.GetSpanID()
				}
			}
		}
	}
	return &span
}

func (ts *Traces) Stop() {

	for _, t := range ts.tracers {
		t.Stop()
	}
}

func NewTraces() *Traces {

	ts := Traces{}
	return &ts
}
