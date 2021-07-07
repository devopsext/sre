package common

type Logger interface {
	Info(obj interface{}, args ...interface{}) Logger
	SpanInfo(span TracerSpan, obj interface{}, args ...interface{}) Logger
	Warn(obj interface{}, args ...interface{}) Logger
	SpanWarn(span TracerSpan, obj interface{}, args ...interface{}) Logger
	Error(obj interface{}, args ...interface{}) Logger
	SpanError(span TracerSpan, obj interface{}, args ...interface{}) Logger
	Debug(obj interface{}, args ...interface{}) Logger
	SpanDebug(span TracerSpan, obj interface{}, args ...interface{}) Logger
	Panic(obj interface{}, args ...interface{})
	SpanPanic(span TracerSpan, obj interface{}, args ...interface{})
	Stack(offset int) Logger
}
