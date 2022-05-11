package incr

// Status is an incremental stabilize status.
type Status int

// Status constants
const (
	StatusNotStabilizing        Status = 0
	StatusStabilizing                  = 1
	StatusRunningUpdateHandlers        = 2
)
