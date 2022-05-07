go-incremental
==============

_CAVEAT:_ this library is "pre-release", that is, it is not ready for production, it's not really ready to be used seriously yet.

`go-incr` is an incremental computation library designed for partial computation of large graphs of operations.

It is useful in situations where you want to compute efficiently the outputs of computations where only a subset of the computation changes for a given pass.

# Inspiration

The inspiration for `go-incr` is Jane Street's [incremental](https://github.com/janestreet/incremental) library.

# Progress

Many of the original library types are implemented, but some of the types are not still, namely:
- Observers
- Stepwise
- Expert

The "core" types are mostly implemented though, specifically:
- Return
- Var
- Map(2,3,If)
- Bind(2,3,If)
- Cutoffs
- Watch

With these, you can create 90% of what I typically needed this library for, though some others would be relatively straightforward to implement given the primitives already implemented.
