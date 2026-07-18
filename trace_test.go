package incr

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_WithTracing(t *testing.T) {
	ctx := context.Background()
	tr := GetTracer(ctx)
	testutil.Nil(t, tr)

	ctx = WithTracing(ctx)
	tr = GetTracer(ctx)
	testutil.NotNil(t, tr)
	testutil.NotNil(t, tr.(*tracer).log)
	testutil.NotNil(t, tr.(*tracer).errLog)
}

func Test_WithTracingOutput(t *testing.T) {
	output := new(bytes.Buffer)
	errOutput := new(bytes.Buffer)

	tr := GetTracer(context.Background())
	testutil.Nil(t, tr)

	ctx := WithTracingOutputs(context.Background(), output, errOutput)
	tr = GetTracer(ctx)
	testutil.NotNil(t, tr)
	testutil.NotNil(t, tr.(*tracer).log)
	testutil.NotNil(t, tr.(*tracer).errLog)

	TracePrintln(ctx, "this is a println test")
	testutil.Equal(t, true, strings.Contains(output.String(), "this is a println test"))
	testutil.Equal(t, "", errOutput.String())

	TraceErrorln(ctx, "this is a errorln test")
	testutil.Equal(t, false, strings.Contains(output.String(), "this is a errorln test"))
	testutil.Equal(t, true, strings.Contains(errOutput.String(), "this is a errorln test"))

	TracePrintf(ctx, "this is a %s test", "printf")
	testutil.Equal(t, true, strings.Contains(output.String(), "this is a printf test"))
	testutil.Equal(t, false, strings.Contains(errOutput.String(), "this is a printf test"))

	TraceErrorf(ctx, "this is a %s test", "errorf")
	testutil.Equal(t, false, strings.Contains(output.String(), "this is a errorf test"))
	testutil.Equal(t, true, strings.Contains(errOutput.String(), "this is a errorf test"))
}
