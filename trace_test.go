package incr

import (
	"bytes"
	"strings"
	"testing"

	. "github.com/wcharczuk/go-incr/testutil"
)

func Test_WithTracingOutput(t *testing.T) {

	g := New()

	output := new(bytes.Buffer)
	errOutput := new(bytes.Buffer)

	g.SetTracer(NewTracer(output, errOutput))

	ItsNotNil(t, g.tracer)
	ItsNotNil(t, g.tracer.(*tracer).log)
	ItsNotNil(t, g.tracer.(*tracer).errLog)

	g.tracePrintln("this is a println test")
	ItsEqual(t, true, strings.Contains(output.String(), "this is a println test"))
	ItsEqual(t, "", errOutput.String())

	g.traceErrorln("this is a errorln test")
	ItsEqual(t, false, strings.Contains(output.String(), "this is a errorln test"))
	ItsEqual(t, true, strings.Contains(errOutput.String(), "this is a errorln test"))

	g.tracePrintf("this is a %s test", "printf")
	ItsEqual(t, true, strings.Contains(output.String(), "this is a printf test"))
	ItsEqual(t, false, strings.Contains(errOutput.String(), "this is a printf test"))

	g.traceErrorf("this is a %s test", "errorf")
	ItsEqual(t, false, strings.Contains(output.String(), "this is a errorf test"))
	ItsEqual(t, true, strings.Contains(errOutput.String(), "this is a errorf test"))
}
