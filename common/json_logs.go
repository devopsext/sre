package common

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/devopsext/utils"
)

type LogMessage struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

func formatLogs(level string, obj interface{}, args ...interface{}) (string, error) {
	rawObj, ok := obj.(string)
	if !ok {
		return "", errors.New("unable to process log")
	}
	message := fmt.Sprintf(rawObj, args...)
	logMessage := LogMessage{
		Level:   level,
		Message: message,
	}

	jsonMessage, err := json.Marshal(logMessage)
	if err != nil {
		return "", errors.New("unable to process log")
	}
	return string(jsonMessage), nil
}

type JsonLogs struct {
	loggers []Logger
}

func (ls *JsonLogs) Info(obj interface{}, args ...interface{}) Logger {
	logMessage, err := formatLogs("info", obj, args...)
	if err != nil {
		return nil
	}
	for _, l := range ls.loggers {
		l.Info(logMessage)
	}
	return ls
}

func (ls *JsonLogs) SpanInfo(span TracerSpan, obj interface{}, args ...interface{}) Logger {
	for _, l := range ls.loggers {
		l.SpanInfo(span, obj, args...)
	}
	return ls
}

func (ls *JsonLogs) Warn(obj interface{}, args ...interface{}) Logger {
	logMessage, err := formatLogs("warn", obj, args...)
	if err != nil {
		return nil
	}
	for _, l := range ls.loggers {
		l.Warn(logMessage)
	}
	return ls
}

func (ls *JsonLogs) SpanWarn(span TracerSpan, obj interface{}, args ...interface{}) Logger {
	for _, l := range ls.loggers {
		l.SpanWarn(span, obj, args...)
	}
	return ls
}

func (ls *JsonLogs) Error(obj interface{}, args ...interface{}) Logger {
	logMessage, err := formatLogs("error", obj, args...)
	if err != nil {
		return nil
	}
	for _, l := range ls.loggers {
		l.Error(logMessage)
	}
	return ls
}

func (ls *JsonLogs) SpanError(span TracerSpan, obj interface{}, args ...interface{}) Logger {
	for _, l := range ls.loggers {
		l.SpanError(span, obj, args...)
	}
	if span != nil && obj != nil {

		message := ""
		switch v := obj.(type) {
		case error:
			message = v.Error()
		case string:
			message = v
		default:
			message = "not implemented"
		}

		if !utils.IsEmpty(message) {
			span.Error(errors.New(message))
		}
	}
	return ls
}

func (ls *JsonLogs) Debug(obj interface{}, args ...interface{}) Logger {
	logMessage, err := formatLogs("debug", obj, args...)
	if err != nil {
		return nil
	}
	for _, l := range ls.loggers {
		l.Debug(logMessage)
	}
	return ls
}

func (ls *JsonLogs) SpanDebug(span TracerSpan, obj interface{}, args ...interface{}) Logger {
	for _, l := range ls.loggers {
		l.SpanDebug(span, obj, args...)
	}
	return ls
}

func (ls *JsonLogs) Panic(obj interface{}, args ...interface{}) {
	logMessage, err := formatLogs("panic", obj, args...)
	if err != nil {
		return
	}
	for _, l := range ls.loggers {
		l.Panic(logMessage)
	}
}

func (ls *JsonLogs) SpanPanic(span TracerSpan, obj interface{}, args ...interface{}) {
	for _, l := range ls.loggers {
		l.SpanPanic(span, obj, args...)
	}
}

func (ls *JsonLogs) Stack(offset int) Logger {
	for _, l := range ls.loggers {
		l.Stack(offset)
	}
	return ls
}

func (ls *JsonLogs) Stop() {
	for _, l := range ls.loggers {
		l.Stop()
	}
}

func (ls *JsonLogs) Register(l Logger) {
	if l != nil {
		ls.loggers = append(ls.loggers, l)
	}
}

func NewJsonLogs() *JsonLogs {
	return &JsonLogs{}
}
