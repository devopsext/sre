package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/devopsext/sre/common"
	"github.com/devopsext/utils"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type OpentelemetryOptions struct {
	ServiceName string
	Version     string
	Environment string
	Attributes  string
	Host        string
	Port        int
}

type OpentelemetryTracerOptions struct {
	OpentelemetryOptions
}

type OpentelemetryTracerSpanContext struct {
}

type OpentelemetryTracerSpan struct {
	span        trace.Span
	spanContext *OpentelemetryTracerSpanContext
	context     context.Context
	tracer      *OpentelemetryTracer
}

type OpentelemetryTracer struct {
	enabled      bool
	options      OpentelemetryTracerOptions
	logger       common.Logger
	tracer       trace.Tracer
	attributes   []attribute.KeyValue
	callerOffset int
}

func (ottsc OpentelemetryTracerSpanContext) GetTraceID() uint64 {

	/*if ottsc.context == nil {
		return 0
	}
	return ottsc.context.TraceID()*/
	return 0
}

func (ddsc OpentelemetryTracerSpanContext) GetSpanID() uint64 {

	/*if ddsc.context == nil {
		return 0
	}
	return ddsc.context.SpanID()*/
	return 0
}

func (otts OpentelemetryTracerSpan) GetContext() common.TracerSpanContext {
	if otts.span == nil {
		return nil
	}

	if otts.spanContext != nil {
		return otts.spanContext
	}

	otts.spanContext = &OpentelemetryTracerSpanContext{
		//context: otts.span.Context(),
	}
	return otts.spanContext
}

func (otts OpentelemetryTracerSpan) SetCarrier(object interface{}) common.TracerSpan {

	if otts.span == nil {
		return nil
	}

	if reflect.TypeOf(object) != reflect.TypeOf(http.Header{}) {
		otts.tracer.logger.Error(errors.New("Other than http.Header is not supported yet"))
		return otts
	}

	/*var h http.Header = object.(http.Header)
	err := tracer.Inject(otts.span.Context(), tracer.HTTPHeadersCarrier(h))
	if err != nil {
		dds.datadog.logger.Error(err)
	}*/
	return otts
}

func (otts OpentelemetryTracerSpan) SetName(name string) common.TracerSpan {

	if otts.span == nil {
		return nil
	}
	otts.span.SetName(name)
	return otts
}

func (otts OpentelemetryTracerSpan) SetTag(key string, value interface{}) common.TracerSpan {

	if otts.span == nil || value == nil {
		return nil
	}

	attr := attribute.Key(key)
	var v attribute.KeyValue

	switch value.(type) {
	case bool:
		v = attr.Bool(value.(bool))
	case int:
		v = attr.Int(value.(int))
	case int64:
		v = attr.Int64(value.(int64))
	case string:
		v = attr.String(value.(string))
	case float32:
		v = attr.Float64(value.(float64))
	case float64:
		v = attr.Float64(value.(float64))
	default:
		v = attr.String(value.(string))
	}

	otts.span.SetAttributes(v)
	return otts
}

func (otts OpentelemetryTracerSpan) SetBaggageItem(restrictedKey, value string) common.TracerSpan {
	if otts.span == nil {
		return nil
	}
	//otts.span.SetBaggageItem(restrictedKey, value)
	return otts
}

func (otts OpentelemetryTracerSpan) Error(err error) common.TracerSpan {
	if otts.span == nil {
		return nil
	}

	otts.span.RecordError(err)
	return otts
}

func (otts OpentelemetryTracerSpan) Finish() {
	if otts.span == nil {
		return
	}
	otts.span.End()
}

func (ott *OpentelemetryTracer) startSpanFromContext(ctx context.Context, offset int, opts ...trace.SpanStartOption) (trace.Span, context.Context) {

	operation, file, line := common.GetCallerInfo(offset)

	context, span := ott.tracer.Start(ctx, operation, opts...)
	if span != nil {
		fileKey := attribute.Key("file")
		span.SetAttributes(fileKey.String(fmt.Sprintf("%s:%d", file, line)))
	}
	return span, context
}

func (ott *OpentelemetryTracer) startChildOfSpan(ctx context.Context, spanContext *trace.SpanContext) (trace.Span, context.Context) {

	var span trace.Span
	var context context.Context

	if spanContext != nil {
		span, context = ott.startSpanFromContext(ctx, ott.callerOffset+5, trace.WithAttributes(ott.attributes...))
	} else {
		span, context = ott.startSpanFromContext(ctx, ott.callerOffset+5, trace.WithAttributes(ott.attributes...), trace.WithNewRoot())
	}
	return span, context
}

