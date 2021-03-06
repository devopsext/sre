package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/devopsext/sre/common"
	"github.com/devopsext/utils"
	"github.com/opentracing/opentracing-go"
	opentracingLog "github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
)

type JaegerOptions struct {
	ServiceName         string
	AgentHost           string
	AgentPort           int
	Endpoint            string
	User                string
	Password            string
	BufferFlushInterval int
	QueueSize           int
	Tags                string
	Version             string
	Debug               bool
}

type JaegerSpanContext struct {
	context opentracing.SpanContext
}

type JaegerSpan struct {
	span         opentracing.Span
	spanContext  *JaegerSpanContext
	context      context.Context
	tracer       *JaegerTracer
	callerOffset int
}

type JaegerTracer struct {
	options      JaegerOptions
	callerOffset int
	tracer       opentracing.Tracer
	logger       common.Logger
}

type JaegerInternalLogger struct {
	logger common.Logger
}

func (jsc *JaegerSpanContext) GetTraceID() string {

	if jsc.context == nil {
		return ""
	}

	jaegerSpanCtx, ok := jsc.context.(jaeger.SpanContext)
	if !ok {
		return ""
	}
	return common.TraceIDUint64ToHex(jaegerSpanCtx.TraceID().Low)
}

func (jsc *JaegerSpanContext) GetSpanID() string {

	if jsc.context == nil {
		return ""
	}

	jaegerSpanCtx, ok := jsc.context.(jaeger.SpanContext)
	if !ok {
		return ""
	}
	return common.SpanIDUint64ToHex(uint64(jaegerSpanCtx.SpanID()))
}

func (js *JaegerSpan) GetContext() common.TracerSpanContext {
	if js.span == nil {
		return nil
	}

	if js.spanContext != nil {
		return js.spanContext
	}

	js.spanContext = &JaegerSpanContext{
		context: js.span.Context(),
	}
	return js.spanContext
}

func (js *JaegerSpan) SetCarrier(object interface{}) common.TracerSpan {

	if js.span == nil {
		return nil
	}

	h, ok := object.(http.Header)
	if ok {
		err := js.tracer.tracer.Inject(js.span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(h))
		if err != nil {
			js.tracer.logger.Error(err)
		}
	}
	return js
}

func (js *JaegerSpan) SetName(name string) common.TracerSpan {

	if js.span == nil {
		return nil
	}

	js.span.SetOperationName(name)
	return js
}

func (js *JaegerSpan) SetTag(key string, value interface{}) common.TracerSpan {

	if js.span == nil {
		return nil
	}

	js.span.SetTag(key, value)
	return js
}

func (js *JaegerSpan) LogFields(fields map[string]interface{}) common.TracerSpan {

	if js.span == nil {
		return nil
	}

	if len(fields) <= 0 {
		return js
	}

	var logFields []opentracingLog.Field

	for k, v := range fields {

		if v != nil {

			var logField opentracingLog.Field
			switch v.(type) {
			case string:
				logField = opentracingLog.String(k, v.(string))
				/*case bool:
					logField = opentracingLog.Bool(k, v.(bool))
				case int:
					logField = opentracingLog.Int(k, v.(int))
				case int64:
					logField = opentracingLog.Int64(k, v.(int64))
				case string:
					logField = opentracingLog.String(k, v.(string))
				case float32:
					logField = opentracingLog.Float32(k, v.(float32))
				case float64:
					logField = opentracingLog.Float64(k, v.(float64))
				case error:
					logField = opentracingLog.Error(v.(error))*/
			}

			logFields = append(logFields, logField)
		}
	}

	if len(logFields) > 0 {
		js.span.LogFields(logFields...)
	}
	return js
}

func (js *JaegerSpan) Error(err error) common.TracerSpan {

	if js.span == nil {
		return nil
	}

	js.SetTag("error", true)
	js.LogFields(map[string]interface{}{
		"error.message": err.Error(),
	})
	return js
}

func (js *JaegerSpan) SetBaggageItem(restrictedKey, value string) common.TracerSpan {

	if js.span == nil {
		return nil
	}

	js.span.SetBaggageItem(restrictedKey, value)
	return js
}

func (js JaegerSpan) Finish() {
	if js.span == nil {
		return
	}
	js.span.Finish()
}

