package provider

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	ddClient "github.com/DataDog/datadog-api-client-go/api/v1/datadog"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/devopsext/sre/common"
	utils "github.com/devopsext/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type DataDogOptions struct {
	ApiKey      string
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

type DataDogEventerOptions struct {
	DataDogOptions
	Site string
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
}

type DataDogGauge struct {
	meter       *DataDogMeter
	name        string
	description string
	labels      []string
}

type DataDogMeter struct {
	options      DataDogMeterOptions
	logger       common.Logger
	callerOffset int
	client       *statsd.Client
}

type DataDogEventer struct {
	options DataDogEventerOptions
	logger  common.Logger
	client  *ddClient.APIClient
	ctx     context.Context
	tags    []string
}

func (ddsc *DataDogTracerSpanContext) GetTraceID() string {

	if ddsc.context == nil {
		return ""
	}
	return common.TraceIDUint64ToHex(ddsc.context.TraceID())
}

func (ddsc *DataDogTracerSpanContext) GetSpanID() string {

	if ddsc.context == nil {
		return ""
	}
	return common.SpanIDUint64ToHex(ddsc.context.SpanID())
}

func (dds *DataDogTracerSpan) GetContext() common.TracerSpanContext {
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

func (dds *DataDogTracerSpan) SetCarrier(object interface{}) common.TracerSpan {

	if dds.span == nil {
		return nil
	}

	h, ok := object.(http.Header)
	if ok {
		err := tracer.Inject(dds.span.Context(), tracer.HTTPHeadersCarrier(h))
		if err != nil {
			dds.tracer.logger.Error(err)
		}
	}
	return dds
}

func (dds *DataDogTracerSpan) SetName(name string) common.TracerSpan {

	if dds.span == nil {
		return nil
	}
	dds.span.SetOperationName(name)
	return dds
}

func (dds *DataDogTracerSpan) SetTag(key string, value interface{}) common.TracerSpan {

	if dds.span == nil {
		return nil
	}
	dds.span.SetTag(key, value)
	return dds
}

func (dds *DataDogTracerSpan) SetBaggageItem(restrictedKey, value string) common.TracerSpan {
	if dds.span == nil {
		return nil
	}
	dds.span.SetBaggageItem(restrictedKey, value)
	return dds
}

func (dds *DataDogTracerSpan) Error(err error) common.TracerSpan {

	if dds.span == nil {
		return nil
	}

	dds.SetTag("error", true)
	return dds
}

func (dds *DataDogTracerSpan) Finish() {
	if dds.span == nil {
		return
	}
	dds.span.Finish()
}

func (ddtl *DataDogInternalLogger) Log(msg string) {
	ddtl.logger.Info(msg)
}

func (dd *DataDogTracer) startSpanFromContext(ctx context.Context, offset int, opts ...tracer.StartSpanOption) (ddtrace.Span, context.Context) {

	operation, file, line := utils.CallerGetInfo(offset)

	span, sContext := tracer.StartSpanFromContext(ctx, operation, opts...)
	if span != nil {
		span.SetTag("file", fmt.Sprintf("%s:%d", file, line))
	}
	return span, sContext
}

func (dd *DataDogTracer) startChildOfSpan(ctx context.Context, spanContext ddtrace.SpanContext) (ddtrace.Span, context.Context) {

	var span ddtrace.Span
	var sContext context.Context
	if spanContext != nil {
		span, sContext = dd.startSpanFromContext(ctx, dd.callerOffset+5, tracer.ChildOf(spanContext))
	} else {
		span, sContext = dd.startSpanFromContext(ctx, dd.callerOffset+5)
	}
	return span, sContext
}

func (dd *DataDogTracer) StartSpan() common.TracerSpan {

	s, ctx := dd.startSpanFromContext(context.Background(), dd.callerOffset+4)
	return &DataDogTracerSpan{
		span:    s,
		context: ctx,
		tracer:  dd,
	}
}

func (dd *DataDogTracer) StartSpanWithTraceID(traceID, spanID string) common.TracerSpan {

	tID := common.TraceIDHexToUint64(traceID)
	if tID == 0 {
		dd.logger.Error(errors.New("invalid trace ID"))
		return nil
	}

	sID := common.SpanIDHexToUint64(spanID)
	if sID == 0 {
		sID = tID
	}

	s, ctx := dd.startSpanFromContext(context.Background(), dd.callerOffset+4,
		tracer.WithSpanID(sID),
		tracer.WithTraceID(tID),
	)
	return &DataDogTracerSpan{
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
	return &DataDogTracerSpan{
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
	return &DataDogTracerSpan{
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

	m := utils.MapGetKeyValues(options.Tags)
	for k, v := range m {
		tag := tracer.WithGlobalTag(k, v)
		opts = append(opts, tag)
	}

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

	// we need to put int64 for DataDog to have it properly correlated with traces
	fields["dd.trace_id"] = common.TraceIDHexToUint64(ctx.GetTraceID())
	fields["dd.span_id"] = common.SpanIDHexToUint64(ctx.GetSpanID())

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

	function, file, line := utils.CallerGetInfo(dd.callerOffset + 5)
	fields := logrus.Fields{
		"file":    fmt.Sprintf("%s:%d", file, line),
		"func":    function,
		"service": dd.options.ServiceName,
		"version": dd.options.Version,
		"env":     dd.options.Environment,
	}

	m := utils.MapGetKeyValues(dd.options.Tags)
	for k, v := range m {
		fields[k] = v
	}

	return true, fields, message
}

func (dd *DataDogLogger) Stop() {
	if dd.connection != nil {
		dd.connection.Close()
	}
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

func (ddm *DataDogMeter) getGlobalTags() []string {

	var tags []string

	for _, v := range strings.Split(ddm.options.Tags, ",") {
		tags = append(tags, strings.Replace(v, "=", ":", 1))
	}
	return tags
}

func (ddm *DataDogMeter) getLabelTags(labelValues ...string) []string {

	var tags []string

	tags = append(tags, ddm.getGlobalTags()...)
	tags = append(tags, fmt.Sprintf("dd.service:%s", ddm.options.ServiceName))
	tags = append(tags, fmt.Sprintf("dd.version:%s", ddm.options.Version))
	tags = append(tags, fmt.Sprintf("dd.env:%s", ddm.options.Environment))

	return tags
}

func (ddm *DataDogMeter) SetCallerOffset(offset int) {
	ddm.callerOffset = offset
}

func (ddmc *DataDogCounter) Inc() common.Counter {

	/*newValues := ddmc.meter.getLabelTags(labelValues...)
	for k, v := range ddmc.labels {

		value := ""
		if len(labelValues) > (k - 1) {
			value = labelValues[k]
			tag := fmt.Sprintf("%s:%s", v, value)
			newValues = append(newValues, tag)
		}
	}

	_, file, line := utils.CallerGetInfo(ddmc.meter.callerOffset + 3)
	newValues = append(newValues, fmt.Sprintf("file:%s", fmt.Sprintf("%s:%d", file, line)))

	err := ddmc.meter.client.Incr(ddmc.name, newValues, 1)
	if err != nil {
		ddmc.meter.logger.Error(err)
	}
	return ddmc*/
	return nil
}

func (ddm *DataDogMeter) Counter(name, description string, labels common.Labels, prefixes ...string) common.Counter {

	/*var names []string

	if !utils.IsEmpty(ddm.options.Prefix) {
		names = append(names, ddm.options.Prefix)
	}

	if len(prefixes) > 0 {
		names = append(names, strings.Join(prefixes, "_"))
	}

	newName := name
	if len(names) > 0 {
		newName = fmt.Sprintf("%s.%s", strings.Join(names, "."), newName)
	}

	return &DataDogCounter{
		meter:       ddm,
		name:        newName,
		description: description,
		labels:      labels,
	}*/
	return nil
}

func (ddmg *DataDogGauge) Set(value float64) common.Gauge {

	/*newValues := ddmg.meter.getLabelTags(labelValues...)
	for k, v := range ddmg.labels {

		value := ""
		if len(labelValues) > (k - 1) {
			value = labelValues[k]
			tag := fmt.Sprintf("%s:%s", v, value)
			newValues = append(newValues, tag)
		}
	}

	_, file, line := utils.CallerGetInfo(ddmg.meter.callerOffset + 3)
	newValues = append(newValues, fmt.Sprintf("file:%s", fmt.Sprintf("%s:%d", file, line)))

	err := ddmg.meter.client.Gauge(ddmg.name, value, newValues, value)
	if err != nil {
		ddmg.meter.logger.Error(err)
	}
	return ddmg*/
	return nil
}

func (ddm *DataDogMeter) Gauge(name, description string, labels common.Labels, prefixes ...string) common.Gauge {

	/*var names []string

	if !utils.IsEmpty(ddm.options.Prefix) {
		names = append(names, ddm.options.Prefix)
	}

	if len(prefixes) > 0 {
		names = append(names, strings.Join(prefixes, "_"))
	}

	newName := name
	if len(names) > 0 {
		newName = fmt.Sprintf("%s.%s", strings.Join(names, "."), newName)
	}

	return &DataDogGauge{
		meter:       ddm,
		name:        newName,
		description: description,
		labels:      labels,
	}*/
	return nil
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

func (dde *DataDogEventer) Interval(name string, attributes map[string]string, begin, end time.Time) error {

	dateHappened := begin.UTC().Unix()

	body := ddClient.EventCreateRequest{
		AggregationKey: nil,
		AlertType:      nil,
		DateHappened:   &dateHappened,
		DeviceName:     nil,
		Host:           nil,
		Id:             nil,
		Payload:        nil,
		Priority:       nil,
		RelatedEventId: nil,
		SourceTypeName: nil,
		Tags:           &dde.tags,
		Text:           name,
		Title:          name,
		Url:            nil,
		UnparsedObject: nil,
	}

	tags := dde.tags

	for key, attr := range attributes {
		switch key {
		case "text":
			body.SetText(attr)
		case "aggregation_key":
			body.SetAggregationKey(attr)
		case "alert_type":
			body.SetAlertType(ddClient.EventAlertType(attr))
		case "device_name":
			body.SetDeviceName(attr)
		case "host":
			body.SetHost(attr)
		case "priority":
			body.SetPriority(ddClient.EventPriority(attr))
		case "related_event_id":
			relatedEventId, err := strconv.ParseInt(attr, 10, 64)
			if err != nil {
				dde.logger.Warn("wrong related_event_id format")
			} else {
				body.SetRelatedEventId(relatedEventId)
			}
		case "source_type_name":
			body.SetSourceTypeName(attr)
		default:
			tags = append(tags, fmt.Sprintf("%s:%s", key, attr))
		}
	}

	body.SetTags(tags)

	resp, r, err := dde.client.EventsApi.CreateEvent(dde.ctx, body)

	if err != nil {
		dde.logger.Error(err)
		dde.logger.Error("Full HTTP response:", r)
		return err
	}
	dde.logger.Debug(fmt.Sprintf("%v", resp))
	return nil
}

func (dde *DataDogEventer) Now(name string, attributes map[string]string) error {
	return dde.At(name, attributes, time.Now())
}

func (dde *DataDogEventer) At(name string, attributes map[string]string, when time.Time) error {
	return dde.Interval(name, attributes, when, when)
}

func (dde *DataDogEventer) Stop() {
	dde.logger.Info("DataDog eventer is stopped.")
}

func NewDataDogEventer(options DataDogEventerOptions, logger common.Logger, stdout *Stdout) *DataDogEventer {

	if logger == nil {
		logger = stdout
	}

	if utils.IsEmpty(options.Site) {
		stdout.Debug("DataDog eventer is disabled.")
		return nil
	}

	configuration := ddClient.NewConfiguration()

	ctx := context.WithValue(
		context.Background(),
		ddClient.ContextAPIKeys,
		map[string]ddClient.APIKey{
			"apiKeyAuth": {
				Key: options.ApiKey,
			},
			"appKeyAuth": {
				Key: options.ApiKey,
			},
		},
	)

	ctx = context.WithValue(ctx,
		ddClient.ContextServerVariables,
		map[string]string{
			"site": options.Site, // "datadoghq.eu"
		})

	logger.Info("DataDog eventer is up...")

	return &DataDogEventer{
		options: options,
		logger:  logger,
		tags:    utils.MapToArrayWithSeparator(utils.MapGetKeyValues(options.Tags), ":"),
		client:  ddClient.NewAPIClient(configuration),
		ctx:     ctx,
	}
}
