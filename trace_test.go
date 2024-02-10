package incr

import (
	"bytes"
	"context"
	"strings"
	"testing"

	. "github.com/wcharczuk/go-incr/testutil"
)

func Test_WithTracing(t *testing.T) {
	ctx := context.Background()
	tr := GetTracer(ctx)
	Nil(t, tr)

	ctx = WithTracing(ctx)
	tr = GetTracer(ctx)
	NotNil(t, tr)
	NotNil(t, tr.(*tracer).log)
	NotNil(t, tr.(*tracer).errLog)
}

func Test_WithTracingOutput(t *testing.T) {
	output := new(bytes.Buffer)
	errOutput := new(bytes.Buffer)

	tr := GetTracer(context.Background())
	Nil(t, tr)

	ctx := WithTracingOutputs(context.Background(), output, errOutput)
	tr = GetTracer(ctx)
	NotNil(t, tr)
	NotNil(t, tr.(*tracer).log)
	NotNil(t, tr.(*tracer).errLog)

	TracePrintln(ctx, "this is a println test")
	Equal(t, true, strings.Contains(output.String(), "this is a println test"))
	Equal(t, "", errOutput.String())

	TraceErrorln(ctx, "this is a errorln test")
	Equal(t, false, strings.Contains(output.String(), "this is a errorln test"))
	Equal(t, true, strings.Contains(errOutput.String(), "this is a errorln test"))

	TracePrintf(ctx, "this is a %s test", "printf")
	Equal(t, true, strings.Contains(output.String(), "this is a printf test"))
	Equal(t, false, strings.Contains(errOutput.String(), "this is a printf test"))

	TraceErrorf(ctx, "this is a %s test", "errorf")
	Equal(t, false, strings.Contains(output.String(), "this is a errorf test"))
	Equal(t, true, strings.Contains(errOutput.String(), "this is a errorf test"))
}
