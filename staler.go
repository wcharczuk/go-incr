package incr

// Staler is a type that can return if it's stale or not.
type Staler interface {
	Stale() bool
}
