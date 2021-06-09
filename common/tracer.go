package common

type TracerSpanContext interface {
	GetTraceID() uint64
}

type TracerSpan interface {
	GetContext() TracerSpanContext
	SetCarrier(object interface{}) TracerSpan
	SetTag(key string, value interface{}) TracerSpan
	Error(err error) TracerSpan
	SetBaggageItem(restrictedKey, value string) TracerSpan
	Finish()
}

type Tracer interface {
	Enabled() bool
	StartSpan() TracerSpan
	StartSpanWithTraceID(traceID uint64) TracerSpan
	StartChildSpan(object interface{}) TracerSpan
	StartFollowSpan(object interface{}) TracerSpan
}
