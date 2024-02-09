package incr

import (
	"fmt"
	"io"
	"slices"
	"strings"
)

// Dot formats a graph from a given node in the dot format
// so that you can export the graph as an image.
//
// To do so, you'll want to make sure you have `graphviz` installed
// locally, and then you'll want to run:
//
//	> go run ??? | dot -Tpng > /home/foo/graph.png
//
// As an for an example of a program that renders a graph with `Dot`,
// look at `examples/benchmark/main.go`.
func Dot(wr io.Writer, g *Graph) (err error) {
	// NOTE(wc): a word on the below
	// basically we panic anywhere we use the `writef` helper
	// specifically where it can error.
	// we then panic if there is an error and recover here.
	defer func() {
		err, _ = recover().(error)
	}()

	writef := func(indent int, format string, args ...any) {
		_, writeErr := io.WriteString(wr, strings.Repeat("\t", indent)+fmt.Sprintf(format, args...)+"\n")
		if writeErr != nil {
			panic(writeErr)
		}
	}

	writef(0, "digraph {")
	nodes := make([]INode, 0, len(g.observed)+len(g.observers))
	for _, n := range g.observed {
		nodes = append(nodes, n)
	}
	for _, n := range g.observers {
		nodes = append(nodes, n)
	}
	slices.SortStableFunc(nodes, nodeSorter)
	nodeLabels := make(map[Identifier]string)
	for index, n := range nodes {
		nodeLabel := fmt.Sprintf("n%d", index+1)
		label := fmt.Sprintf(`label="%v" shape="rect"`, n)
		color := ` fillcolor = "white" style="filled" fontcolor="black"`
		if n.Node().setAt > 0 {
			color = ` fillcolor = "red" style="filled" fontcolor="white"`
		} else if n.Node().changedAt > 0 {
			color = ` fillcolor = "pink" style="filled" fontcolor="black"`
		}
		writef(1, "node [%s%s]; %s", label, color, nodeLabel)
		nodeLabels[n.Node().id] = nodeLabel
	}
	for _, n := range nodes {
		nodeLabel := nodeLabels[n.Node().id]
		for _, p := range n.Node().parents {
			parentLabel, ok := nodeLabels[p.Node().id]
			if ok {
				writef(1, "%s -> %s;", nodeLabel, parentLabel)
			}
		}
	}
	writef(0, "}")
	return
}

func nodeSorter(a, b INode) int {
	if a.Node().height == b.Node().height {
		aID := a.Node().ID().String()
		bID := b.Node().ID().String()
		if aID == bID {
			return 0
		} else if aID > bID {
			return -1
		}
		return 1
	} else if a.Node().height > b.Node().height {
		return -1
	}
	return 1
}
