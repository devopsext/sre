package provider

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/devopsext/sre/common"
	utils "github.com/devopsext/utils"
	"github.com/sirupsen/logrus"
)

type NewRelicOptions struct {
	ServiceName string
	Environment string
	Version     string
	Labels      string
	Debug       bool
}

type NewRelicLoggerOptions struct {
	NewRelicOptions
	AgentHost string
	AgentPort int
	Level     string
}

type NewRelicMeterOptions struct {
	NewRelicOptions
	AgentHost string
	AgentPort int
	Prefix    string
}

type NewRelicLogger struct {
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
	prefix      string
}

type NewRelicMeter struct {
	options      NewRelicMeterOptions
	logger       common.Logger
	callerOffset int
	//	client       *statsd.Client
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

func (nr *NewRelicLogger) Info(obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := nr.exists(logrus.InfoLevel, obj, args...); exists {
		nr.log.WithFields(fields).Infoln(message)
	}
	return nr
}

func (nr *NewRelicLogger) SpanInfo(span common.TracerSpan, obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := nr.exists(logrus.InfoLevel, obj, args...); exists {
		fields = nr.addSpanFields(span, fields)
		nr.log.WithFields(fields).Infoln(message)
	}
	return nr
}

func (nr *NewRelicLogger) Warn(obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := nr.exists(logrus.WarnLevel, obj, args...); exists {
		nr.log.WithFields(fields).Warnln(message)
	}
	return nr
}

func (nr *NewRelicLogger) SpanWarn(span common.TracerSpan, obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := nr.exists(logrus.WarnLevel, obj, args...); exists {
		fields = nr.addSpanFields(span, fields)
		nr.log.WithFields(fields).Warnln(message)
	}
	return nr
}

func (nr *NewRelicLogger) Error(obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := nr.exists(logrus.ErrorLevel, obj, args...); exists {
		nr.log.WithFields(fields).Errorln(message)
	}
	return nr
}

func (nr *NewRelicLogger) SpanError(span common.TracerSpan, obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := nr.exists(logrus.ErrorLevel, obj, args...); exists {
		fields = nr.addSpanFields(span, fields)
		nr.log.WithFields(fields).Errorln(message)
	}
	return nr
}

func (nr *NewRelicLogger) Debug(obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := nr.exists(logrus.DebugLevel, obj, args...); exists {
		nr.log.WithFields(fields).Debugln(message)
	}
	return nr
}

func (nr *NewRelicLogger) SpanDebug(span common.TracerSpan, obj interface{}, args ...interface{}) common.Logger {

	if exists, fields, message := nr.exists(logrus.DebugLevel, obj, args...); exists {
		fields = nr.addSpanFields(span, fields)
		nr.log.WithFields(fields).Debugln(message)
	}
	return nr
}

func (nr *NewRelicLogger) Panic(obj interface{}, args ...interface{}) {

	if exists, fields, message := nr.exists(logrus.PanicLevel, obj, args...); exists {
		nr.log.WithFields(fields).Panicln(message)
	}
}

func (nr *NewRelicLogger) SpanPanic(span common.TracerSpan, obj interface{}, args ...interface{}) {

	if exists, fields, message := nr.exists(logrus.PanicLevel, obj, args...); exists {
		fields = nr.addSpanFields(span, fields)
		nr.log.WithFields(fields).Panicln(message)
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

	function, file, line := common.GetCallerInfo(nr.callerOffset + 5)
	fields := logrus.Fields{
		"file":    fmt.Sprintf("%s:%d", file, line),
		"func":    function,
		"service": nr.options.ServiceName,
		"version": nr.options.Version,
		"env":     nr.options.Environment,
	}

	m := common.GetKeyValues(nr.options.Labels)
	for k, v := range m {
		fields[k] = v
	}

	return true, fields, message
}

func NewNewRelicLogger(options NewRelicLoggerOptions, logger common.Logger, stdout *Stdout) *NewRelicLogger {

	if logger == nil {
		logger = stdout
	}

	if utils.IsEmpty(options.AgentHost) {
		stdout.Debug("NewRelic logger is disabled.")
		return nil
	}

	address := fmt.Sprintf("%s:%d", options.AgentHost, options.AgentPort)
	serverAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		stdout.Error(err)
		return nil
	}

	connection, err := net.DialTCP("tcp", nil, serverAddr)
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

	logger.Info("NewRelic logger is up...")

	return &NewRelicLogger{
		connection:   connection,
		stdout:       stdout,
		log:          log,
		options:      options,
		callerOffset: 1,
	}
}

func (nrc *NewRelicCounter) Inc(labelValues ...string) common.Counter {

	/*newName := ddmc.name
	if !utils.IsEmpty(ddmc.prefix) {
		newName = fmt.Sprintf("%s.%s", ddmc.prefix, newName)
	}

	newValues := ddmc.getLabelTags(labelValues...)
	_, file, line := common.GetCallerInfo(ddmc.meter.callerOffset + 3)
	newValues = append(newValues, fmt.Sprintf("file:%s", fmt.Sprintf("%s:%d", file, line)))

	err := ddmc.meter.client.Incr(newName, newValues, 1)
	if err != nil {
		ddmc.meter.logger.Error(err)
	}*/
	return nrc
}

func (nrm *NewRelicMeter) SetCallerOffset(offset int) {
	nrm.callerOffset = offset
}

func (nrm *NewRelicMeter) Counter(name, description string, labels []string, prefixes ...string) common.Counter {

	var names []string

	/*	if !utils.IsEmpty(ddm.options.Prefix) {
			names = append(names, ddm.options.Prefix)
		}

		if len(prefixes) > 0 {
			names = append(names, strings.Join(prefixes, "_"))
		}
	*/
	return &NewRelicCounter{
		meter:       nrm,
		name:        name,
		description: description,
		labels:      labels,
		prefix:      strings.Join(names, "."),
	}
}

func (nrm *NewRelicMeter) Stop() {
	// nothing here
}

func NewNewRelicMeter(options NewRelicMeterOptions, logger common.Logger, stdout *Stdout) *NewRelicMeter {

	if logger == nil {
		logger = stdout
	}

	if utils.IsEmpty(options.AgentHost) {
		stdout.Debug("NewRelic meter is disabled.")
		return nil
	}

	/*client, err := statsd.New(fmt.Sprintf("%s:%d", options.AgentHost, options.AgentPort))
	if err != nil {
		logger.Error(err)
		return nil
	}*/

	logger.Info("NewRelic meter is up...")

	return &NewRelicMeter{
		options:      options,
		logger:       logger,
		callerOffset: 1,
		//	client:       client,
	}
}
