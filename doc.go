/*
Package incr implements incremental computation: a graph of values where changing an
input recomputes only what depends on it, rather than everything.

The model follows Jane Street's OCaml incremental library. You build nodes from inputs
([Var]) and combinators ([Map], [Bind], and the rest), observe the ones whose values you
want with [Observe], and call [Graph.Stabilize] to bring them up to date. A node is
recomputed only if it is observed, directly or transitively, and only if one of its
inputs has changed since it last ran.

# Choosing a combinator

Most of what decides whether an incremental graph stays fast as it grows is which
combinator you reach for. Several ways of writing the same aggregate differ by orders of
magnitude, and the difference does not show up until the collection is large:

	aggregating n inputs, cost of one input changing

	  MapN                  O(n)       every value is handed to the function
	  ReduceBalanced        O(log n)   only the path from the changed leaf to the root
	  UnorderedArrayFold    O(1)       the accumulator is adjusted in place

Measured at 4096 inputs those are 17.5us, 330ns and 112ns.

The rule of thumb:

  - Aggregating many inputs with an operation that has an inverse -- a sum, a count --
    use [UnorderedArrayFold], which withdraws the changed input's old contribution and
    applies its new one.
  - Aggregating with an operation that has no inverse -- a maximum, a minimum, a
    concatenation -- use [ReduceBalanced]. Nothing can be withdrawn from a maximum, so a
    running accumulator cannot work, but a balanced tree only recomputes one path.
  - [MapN] and [All] read every input on every pass by construction. Reach for them when
    you genuinely need all the values, not to aggregate them.
  - [Map] through [Map8] take a fixed number of inputs and are the ordinary case.

# Keyed collections

A computation over a map has the same trap in a sharper form: comparing two of Go's
builtin maps costs O(n) however few keys changed, because a hash table shares no
structure with the map it was copied from and cannot say what differs.

Package incrutil/pmap provides an immutable ordered map with structural sharing, whose
symmetric difference is proportional to the number of changes rather than to the size of
the map. Package incrutil/mapi builds the incremental operators over it: MapValues,
FilterMapValues, Merge, UnorderedFold, Reduce, Subrange, Partition, Join, and aggregates
such as Sum and Cardinality. Where a collection has many independent watchers -- a view
per row, a subscription per instrument -- mapi.Selector hands out an incremental per key,
so one key changing recomputes the consumers of that key rather than all of them. Diffing eight changes in a 65536 entry map costs about 700ns and
does not grow with the map.

# Dynamic graphs

[Bind] replaces a whole subgraph based on a value, which is how the shape of a
computation can depend on data rather than being fixed when it is built. Nodes created
inside a bind's scope are released when the bind rewrites its right-hand side. [Join]
flattens an incremental whose value is itself an incremental.

# Keeping the graph's shape stable

The largest performance decision here is whether the shape of the graph changes rather
than which combinator is used. [Bind] rebuilds its right-hand side whenever its input
changes, and each node in that subgraph is allocated, linked, walked for necessity and
heights, and torn down on the next swap. The same computation as a fixed shape that simply
recomputes measured 0.7us and no allocations against 9.1us and 34 allocations.

Reach for [Bind] when the shape genuinely depends on the data, not to express a
conditional over a known set of inputs -- that is better as a fixed graph with a cutoff, or
[MapIf]. incrutil.BindMemoized caches right-hand sides by key when they are expensive to
build, though it reduces allocation rather than time, since a cached subgraph is still
relinked on every swap.

# Cutoffs

By default a node propagates whenever it recomputes, including when its new value equals
its old one, so writing a var the value it already holds costs a full propagation.

There is no graph-wide setting to change that, and there cannot be: deciding whether two
values are equal requires them to be comparable, and an [Incr][A] holds any type at all.
The constraint has to appear where the type is known, which means at construction:

  - [VarEqual] is a var that ignores being set to the value it already holds. This is the
    cheapest place to stop a no-op write, because nothing downstream is consulted at all
    -- for an input with a few thousand dependents the difference is microseconds against
    nanoseconds.
  - [CutoffEqual] stops propagation further down the graph, for values that are computed
    rather than set. [Cutoff] and [Cutoff2] take an arbitrary predicate, for when
    "changed enough to matter" is a judgement rather than an equality.
  - [CutoffAlways] and [CutoffNever] are the degenerate cases, worth naming so a graph can
    say it means them.

[VarEqualFunc] and [CutoffEqualFunc] take the comparison, for types with no ==.

# Time

Nodes that read the wall clock are hard to test and hard to reason about, since two
nodes in one pass can see different instants. [Clock] separates what time it is from
time passing: it is advanced explicitly with [Clock.Advance], wakes only the nodes whose
trigger has passed, and holds still for the duration of a stabilization. [At],
[AtIntervals], [Snapshot] and [StepFunction] are built on it. A step function is how to
express something scheduled -- a rate that changes at market open, a limit that relaxes
overnight -- as an input rather than as something a caller has to remember to poll.

# Node lifecycle

A node can stop being part of the computation without its value ever changing, which is
invisible through update handlers. [Node.OnBecameNecessary], [Node.OnBecameUnnecessary]
and [Node.OnInvalidated] report those transitions, and are how a node releases something
it was holding on the graph's behalf.

# Errors and cancellation

If a node's stabilization returns an error the pass stops and the error is returned. The
node that failed and every node not yet reached stay in the recompute heap, so a later
pass tries them again: a transient failure does not strand a value. Canceling the
context has the same shape and returns the context's cause, which makes a long pass over a
large graph interruptible. It stops within a bounded amount of work rather than instantly:
serial stabilization checks every 64 nodes, since checking every node would cost a
measurable share of a cheap node's recompute, and parallel stabilization checks once per
height block.

A panic in a node's computation is reported the same way, as a [PanicError] carrying the
panic value, the stack, and the node responsible. This is not only for convenience: under
[Graph.ParallelStabilize] nodes are recomputed in worker goroutines, where a panic cannot
be recovered by the caller and would otherwise end the process.

# Stabilizing in parallel

[Graph.ParallelStabilize] recomputes nodes at the same height concurrently. It is worth
using when individual node computations are expensive; for cheap nodes the coordination
costs more than it saves, since serial stabilization can skip the locking entirely.
*/
package incr
