package incr

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Dot(t *testing.T) {
	golden := `digraph {
	node [label="map2[563b29ae]" shape="rect" fillcolor = "white" style="filled" fontcolor="black"]; n1
	node [label="var[884f9774]" shape="rect" fillcolor = "white" style="filled" fontcolor="black"]; n2
	node [label="var[7f8a4e21]" shape="rect" fillcolor = "white" style="filled" fontcolor="black"]; n3
	n2 -> n1;
	n3 -> n1;
}
`
	v0 := Var("foo")
	v0.Node().id, _ = ParseIdentifier("165382c219e24e3db77fd41a884f9774")
	v1 := Var("bar")
	v1.Node().id, _ = ParseIdentifier("a985936bed8c48b99801a5bd7f8a4e21")
	m0 := Map2(v0, v1, concat)
	m0.Node().id, _ = ParseIdentifier("fc45f4a7b5c7456f852f2298563b29ae")

	buf := new(bytes.Buffer)
	err := Dot(buf, m0)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, golden, buf.String())
}

type errorWriter struct {
	e error
}

func (ew errorWriter) Write(_ []byte) (int, error) {
	return 0, ew.e
}

func Test_Dot_writeError(t *testing.T) {
	v0 := Var("foo")
	v0.Node().id, _ = ParseIdentifier("165382c219e24e3db77fd41a884f9774")
	v1 := Var("bar")
	v1.Node().id, _ = ParseIdentifier("a985936bed8c48b99801a5bd7f8a4e21")
	m0 := Map2(v0, v1, concat)
	m0.Node().id, _ = ParseIdentifier("fc45f4a7b5c7456f852f2298563b29ae")

	buf := errorWriter{fmt.Errorf("this is just a test")}
	err := Dot(buf, m0)
	testutil.ItsNotNil(t, err)
}

func Test_Dot_setAt(t *testing.T) {
	golden := `digraph {
	node [label="map2[563b29ae]" shape="rect" fillcolor = "white" style="filled" fontcolor="black"]; n1
	node [label="var[884f9774]" shape="rect" fillcolor = "white" style="filled" fontcolor="black"]; n2
	node [label="var[7f8a4e21]" shape="rect" fillcolor = "red" style="filled" fontcolor="white"]; n3
	n2 -> n1;
	n3 -> n1;
}
`
	v0 := Var("foo")
	v0.Node().id, _ = ParseIdentifier("165382c219e24e3db77fd41a884f9774")
	v1 := Var("bar")
	v1.Node().id, _ = ParseIdentifier("a985936bed8c48b99801a5bd7f8a4e21")
	v1.Node().setAt = 1
	m0 := Map2(v0, v1, concat)
	m0.Node().id, _ = ParseIdentifier("fc45f4a7b5c7456f852f2298563b29ae")

	buf := new(bytes.Buffer)
	err := Dot(buf, m0)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, golden, buf.String())
}

func Test_Dot_changedAt(t *testing.T) {
	golden := `digraph {
	node [label="map2[563b29ae]" shape="rect" fillcolor = "pink" style="filled" fontcolor="black"]; n1
	node [label="var[884f9774]" shape="rect" fillcolor = "white" style="filled" fontcolor="black"]; n2
	node [label="var[7f8a4e21]" shape="rect" fillcolor = "white" style="filled" fontcolor="black"]; n3
	n2 -> n1;
	n3 -> n1;
}
`
	v0 := Var("foo")
	v0.Node().id, _ = ParseIdentifier("165382c219e24e3db77fd41a884f9774")
	v1 := Var("bar")
	v1.Node().id, _ = ParseIdentifier("a985936bed8c48b99801a5bd7f8a4e21")
	m0 := Map2(v0, v1, concat)
	m0.Node().id, _ = ParseIdentifier("fc45f4a7b5c7456f852f2298563b29ae")
	m0.Node().changedAt = 1

	buf := new(bytes.Buffer)
	err := Dot(buf, m0)
	testutil.ItsNil(t, err)
	testutil.ItsEqual(t, golden, buf.String())
}
