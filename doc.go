/*
Incr implements an incremental computation graph.

This graph is useful for partially recomputing a small subset of a very large graph of computation nodes.

It is largely based off Jane Street's `incremental` ocaml library, with some go specific changes.
*/
package incr
