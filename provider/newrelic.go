package provider

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/devopsext/sre/common"
	utils "github.com/devopsext/utils"
	telemetry "github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
	"github.com/sirupsen/logrus"
)

type NewRelicOptions struct {
	ApiKey      string
	ServiceName string
	Environment string
	Version     string
	Attributes  string
	Debug       bool
}

type NewRelicTracerOptions struct {
	NewRelicOptions
	Endpoint      string
	HeaderTraceID string
}

type NewRelicLoggerOptions struct {
	NewRelicOptions
	Endpoint  string
	AgentHost string
	AgentPort int
	Level     string
}

type NewRelicMeterOptions struct {
	NewRelicOptions
	Endpoint string
	Prefix   string
}

type NewRelicEventerOptions struct {
	NewRelicOptions
	Endpoint string
}

type NewRelicTracerSpanContext struct {
	tracerSpan *NewRelicTracerSpan
}

type NewRelicTracerSpan struct {
	tracer            *NewRelicTracer
	traceID           string
	spanID            string
	parentID          string
	operation         string
	timestamp         time.Time
	attributes        map[string]interface{}
	events            []telemetry.Event
	tracerSpanContext *NewRelicTracerSpanContext
}

type NewRelicTracer struct {
	options      NewRelicTracerOptions
	harvester    *telemetry.Harvester
	logger       common.Logger
	callerOffset int
}

type NewRelicLogger struct {
	harvester    *telemetry.Harvester
	connection   *net.TCPConn
	stdout       *Stdout
	log          *logrus.Logger
	options      NewRelicLoggerOptions
	callerOffset int
}

type NewRelicCounter struct {
	meter       *NewRelicMeter
	name        string
	description string
	labels      []string
}

type NewRelicGauge struct {
	meter       *NewRelicMeter
	name        string
	description string
	labels      []string
}

type NewRelicMeter struct {
	harvester    *telemetry.Harvester
	options      NewRelicMeterOptions
	logger       common.Logger
	callerOffset int
}

type NewRelicEventer struct {
	harvester  *telemetry.Harvester
	options    NewRelicEventerOptions
	logger     common.Logger
	attributes map[string]interface{}
}

func (nrtsc *NewRelicTracerSpanContext) GetTraceID() string {

	return nrtsc.tracerSpan.traceID
}

func (nrtsc *NewRelicTracerSpanContext) GetSpanID() string {

	return nrtsc.tracerSpan.spanID
}

func (nrts *NewRelicTracerSpan) GetContext() common.TracerSpanContext {

	if nrts.tracerSpanContext != nil {
		return nrts.tracerSpanContext
	}

	nrts.tracerSpanContext = &NewRelicTracerSpanContext{
		tracerSpan: nrts,
	}
	return nrts.tracerSpanContext
}

func (nrts *NewRelicTracerSpan) SetCarrier(object interface{}) common.TracerSpan {

	/*if nrts.span == nil {
		return nil
	}*/

	h, ok := object.(http.Header)
	if ok {

		ctx := nrts.GetContext()
		if ctx != nil {
			name := http.CanonicalHeaderKey(headerTraceID)
			if len(h[name]) == 0 {
				h[name] = append(h[name], ctx.GetTraceID())
			}
		}
		if ctx != nil {
			name := http.CanonicalHeaderKey(headerSpanID)
			if len(h[name]) == 0 {
				h[name] = append(h[name], ctx.GetTraceID())
			}
		}
	}
	return nrts
}

func (nrts *NewRelicTracerSpan) SetName(name string) common.TracerSpan {

	nrts.operation = name
	return nrts
}

func (nrts *NewRelicTracerSpan) SetTag(key string, value interface{}) common.TracerSpan {

	nrts.attributes[key] = value
	return nrts
}

