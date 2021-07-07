package provider

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"text/template"

	"github.com/devopsext/sre/common"
	"github.com/sirupsen/logrus"
)

type StdoutOptions struct {
	Format          string
	Level           string
	Template        string
	TimestampFormat string
	Version         string
	TextColors      bool
}

type Stdout struct {
	log          *logrus.Logger
	options      StdoutOptions
	callerOffset int
}

type templateFormatter struct {
	template        *template.Template
	timestampFormat string
}

func (f *templateFormatter) Format(entry *logrus.Entry) ([]byte, error) {

	r := entry.Message
	m := make(map[string]interface{})

	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			m[k] = v.Error()
		default:
			m[k] = v
		}
	}

	m["msg"] = entry.Message
	m["time"] = entry.Time.Format(f.timestampFormat)
	m["level"] = entry.Level.String()

	var err error

	if f.template != nil {

		var b bytes.Buffer
		err = f.template.Execute(&b, m)
		if err == nil {

			r = fmt.Sprintf("%s\n", b.String())
		}
	}

	return []byte(r), err
}

func (so *Stdout) addSpanFields(span common.TracerSpan, fields logrus.Fields) logrus.Fields {

	if span == nil {
		return fields
	}

	ctx := span.GetContext()
	if ctx == nil {
		return fields
	}

	fields["trace_id"] = strconv.FormatUint(ctx.GetTraceID(), 10)
	return fields
}

func (so *Stdout) addCallerFields(offset int) logrus.Fields {

	function, file, line := common.GetCallerInfo(so.callerOffset + offset)
	return logrus.Fields{
		"file": fmt.Sprintf("%s:%d", file, line),
		"func": function,
	}
}

func prepare(message string, args ...interface{}) string {

	if len(args) > 0 {
		return fmt.Sprintf(message, args...)
	} else {
		return message
	}
}

func (so *Stdout) exists(level logrus.Level, obj interface{}, args ...interface{}) (bool, string) {

	if obj == nil {
		return false, ""
	}

	message := ""

	switch v := obj.(type) {
	case error:
		message = v.Error()
	case string:
		message = v
	default:
		message = "not implemented"
	}

	flag := message != "" && so.log.IsLevelEnabled(level)
	if flag {
		message = prepare(message, args...)
	}
	return flag, message
}

func (so *Stdout) Info(obj interface{}, args ...interface{}) common.Logger {

	if exists, message := so.exists(logrus.InfoLevel, obj, args...); exists {
		so.log.WithFields(so.addCallerFields(3)).Infoln(message)
	}
	return so
}

func (so *Stdout) SpanInfo(span common.TracerSpan, obj interface{}, args ...interface{}) common.Logger {

	if exists, message := so.exists(logrus.InfoLevel, obj, args...); exists {
		fields := so.addSpanFields(span, so.addCallerFields(3))
		so.log.WithFields(fields).Infoln(message)
	}
	return so
}

func (so *Stdout) Warn(obj interface{}, args ...interface{}) common.Logger {

	if exists, message := so.exists(logrus.WarnLevel, obj, args...); exists {
		so.log.WithFields(so.addCallerFields(3)).Warnln(message)
	}
	return so
}

func (so *Stdout) SpanWarn(span common.TracerSpan, obj interface{}, args ...interface{}) common.Logger {

	if exists, message := so.exists(logrus.WarnLevel, obj, args...); exists {
		fields := so.addSpanFields(span, so.addCallerFields(3))
		so.log.WithFields(fields).Warnln(message)
	}
	return so
}

func (so *Stdout) Error(obj interface{}, args ...interface{}) common.Logger {

	if exists, message := so.exists(logrus.ErrorLevel, obj, args...); exists {
		so.log.WithFields(so.addCallerFields(3)).Errorln(message)
	}
	return so
}

func (so *Stdout) SpanError(span common.TracerSpan, obj interface{}, args ...interface{}) common.Logger {

	if exists, message := so.exists(logrus.ErrorLevel, obj, args...); exists {
		fields := so.addSpanFields(span, so.addCallerFields(3))
		so.log.WithFields(fields).Errorln(message)
	}
	return so
}

func (so *Stdout) Debug(obj interface{}, args ...interface{}) common.Logger {

	if exists, message := so.exists(logrus.DebugLevel, obj, args...); exists {
		so.log.WithFields(so.addCallerFields(3)).Debugln(message)
	}
	return so
}

func (so *Stdout) SpanDebug(span common.TracerSpan, obj interface{}, args ...interface{}) common.Logger {

	if exists, message := so.exists(logrus.DebugLevel, obj, args...); exists {
		fields := so.addSpanFields(span, so.addCallerFields(3))
		so.log.WithFields(fields).Debugln(message)
	}
	return so
}

func (so *Stdout) Panic(obj interface{}, args ...interface{}) {

	if exists, message := so.exists(logrus.PanicLevel, obj, args...); exists {
		so.log.WithFields(so.addCallerFields(3)).Panicln(message)
	}
}

func (so *Stdout) SpanPanic(span common.TracerSpan, obj interface{}, args ...interface{}) {

	if exists, message := so.exists(logrus.PanicLevel, obj, args...); exists {
		fields := so.addSpanFields(span, so.addCallerFields(3))
		so.log.WithFields(fields).Panicln(message)
	}
}

func (so *Stdout) Stack(offset int) common.Logger {
	so.callerOffset = so.callerOffset - offset
	return so
}

func newLog(options StdoutOptions) *logrus.Logger {

	log := logrus.New()

	switch options.Format {
	case "json":
		formatter := &logrus.JSONFormatter{}
		formatter.TimestampFormat = options.TimestampFormat
		log.SetFormatter(formatter)
	case "text":
		formatter := &logrus.TextFormatter{}
		formatter.TimestampFormat = options.TimestampFormat
		formatter.ForceColors = options.TextColors
		formatter.FullTimestamp = true
		log.SetFormatter(formatter)
	case "template":
		t, err := template.New("").Parse(options.Template)
		if err != nil {
			log.Panic(err)
		}
		log.SetFormatter(&templateFormatter{template: t, timestampFormat: options.TimestampFormat})
	default:
		formatter := &logrus.TextFormatter{}
		formatter.TimestampFormat = options.TimestampFormat
		formatter.ForceColors = options.TextColors
		formatter.FullTimestamp = true
		log.SetFormatter(formatter)
	}

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

	log.SetOutput(os.Stdout)
	return log
}

func (so *Stdout) SetCallerOffset(offset int) {
	so.callerOffset = offset
}

func NewStdout(options StdoutOptions) *Stdout {

	log := newLog(options)

	return &Stdout{
		log:          log,
		options:      options,
		callerOffset: 1,
	}
}
