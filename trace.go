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

const defaultLoggerFlags = log.LUTC | log.Lshortfile | log.Ldate | log.Lmicroseconds

// WithTracingOutputs adds a tracer to a given context with given outputs.
func WithTracingOutputs(ctx context.Context, output, errOutput io.Writer) context.Context {
	tracer := &tracer{
		log:    log.New(output, "incr.trace|", defaultLoggerFlags),
		errLog: log.New(errOutput, "incr.trace.err|", defaultLoggerFlags),
	}
	return WithTracer(ctx, tracer)
}

// WithTracer adds a tracer to a given context.
func WithTracer(ctx context.Context, tracer Tracer) context.Context {
	return context.WithValue(ctx, tracerKey{}, tracer)
}

// GetTracer returns the tracer from a given context, and nil if one is not present.
func GetTracer(ctx context.Context) Tracer {
	if value := ctx.Value(tracerKey{}); value != nil {
		if typed, ok := value.(Tracer); ok {
			return typed
		}
	}
	return nil
}

// TracePrintln prints a line to the tracer on a given context.
func TracePrintln(ctx context.Context, args ...any) {
	if tracer := GetTracer(ctx); tracer != nil {
		tracer.Print(FormatStabilizationNumber(ctx) + fmt.Sprint(args...))
	}
}

// TracePrintf prints a line to the tracer on a given
// context with a given format and args.
func TracePrintf(ctx context.Context, format string, args ...any) {
	if tracer := GetTracer(ctx); tracer != nil {
		tracer.Print(FormatStabilizationNumber(ctx) + fmt.Sprintf(format, args...))
	}
}

// TraceErrorln prints a line to the error output of a tracer on a given context.
func TraceErrorln(ctx context.Context, args ...any) {
	if tracer := GetTracer(ctx); tracer != nil {
		tracer.Error(FormatStabilizationNumber(ctx) + fmt.Sprint(args...))
	}
}

// TraceErrorf prints a line to the error output of a tracer
// on a given context with a given format and args.
func TraceErrorf(ctx context.Context, format string, args ...any) {
	if tracer := GetTracer(ctx); tracer != nil {
		tracer.Error(FormatStabilizationNumber(ctx) + fmt.Sprintf(format, args...))
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

type stabilizationNumberKey struct{}

// WithStabilizationNumber adds a stabilization number to a context.
func WithStabilizationNumber(ctx context.Context, stabilizationNumber uint64) context.Context {
	return context.WithValue(ctx, stabilizationNumberKey{}, stabilizationNumber)
}

// GetStabilizationNumber gets the stabilization number from a context.
func GetStabilizationNumber(ctx context.Context) (stabilizationNumber uint64, ok bool) {
	if value := ctx.Value(stabilizationNumberKey{}); value != nil {
		stabilizationNumber, ok = value.(uint64)
		return
	}
	return
}

// FormatStabilizationNumber formats the stabilization number from a context for tracing output.
func FormatStabilizationNumber(ctx context.Context) (output string) {
	num, ok := GetStabilizationNumber(ctx)
	if !ok {
		return
	}
	return fmt.Sprintf("%d: ", num)
}
