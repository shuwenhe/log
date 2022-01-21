package log

import (
	"context"
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	_log *Logger
	once sync.Once
)

type Logger struct {
	*logrus.Entry
}

// MustInit if fail panic
func New(serviceName string, logFile string) *Logger {
	once.Do(func() {
		l := logrus.New()
		l.SetFormatter(&logrus.JSONFormatter{})
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			l.Out = io.MultiWriter(file, os.Stdout)
		} else {
			l.Warn("Failed to log to file, using default stderr")
		}
		_log = &Logger{Entry: l.WithField(serviceNameField, serviceName)}
	})
	return _log
}

type requestID struct{}

const (
	reqIDName        = "request_id"
	serviceNameField = "service_name"
)

func (l *Logger) StartTracerFromCtx(ctx context.Context, reqID uint64) (context.Context, *logrus.Entry) {
	return context.WithValue(ctx, requestID{}, requestID{}), l.WithField(reqIDName, reqID)
}

func (l *Logger) StartTracerFromNewCtx(reqID uint64) (context.Context, *logrus.Entry) {
	return context.WithValue(context.Background(), requestID{}, reqID), l.WithField(reqIDName, reqID)
}

func (l *Logger) TracerFromCtx(ctx context.Context) *logrus.Entry {
	traceID, ok := ctx.Value(requestID{}).(uint64)
	if ok {
		return l.WithField(reqIDName, traceID)
	}
	return l.Entry
}
