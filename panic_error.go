package incr

import (
	"fmt"
	"runtime/debug"
)

// PanicError is returned by [Graph.Stabilize] and [Graph.ParallelStabilize] when a node's
// computation panicked.
//
// A panic in user code is turned into an error rather than being allowed to unwind through
// the stabilization. Letting it through would leave the recompute heap mid-pass with no
// record of which node was responsible, and under [Graph.ParallelStabilize] it would end
// the process outright, since a panic in a worker goroutine cannot be recovered by the
// caller. As an error it is reported like any other failure: the node is returned to the
// recompute heap, and a later pass tries it again.
//
// The panic value and the stack as it was at the point of the panic are both kept, because
// the stack is the whole value of a panic and recovering without it would trade a crash for
// a mystery.
type PanicError struct {
	// Value is what was passed to panic.
	Value any
	// Stack is the stack trace captured where the panic was recovered.
	Stack []byte
	// Node describes the node whose computation panicked, or is empty when the panic did
	// not come from a node.
	Node string
}

func (pe *PanicError) Error() string {
	if pe.Node == "" {
		return fmt.Sprintf("incr: panicked during stabilization: %v\n%s", pe.Value, pe.Stack)
	}
	return fmt.Sprintf("incr: node %s panicked: %v\n%s", pe.Node, pe.Value, pe.Stack)
}

// Unwrap returns the panic value when it was itself an error, so that a caller can match
// against it with [errors.Is] and [errors.As].
func (pe *PanicError) Unwrap() error {
	err, _ := pe.Value.(error)
	return err
}

// newPanicError captures a recovered panic. It is only reached when a panic actually
// happened, so the cost of taking a stack trace here does not appear on the hot path.
func newPanicError(n INode, recovered any) *PanicError {
	pe := &PanicError{
		Value: recovered,
		Stack: debug.Stack(),
	}
	if n != nil {
		pe.Node = n.Node().String()
	}
	return pe
}
