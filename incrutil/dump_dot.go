package incrutil

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"github.com/wcharczuk/go-incr"
)

// DumpDot dumps a graph as accessed from a node as a png
// to a given path (which will be expanded by env).
//
// You _must_ have `graphviz` installed to use this;
// this can be installed with `brew install graphviz`
func DumpDot(root incr.INode, path string) error {
	dotContents := new(bytes.Buffer)
	if err := incr.Dot(dotContents, root); err != nil {
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

	cmd := exec.Command(dotFullPath, "-Tpng")
	cmd.Stdin = dotContents
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	_, _ = io.Copy(dotOutput, bytes.NewReader(output))
	return nil
}
