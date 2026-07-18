package incr

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Clock is the source of time for time-dependent nodes, and can be advanced
// explicitly.
//
// Nodes that depend on the wall clock cannot be tested: the assertion has to wait for
// real time to pass, and a graph that recomputes on a schedule has no reproducible
// behavior to assert at all. A Clock separates "what time is it" from "time passes",
// so a test can step time forward and stabilize, while production advances it from the
// real clock.
//
// A Clock is safe for concurrent use. Advancing it marks the nodes whose trigger has
// passed as stale; the values change on the next stabilization, not during the
// advance.
type Clock struct {
	mu      sync.Mutex
	now     time.Time
	entries []*clockEntry
}

// clockEntry is a node's registration with the clock.
type clockEntry struct {
	node INode
	// at is when the node next needs to recompute. A zero time means the node has
	// nothing further scheduled and will not be woken again.
	at time.Time
}

// NewClock returns a clock reading start, which advances only when told to.
func NewClock(start time.Time) *Clock {
	return &Clock{now: start}
}

// Now returns the current time.
func (c *Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// Advance moves the clock to a later time and marks any node whose trigger has now
// passed as stale, so that the next stabilization recomputes it.
//
// Moving backwards is ignored: a node that has already fired cannot un-fire, so
// rewinding would leave the graph describing a past it has already moved through.
func (c *Clock) Advance(to time.Time) {
	c.mu.Lock()
	if to.Before(c.now) {
		c.mu.Unlock()
		return
	}
	c.now = to
	// Registrations are scanned rather than kept in a heap: a graph has few
	// time-dependent nodes next to its total size, and a scan keeps re-arming a
	// repeating node to a single field write.
	due := make([]INode, 0, len(c.entries))
	for _, entry := range c.entries {
		if !entry.at.IsZero() && !entry.at.After(to) {
			due = append(due, entry.node)
		}
	}
	c.mu.Unlock()

	// outside the lock, since marking a node stale reaches into the graph
	for _, node := range due {
		if graph := GraphForNode(node); graph != nil {
			graph.SetStale(node)
		}
	}
}

// AdvanceBy moves the clock forward by a duration.
func (c *Clock) AdvanceBy(d time.Duration) { c.Advance(c.Now().Add(d)) }

// register adds a node to be woken at a time, returning its entry so the node can
// re-arm itself.
func (c *Clock) register(node INode, at time.Time) *clockEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry := &clockEntry{node: node, at: at}
	c.entries = append(c.entries, entry)
	return entry
}

// rearm sets when an entry should next fire, or clears it with a zero time.
func (c *Clock) rearm(entry *clockEntry, at time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry.at = at
}

// SystemClock returns a clock that follows the wall clock.
//
// Advancing it to the current time is the caller's job -- typically a ticker calling
// [Clock.Advance] then stabilizing -- which keeps the graph's notion of time moving in
// discrete, observable steps rather than changing underneath a stabilization.
func SystemClock() *Clock { return NewClock(time.Now().UTC()) }

// At returns an incremental that is false before a time and true from that time on.
//
// The node is woken once, when the clock first reaches when.
func At(scope Scope, clock *Clock, when time.Time) Incr[bool] {
	a := &atIncr{
		clock: clock,
		when:  when,
	}
	a.n = scope.newNode(KindAt)
	WithinScope(scope, a)
	a.entry = clock.register(a, when)
	return a
}

var (
	_ Incr[bool]   = (*atIncr)(nil)
	_ IStabilize   = (*atIncr)(nil)
	_ fmt.Stringer = (*atIncr)(nil)
)

type atIncr struct {
	n     *Node
	clock *Clock
	when  time.Time
	entry *clockEntry
	value bool
}

func (a *atIncr) Node() *Node { return a.n }

func (a *atIncr) Value() bool { return a.value }

func (a *atIncr) Stabilize(_ context.Context) error {
	a.value = !a.clock.Now().Before(a.when)
	if a.value {
		// nothing further to wake for
		a.clock.rearm(a.entry, time.Time{})
	}
	return nil
}

