package incr

import (
	"fmt"
	"io"
	"log"
)

// NewTracer returns a new tracer from a pair of output writers.
func NewTracer(output, errOutput io.Writer) Tracer {
	tracer := &tracer{
		log:    log.New(output, "incr.trace|", log.LUTC|log.Lshortfile|log.Ldate|log.Lmicroseconds),
		errLog: log.New(errOutput, "incr.trace.err|", log.LUTC|log.Lshortfile|log.Ldate|log.Lmicroseconds),
	}
	return tracer
}

// Tracer is a type that can implement a tracer.
type Tracer interface {
	Print(...any)
	Error(...any)
}

// tracePrintln prints a line to the tracer on a given context.
func (g *Graph) tracePrintln(args ...any) {
	if g.tracer != nil {
		g.tracer.Print(args...)
	}
}

// tracePrintf prints a line to the tracer on a given
// context with a given format and args.
func (g *Graph) tracePrintf(format string, args ...any) {
	if g.tracer != nil {
		g.tracer.Print(fmt.Sprintf(format, args...))
	}
}

// traceErrorln prints a line to the error output of a tracer on a given context.
func (g *Graph) traceErrorln(args ...any) {
	if g.tracer != nil {
		g.tracer.Error(args...)
	}
}

// traceErrorf prints a line to the error output of a tracer
// on a given context with a given format and args.
func (g *Graph) traceErrorf(format string, args ...any) {
	if g.tracer != nil {
		g.tracer.Error(fmt.Sprintf(format, args...))
	}
}

type tracer struct {
	log    *log.Logger
	errLog *log.Logger
}

func (t *tracer) Print(args ...any) {
	_ = t.log.Output(4, fmt.Sprint(args...))
}

func (t *tracer) Error(args ...any) {
	_ = t.errLog.Output(4, fmt.Sprint(args...))
}