func (nrts *NewRelicTracerSpan) SetBaggageItem(restrictedKey, value string) common.TracerSpan {

	attributes := make(map[string]interface{})
	attributes["event"] = "baggage"
	attributes["key"] = restrictedKey
	attributes["value"] = value

	nrts.events = append(nrts.events, telemetry.Event{
		EventType:  "baggage",
		Attributes: attributes,
	})

	return nrts
}

func (nrts *NewRelicTracerSpan) Error(err error) common.TracerSpan {

	nrts.attributes["status"] = "error"

	attributes := make(map[string]interface{})
	attributes["error.message"] = err.Error()

	nrts.events = append(nrts.events, telemetry.Event{
		EventType:  "error",
		Attributes: attributes,
	})
	return nrts
}

func (nrts *NewRelicTracerSpan) Finish() {

	span := telemetry.Span{
		TraceID:     nrts.traceID,
		ID:          nrts.spanID,
		ParentID:    nrts.parentID,
		Name:        nrts.operation,
		Timestamp:   nrts.timestamp,
		Duration:    time.Second,
		ServiceName: nrts.tracer.options.ServiceName,
		Attributes:  nrts.attributes,
		Events:      nrts.events,
	}

	err := nrts.tracer.harvester.RecordSpan(span)
	if err != nil {
		nrts.tracer.logger.Error(err)
	}
}

func (nrt *NewRelicTracer) getSpanAttributes() (string, map[string]interface{}) {

	operation, file, line := utils.CallerGetInfo(nrt.callerOffset + 4)

	attribites := make(map[string]interface{})
	attribites["file"] = fmt.Sprintf("%s:%d", file, line)

	return operation, attribites
}

func (nrt *NewRelicTracer) StartSpan() common.TracerSpan {

	operation, attributes := nrt.getSpanAttributes()

	return &NewRelicTracerSpan{
		traceID:    common.NewTraceID(),
		spanID:     common.NewSpanID(),
		operation:  operation,
		timestamp:  time.Now(),
		attributes: attributes,
		tracer:     nrt,
	}
}

func (nrt *NewRelicTracer) StartSpanWithTraceID(traceID, spanID string) common.TracerSpan {

	operation, attributes := nrt.getSpanAttributes()

	return &NewRelicTracerSpan{
		traceID:    traceID,
		spanID:     spanID,
		operation:  operation,
		timestamp:  time.Now(),
		attributes: attributes,
		tracer:     nrt,
	}
}

func (nrt *NewRelicTracer) getParentSpanID(object interface{}) (string, string) {

	h, ok := object.(http.Header)
	if ok {

		traceID := ""
		spanID := ""

		arr := h[http.CanonicalHeaderKey(headerTraceID)]
		if len(arr) > 0 {
			traceID = arr[len(arr)-1]
		}

		arr = h[http.CanonicalHeaderKey(headerSpanID)]
		if len(arr) > 0 {
			spanID = arr[len(arr)-1]
		}

		return spanID, traceID
	}

	nrtsc, ok := object.(*NewRelicTracerSpanContext)
	if ok {
		return nrtsc.tracerSpan.spanID, nrtsc.tracerSpan.traceID
	}
	return "", ""
}

func (nrt *NewRelicTracer) StartChildSpan(object interface{}) common.TracerSpan {

	parentID, traceID := nrt.getParentSpanID(object)
	operation, attributes := nrt.getSpanAttributes()

	return &NewRelicTracerSpan{
		traceID:    traceID,
		spanID:     common.NewSpanID(),
		parentID:   parentID,
		operation:  operation,
		timestamp:  time.Now(),
		attributes: attributes,
		tracer:     nrt,
	}
}

func (nrt *NewRelicTracer) StartFollowSpan(object interface{}) common.TracerSpan {

	parentID, traceID := nrt.getParentSpanID(object)
	operation, attributes := nrt.getSpanAttributes()

	return &NewRelicTracerSpan{
		traceID:    traceID,
		spanID:     common.NewSpanID(),
		parentID:   parentID,
		operation:  operation,
		timestamp:  time.Now(),
		attributes: attributes,
		tracer:     nrt,
	}
}

