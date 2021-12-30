package incr

import (
	"context"
	"log"
	"os"
)

// Tracer is a type that can implement a tracer.
type Tracer interface {
	Println(...any)
	Printf(string, ...any)
	Errorln(...any)
	Errorf(string, ...any)
}

type tracerKey struct{}

// WithTracing adds a tracer to a given context.
func WithTracing(ctx context.Context) context.Context {
	tracer := &tracer{
		log:    log.New(os.Stderr, "incr.trace|", log.LUTC|log.Lshortfile|log.Ldate|log.Lmicroseconds),
		errLog: log.New(os.Stderr, "incr.trace.err|", log.LUTC|log.Lshortfile|log.Ldate|log.Lmicroseconds),
	}
	return context.WithValue(ctx, tracerKey{}, tracer)
}

func getTracer(ctx context.Context) Tracer {
	if value := ctx.Value(tracerKey{}); value != nil {
		if typed, ok := value.(Tracer); ok {
			return typed
		}
	}
	return nil
}

func tracePrintln(ctx context.Context, args ...any) {
	if tracer := getTracer(ctx); tracer != nil {
		tracer.Println(args...)
	}
}

func tracePrintf(ctx context.Context, format string, args ...any) {
	if tracer := getTracer(ctx); tracer != nil {
		tracer.Printf(format, args...)
	}
}

func traceErrorln(ctx context.Context, args ...any) {
	if tracer := getTracer(ctx); tracer != nil {
		tracer.Errorln(args...)
	}
}

func traceErrorf(ctx context.Context, format string, args ...any) {
	if tracer := getTracer(ctx); tracer != nil {
		tracer.Errorf(format, args...)
	}
}

type tracer struct {
	log    *log.Logger
	errLog *log.Logger
}

func (t *tracer) Println(args ...any) {
	t.log.Println(args...)
}

func (t *tracer) Printf(format string, args ...any) {
	t.log.Printf(format, args...)
}

func (t *tracer) Errorln(args ...any) {
	t.errLog.Println(args...)
}

func (t *tracer) Errorf(format string, args ...any) {
	t.errLog.Printf(format, args...)
}