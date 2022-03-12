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
	itsNil(t, tr)

	ctx := WithTracingOutput(context.Background(), output, errOutput)
	tr = getTracer(ctx)
	itsNotNil(t, tr)
	itsNotNil(t, tr.(*tracer).log)
	itsNotNil(t, tr.(*tracer).errLog)

	tracePrintln(ctx, "this is a println test")
	itsEqual(t, true, strings.Contains(output.String(), "this is a println test"))
	itsEqual(t, "", errOutput.String())

	traceErrorln(ctx, "this is a errorln test")
	itsEqual(t, false, strings.Contains(output.String(), "this is a errorln test"))
	itsEqual(t, true, strings.Contains(errOutput.String(), "this is a errorln test"))

	tracePrintf(ctx, "this is a %s test", "printf")
	itsEqual(t, true, strings.Contains(output.String(), "this is a printf test"))
	itsEqual(t, false, strings.Contains(errOutput.String(), "this is a printf test"))

	traceErrorf(ctx, "this is a %s test", "errorf")
	itsEqual(t, false, strings.Contains(output.String(), "this is a errorf test"))
	itsEqual(t, true, strings.Contains(errOutput.String(), "this is a errorf test"))
}