func (nrt *NewRelicTracer) SetCallerOffset(offset int) {
	nrt.callerOffset = offset
}

func (nrt *NewRelicTracer) Stop() {

	if nrt.harvester != nil {
		nrt.harvester.HarvestNow(context.Background())
	}
}

func NewNewRelicTracer(options NewRelicTracerOptions, logger common.Logger, stdout *Stdout) *NewRelicTracer {

	if logger == nil {
		logger = stdout
	}

	if utils.IsEmpty(options.Endpoint) {
		stdout.Debug("NewRelic tracer is disabled.")
		return nil
	}

	attribites := make(map[string]interface{})
	m := utils.MapGetKeyValues(options.Attributes)
	for k, v := range m {
		attribites[k] = v
	}

	var cfgs []func(*telemetry.Config)
	cfgs = append(cfgs,
		telemetry.ConfigAPIKey(options.ApiKey),
		telemetry.ConfigSpansURLOverride(options.Endpoint),
		telemetry.ConfigCommonAttributes(attribites),
	)

	if options.Debug {
		cfgs = append(cfgs,
			telemetry.ConfigBasicErrorLogger(stdout.log.Writer()),
			telemetry.ConfigBasicDebugLogger(stdout.log.Writer()),
		)
	}

	harvester, err := telemetry.NewHarvester(cfgs...)
	if err != nil {
		stdout.Error(err)
		return nil
	}

	logger.Info("NewRelic tracer is up...")

	return &NewRelicTracer{
		options:      options,
		harvester:    harvester,
		logger:       logger,
		callerOffset: 1,
	}
}

func (nr *NewRelicLogger) addSpanFields(span common.TracerSpan, fields logrus.Fields) logrus.Fields {

	if span == nil {
		return fields
	}

	ctx := span.GetContext()
	if ctx == nil {
		return fields
	}

	fields["trace.id"] = ctx.GetTraceID()
	fields["span.id"] = ctx.GetSpanID()

	return fields
}

func (nr *NewRelicLogger) logToApi(level, message string, fields logrus.Fields) bool {

	if nr.harvester != nil {

		attributes := fields
		if attributes != nil {
			attributes["level"] = level
		}

		err := nr.harvester.RecordLog(telemetry.Log{
			Timestamp:  time.Now(),
			Message:    message,
			Attributes: attributes,
		})
		if err != nil {
			nr.stdout.Error(err)
			return false
		}
		return true
	}
	return false
}

func (nr *NewRelicLogger) Info(obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := nr.exists(logrus.InfoLevel, obj, args...); exists {
		if nr.log != nil {
			nr.log.WithFields(fields).Infoln(message)
		} else {
			nr.logToApi("info", message, fields)
		}
	}
	return nr
}

func (nr *NewRelicLogger) SpanInfo(span common.TracerSpan, obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := nr.exists(logrus.InfoLevel, obj, args...); exists {
		fields = nr.addSpanFields(span, fields)
		if nr.log != nil {
			nr.log.WithFields(fields).Infoln(message)
		} else {
			nr.logToApi("info", message, fields)
		}
	}
	return nr
}

func (nr *NewRelicLogger) Warn(obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := nr.exists(logrus.WarnLevel, obj, args...); exists {
		if nr.log != nil {
			nr.log.WithFields(fields).Warnln(message)
		} else {
			nr.logToApi("warn", message, fields)
		}
	}
	return nr
}

func (nr *NewRelicLogger) SpanWarn(span common.TracerSpan, obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := nr.exists(logrus.WarnLevel, obj, args...); exists {
		fields = nr.addSpanFields(span, fields)
		if nr.log != nil {
			nr.log.WithFields(fields).Warnln(message)
		} else {
			nr.logToApi("warn", message, fields)
		}
	}
	return nr
}

