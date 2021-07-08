package provider

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/devopsext/sre/common"
	"github.com/devopsext/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type DataDogOptions struct {
	ServiceName string
	Environment string
	Version     string
	Tags        string
	Debug       bool
}

type DataDogTracerOptions struct {
	DataDogOptions
	AgentHost string
	AgentPort int
}

type DataDogLoggerOptions struct {
	DataDogOptions
	AgentHost string
	AgentPort int
	Level     string
}

type DataDogMeterOptions struct {
	DataDogOptions
	AgentHost string
	AgentPort int
	Prefix    string
}

type DataDogTracerSpanContext struct {
	context ddtrace.SpanContext
}

type DataDogTracerSpan struct {
	span        ddtrace.Span
	spanContext *DataDogTracerSpanContext
	context     context.Context
	tracer      *DataDogTracer
}

type DataDogInternalLogger struct {
	logger common.Logger
}

type DataDogTracer struct {
	options      DataDogTracerOptions
	logger       common.Logger
	callerOffset int
}

type DataDogLogger struct {
	connection   *net.UDPConn
	stdout       *Stdout
	log          *logrus.Logger
	options      DataDogLoggerOptions
	callerOffset int
}

type DataDogCounter struct {
	meter       *DataDogMeter
	name        string
	description string
	labels      []string
	prefix      string
}

type DataDogMeter struct {
	options      DataDogMeterOptions
	logger       common.Logger
	callerOffset int
	client       *statsd.Client
}

func (ddsc DataDogTracerSpanContext) GetTraceID() string {

	if ddsc.context == nil {
		return ""
	}
	return fmt.Sprintf("%d", ddsc.context.TraceID())
}

func (ddsc DataDogTracerSpanContext) GetSpanID() string {

	if ddsc.context == nil {
		return ""
	}
	return fmt.Sprintf("%d", ddsc.context.SpanID())
}

func (dds DataDogTracerSpan) GetContext() common.TracerSpanContext {
	if dds.span == nil {
		return nil
	}

	if dds.spanContext != nil {
		return dds.spanContext
	}

	dds.spanContext = &DataDogTracerSpanContext{
		context: dds.span.Context(),
	}
	return dds.spanContext
}

func (dds DataDogTracerSpan) SetCarrier(object interface{}) common.TracerSpan {

	if dds.span == nil {
		return nil
	}

	if reflect.TypeOf(object) != reflect.TypeOf(http.Header{}) {
		dds.tracer.logger.Error(errors.New("other than http.Header is not supported yet"))
		return dds
	}

	var h http.Header = object.(http.Header)
	err := tracer.Inject(dds.span.Context(), tracer.HTTPHeadersCarrier(h))
	if err != nil {
		dds.tracer.logger.Error(err)
	}
	return dds
}

func (dds DataDogTracerSpan) SetName(name string) common.TracerSpan {

	if dds.span == nil {
		return nil
	}
	dds.span.SetOperationName(name)
	return dds
}

func (dds DataDogTracerSpan) SetTag(key string, value interface{}) common.TracerSpan {

	if dds.span == nil {
		return nil
	}
	dds.span.SetTag(key, value)
	return dds
}

func (dds DataDogTracerSpan) SetBaggageItem(restrictedKey, value string) common.TracerSpan {
	if dds.span == nil {
		return nil
	}
	dds.span.SetBaggageItem(restrictedKey, value)
	return dds
}

func (dds DataDogTracerSpan) Error(err error) common.TracerSpan {

	if dds.span == nil {
		return nil
	}

	dds.SetTag("error", true)
	return dds
}

func (dds DataDogTracerSpan) Finish() {
	if dds.span == nil {
		return
	}
	dds.span.Finish()
}

func (ddtl *DataDogInternalLogger) Log(msg string) {
	ddtl.logger.Info(msg)
}

func (dd *DataDogTracer) startSpanFromContext(ctx context.Context, offset int, opts ...tracer.StartSpanOption) (ddtrace.Span, context.Context) {

	operation, file, line := common.GetCallerInfo(offset)

	span, context := tracer.StartSpanFromContext(ctx, operation, opts...)
	if span != nil {
		span.SetTag("file", fmt.Sprintf("%s:%d", file, line))
	}
	return span, context
}

