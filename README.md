go-incr
==============

[![Continuous Integration](https://github.com/wcharczuk/go-incr/actions/workflows/ci.yml/badge.svg)](https://github.com/wcharczuk/go-incr/actions/workflows/ci.yml)

![Graph](https://github.com/wcharczuk/go-incr/blob/main/_assets/small_graph.png)

_CAVEAT:_ this library is "pre-release", that is, it is not ready for production.

`go-incr` is an incremental computation library designed for partial computation of large graphs of operations.

It is useful in situations where you want to compute efficiently the outputs of computations where only a subset of the computation changes for a given pass.

# Inspiration

The inspiration for `go-incr` is Jane Street's [incremental](https://github.com/janestreet/incremental) library.

# Usage

Given an exmaple computation:

```go
v0 := incr.Var("foo")
v1 := incr.Var("bar")

output := incr.Map2(v0, v1, func(a, b string) string { return a + " and " + b })
```

In order to realize the values, we need to observe nodes in a graph, and then call `Stabilize` on the graph:

```go
g := incr.New()
o := incr.Observe(g, output)
if err := g.Stabilize(context.Background()); err != nil {
  // ... handle error if it comes up
}
```

`Stabilize` then does the full recomputation, with the "observer" `o` marking the graph up from the `output` map as observed.

# Design Choices

There is some consideration with this library on the balance between hiding mutable implemenation details to protect against [Hyrum's Law](https://www.hyrumslaw.com/) issues, and surfacing enough utility helpers to allow users to extend this library for their own use cases.

As a result, most of the features of this library can be leveraged externally, but little of the internal state and recomputation mechanism is exposed to the user.

Specific implications of this are, the `INode` interface includes a function that returns the `Node` metadata, but this `Node` struct has 0 exported fields on it, and users of this library should not really concern themselves with what's on it, just that it gets supplied to `Incr` through the interface implementation.

# Implementation details

Internally `go-incr` uses a pseudo-height recomputation adjacency list similar to the mainline library. This offers "good enough" approximation of a heap while also allowing for fast iteration between results (e.g. faster than a traditional heap's O(log(n)) performance).

`go-incr` also supports multiple "graphs" rooted in specific variables.

# A word on `Bind`

The `Bind` incremental is by far the most powerful, and as a result complicated, node type in the ecosystem.

To explain its purpose; there are situations where you want to dynamically swap out parts of a graph rooted at a given node. A concrete example might be in a user interface swapping out the displayed controls, which may be made up of individual computations.

The effect of `Bind` is that "children" of a `Bind` node may have their heights change significantly depending on the changes made to the "bound" nodes, and this has implications for the recomputation heap as a result.

# Progress

Many of the original library types are implemented, including:
- Always
- Bind(If)
- Cutoff(2)
- Freeze
- Map(2,3,If,N)
- Observe
- Return
- Var
- Watch

With these, you can create 90% of what I typically needed this library for, though some others would be relatively straightforward to implement given the primitives already implemented.