func (nr *NewRelicLogger) Error(obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := nr.exists(logrus.ErrorLevel, obj, args...); exists {
		if nr.log != nil {
			nr.log.WithFields(fields).Errorln(message)
		} else {
			nr.logToApi("error", message, fields)
		}
	}
	return nr
}

func (nr *NewRelicLogger) SpanError(span common.TracerSpan, obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := nr.exists(logrus.ErrorLevel, obj, args...); exists {
		fields = nr.addSpanFields(span, fields)
		if nr.log != nil {
			nr.log.WithFields(fields).Errorln(message)
		} else {
			nr.logToApi("error", message, fields)
		}
	}
	return nr
}

func (nr *NewRelicLogger) Debug(obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := nr.exists(logrus.DebugLevel, obj, args...); exists {
		if nr.log != nil {
			nr.log.WithFields(fields).Debugln(message)
		} else {
			nr.logToApi("debug", message, fields)
		}
	}
	return nr
}

func (nr *NewRelicLogger) SpanDebug(span common.TracerSpan, obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := nr.exists(logrus.DebugLevel, obj, args...); exists {
		fields = nr.addSpanFields(span, fields)
		if nr.log != nil {
			nr.log.WithFields(fields).Debugln(message)
		} else {
			nr.logToApi("debug", message, fields)
		}
	}
	return nr
}

func (nr *NewRelicLogger) Panic(obj interface{}, args ...interface{}) {

	if exists, fields, message := nr.exists(logrus.PanicLevel, obj, args...); exists {
		if nr.log != nil {
			nr.log.WithFields(fields).Panicln(message)
		} else {
			nr.logToApi("panic", message, fields)
			nr.stdout.Panic(message)
		}
	}
}

func (nr *NewRelicLogger) SpanPanic(span common.TracerSpan, obj interface{}, args ...interface{}) {

	if exists, fields, message := nr.exists(logrus.PanicLevel, obj, args...); exists {
		fields = nr.addSpanFields(span, fields)
		if nr.log != nil {
			nr.log.WithFields(fields).Panicln(message)
		} else {
			nr.logToApi("panic", message, fields)
			nr.stdout.SpanPanic(span, message)
		}
	}
}

func (nr *NewRelicLogger) Stack(offset int) common.Logger {
	nr.callerOffset = nr.callerOffset - offset
	return nr
}

func (nr *NewRelicLogger) exists(level logrus.Level, obj interface{}, args ...interface{}) (bool, logrus.Fields, string) {

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

	if utils.IsEmpty(message) && !nr.log.IsLevelEnabled(level) {
		return false, nil, ""
	}

	function, file, line := utils.CallerGetInfo(nr.callerOffset + 5)
	fields := logrus.Fields{
		"file":    fmt.Sprintf("%s:%d", file, line),
		"func":    function,
		"service": nr.options.ServiceName,
		"version": nr.options.Version,
		"env":     nr.options.Environment,
	}

	m := utils.MapGetKeyValues(nr.options.Attributes)
	for k, v := range m {
		fields[k] = v
	}

	return true, fields, message
}

func (nr *NewRelicLogger) Stop() {
	if nr.connection != nil {
		nr.connection.Close()
	}
	if nr.harvester != nil {
		nr.harvester.HarvestNow(context.Background())
	}
}