func (dd *DataDogTracer) startChildOfSpan(ctx context.Context, spanContext ddtrace.SpanContext) (ddtrace.Span, context.Context) {

	var span ddtrace.Span
	var context context.Context
	if spanContext != nil {
		span, context = dd.startSpanFromContext(ctx, dd.callerOffset+5, tracer.ChildOf(spanContext))
	} else {
		span, context = dd.startSpanFromContext(ctx, dd.callerOffset+5)
	}
	return span, context
}

func (dd *DataDogTracer) StartSpan() common.TracerSpan {

	s, ctx := dd.startSpanFromContext(context.Background(), dd.callerOffset+4)
	return DataDogTracerSpan{
		span:    s,
		context: ctx,
		tracer:  dd,
	}
}

func (dd *DataDogTracer) StartSpanWithTraceID(traceID string) common.TracerSpan {

	iTraceID, err := strconv.ParseInt(traceID, 10, 64)
	if err != nil {
		dd.logger.Error(err)
		return nil
	}

	s, ctx := dd.startSpanFromContext(context.Background(), dd.callerOffset+4,
		tracer.WithSpanID(uint64(iTraceID)),
	)
	return DataDogTracerSpan{
		span:    s,
		context: ctx,
		tracer:  dd,
	}
}

func (dd *DataDogTracer) getSpanContext(object interface{}) ddtrace.SpanContext {

	h, ok := object.(http.Header)
	if ok {
		spanContext, err := tracer.Extract(tracer.HTTPHeadersCarrier(h))
		if err != nil {
			dd.logger.Error(err)
			return nil
		}
		return spanContext
	}

	ddsc, ok := object.(*DataDogTracerSpanContext)
	if ok {
		return ddsc.context
	}
	return nil
}

func (dd *DataDogTracer) StartChildSpan(object interface{}) common.TracerSpan {

	spanContext := dd.getSpanContext(object)
	if spanContext == nil {
		return nil
	}

	s, ctx := dd.startChildOfSpan(context.Background(), spanContext)
	return DataDogTracerSpan{
		span:    s,
		context: ctx,
		tracer:  dd,
	}
}

func (dd *DataDogTracer) StartFollowSpan(object interface{}) common.TracerSpan {

	spanContext := dd.getSpanContext(object)
	if spanContext == nil {
		return nil
	}

	s, ctx := dd.startChildOfSpan(context.Background(), spanContext)
	return DataDogTracerSpan{
		span:    s,
		context: ctx,
		tracer:  dd,
	}
}

func (dd *DataDogTracer) SetCallerOffset(offset int) {
	dd.callerOffset = offset
}

func (dd *DataDogTracer) Stop() {
	tracer.Stop()
}

func startDataDogTracer(options DataDogTracerOptions, logger common.Logger) bool {

	disabled := utils.IsEmpty(options.AgentHost)
	if disabled {
		return false
	}

	addr := net.JoinHostPort(
		options.AgentHost,
		strconv.Itoa(options.AgentPort),
	)

	var opts []tracer.StartOption
	opts = append(opts, tracer.WithAgentAddr(addr))
	opts = append(opts, tracer.WithServiceName(options.ServiceName))
	opts = append(opts, tracer.WithServiceVersion(options.Version))
	opts = append(opts, tracer.WithEnv(options.Environment))

	if options.Debug {
		opts = append(opts, tracer.WithLogger(&DataDogInternalLogger{logger: logger}))
	}

	opts = setDataDogTracerTags(opts, options.Tags)

	tracer.Start(opts...)
	return true
}

func NewDataDogTracer(options DataDogTracerOptions, logger common.Logger, stdout *Stdout) *DataDogTracer {

	if logger == nil {
		logger = stdout
	}

	enabled := startDataDogTracer(options, logger)
	if !enabled {
		stdout.Debug("DataDog tracer is disabled.")
		return nil
	}

	logger.Info("DataDog tracer is up...")

	return &DataDogTracer{
		options:      options,
		callerOffset: 1,
		logger:       logger,
	}
}