func (j *JaegerTracer) startSpanFromContext(ctx context.Context, offset int, opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {

	operation, file, line := utils.CallerGetInfo(offset)

	span, context := opentracing.StartSpanFromContextWithTracer(ctx, j.tracer, operation, opts...)
	if span != nil {
		span.SetTag("file", fmt.Sprintf("%s:%d", file, line))
	}
	return span, context
}

func (j *JaegerTracer) startChildOfSpan(ctx context.Context, spanContext opentracing.SpanContext) (opentracing.Span, context.Context) {

	var span opentracing.Span
	var context context.Context
	if spanContext != nil {
		span, context = j.startSpanFromContext(ctx, j.callerOffset+5, opentracing.ChildOf(spanContext))
	} else {
		span, context = j.startSpanFromContext(ctx, j.callerOffset+5)
	}
	return span, context
}

func (j *JaegerTracer) startFollowsFromSpan(ctx context.Context, spanContext opentracing.SpanContext) (opentracing.Span, context.Context) {

	var span opentracing.Span
	var context context.Context
	if spanContext != nil {
		span, context = j.startSpanFromContext(ctx, j.callerOffset+5, opentracing.FollowsFrom(spanContext))
	} else {
		span, context = j.startSpanFromContext(ctx, j.callerOffset+5)
	}
	return span, context
}

func (j *JaegerTracer) StartSpan() common.TracerSpan {

	s, ctx := j.startSpanFromContext(context.Background(), j.callerOffset+4)
	return &JaegerSpan{
		span:         s,
		context:      ctx,
		tracer:       j,
		callerOffset: j.callerOffset,
	}
}

func (j *JaegerTracer) StartSpanWithTraceID(traceID, spanID string) common.TracerSpan {

	tID := common.TraceIDHexToUint64(traceID)
	if tID == 0 {
		j.logger.Error(errors.New("invalid trace ID"))
		return nil
	}

	sID := common.SpanIDHexToUint64(spanID)
	if sID == 0 {
		sID = tID
	}

	newTraceID := jaeger.TraceID{
		Low:  tID, // set your own trace ID
		High: 0,
	}
	var newSpanID jaeger.SpanID = jaeger.SpanID(sID)
	parentID := jaeger.SpanID(0)
	sampled := true

	baggage := make(map[string]string)

	newJaegerSpanCtx := jaeger.NewSpanContext(newTraceID, newSpanID, parentID, sampled, baggage)

	s, ctx := j.startSpanFromContext(context.Background(), j.callerOffset+4, jaeger.SelfRef(newJaegerSpanCtx))

	return &JaegerSpan{
		span:         s,
		context:      ctx,
		tracer:       j,
		callerOffset: j.callerOffset,
	}
}

func (j *JaegerTracer) getSpanContext(object interface{}) opentracing.SpanContext {

	h, ok := object.(http.Header)
	if ok {
		spanContext, err := j.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(h))
		if err != nil {
			j.logger.Error(err)
			return nil
		}
		return spanContext
	}

	sc, ok := object.(*JaegerSpanContext)
	if ok {
		return sc.context
	}
	return nil
}

func (j *JaegerTracer) StartChildSpan(object interface{}) common.TracerSpan {

	spanContext := j.getSpanContext(object)
	if spanContext == nil {
		return nil
	}

	s, ctx := j.startChildOfSpan(context.Background(), spanContext)
	return &JaegerSpan{
		span:    s,
		context: ctx,
		tracer:  j,
	}
}

func (j *JaegerTracer) StartFollowSpan(object interface{}) common.TracerSpan {

	spanContext := j.getSpanContext(object)
	if spanContext == nil {
		return nil
	}

	s, ctx := j.startFollowsFromSpan(context.Background(), spanContext)
	return &JaegerSpan{
		span:    s,
		context: ctx,
		tracer:  j,
	}
}

func (j *JaegerTracer) SetCallerOffset(offset int) {
	j.callerOffset = offset
}

func (j *JaegerInternalLogger) Error(msg string) {
	j.logger.Stack(-2).Error(msg).Stack(2)
}

func (j *JaegerInternalLogger) Infof(msg string, args ...interface{}) {

	if utils.IsEmpty(msg) {
		return
	}

	msg = strings.TrimSpace(msg)
	if args != nil {
		j.logger.Stack(-2).Info(msg, args...).Stack(2)
	} else {
		j.logger.Stack(-2).Info(msg).Stack(2)
	}
}

func (j *JaegerTracer) Stop() {
	// nothing here
}

func newJaegerTracer(options JaegerOptions, logger common.Logger, stdout *Stdout) opentracing.Tracer {

	disabled := utils.IsEmpty(options.AgentHost) && utils.IsEmpty(options.Endpoint)
	if disabled {
		return nil
	}

	tags := make([]opentracing.Tag, 0)
	m := utils.MapGetKeyValues(options.Tags)
	for k, v := range m {
		tag := opentracing.Tag{Key: k, Value: v}
		tags = append(tags, tag)
	}

	tags = append(tags, opentracing.Tag{
		Key:   "version",
		Value: options.Version,
	})

	cfg := &jaegerConfig.Configuration{

		ServiceName: options.ServiceName,
		Disabled:    disabled,
		Tags:        tags,

		// Use constant sampling to sample every trace
		Sampler: &jaegerConfig.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},

		// Enable LogSpan to log every span via configured Logger
		Reporter: &jaegerConfig.ReporterConfig{
			LogSpans:            true,
			User:                options.User,
			Password:            options.Password,
			LocalAgentHostPort:  fmt.Sprintf("%s:%d", options.AgentHost, options.AgentPort),
			CollectorEndpoint:   options.Endpoint,
			BufferFlushInterval: time.Duration(options.BufferFlushInterval) * time.Second,
			QueueSize:           options.QueueSize,
		},
	}

	var configOpts []jaegerConfig.Option

	if options.Debug {
		configOpts = append(configOpts, jaegerConfig.Logger(&JaegerInternalLogger{logger: logger}))
	}

	tracer, _, err := cfg.NewTracer(configOpts...)
	if err != nil {
		stdout.Error(err)
		return nil
	}
	opentracing.SetGlobalTracer(tracer)
	return tracer
}

func NewJaegerTracer(options JaegerOptions, logger common.Logger, stdout *Stdout) *JaegerTracer {

	if logger == nil {
		logger = stdout
	}

	tracer := newJaegerTracer(options, logger, stdout)
	if tracer == nil {
		stdout.Debug("Jaeger tracer is disabled.")
		return nil
	}

	logger.Info("Jaeger tracer is up...")

	return &JaegerTracer{
		options:      options,
		callerOffset: 1,
		tracer:       tracer,
		logger:       logger,
	}
}