func (a *atIncr) String() string { return a.n.String() }

// AtIntervals returns an incremental counting how many whole intervals have elapsed
// since it was created.
//
// The count is derived from the clock rather than from how often the graph
// stabilized, so a stabilization that happens late reports the interval it is late
// for, and several intervals passing between stabilizations advances the count by
// several rather than by one.
func AtIntervals(scope Scope, clock *Clock, every time.Duration) Incr[int] {
	if every <= 0 {
		panic("incr: AtIntervals requires a positive interval")
	}
	a := &atIntervalsIncr{
		clock: clock,
		start: clock.Now(),
		every: every,
	}
	a.n = scope.newNode(KindAtIntervals)
	WithinScope(scope, a)
	a.entry = clock.register(a, a.start.Add(every))
	return a
}

var (
	_ Incr[int]    = (*atIntervalsIncr)(nil)
	_ IStabilize   = (*atIntervalsIncr)(nil)
	_ fmt.Stringer = (*atIntervalsIncr)(nil)
)

type atIntervalsIncr struct {
	n     *Node
	clock *Clock
	start time.Time
	every time.Duration
	entry *clockEntry
	value int
}

func (a *atIntervalsIncr) Node() *Node { return a.n }

func (a *atIntervalsIncr) Value() int { return a.value }

func (a *atIntervalsIncr) Stabilize(_ context.Context) error {
	elapsed := a.clock.Now().Sub(a.start)
	if elapsed < 0 {
		a.value = 0
	} else {
		a.value = int(elapsed / a.every)
	}
	// arm for the interval after the one just reported, so a long gap between
	// stabilizations does not queue up a wake per interval it skipped
	a.clock.rearm(a.entry, a.start.Add(time.Duration(a.value+1)*a.every))
	return nil
}

func (a *atIntervalsIncr) String() string { return a.n.String() }

// Snapshot returns an incremental holding before until a time, and from then on the
// value its input had at the moment the clock first reached that time.
//
// The captured value does not change afterwards even though the input keeps changing,
// which is what distinguishes this from reading the input: it records what was true
// then rather than what is true now.
func Snapshot[A any](scope Scope, clock *Clock, input Incr[A], at time.Time, before A) Incr[A] {
	s := &snapshotIncr[A]{
		clock: clock,
		input: input,
		at:    at,
		value: before,
	}
	s.n = scope.newNode(KindSnapshot)
	s.parents[0] = input
	WithinScope(scope, s)
	s.entry = clock.register(s, at)
	return s
}

var (
	_ Incr[int]    = (*snapshotIncr[int])(nil)
	_ IStabilize   = (*snapshotIncr[int])(nil)
	_ IParents     = (*snapshotIncr[int])(nil)
	_ fmt.Stringer = (*snapshotIncr[int])(nil)
)

type snapshotIncr[A any] struct {
	n       *Node
	clock   *Clock
	input   Incr[A]
	at      time.Time
	entry   *clockEntry
	value   A
	taken   bool
	parents [1]INode
}

func (s *snapshotIncr[A]) Parents() []INode { return s.parents[:] }

func (s *snapshotIncr[A]) Node() *Node { return s.n }

func (s *snapshotIncr[A]) Value() A { return s.value }

// Taken reports whether the snapshot has been captured.
func (s *snapshotIncr[A]) Taken() bool { return s.taken }

func (s *snapshotIncr[A]) Stabilize(_ context.Context) error {
	if s.taken {
		// the input is still an input, so this node recomputes when it changes; the
		// captured value deliberately does not follow it.
		return nil
	}
	if s.clock.Now().Before(s.at) {
		return nil
	}
	s.value = s.input.Value()
	s.taken = true
	s.clock.rearm(s.entry, time.Time{})
	return nil
}

func (s *snapshotIncr[A]) String() string { return s.n.String() }