func (dd *DataDogLogger) addSpanFields(span common.TracerSpan, fields logrus.Fields) logrus.Fields {

	if span == nil {
		return fields
	}

	ctx := span.GetContext()
	if ctx == nil {
		return fields
	}

	fields["dd.trace_id"] = ctx.GetTraceID()
	return fields
}

func (dd *DataDogLogger) Info(obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := dd.exists(logrus.InfoLevel, obj, args...); exists {
		dd.log.WithFields(fields).Infoln(message)
	}
	return dd
}

func (dd *DataDogLogger) SpanInfo(span common.TracerSpan, obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := dd.exists(logrus.InfoLevel, obj, args...); exists {
		fields = dd.addSpanFields(span, fields)
		dd.log.WithFields(fields).Infoln(message)
	}
	return dd
}

func (dd *DataDogLogger) Warn(obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := dd.exists(logrus.WarnLevel, obj, args...); exists {
		dd.log.WithFields(fields).Warnln(message)
	}
	return dd
}

func (dd *DataDogLogger) SpanWarn(span common.TracerSpan, obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := dd.exists(logrus.WarnLevel, obj, args...); exists {
		fields = dd.addSpanFields(span, fields)
		dd.log.WithFields(fields).Warnln(message)
	}
	return dd
}

func (dd *DataDogLogger) Error(obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := dd.exists(logrus.ErrorLevel, obj, args...); exists {
		dd.log.WithFields(fields).Errorln(message)
	}
	return dd
}

func (dd *DataDogLogger) SpanError(span common.TracerSpan, obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := dd.exists(logrus.ErrorLevel, obj, args...); exists {
		fields = dd.addSpanFields(span, fields)
		dd.log.WithFields(fields).Errorln(message)
	}
	return dd
}

func (dd *DataDogLogger) Debug(obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := dd.exists(logrus.DebugLevel, obj, args...); exists {
		dd.log.WithFields(fields).Debugln(message)
	}
	return dd
}

func (dd *DataDogLogger) SpanDebug(span common.TracerSpan, obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := dd.exists(logrus.DebugLevel, obj, args...); exists {
		fields = dd.addSpanFields(span, fields)
		dd.log.WithFields(fields).Debugln(message)
	}
	return dd
}

func (dd *DataDogLogger) Panic(obj interface{}, args ...interface{}) {

	if exists, fields, message := dd.exists(logrus.PanicLevel, obj, args...); exists {
		dd.log.WithFields(fields).Panicln(message)
	}
}

func (dd *DataDogLogger) SpanPanic(span common.TracerSpan, obj interface{}, args ...interface{}) {

	if exists, fields, message := dd.exists(logrus.PanicLevel, obj, args...); exists {
		fields = dd.addSpanFields(span, fields)
		dd.log.WithFields(fields).Panicln(message)
	}
}

func (dd *DataDogLogger) Stack(offset int) common.Logger {
	dd.callerOffset = dd.callerOffset - offset
	return dd
}

func (dd *DataDogLogger) exists(level logrus.Level, obj interface{}, args ...interface{}) (bool, logrus.Fields, string) {

	message := ""

	switch v := obj.(type) {
	case error:
		message = v.Error()
	case string:
		message = v
	default:
		message = "not implemented"
	}

	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	if utils.IsEmpty(message) && !dd.log.IsLevelEnabled(level) {
		return false, nil, ""
	}

	function, file, line := common.GetCallerInfo(dd.callerOffset + 5)
	fields := logrus.Fields{
		"file":    fmt.Sprintf("%s:%d", file, line),
		"func":    function,
		"service": dd.options.ServiceName,
		"version": dd.options.Version,
		"env":     dd.options.Environment,
	}
	return true, fields, message
}

func setDataDogTracerTags(opts []tracer.StartOption, sTags string) []tracer.StartOption {

	env := utils.GetEnvironment()
	pairs := strings.Split(sTags, ",")

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

		tag := tracer.WithGlobalTag(k, v)
		opts = append(opts, tag)
	}
	return opts
}

