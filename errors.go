package incr

import "errors"

var (
	// ErrAlreadyStabilizing is returned if you're already stabilizing a graph.
	ErrAlreadyStabilizing = errors.New("stabilize; already stabilizing, cannot continue")
)