func NewNewRelicLogger(options NewRelicLoggerOptions, logger common.Logger, stdout *Stdout) *NewRelicLogger {

	if logger == nil {
		logger = stdout
	}

	if utils.IsEmpty(options.Endpoint) && utils.IsEmpty(options.AgentHost) {
		stdout.Debug("NewRelic logger is disabled.")
		return nil
	}

	var connection *net.TCPConn = nil
	var log *logrus.Logger = nil

	if utils.IsEmpty(options.Endpoint) && !utils.IsEmpty(options.AgentHost) {

		address := fmt.Sprintf("%s:%d", options.AgentHost, options.AgentPort)
		serverAddr, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			stdout.Error(err)
			return nil
		}

		connection, err = net.DialTCP("tcp", nil, serverAddr)
		if err != nil {
			stdout.Error(err)
			return nil
		}

		formatter := &logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		}
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

		if connection != nil {
			log.SetOutput(connection)
		}
	}

	var harvester *telemetry.Harvester = nil

	if !utils.IsEmpty(options.Endpoint) {

		attribites := make(map[string]interface{})
		m := utils.MapGetKeyValues(options.Attributes)
		for k, v := range m {
			attribites[k] = v
		}

		var cfgs []func(*telemetry.Config)
		cfgs = append(cfgs,
			telemetry.ConfigAPIKey(options.ApiKey),
			telemetry.ConfigLogsURLOverride(options.Endpoint),
			telemetry.ConfigCommonAttributes(attribites),
		)

		if options.Debug {
			cfgs = append(cfgs,
				telemetry.ConfigBasicErrorLogger(stdout.log.Writer()),
				telemetry.ConfigBasicDebugLogger(stdout.log.Writer()),
			)
		}

		h, err := telemetry.NewHarvester(cfgs...)
		if err != nil {
			stdout.Error(err)
			return nil
		}
		harvester = h
	}

	logger.Info("NewRelic logger is up...")

	return &NewRelicLogger{
		harvester:    harvester,
		connection:   connection,
		stdout:       stdout,
		log:          log,
		options:      options,
		callerOffset: 1,
	}
}

func (nrc *NewRelicCounter) getGlobalTags(labelValues ...string) map[string]interface{} {

	m := make(map[string]interface{})
	l := len(labelValues)

	for index, name := range nrc.labels {
		if l > index {
			m[name] = labelValues[index]
		}
	}
	return m
}

func (nrc *NewRelicCounter) Inc(labelValues ...string) common.Counter {

	attributes := nrc.getGlobalTags(labelValues...)
	_, file, line := utils.CallerGetInfo(nrc.meter.callerOffset + 3)
	attributes["file"] = fmt.Sprintf("%s:%d", file, line)

	nrc.meter.harvester.RecordMetric(telemetry.Count{
		Timestamp:  time.Now(),
		Name:       nrc.name,
		Value:      1,
		Attributes: attributes,
	})

	return nrc
}

func (nrm *NewRelicMeter) Counter(name, description string, labels []string, prefixes ...string) common.Counter {

	var names []string

	if !utils.IsEmpty(nrm.options.Prefix) {
		names = append(names, nrm.options.Prefix)
	}

	names = append(names, prefixes...)
	names = append(names, name)
	newName := strings.Join(names, ".")

	return &NewRelicCounter{
		meter:       nrm,
		name:        newName,
		description: description,
		labels:      labels,
	}
}

func (nrg *NewRelicGauge) getGlobalTags(labelValues ...string) map[string]interface{} {

	m := make(map[string]interface{})
	l := len(labelValues)

	for index, name := range nrg.labels {
		if l > index {
			m[name] = labelValues[index]
		}
	}
	return m
}

func (nrg *NewRelicGauge) Set(value float64, labelValues ...string) common.Gauge {

	attributes := nrg.getGlobalTags(labelValues...)
	_, file, line := utils.CallerGetInfo(nrg.meter.callerOffset + 3)
	attributes["file"] = fmt.Sprintf("%s:%d", file, line)

	nrg.meter.harvester.RecordMetric(telemetry.Gauge{
		Timestamp:  time.Now(),
		Name:       nrg.name,
		Value:      value,
		Attributes: attributes,
	})

	return nrg
}

func (nrm *NewRelicMeter) Gauge(name, description string, labels []string, prefixes ...string) common.Gauge {

	var names []string

	if !utils.IsEmpty(nrm.options.Prefix) {
		names = append(names, nrm.options.Prefix)
	}

	names = append(names, prefixes...)
	names = append(names, name)
	newName := strings.Join(names, ".")

	return &NewRelicGauge{
		meter:       nrm,
		name:        newName,
		description: description,
		labels:      labels,
	}
}

