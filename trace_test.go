package incr

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func Test_WithTracingOutput(t *testing.T) {
	output := new(bytes.Buffer)
	errOutput := new(bytes.Buffer)

	tr := getTracer(context.Background())
	ItsNil(t, tr)

	ctx := WithTracingOutputs(context.Background(), output, errOutput)
	tr = getTracer(ctx)
	ItsNotNil(t, tr)
	ItsNotNil(t, tr.(*tracer).log)
	ItsNotNil(t, tr.(*tracer).errLog)

	tracePrintln(ctx, "this is a println test")
	ItsEqual(t, true, strings.Contains(output.String(), "this is a println test"))
	ItsEqual(t, "", errOutput.String())

	traceErrorln(ctx, "this is a errorln test")
	ItsEqual(t, false, strings.Contains(output.String(), "this is a errorln test"))
	ItsEqual(t, true, strings.Contains(errOutput.String(), "this is a errorln test"))

	tracePrintf(ctx, "this is a %s test", "printf")
	ItsEqual(t, true, strings.Contains(output.String(), "this is a printf test"))
	ItsEqual(t, false, strings.Contains(errOutput.String(), "this is a printf test"))

	traceErrorf(ctx, "this is a %s test", "errorf")
	ItsEqual(t, false, strings.Contains(output.String(), "this is a errorf test"))
	ItsEqual(t, true, strings.Contains(errOutput.String(), "this is a errorf test"))
}