func (ott *OpentelemetryTracer) StartSpan() common.TracerSpan {

	s, ctx := ott.startSpanFromContext(context.Background(), ott.callerOffset+4,
		trace.WithAttributes(ott.attributes...),
		trace.WithNewRoot())

	return OpentelemetryTracerSpan{
		span:    s,
		context: ctx,
		tracer:  ott,
	}
}

func (ott *OpentelemetryTracer) StartSpanWithTraceID(traceID uint64) common.TracerSpan {

	sSpanID := fmt.Sprintf("%x", i)
	sTraceID := strconv.FormatInt(int64(traceID), 16)

	trace.

	traceSpanID, err := trace.SpanIDFromHex(sSpanID)
	if err != nil {
		ott.logger.Error(err)
	}

	traceTraceID, err := trace.TraceIDFromHex(sTraceID)
	if err != nil {
		ott.logger.Error(err)
	}

	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		SpanID:  traceSpanID,
		TraceID: traceTraceID,
	})

	parentCtx := trace.ContextWithSpanContext(context.Background(), spanContext)

	s, ctx := ott.startSpanFromContext(parentCtx, ott.callerOffset+4,
		trace.WithAttributes(ott.attributes...),
	)
	return OpentelemetryTracerSpan{
		span:    s,
		context: ctx,
		tracer:  ott,
	}
}

func (ott *OpentelemetryTracer) getOpentracingSpanContext(object interface{}) *trace.SpanContext {

	/*	h, ok := object.(http.Header)
		if ok {
			spanContext, err := trace.Extract(tracer.HTTPHeadersCarrier(h))
			if err != nil {
				ott.logger.Error(err)
				return nil
			}
			return spanContext
		}

		ottsc, ok := object.(*OpentelemetryTracerSpanContext)
		if ok {
			return ottsc.context
		}*/
	return nil
}

func (ott *OpentelemetryTracer) StartChildSpan(object interface{}) common.TracerSpan {

	spanContext := ott.getOpentracingSpanContext(object)
	if spanContext == nil {
		return nil
	}

	s, ctx := ott.startChildOfSpan(context.Background(), spanContext)
	return OpentelemetryTracerSpan{
		span:    s,
		context: ctx,
		tracer:  ott,
	}
}

func (ott *OpentelemetryTracer) StartFollowSpan(object interface{}) common.TracerSpan {

	spanContext := ott.getOpentracingSpanContext(object)
	if spanContext == nil {
		return nil
	}

	s, ctx := ott.startChildOfSpan(context.Background(), spanContext)
	return OpentelemetryTracerSpan{
		span:    s,
		context: ctx,
		tracer:  ott,
	}
}

func (ott *OpentelemetryTracer) SetCallerOffset(offset int) {
	ott.callerOffset = offset
}

func (ott *OpentelemetryTracer) Enabled() bool {
	return ott.enabled
}

func parseOpentelemetryAttrributes(sAttributes string) []attribute.KeyValue {

	env := utils.GetEnvironment()
	pairs := strings.Split(sAttributes, ",")
	attributes := make([]attribute.KeyValue, 0)
	for _, p := range pairs {

		if utils.IsEmpty(p) {
			continue
		}
		kv := strings.SplitN(p, "=", 2)
		k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])

		if strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") {
			ed := strings.SplitN(v[2:len(v)-1], ":", 2)
			e, d := ed[0], ed[1]
			v = env.Get(e, "").(string)
			if v == "" && d != "" {
				v = d
			}
		}

		attribute := attribute.String(k, v)
		attributes = append(attributes, attribute)
	}
	return attributes
}

func startOpentelemtryTracer(options OpentelemetryTracerOptions, logger common.Logger, stdout *Stdout) (trace.Tracer, bool) {

	disabled := utils.IsEmpty(options.Host)
	if disabled {
		stdout.Debug("Opentelemetry tracer is disabled.")
	}

	var err error
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(options.ServiceName),
			semconv.ServiceVersionKey.String(options.Version),
			semconv.DeploymentEnvironmentKey.String(options.Environment),
		),
	)
	if err != nil {
		stdout.Error(err)
		return nil, false
	}

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%d", options.Host, options.Port)),
	)
	if err != nil {
		stdout.Error(err)
		return nil, false
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	_, file, _ := common.GetCallerInfo(1)

	tracer := otel.Tracer(file)

	//defer func() { tracerProvider.Shutdown(ctx) }()

	return tracer, !disabled
}

func NewOpentelemetryTracer(options OpentelemetryTracerOptions, logger common.Logger, stdout *Stdout) *OpentelemetryTracer {

	tracer, enabled := startOpentelemtryTracer(options, logger, stdout)
	attributes := parseOpentelemetryAttrributes(options.Attributes)

	return &OpentelemetryTracer{
		enabled:      enabled,
		options:      options,
		logger:       logger,
		tracer:       tracer,
		attributes:   attributes,
		callerOffset: 1,
	}
}
