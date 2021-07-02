package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/devopsext/sre/common"
	"github.com/devopsext/utils"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
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
}

const headerTraceID string = "X-Trace-ID"

type OpentelemetryTracerOptions struct {
	OpentelemetryOptions
	AgentHost     string
	AgentPort     int
	HeaderTraceID string
}

type OpentelemetryMeterOptions struct {
	OpentelemetryOptions
	AgentHost     string
	AgentPort     int
	Prefix        string
	CollectPeriod int64
}

type OpentelemetryTracerSpanContext struct {
	tracerSpan OpentelemetryTracerSpan
	context    *trace.SpanContext
}

type OpentelemetryTracerSpan struct {
	span              trace.Span
	tracerSpanContext *OpentelemetryTracerSpanContext
	context           context.Context
	tracer            *OpentelemetryTracer
}

type OpentelemetryTracer struct {
	options      OpentelemetryTracerOptions
	logger       common.Logger
	tracer       trace.Tracer
	provider     *sdktrace.TracerProvider
	attributes   []attribute.KeyValue
	callerOffset int
}

type OpentelemetryCounter struct {
	meter   *OpentelemetryMeter
	counter *metric.Int64Counter
	labels  []string
}

type OpentelemetryMeter struct {
	options      OpentelemetryMeterOptions
	logger       common.Logger
	meter        *metric.Meter
	controller   *controller.Controller
	exporter     *otlpmetric.Exporter
	attributes   []attribute.KeyValue
	callerOffset int
}

func (ottsc OpentelemetryTracerSpanContext) GetTraceID() uint64 {

	if ottsc.context == nil || !ottsc.context.HasTraceID() {
		return 0
	}

	traceID := ottsc.context.TraceID()
	sTraceID := traceID.String()
	dec, err := strconv.ParseInt(sTraceID, 16, 32)
	if err != nil {
		ottsc.tracerSpan.tracer.logger.Error(err)
		return 0
	}
	return uint64(dec)
}

func (ottsc OpentelemetryTracerSpanContext) GetSpanID() uint64 {

	if ottsc.context == nil || !ottsc.context.HasSpanID() {
		return 0
	}

	spanID := ottsc.context.SpanID()
	sSpanID := spanID.String()
	dec, err := strconv.ParseInt(sSpanID, 16, 32)
	if err != nil {
		ottsc.tracerSpan.tracer.logger.Error(err)
		return 0
	}
	return uint64(dec)
}

func (otts OpentelemetryTracerSpan) GetContext() common.TracerSpanContext {
	if otts.span == nil {
		return nil
	}

	if otts.tracerSpanContext != nil {
		return otts.tracerSpanContext
	}

	context := otts.span.SpanContext()
	otts.tracerSpanContext = &OpentelemetryTracerSpanContext{
		context:    &context,
		tracerSpan: otts,
	}
	return otts.tracerSpanContext
}

func (otts OpentelemetryTracerSpan) SetCarrier(object interface{}) common.TracerSpan {

	if otts.span == nil {
		return nil
	}

	if reflect.TypeOf(object) != reflect.TypeOf(http.Header{}) {
		otts.tracer.logger.Error(errors.New("other than http.Header is not supported yet"))
		return otts
	}

	// take a look at TextMapPropagator instead
	// otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))

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
	// may be SetBaggage should be replaceb by AddEvent
	otts.span.AddEvent("",
		trace.WithAttributes(
			attribute.String("event", "baggage"),
			attribute.String("key", restrictedKey),
			attribute.String("value", value)))
	return otts
}

func (otts OpentelemetryTracerSpan) Error(err error) common.TracerSpan {
	if otts.span == nil {
		return nil
	}
	otts.span.SetStatus(codes.Error, "")
	otts.span.AddEvent("",
		trace.WithAttributes(
			attribute.String("error.message", err.Error())))
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
		trace.WithNewRoot(),
	)

	return OpentelemetryTracerSpan{
		span:    s,
		context: ctx,
		tracer:  ott,
	}
}