func NewDataDogLogger(options DataDogLoggerOptions, logger common.Logger, stdout *Stdout) *DataDogLogger {

	if logger == nil {
		logger = stdout
	}

	if utils.IsEmpty(options.AgentHost) {
		stdout.Debug("DataDog logger is disabled.")
		return nil
	}

	address := fmt.Sprintf("%s:%d", options.AgentHost, options.AgentPort)
	serverAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		stdout.Error(err)
		return nil
	}

	connection, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		stdout.Error(err)
		return nil
	}

	formatter := &logrus.JSONFormatter{}
	formatter.TimestampFormat = time.RFC3339Nano

	log := logrus.New()
	log.SetFormatter(formatter)

	switch options.Level {
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	case "panic":
		log.SetLevel(logrus.PanicLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	log.SetOutput(connection)

	logger.Info("DataDog logger is up...")

	return &DataDogLogger{
		connection:   connection,
		stdout:       stdout,
		log:          log,
		options:      options,
		callerOffset: 1,
	}
}

func (ddmc *DataDogCounter) getGlobalTags() []string {

	var tags []string

	for _, v := range strings.Split(ddmc.meter.options.Tags, ",") {
		tags = append(tags, strings.Replace(v, "=", ":", 1))
	}
	return tags
}

func (ddmc *DataDogCounter) getLabelTags(labelValues ...string) []string {

	var tags []string

	tags = append(tags, ddmc.getGlobalTags()...)
	tags = append(tags, fmt.Sprintf("dd.service:%s", ddmc.meter.options.ServiceName))
	tags = append(tags, fmt.Sprintf("dd.version:%s", ddmc.meter.options.Version))
	tags = append(tags, fmt.Sprintf("dd.env:%s", ddmc.meter.options.Environment))

	for k, v := range ddmc.labels {

		value := ""
		if len(labelValues) > (k - 1) {
			value = labelValues[k]
			tag := fmt.Sprintf("%s:%s", v, value)
			tags = append(tags, tag)
		}
	}
	return tags
}

func (ddmc *DataDogCounter) Inc(labelValues ...string) common.Counter {

	newName := ddmc.name
	if !utils.IsEmpty(ddmc.prefix) {
		newName = fmt.Sprintf("%s.%s", ddmc.prefix, newName)
	}

	newValues := ddmc.getLabelTags(labelValues...)
	_, file, line := common.GetCallerInfo(ddmc.meter.callerOffset + 3)
	newValues = append(newValues, fmt.Sprintf("file:%s", fmt.Sprintf("%s:%d", file, line)))

	err := ddmc.meter.client.Incr(newName, newValues, 1)
	if err != nil {
		ddmc.meter.logger.Error(err)
	}
	return ddmc
}

func (ddm *DataDogMeter) SetCallerOffset(offset int) {
	ddm.callerOffset = offset
}

func (ddm *DataDogMeter) Counter(name, description string, labels []string, prefixes ...string) common.Counter {

	var names []string

	if !utils.IsEmpty(ddm.options.Prefix) {
		names = append(names, ddm.options.Prefix)
	}

	if len(prefixes) > 0 {
		names = append(names, strings.Join(prefixes, "_"))
	}

	return &DataDogCounter{
		meter:       ddm,
		name:        name,
		description: description,
		labels:      labels,
		prefix:      strings.Join(names, "."),
	}
}

func (ddm *DataDogMeter) Stop() {
	// nothing here
}

func NewDataDogMeter(options DataDogMeterOptions, logger common.Logger, stdout *Stdout) *DataDogMeter {

	if logger == nil {
		logger = stdout
	}

	if utils.IsEmpty(options.AgentHost) {
		stdout.Debug("DataDog meter is disabled.")
		return nil
	}

	client, err := statsd.New(fmt.Sprintf("%s:%d", options.AgentHost, options.AgentPort))
	if err != nil {
		logger.Error(err)
		return nil
	}

	logger.Info("DataDog meter is up...")

	return &DataDogMeter{
		options:      options,
		logger:       logger,
		callerOffset: 1,
		client:       client,
	}
}
