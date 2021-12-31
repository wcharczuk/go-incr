package incr

var (
	_ error = (*Error)(nil)
)

// Error is an error
type Error string

// Error implements error
func (e Error) Error() string { return string(e) }

const (
	// ErrStabilizeCycle is returned if a cycle is detected during Stabilize.
	ErrStabilizeCycle Error = "cycle detected"
)
