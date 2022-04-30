package incr

// Status is an incremental stabilize status.
type Status int

// Status constants
const (
	StatusStabilizing Status = iota
	StatusRunningUpdateHandlers
	SStatusNotStabilizing
)
