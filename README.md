go-incremental
==============

_CAVEAT:_ this library is "pre-release", that is, it is not ready for production, it's not really ready to be used seriously yet.

`go-incr` is an incremental computation library designed for partial computation of large graphs of operations.

It is useful in situations where you want to compute efficiently the outputs of computations where only a subset of the computation changes for a given pass.

# Inspiration

The inspiration for `go-incr` is Jane Street's [incremental](https://github.com/janestreet/incremental) library.

# Usage

Given an exmaple computation:

```go
v0 := incr.Var("foo")
v1 := incr.Var("bar")

output := incr.Apply2(v0.Read(), v1.Read(), func(a, b string) string { return a + " and " + b})
```

In order to realize the values, we need to call `Stabilize` on it:

```go
if err := incr.Stabilize(context.Background(), output); err != nil {
  // ... handle error
}
```

`Stabilize` both initializes internal metadata for the computation graph, and does the full recomputation in subsequent cycles. 

# Design Choices

There is some consideration with this library on the balance between hiding mutable implemenation details to protect against [Hyrum's Law](https://www.hyrumslaw.com/) issues, and surfacing enough utility helpers to allow users to extend this library for their own use cases.

As a result, most of the features of this library can be leveraged externally, but little of the internal state and recomputation mechanism is exposed to the user. 

Specific implications of this are, the `INode` interface includes a function that returns the `Node` metadata, but this `Node` struct has 0 exported members on it, and users of this library should not really concern themselves with what's on it, just that it gets supplied to `Incr` through the interface implementation.

# Progress

Many of the original library types are implemented, but some of the types are not still, namely:
- Observers
- Stepwise
- Expert

The "core" types are mostly implemented though, specifically:
- Return
- Var
- Apply(2,3,If)
- Bind(2,3,If)
- Cutoffs
- Watch

With these, you can create 90% of what I typically needed this library for, though some others would be relatively straightforward to implement given the primitives already implemented.
