`go-incr`
==============

[![Continuous Integration](https://github.com/wcharczuk/go-incr/actions/workflows/ci.yml/badge.svg)](https://github.com/wcharczuk/go-incr/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/wcharczuk/go-incr)](https://goreportcard.com/report/github.com/wcharczuk/go-incr)

![Graph](https://github.com/wcharczuk/go-incr/blob/main/_assets/basic_graph.png)

`go-incr` is an optimization tool to help implement partial computation of large graphs of operations.

It's useful in situations where you want to efficiently compute the outputs of computations where only a subset of the computation changes based on changes in inputs, but also specifically when the graph might be _dynamic_ over the course of the computation's lifetime.

Think Excel spreadsheets and formulas, but in code, which can change over time.

# Important caveat

Very often code is faster _without_ incremental. Only reach for this library if you've hit a wall and need to add incrementality.

# Inspiration & differences from the original

The inspiration for `go-incr` is Jane Street's [incremental](https://github.com/janestreet/incremental) library.

Three differences are worth knowing if you are coming from that library:

- **No cutoff by default.** The original stops propagating when a recomputed value is
  physically equal to the previous one. This one cannot: an `Incr[A]` holds any type, so
  equality is not available to it. Ask for it where the type is known, with `VarEqual` on an
  input or `CutoffEqual` further down. This is the difference most likely to surprise you,
  and the one most likely to cost you performance if you miss it.
- **Parallelism.** `ParallelStabilize` recomputes nodes of the same height concurrently,
  which is worth reaching for when individual nodes are expensive -- network calls, or other
  work that is not CPU bound. For cheap nodes the coordination costs more than it saves.
- **Errors, contexts and panics.** A node's computation can take a `context.Context` and
  return an error; a failing or panicking node stops the pass and is retried on the next
  one. See "Error handling and context propagation" below.

# Performance relative to the original

The two libraries implement the same algorithms. The benchmark suite in `_bench` refuses to
report any timing until both harnesses agree on every observed value, so a graph that
stabilized nothing cannot look fast rather than wrong. It also has a separate `-stats` mode
that reports how many nodes each shape recomputes; those counts agree except in three
understood places, listed in `_bench/ALGORITHMS.md`. What is left is constant factors, and
they fall differently depending on the shape:

| shape | vs OCaml `incremental` |
| --- | --- |
| `Bind` swapping a large subgraph | go-incr 1.25x faster |
| writing an input the value it already holds | go-incr 2.9x faster with `VarEqual` |
| every input changes, wide graph | go-incr 1.15x faster |
| building a graph | go-incr 1.1x faster; the widest case straddles parity |
| one input changes, shallow graph | go-incr 1.1x faster |
| `Bind` swapping a small subgraph | 1.1 - 1.45x slower |
| one input changes, wide graph | 1.2x slower at 1k nodes, 1.35x at 64k |
| one input changes, deep graph | 1.3 - 1.4x slower |
| writing the same value through a plain `Var` | 4.3 - 5.6x slower |

The last two rows are the same difference in defaults, seen from both sides. OCaml
`incremental` cuts off on physical equality by default; go-incr propagates unless you ask it
not to, because an `Incr[A]` holds any type and equality cannot be assumed. Through a plain
`Var` that costs 4.3 - 5.6x. Through `VarEqual` a no-op write costs 26ns against the reference's
74ns, because the write is rejected before anything downstream is consulted at all.

Where go-incr is ahead is allocation-bound work and fixed per-pass cost. Node metadata is
carved from contiguous per-scope chunks rather than allocated one at a time, and a bind reuses
its chunks across rebuilds, so building a graph and swapping a large subgraph both cost less
than the reference and nodes created together sit next to each other in memory. A
stabilization that finds nothing to do costs 22ns.

Where it is behind is propagating a *single* input change through many nodes, and that one
gets worse as the graph grows: 1.2x at a thousand nodes, 1.35x at sixty-five thousand. Nothing
else in the suite does that -- construction and bulk updates hold their ratio flat across a 64x
size range. The cause is footprint rather than work, and both libraries recompute the same set
of nodes. A single-input change walks one node per tree level, so at 64k nodes it touches
eighteen nodes scattered through about 86MB, and nearly every one of those is a cache miss.
`Node` is 384 bytes here and the reference's node record is smaller, so the same walk crosses
less memory. Padding `Node` by 128 bytes and re-measuring confirms it: no effect at a thousand
nodes, 8-13% at sixteen and sixty-five thousand.

So if your graphs are large and your updates touch one input at a time, that is the shape
where this library is furthest behind, and per-node footprint is the lever rather than
per-node instruction count.

`_bench/ALGORITHMS.md` has the full analysis, including several optimizations that measured
worse and were reverted, which are usually more informative than the ones that landed.

If a workload churns the graph heavily and cares about throughput rather than latency,
`debug.SetGCPercent(400)` is worth about 11% on bind swaps and 6% on construction, ideally
with `debug.SetMemoryLimit` to bound the resulting peak. Note that turning collection off
entirely is *slower* than the default, by 11-26%: nothing gets reused, the heap grows, and
the working set stops fitting. This library does not tune the collector on a caller's behalf.

Two caveats if you run the suite yourself. Compare interleaved and never across sessions:
the same unmodified binary drifts by up to 2x between runs on the same machine, which is
enough to invent a regression or hide one. And establish the noise floor before believing a
delta -- building identical source twice and comparing gives 1-2%, and `wide/update_all` at
16k is unstable enough on its own to need a distribution rather than a single minimum. The
numbers above are the agreement of two full runs on an otherwise idle machine.

# Choosing a combinator

Most of what decides whether an incremental graph stays fast as it grows is which
combinator you reach for. Several ways of writing the same aggregate differ by orders of
magnitude, and the difference does not appear until the collection is large:

| aggregating n inputs | cost of one input changing | at 4096 inputs |
| --- | --- | --- |
| `MapN` | O(n) — every value is passed to the function | 17.5us |
| `ReduceBalanced` | O(log n) — only the path from the changed leaf to the root | 330ns |
| `UnorderedArrayFold` | O(1) — the accumulator is adjusted in place | 112ns |

The rule of thumb:

- Aggregating with an operation that **has an inverse** (a sum, a count) — use
  `UnorderedArrayFold`, which withdraws the old contribution and applies the new one.
- Aggregating with an operation that has **no inverse** (a maximum, a concatenation) —
  use `ReduceBalanced`. Nothing can be withdrawn from a maximum, so a running
  accumulator cannot work, but a balanced tree only recomputes one path.
- `MapN` and `All` read every input on every pass by construction. Use them when you
  genuinely want all the values, not to aggregate them.

# Incremental maps

A computation over a map has the same trap in a sharper form: comparing two of Go's
builtin maps costs O(n) however few keys changed, because a hash table shares no
structure with the map it was copied from and cannot say what differs.

`incrutil/pmap` is an immutable ordered map with structural sharing, whose symmetric
difference is proportional to the number of changes rather than the size of the map —
diffing 8 changes in a 65536 entry map costs about 700ns and does not grow with the map.
`incrutil/mapi` builds the incremental operators over it:

| operator | does |
| --- | --- |
| `MapValues`, `FilterMapValues` | per-key transform, recomputing only changed keys |
| `Merge` | combine two maps, recomputing the union of their changes |
| `UnorderedFold` | aggregate with an inverse, O(1) per changed key |
| `Reduce`, `MaxValue`, `MinValue` | aggregate without an inverse, O(log n) per change |
| `Subrange` | a window over a sorted map, with incremental bounds |
| `Partition` | split by a predicate into two maps |
| `Join` | a map of incrementals becomes an incremental map |
| `Selector` | an incremental per key, so one key changing wakes only that key's consumers |
| `Sum`, `Cardinality`, `Counti`, `Keys` | common aggregates |

`FromGoMap` and `ToGoMap` bridge to builtin maps at the edges.

`examples/incremental_map` is a worked example: a book of 5000 orders with a per-order
transform, a running total, a maximum and a moving window over it. Repricing one order
recomputes one line and adjusts the total in constant time; filling the largest order
finds the new maximum without rereading the book.

# Keeping the graph's shape stable

The single largest performance decision in this library is not which combinator you pick,
it is whether the shape of the graph changes. `Bind` rebuilds its right-hand side every
time its input changes, and every node in that subgraph is allocated, linked, walked for
necessity and heights, then torn down again on the next swap.

The same computation, once as a bind that rebuilds and once as a fixed shape whose nodes
are always present and simply recompute:

    bind rebuild    9.1us    34 allocations
    stable shape    0.7us     0 allocations

13x the time, and the stable shape allocates nothing at all once it is built. Nothing else
in this library is worth that much.

So reach for `Bind` when the shape genuinely depends on the data -- when you cannot know
the set of nodes until you see a value -- and not as a way to express a conditional. A
condition over a known set of inputs is better as a fixed graph plus a cutoff, or as
`MapIf`. If the right-hand side is expensive to build and the input takes a small set of
values, `incrutil.BindMemoized` caches the subgraphs by key; note that it reduces
allocation rather than time, since a cached subgraph still has to be relinked on every swap
-- measured, it was slower than rebuilding.

Two related habits, both measured elsewhere in this file: prefer `UnorderedArrayFold` or
`ReduceBalanced` to `MapN` for aggregates, and use `VarEqual` for inputs that are written
repeatedly with the value they already hold.

# Cutoffs

By default a node propagates whenever it recomputes, including when its new value equals
its old one — so writing a var the value it already holds costs a full propagation. There
is no graph-wide setting for this and there cannot be: comparing two values requires them
to be comparable, and an `Incr[A]` holds any type. The constraint appears at construction
instead:

- `VarEqual` ignores being set to the value it already holds. This is the cheapest place
  to stop a no-op write, since nothing downstream is consulted: measured over 4096
  dependents, 329us for a plain `Var` against 87ns. (The performance section above quotes
  26ns for the same operation; that is a different graph -- a no-op write with a handful of
  dependents rather than a few thousand.)
- `CutoffEqual` stops propagation further down, for computed rather than set values.
- `Cutoff`/`Cutoff2` take an arbitrary predicate, for when "changed enough to matter" is a
  judgement rather than equality.

`VarEqualFunc` and `CutoffEqualFunc` take the comparison, for types with no `==`.

# Time

Nodes that read the wall clock are hard to test and hard to reason about, since two nodes
in one pass can see different instants. `Clock` separates what time it is from time
passing: advance it explicitly with `Advance`, and it wakes only the nodes whose trigger
has passed while holding still for the duration of a stabilization. `At`, `AtIntervals`,
`Snapshot` and `StepFunction` are built on it.

For code that reads the real clock — including the older `Timer` node — `testing/synctest`
makes it testable without an abstraction; see `timer_synctest_test.go`.

# Node lifecycle

A node can stop being part of the computation without its value ever changing, which is
invisible through update handlers. `Node.OnBecameNecessary`, `OnBecameUnnecessary` and
`OnInvalidated` report those transitions, and are how a node releases something it was
holding on the graph's behalf.

# Examples

Each of these is a runnable program that reports what work it actually did, rather than
just producing an answer.

| example | shows |
| --- | --- |
| `examples/aggregation` | the cost of `MapN` vs `ReduceBalanced` vs `UnorderedArrayFold` on one workload: 4096 function calls vs 12 vs 1 |
| `examples/build_graph` | an incremental build: a cutoff on canonical content so a comment-only edit rebuilds nothing, and `Bind` for dependencies discovered by parsing |
| `examples/incremental_map` | a book of 5000 orders with a per-order transform, running total, maximum and moving window over it |
| `examples/monitoring` | staleness detection and alerting driven by `Clock`, stepping through half an hour deterministically |
| `examples/subscriptions` | lifecycle handlers opening and closing real resources as a dashboard's panels change |
| `examples/basic`, `examples/orders`, `examples/streaming` | smaller introductions to the core API |

# Basic concepts

A computation in `go-incr` is composed of three "meta" node types.

- "Input" types that act as entry-points for data that we'll need in our computation (think, raw cell values in Excel)
- "Mutator" types that take other nodes as inputs, and transform those input values into new values (think, formulas in Excel) or even new sub-graph components.
- "Observe" type that indicates which values we care about, and ultimately want to display or return.

"Input" types for `go-incr` are `Var` and `Return` typically.

"Mutator" types for `go-incr` are nodes like `Map` and `Bind`.

The observer node is special in that there is only (1) type of observer, and it has a very specific role and cannot be passed as an input to other nodes.

# Usage

A mini-worked example:

```go
import "github.com/wcharczuk/go-incr"
...
g := incr.New()
v0 := incr.Var(g, "hello")
v1 := incr.Var(g, "world!")
m := incr.Map2(g, v0, v1, func(a, b string) string {
  return a+" "+b
})
om := incr.MustObserve(g, m)
_ = g.Stabilize(context.Background())
fmt.Println(om.Value()) // prints "hello world!"
```

This is not intended as a "real" use case for the library, simply as a worked example to show the syntax.

You can see more sample use cases in the `examples/` directory in this repository.

# API compatibility guarantees

As of v1.xxx you should assume that the functions and types exported by this library will maintain forward compatibility until some future v2 necessitates changing things meaningfully, at which point we'll integrate [semantic import versioning](https://go.googlesource.com/proposal/+/master/design/24301-versioned-go.md) to create a new `/v2/` package. The goal will be to put off a v2 for as long as possible.

An exception to the preceeding are the `incr.Expert...` functions, for which there are no guarantees and the API may change between refs without notice.

# A word on how to use this library effectively

It can be tempting to, when looking at what this library can do, make every operation in a slow piece of code incrementally computed.

This will typically make performance _worse_, as making a computation incrementally computed adds overhead for each operation.

A more ideal balance is to write your code as normal assuming nothing is incrementally computed, then do a coarsed-grain pass, specifically breaking up chunks that are mostly atomic, and making those chunks incrementally computed.

# Error handling and context propagation

Every combinator has a `...Context` form -- `MapContext`, `Map2Context` through
`Map8Context`, `MapNContext`, `BindContext`, `CutoffContext`, `SentinelContext` -- whose
function takes a `context.Context` and may return an error. The plain forms are for
computations that cannot fail.

**An error halts the pass.** Stabilizing serially returns immediately and recomputes nothing
further; stabilizing in parallel finishes the current height block and then stops.

**A failed node is retried.** It goes back on the recompute heap, so the next pass tries it
again even if no input changed. A transient failure -- a request that timed out, a query that
deadlocked -- does not strand the node's old value. `OptGraphClearRecomputeHeapOnError(true)`
opts out, abandoning the pass instead and running the aborted handlers.

**A panic becomes an error.** A panic in your code is reported as a `*PanicError` carrying the
panic value, the stack, and the node responsible, and the node is retried like any other
failure. `errors.As` gets you the `*PanicError`; `errors.Is` sees through to the panicked
value when it was itself an error. This matters most under `ParallelStabilize`, where nodes
run in worker goroutines and a panic could not be recovered by the caller at all.

**Cancelling the context stops the pass.** It returns the context's cause, and the nodes not
yet recomputed stay on the heap, so the state left behind is the same as an error abort and
stabilizing again continues from where it stopped. A context that cannot be cancelled costs
nothing.

Per-node handlers are available for all of this: `Node().OnError`, `OnAborted`, and
`OnUpdate`. Graph-wide, `OnStabilizationStart` and `OnStabilizationEnd` bracket each pass.

# Design Choices

There is some consideration with this library on the balance between hiding mutable implemenation details to protect against [Hyrum's Law](https://www.hyrumslaw.com/) issues, and surfacing enough utility helpers to allow users to extend this library for their own use cases (specifically through `incr.Expert...` types.)

As a result, most of the features of this library can be leveraged externally, but little of the internal state and recomputation mechanism is exposed to the user.

Specific implications of this are, the `INode` interface includes a function that returns the `Node` metadata, but this `Node` struct has few exported fields on it, and users of this library should not really concern themselves with what's on it, just that it gets supplied to `Incr` through the interface implementation.

# Implementation details

To determine which nodes to recompute `go-incr` uses a partial-order pseudo-height adjacency list similar to the Jane Street implementation. This offers "good enough" approximation of a heap while also allowing for fast iteration (e.g. faster than a traditional heap's O(log(n)) performance).

The underlying assumption is that nodes with the same pseudo-height _do not depend on each other_, and as a result can be recomputed in parallel.

# A word on `Bind`

The `Bind` incremental is by far the most powerful, and as a result complicated, node type in the ecosystem.

To explain its purpose; there are situations where you want to dynamically swap out parts of a graph rooted at a given node. A concrete example might be in a user interface swapping out the displayed controls, which may be made up of individual computations.

The effect of `Bind` is that "children" of a `Bind` node may have their heights change significantly depending on the changes made to the "bound" nodes, and this has implications for the recomputation heap as a result, specifically that it has to handle updating the reflected height of nodes that may already be in the heap.

`Bind` nodes may also return `Bind` nodes themselves, creating fun and interesting implications for how an inner right-hand-side incremental needs to be propagated through multiple layers of binds, and reflect both its and the original bind children's true height through recomputations. To do this, we adopt the trick from the main ocaml library of creating two new pieces of state; a `bind-lhs-change` node that links the original bind input, and the right-hand-side incremental of the bind, making sure that the rhs respects the height of the transitive dependency of the bind's input. We also maintain a "scope", or a list of all the nodes that were created in respect to the rhs, and when the outer bind stabilizes, we also propagate changes to inner binds, regardless of where they are in the rhs graph.

If this sounds complicated, it is!

# A word on `Scopes`

Because `Bind` nodes rely on scopes to operate correctly, and specifically know which nodes it was responsible for creating, the function you must provide to the `Bind` node constructor is passed `Scope` argument. This scope argument should be passed to node constructors within the bind function.

An example of a use case for bind might be:

```
g := incr.New()
t1 := incr.Map(g, Return(g, "hello"), func(v string) string { return v + " world!" })
t2v := incr.Var(g, "a")
t2 := incr.Bind(g, t2v, func(scope incr.Scope, t2vv string) Incr[string] {
  return Map(scope, t1, func(v string) string { return v + " Ipsum" })
})
```

Here `t1` is _not_ created within a bind scope (it's created in the top level scope by passing in the `Graph` reference), but the map that adds `" Ipsum"` to the value _is_ created within a bind scope. This is done transparently by passing the scope through the `Map` constructor within the bind.

# Progress

The combinators from the original library are implemented, plus a few this one adds:

- **Inputs** — `Var`, `VarEqual`, `VarEqualFunc`, `Return`, `Func`
- **Mapping** — `Map` through `Map8`, `MapN` (which accepts inputs added after construction
  via `AddInput`/`RemoveInput`), `MapIf`, and a `...Context` form of each that can take a
  context and fail
- **Aggregating** — `ReduceBalanced`, `UnorderedArrayFold`, `ArrayFold`, `All`, `ForAll`,
  `Exists`
- **Dynamic shape** — `Bind`, `Bind2`, `Bind3`, `Bind4`, `BindIf`, `Join`, and
  `incrutil.BindMemoized` for right-hand sides that are expensive to build
- **Cutoffs** — `Cutoff`, `Cutoff2`, `CutoffEqual`, `CutoffEqualFunc`, `CutoffAlways`,
  `CutoffNever`
- **Time** — `Clock`, `At`, `AtIntervals`, `Snapshot`, `StepFunction`, `Timer`
- **Other** — `Observe`, `Freeze`, `Always`, `Watch`, `Sentinel`, `DependOn`
- **Keyed collections** — `incrutil/pmap` and `incrutil/mapi`, described above
- **Escape hatches** — `ExpertGraph`, `ExpertNode`, `ExpertScope`, `ExpertVar` for
  assembling or driving a graph from outside the ordinary constructors, and
  `ExpertGraph.CheckInvariants` to verify the result

Not implemented: `Incr_map`'s `transpose` and the nested-map `collapse`/`expand` operators,
which are the most niche of that family. `ParallelStabilize` has no counterpart in the
original.
