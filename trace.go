package incr

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
)

// Tracer is a type that can implement a tracer.
type Tracer interface {
	Print(...any)
	Error(...any)
}

type tracerKey struct{}

// WithTracing adds a default tracer to a given context.
func WithTracing(ctx context.Context) context.Context {
	return WithTracingOutputs(ctx, os.Stderr, os.Stderr)
}

// WithTracingOutputs adds a tracer to a given context with given outputs.
func WithTracingOutputs(ctx context.Context, output, errOutput io.Writer) context.Context {
	tracer := &tracer{
		log:    log.New(output, "incr.trace|", log.LUTC|log.Lshortfile|log.Ldate|log.Lmicroseconds),
		errLog: log.New(errOutput, "incr.trace.err|", log.LUTC|log.Lshortfile|log.Ldate|log.Lmicroseconds),
	}
	return WithTracer(ctx, tracer)
}

// WithTracer adds a tracer to a given context.
func WithTracer(ctx context.Context, tracer Tracer) context.Context {
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
		tracer.Print(args...)
	}
}

func tracePrintf(ctx context.Context, format string, args ...any) {
	if tracer := getTracer(ctx); tracer != nil {
		tracer.Print(fmt.Sprintf(format, args...))
	}
}

func traceErrorln(ctx context.Context, args ...any) {
	if tracer := getTracer(ctx); tracer != nil {
		tracer.Error(args...)
	}
}

func traceErrorf(ctx context.Context, format string, args ...any) {
	if tracer := getTracer(ctx); tracer != nil {
		tracer.Error(fmt.Sprintf(format, args...))
	}
}

type tracer struct {
	log    *log.Logger
	errLog *log.Logger
}

func (t *tracer) Print(args ...any) {
	_ = t.log.Output(3, fmt.Sprint(args...))
}

func (t *tracer) Error(args ...any) {
	_ = t.errLog.Output(3, fmt.Sprint(args...))
}
