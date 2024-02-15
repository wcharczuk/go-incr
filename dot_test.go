package incr

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Dot(t *testing.T) {
	golden := `digraph {
	node [label="observer[033257be]@2" shape="rect" fillcolor = "white" style="filled" fontcolor="black"]; n1
	node [label="map2[563b29ae]@1" shape="rect" fillcolor = "white" style="filled" fontcolor="black"]; n2
	node [label="var[7f8a4e21]@0" shape="rect" fillcolor = "white" style="filled" fontcolor="black"]; n3
	node [label="var[884f9774]@0" shape="rect" fillcolor = "white" style="filled" fontcolor="black"]; n4
	n2 -> n1;
	n3 -> n2;
	n4 -> n2;
}
`

	g := New()
	v0 := Var(g, "foo")
	ExpertNode(v0).SetID(MustParseIdentifier("165382c219e24e3db77fd41a884f9774"))
	v1 := Var(g, "bar")
	ExpertNode(v1).SetID(MustParseIdentifier("a985936bed8c48b99801a5bd7f8a4e21"))
	m0 := Map2(g, v0, v1, concat)
	ExpertNode(m0).SetID(MustParseIdentifier("fc45f4a7b5c7456f852f2298563b29ae"))
	o := MustObserve(g, m0)
	ExpertNode(o).SetID(MustParseIdentifier("507dd07419724979bb34f2ca033257be"))

	buf := new(bytes.Buffer)
	err := Dot(buf, g)
	testutil.Nil(t, err)
	testutil.Equal(t, golden, buf.String())
}

type errorWriter struct {
	e error
}

func (ew errorWriter) Write(_ []byte) (int, error) {
	return 0, ew.e
}

func Test_Dot_writeError(t *testing.T) {
	g := New()
	v0 := Var(g, "foo")
	ExpertNode(v0).SetID(MustParseIdentifier("165382c219e24e3db77fd41a884f9774"))
	v1 := Var(g, "bar")
	ExpertNode(v1).SetID(MustParseIdentifier("a985936bed8c48b99801a5bd7f8a4e21"))
	m0 := Map2(g, v0, v1, concat)
	ExpertNode(m0).SetID(MustParseIdentifier("fc45f4a7b5c7456f852f2298563b29ae"))
	o := MustObserve(g, m0)
	ExpertNode(o).SetID(MustParseIdentifier("507dd07419724979bb34f2ca033257be"))

	_ = MustObserve(g, m0)

	buf := errorWriter{fmt.Errorf("this is just a test")}
	err := Dot(buf, g)
	testutil.NotNil(t, err)
}

func Test_Dot_setAt(t *testing.T) {
	golden := `digraph {
	node [label="observer[033257be]@2" shape="rect" fillcolor = "white" style="filled" fontcolor="black"]; n1
	node [label="map2[563b29ae]@1" shape="rect" fillcolor = "white" style="filled" fontcolor="black"]; n2
	node [label="var[7f8a4e21]@0" shape="rect" fillcolor = "red" style="filled" fontcolor="white"]; n3
	node [label="var[884f9774]@0" shape="rect" fillcolor = "white" style="filled" fontcolor="black"]; n4
	n2 -> n1;
	n3 -> n2;
	n4 -> n2;
}
`

	g := New()
	v0 := Var(g, "foo")
	ExpertNode(v0).SetID(MustParseIdentifier("165382c219e24e3db77fd41a884f9774"))
	v1 := Var(g, "bar")
	ExpertNode(v1).SetID(MustParseIdentifier("a985936bed8c48b99801a5bd7f8a4e21"))
	v1.Node().setAt = 1
	m0 := Map2(g, v0, v1, concat)
	ExpertNode(m0).SetID(MustParseIdentifier("fc45f4a7b5c7456f852f2298563b29ae"))
	o := MustObserve(g, m0)
	ExpertNode(o).SetID(MustParseIdentifier("507dd07419724979bb34f2ca033257be"))

	buf := new(bytes.Buffer)
	err := Dot(buf, g)
	testutil.Nil(t, err)
	testutil.Equal(t, golden, buf.String())
}

func Test_Dot_changedAt(t *testing.T) {
	golden := `digraph {
	node [label="observer[033257be]@2" shape="rect" fillcolor = "white" style="filled" fontcolor="black"]; n1
	node [label="map2[563b29ae]@1" shape="rect" fillcolor = "pink" style="filled" fontcolor="black"]; n2
	node [label="var[7f8a4e21]@0" shape="rect" fillcolor = "white" style="filled" fontcolor="black"]; n3
	node [label="var[884f9774]@0" shape="rect" fillcolor = "white" style="filled" fontcolor="black"]; n4
	n2 -> n1;
	n3 -> n2;
	n4 -> n2;
}
`

	g := New()
	v0 := Var(g, "foo")
	ExpertNode(v0).SetID(MustParseIdentifier("165382c219e24e3db77fd41a884f9774"))
	v1 := Var(g, "bar")
	ExpertNode(v1).SetID(MustParseIdentifier("a985936bed8c48b99801a5bd7f8a4e21"))
	m0 := Map2(g, v0, v1, concat)
	ExpertNode(m0).SetID(MustParseIdentifier("fc45f4a7b5c7456f852f2298563b29ae"))
	m0.Node().changedAt = 1
	o := MustObserve(g, m0)
	ExpertNode(o).SetID(MustParseIdentifier("507dd07419724979bb34f2ca033257be"))

	buf := new(bytes.Buffer)
	err := Dot(buf, g)
	testutil.Nil(t, err)
	testutil.Equal(t, golden, buf.String())
}
