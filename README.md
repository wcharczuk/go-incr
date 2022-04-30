go-incremental
==============

_NOTE:_ this library is "pre-release", that is, it is not ready for production.

`go-incremental` is an incremental computation library designed for parallel computation of large graphs of operations.

It is useful in situations where you want to compute efficiently the outputs of computations where only a subset of the computation changes for a given pass.

# Inspiration

The inspiration for `go-incremental` is Jane Street's [incremental](https://github.com/janestreet/incremental) library.
