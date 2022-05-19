package incr

import (
	"fmt"
	"io"
	"strings"
)

// Dot formats a graph from a given node in the dot format
// so that you can export the graph as an image.
//
// To do so, you'll want to make sure you have `graphviz` installed
// locally, and then you'll want to run:
//
//    > go run ??? | dot -Tpng > /home/foo/graph.png
//
// As an for an example of a program that renders a graph with `Dot`,
// look at `examples/benchmark/main.go`.
func Dot(wr io.Writer, node INode) (err error) {
	defer func() {
		err, _ = recover().(error)
	}()

	writeln := func(indent int, format string, args ...any) {
		_, writeErr := io.WriteString(wr, strings.Repeat("\t", indent)+fmt.Sprintf(format, args...)+"\n")
		if writeErr != nil {
			panic(writeErr)
		}
	}

	writeln(0, "digraph {")
	nodes := dotDiscoverNodes(node)
	nodeLabels := make(map[Identifier]string)
	for index, n := range nodes {
		nodeLabel := fmt.Sprintf("n%d", index+1)
		label := fmt.Sprintf(`label="%v"`, n)
		color := ` fillcolor = "white" style="filled" fontcolor="black"`
		if n.Node().setAt > 0 {
			color = ` fillcolor = "red" style="filled" fontcolor="white"`
		} else if n.Node().changedAt > 1 {
			color = ` fillcolor = "pink" style="filled" fontcolor="white"`
		}
		writeln(1, "node [%s%s]; %s", label, color, nodeLabel)
		nodeLabels[n.Node().id] = nodeLabel
	}
	for _, n := range nodes {
		nodeLabel := nodeLabels[n.Node().id]
		for _, p := range n.Node().parents {
			parentLabel := nodeLabels[p.Node().id]
			writeln(1, "%s -> %s;", nodeLabel, parentLabel)
		}
	}
	writeln(0, "}")
	return
}

func dotDiscoverNodes(node INode) (output []INode) {
	seen := make(set[Identifier])
	output = dotDiscoverNodesVisit(seen, node)
	return
}

func dotDiscoverNodesVisit(seen set[Identifier], node INode) (output []INode) {
	if seen.has(node.Node().id) {
		return
	}
	seen.add(node.Node().id)
	output = append(output, node)
	for _, c := range node.Node().children {
		output = append(output, dotDiscoverNodesVisit(seen, c)...)
	}
	for _, p := range node.Node().parents {
		output = append(output, dotDiscoverNodesVisit(seen, p)...)
	}
	return
}
