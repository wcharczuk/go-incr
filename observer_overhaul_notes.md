Observer Overhaul Notes
=======================

- The current implementation of observers is overly aggressive; we almost never really care about parts of a graph that are linked but unobserved (can this even happen?)
- We can, as a result, do away with the hierarchical concept of observers and just use them to anchor points in the graph.

Spedific changes
- Do away with recursive observed steps.
- Observation simply changes the observer(s) list for a node.
- Necessary is a combination of having children and being in a scope.
- When we unobserve, we snap the parent list of the observer and remove it from the observed nodes observer list.
- This does _not_ affect the graph meaningfully. If we swap out a bind graph, we have to remove it from the graph during the scope update.

A node is necessary if:
- it has children
OR
- it has observers

What _is_ the purpose of observers then?
- just to anchor leaves of the graph

What happens if a node becomes unnecessary?
- if we leave it with no parents, and it's unobserved, it's removed from the graph.
- this happens at _unlink_ time
- we then proceed up (or down?) the graph, removing all parents of that newly unnecessary node

Tests will be a pain to change but that's fine.