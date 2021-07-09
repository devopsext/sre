package common

type TracerSpanContext interface {
	GetTraceID() string
	GetSpanID() string
}

type TracerSpan interface {
	GetContext() TracerSpanContext
	SetCarrier(object interface{}) TracerSpan
	SetName(name string) TracerSpan
	SetTag(key string, value interface{}) TracerSpan
	Error(err error) TracerSpan
	SetBaggageItem(restrictedKey, value string) TracerSpan
	Finish()
}

type Tracer interface {
	StartSpan() TracerSpan
	StartSpanWithTraceID(traceID, spanID string) TracerSpan
	StartChildSpan(object interface{}) TracerSpan
	StartFollowSpan(object interface{}) TracerSpan
	Stop()
}
