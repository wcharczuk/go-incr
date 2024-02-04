package incr

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
	nodes := make([]INode, 0, g.observed.Len()+len(g.observers))
	g.observed.Each(func(n INode) {
		nodes = append(nodes, n)
	})
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
		n.Node().children.Each(func(p INode) {
			childLabel, ok := nodeLabels[p.Node().id]
			if ok {
				writef(1, "%s -> %s;", nodeLabel, childLabel)
			}
		})
	}
	writef(0, "}")
	return
}

func homedir(filename string) string {
	return filepath.Join(os.ExpandEnv("/mnt/c/Users/wcharczuk/Desktop"), filename)
}

func dumpDot(g *Graph, path string) error {
	if os.Getenv("INCR_DEBUG_DOT") != "true" {
		return nil
	}

	dotContents := new(bytes.Buffer)
	if err := Dot(dotContents, g); err != nil {
		return err
	}
	dotOutput, err := os.Create(os.ExpandEnv(path))
	if err != nil {
		return err
	}
	defer func() { _ = dotOutput.Close() }()
	dotFullPath, err := exec.LookPath("dot")
	if err != nil {
		return err
	}

	errOut := new(bytes.Buffer)
	cmd := exec.Command(dotFullPath, "-Tpng")
	cmd.Stdin = dotContents
	cmd.Stdout = dotOutput
	cmd.Stderr = errOut
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%v; %w", errOut.String(), err)
	}
	return nil
}