func (nrm *NewRelicMeter) SetCallerOffset(offset int) {
	nrm.callerOffset = offset
}

func (nrm *NewRelicMeter) Stop() {
	if nrm.harvester != nil {
		nrm.harvester.HarvestNow(context.Background())
	}
}

func NewNewRelicMeter(options NewRelicMeterOptions, logger common.Logger, stdout *Stdout) *NewRelicMeter {

	if logger == nil {
		logger = stdout
	}

	if utils.IsEmpty(options.Endpoint) {
		stdout.Debug("NewRelic meter is disabled.")
		return nil
	}

	attribites := make(map[string]interface{})
	m := utils.MapGetKeyValues(options.Attributes)
	for k, v := range m {
		attribites[k] = v
	}

	var cfgs []func(*telemetry.Config)
	cfgs = append(cfgs,
		telemetry.ConfigAPIKey(options.ApiKey),
		telemetry.ConfigMetricsURLOverride(options.Endpoint),
		telemetry.ConfigCommonAttributes(attribites),
	)

	if options.Debug {
		cfgs = append(cfgs,
			telemetry.ConfigBasicErrorLogger(stdout.log.Writer()),
			telemetry.ConfigBasicDebugLogger(stdout.log.Writer()),
		)
	}

	harvester, err := telemetry.NewHarvester(cfgs...)
	if err != nil {
		stdout.Error(err)
		return nil
	}

	logger.Info("NewRelic meter is up...")

	return &NewRelicMeter{
		harvester:    harvester,
		options:      options,
		logger:       logger,
		callerOffset: 1,
	}
}

func (nre *NewRelicEventer) Interval(name string, attributes map[string]string, begin, end time.Time) error {

	attrs := make(map[string]interface{})
	if attributes != nil {
		for k, v := range attributes {
			attrs[k] = v
		}
		for k, v := range nre.attributes {
			attrs[k] = v
		}
	}

	event := telemetry.Event{
		EventType:  name,
		Timestamp:  begin,
		Attributes: attrs,
	}

	err := nre.harvester.RecordEvent(event)
	if err != nil {
		nre.logger.Error(err)
		return err
	}
	return nil
}

func (nre *NewRelicEventer) Now(name string, attributes map[string]string) error {
	return nre.At(name, attributes, time.Now())
}

func (nre *NewRelicEventer) At(name string, attributes map[string]string, when time.Time) error {
	return nre.Interval(name, attributes, when, when)
}

func (nre *NewRelicEventer) Stop() {
	if nre.harvester != nil {
		nre.harvester.HarvestNow(context.Background())
	}
}

func NewNewRelicEventer(options NewRelicEventerOptions, logger common.Logger, stdout *Stdout) *NewRelicEventer {

	if logger == nil {
		logger = stdout
	}

	if utils.IsEmpty(options.Endpoint) {
		stdout.Debug("NewRelic eventer is disabled.")
		return nil
	}

	attribites := make(map[string]interface{})
	m := utils.MapGetKeyValues(options.Attributes)
	for k, v := range m {
		attribites[k] = v
	}

	var cfgs []func(*telemetry.Config)
	cfgs = append(cfgs,
		telemetry.ConfigAPIKey(options.ApiKey),
		telemetry.ConfigEventsURLOverride(options.Endpoint),
		telemetry.ConfigCommonAttributes(attribites),
	)

	if options.Debug {
		cfgs = append(cfgs,
			telemetry.ConfigBasicErrorLogger(stdout.log.Writer()),
			telemetry.ConfigBasicDebugLogger(stdout.log.Writer()),
		)
	}

	harvester, err := telemetry.NewHarvester(cfgs...)
	if err != nil {
		stdout.Error(err)
		return nil
	}

	logger.Info("NewRelic eventer is up...")

	return &NewRelicEventer{
		harvester:  harvester,
		options:    options,
		logger:     logger,
		attributes: attribites,
	}
}