func (ott *OpentelemetryTracer) getSpanTraceID(spanID, traceID uint64) (*trace.SpanID, *trace.TraceID) {

	sSpanID := fmt.Sprintf("%016x", spanID)
	traceSpanID, err := trace.SpanIDFromHex(sSpanID)
	if err != nil {
		ott.logger.Error(err)
		return nil, nil
	}

	sTraceID := fmt.Sprintf("%032x", traceID)
	traceTraceID, err := trace.TraceIDFromHex(sTraceID)
	if err != nil {
		ott.logger.Error(err)
		return &traceSpanID, nil
	}

	return &traceSpanID, &traceTraceID
}

func (ott *OpentelemetryTracer) StartSpanWithTraceID(traceID uint64) common.TracerSpan {

	parentCtx := context.Background()

	var opts []trace.SpanStartOption
	opts = append(opts, trace.WithAttributes(ott.attributes...))
	opts = append(opts, trace.WithNewRoot())

	traceSpanID, traceTraceID := ott.getSpanTraceID(traceID, traceID)
	if traceSpanID != nil && traceTraceID != nil {
		spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
			SpanID:  *traceSpanID,
			TraceID: *traceTraceID,
		})
		parentCtx = trace.ContextWithSpanContext(context.Background(), spanCtx)
	} else {
		opts = append(opts, trace.WithNewRoot())
	}

	s, ctx := ott.startSpanFromContext(parentCtx, ott.callerOffset+4, opts...)
	return OpentelemetryTracerSpan{
		span:    s,
		context: ctx,
		tracer:  ott,
	}
}

func (ott *OpentelemetryTracer) getSpanContext(object interface{}) (context.Context, *trace.SpanContext) {

	h, ok := object.(http.Header)
	if ok {

		// take a look at TextMapPropagator instead
		//ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		headerName := ott.options.HeaderTraceID
		if utils.IsEmpty(headerName) {
			headerName = headerTraceID
		}

		s := h.Get(headerName)
		if utils.IsEmpty(s) {
			ott.logger.Error("Opentelemetry header is not found")
			return nil, nil
		}

		traceID, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			ott.logger.Error(err)
			return nil, nil
		}

		traceSpanID, traceTraceID := ott.getSpanTraceID(0, uint64(traceID))
		if traceSpanID != nil && traceTraceID != nil {
			ott.logger.Error(err)
			return nil, nil
		}

		spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
			SpanID:  *traceSpanID,
			TraceID: *traceTraceID,
		})
		parentCtx := trace.ContextWithRemoteSpanContext(context.Background(), spanCtx)
		return parentCtx, &spanCtx
	}

	ottsc, ok := object.(*OpentelemetryTracerSpanContext)
	if ok {
		return ottsc.tracerSpan.context, ottsc.context
	}
	return nil, nil
}

func (ott *OpentelemetryTracer) StartChildSpan(object interface{}) common.TracerSpan {

	parentCtx, spanContext := ott.getSpanContext(object)
	if spanContext == nil {
		return nil
	}

	s, ctx := ott.startChildOfSpan(parentCtx, spanContext)
	return OpentelemetryTracerSpan{
		span:    s,
		context: ctx,
		tracer:  ott,
	}
}

func (ott *OpentelemetryTracer) StartFollowSpan(object interface{}) common.TracerSpan {

	parentCtx, spanContext := ott.getSpanContext(object)
	if spanContext == nil {
		return nil
	}

	s, ctx := ott.startChildOfSpan(parentCtx, spanContext)
	return OpentelemetryTracerSpan{
		span:    s,
		context: ctx,
		tracer:  ott,
	}
}

func (ott *OpentelemetryTracer) SetCallerOffset(offset int) {
	ott.callerOffset = offset
}

func (ott *OpentelemetryTracer) Stop() {

	ctx := context.Background()

	if ott.provider != nil {
		ott.provider.Shutdown(ctx)
	}
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

func startOpentelemtryTracer(options OpentelemetryTracerOptions, logger common.Logger, stdout *Stdout) (trace.Tracer, *sdktrace.TracerProvider) {

	disabled := utils.IsEmpty(options.AgentHost)
	if disabled {
		return nil, nil
	}

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
		return nil, nil
	}

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%d", options.AgentHost, options.AgentPort)),
	)
	if err != nil {
		stdout.Error(err)
		return nil, nil
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

	return tracer, tracerProvider
}

