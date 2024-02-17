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
	nodes := make([]INode, 0, len(g.nodes)+len(g.observers))
	for _, n := range g.nodes {
		nodes = append(nodes, n)
	}
	for _, o := range g.observers {
		nodes = append(nodes, o)
	}

	slices.SortStableFunc(nodes, nodeSorter)

	nodeLabels := make(map[Identifier]string)
	for index, n := range nodes {
		nodeLabel := fmt.Sprintf("n%d", index+1)

		var nodeInternalLabelParts []string
		nodeInternalLabelParts = append(nodeInternalLabelParts, fmt.Sprintf("%s:%s", n.Node().kind, n.Node().id.Short()))
		if n.Node().label != "" {
			nodeInternalLabelParts = append(nodeInternalLabelParts, fmt.Sprintf("label: %s", n.Node().label))
		}
		if n.Node().height != HeightUnset {
			nodeInternalLabelParts = append(nodeInternalLabelParts, fmt.Sprintf("height: %d", n.Node().height))
		}
		if value := ExpertNode(n).Value(); value != nil {
			nodeInternalLabelParts = append(nodeInternalLabelParts, fmt.Sprintf("value: %v", value))
		}
		nodeInternalLabel := strings.Join(nodeInternalLabelParts, "\n")
		label := fmt.Sprintf(`label = "%s" shape = "box3d"`, escapeForDot(nodeInternalLabel))
		color := ` fillcolor = "white" style="filled" fontcolor="black"`
		if n.Node().setAt >= (g.stabilizationNum - 1) {
			color = ` fillcolor = "red" style="filled" fontcolor="white"`
		} else if n.Node().changedAt >= (g.stabilizationNum - 1) {
			color = ` fillcolor = "pink" style="filled" fontcolor="black"`
		}
		writef(1, "node [%s%s]; %s", label, color, nodeLabel)
		nodeLabels[n.Node().id] = nodeLabel
	}
	for _, n := range nodes {
		nodeLabel := nodeLabels[n.Node().id]
		for _, p := range n.Node().children {
			childLabel, ok := nodeLabels[p.Node().id]
			if ok {
				writef(1, "%s -> %s;", nodeLabel, childLabel)
			}
		}
		for _, o := range n.Node().observers {
			childLabel, ok := nodeLabels[o.Node().id]
			if ok {
				writef(1, "%s -> %s;", nodeLabel, childLabel)
			}
		}
	}
	writef(0, "}")
	return
}

// escapeForDot escapes double quotes and backslashes, and replaces Graphviz's
// "center" character (\n) with a left-justified character.
// See https://graphviz.org/docs/attr-types/escString/ for more info.
func escapeForDot(str string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(str, `\`, `\\`),
			`"`, `\"`),
		`\l`, `\n`)
}
