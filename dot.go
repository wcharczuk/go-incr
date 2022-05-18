package incr

import (
	"fmt"
	"io"
	"strings"
)

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
		label := fmt.Sprintf("n%d", index+1)
		writeln(1, "node [label=\"%v\"]; %s", n, label)
		nodeLabels[n.Node().id] = label
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
