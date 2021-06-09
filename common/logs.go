package common

import (
	"errors"

	"github.com/devopsext/utils"
)

type Logs struct {
	loggers []Logger
}

func (ls *Logs) Info(obj interface{}, args ...interface{}) Logger {
	for _, l := range ls.loggers {
		l.Info(obj, args...)
	}
	return ls
}

func (ls *Logs) SpanInfo(span TracerSpan, obj interface{}, args ...interface{}) Logger {
	for _, l := range ls.loggers {
		l.SpanInfo(span, obj, args...)
	}
	return ls
}

func (ls *Logs) Warn(obj interface{}, args ...interface{}) Logger {
	for _, l := range ls.loggers {
		l.Warn(obj, args...)
	}
	return ls
}

func (ls *Logs) SpanWarn(span TracerSpan, obj interface{}, args ...interface{}) Logger {
	for _, l := range ls.loggers {
		l.SpanWarn(span, obj, args...)
	}
	return ls
}

func (ls *Logs) Error(obj interface{}, args ...interface{}) Logger {
	for _, l := range ls.loggers {
		l.Error(obj, args...)
	}
	return ls
}

func (ls *Logs) SpanError(span TracerSpan, obj interface{}, args ...interface{}) Logger {
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

func (ls *Logs) Debug(obj interface{}, args ...interface{}) Logger {
	for _, l := range ls.loggers {
		l.Debug(obj, args...)
	}
	return ls
}

func (ls *Logs) SpanDebug(span TracerSpan, obj interface{}, args ...interface{}) Logger {
	for _, l := range ls.loggers {
		l.SpanDebug(span, obj, args...)
	}
	return ls
}

func (ls *Logs) Panic(obj interface{}, args ...interface{}) Logger {
	for _, l := range ls.loggers {
		l.Panic(obj, args...)
	}
	return ls
}

func (ls *Logs) SpanPanic(span TracerSpan, obj interface{}, args ...interface{}) Logger {
	for _, l := range ls.loggers {
		l.SpanPanic(span, obj, args...)
	}
	return ls
}

func (ls *Logs) Stack(offset int) Logger {
	for _, l := range ls.loggers {
		l.Stack(offset)
	}
	return ls
}

func (ls *Logs) Register(l Logger) {
	if l != nil {
		ls.loggers = append(ls.loggers, l)
	}
}

func NewLogs() *Logs {
	return &Logs{}
}