func NewOpentelemetryTracer(options OpentelemetryTracerOptions, logger common.Logger, stdout *Stdout) *OpentelemetryTracer {

	tracer, provider := startOpentelemtryTracer(options, logger, stdout)
	if tracer == nil {
		stdout.Debug("Opentelemetry tracer is disabled.")
		return nil
	}

	attributes := parseOpentelemetryAttrributes(options.Attributes)

	logger.Info("Opentelemetry tracer is up...")

	return &OpentelemetryTracer{
		options:      options,
		logger:       logger,
		tracer:       tracer,
		provider:     provider,
		attributes:   attributes,
		callerOffset: 1,
	}
}

func (otc *OpentelemetryCounter) getGlobalTags(labelValues ...string) []attribute.KeyValue {

	var labels []attribute.KeyValue
	l := len(labelValues)

	for index, name := range otc.labels {

		if l > index {
			value := attribute.String(name, labelValues[index])
			labels = append(labels, value)
		}
	}
	return labels
}

func (otc *OpentelemetryCounter) Inc(labelValues ...string) common.Counter {

	labels := otc.getGlobalTags(labelValues...)
	_, file, line := common.GetCallerInfo(otc.meter.callerOffset + 3)
	labels = append(labels, attribute.String("file", fmt.Sprintf("%s:%d", file, line)))

	otc.counter.Add(context.Background(), 1, labels...)
	return otc
}

func (otm *OpentelemetryMeter) Counter(name, description string, labels []string, prefixes ...string) common.Counter {

	var names []string

	if !utils.IsEmpty(otm.options.Prefix) {
		names = append(names, otm.options.Prefix)
	}

	names = append(names, prefixes...)
	names = append(names, name)
	newName := strings.Join(names, ".")

	counter := metric.Must(*otm.meter).NewInt64Counter(newName, metric.WithDescription(description))
	counter.Bind(otm.attributes...)

	return &OpentelemetryCounter{
		meter:   otm,
		counter: &counter,
		labels:  labels,
	}
}

func (otm *OpentelemetryMeter) SetCallerOffset(offset int) {
	otm.callerOffset = offset
}

func (otm *OpentelemetryMeter) Stop() {

	ctx := context.Background()
	if otm.controller != nil {
		otm.controller.Stop(ctx)
	}
	if otm.exporter != nil {
		otm.exporter.Shutdown(ctx)
	}
}

func startOpentelemetryMeter(options OpentelemetryMeterOptions, stdout *Stdout) (*metric.Meter, *controller.Controller, *otlpmetric.Exporter) {

	if utils.IsEmpty(options.AgentHost) {
		return nil, nil, nil
	}

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
		return nil, nil, nil
	}

	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(fmt.Sprintf("%s:%d", options.AgentHost, options.AgentPort)),
	)
	if err != nil {
		stdout.Error(err)
		return nil, nil, nil
	}

	// it can be used for internal logging
	/*stdoutExporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	if err != nil {
		stdout.Error(err)
		return nil, nil, nil
	}*/

	collectPeriod := options.CollectPeriod
	if collectPeriod == 0 {
		collectPeriod = 1000
	}

	cont := controller.New(
		processor.New(
			simple.NewWithExactDistribution(),
			metricExporter,
		),
		controller.WithCollectPeriod(time.Duration(collectPeriod)*time.Millisecond),
		controller.WithExporter(metricExporter),
		//controller.WithExporter(stdoutExporter),
		controller.WithResource(res),
	)

	err = cont.Start(context.Background())
	if err != nil {
		stdout.Error(err)
		return nil, nil, nil
	}
	global.SetMeterProvider(cont.MeterProvider())

	_, file, _ := common.GetCallerInfo(1)
	meter := global.Meter(file)

	return &meter, cont, metricExporter
}

func NewOpentelemetryMeter(options OpentelemetryMeterOptions, logger common.Logger, stdout *Stdout) *OpentelemetryMeter {

	meter, controller, exporter := startOpentelemetryMeter(options, stdout)
	if meter == nil {
		stdout.Debug("Opentelemetry meter is disabled.")
		return nil
	}

	attributes := parseOpentelemetryAttrributes(options.Attributes)

	return &OpentelemetryMeter{
		options:      options,
		logger:       logger,
		meter:        meter,
		controller:   controller,
		exporter:     exporter,
		attributes:   attributes,
		callerOffset: 1,
	}
}
