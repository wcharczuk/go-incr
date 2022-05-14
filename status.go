package incr

// Status constants
const (
	StatusNotStabilizing int32 = iota
	StatusStabilizing
	StatusRunningUpdateHandlers
)
