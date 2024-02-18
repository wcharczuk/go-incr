package incr

// Status constants for graph stabilization status.
const (
	StatusNotStabilizing int32 = iota
	StatusStabilizing
	StatusRunningUpdateHandlers
)
