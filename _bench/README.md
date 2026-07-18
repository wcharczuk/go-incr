# go-incr vs. Jane Street `incremental`

A like-for-like benchmark suite comparing this library to the reference OCaml
implementation it is modeled on, across three graph shapes: **wide**, **deep**,
and **bind-heavy**.

## Running

```bash
./run.sh
```

Requires Go plus an opam switch named `bench` with `incremental` installed:

```bash
opam switch create bench 5.1.1
eval $(opam env --switch=bench)
opam install incremental dune
```

Override the switch with `OPAM_SWITCH=...`, and the output directory with
`OUT_DIR=...`.

## Measuring changes to go-incr

**Never compare numbers captured in different sessions.** This machine's absolute
throughput drifts by up to ~2x on allocation-heavy cases between sessions; an
early round of this work "showed" a 90% regression on construction that was
entirely drift — the unmodified baseline binary reproduced it. Use `ab.py`, which
interleaves variants within one invocation and takes the per-case min:

```bash
git worktree add /tmp/base HEAD && (cd /tmp/base && go build -o /tmp/gobench_base ./_bench/go)
go build -o /tmp/gobench_new ./_bench/go
python3 _bench/ab.py 3 "base=/tmp/gobench_base" "new=/tmp/gobench_new"
```

`ab.py` also accepts the OCaml binary, so the same run can show both the delta
and the remaining gap. Always diff `-verify` output between variants first.

## Methodology

Both harnesses (`go/main.go`, `ocaml/bench.ml`) construct structurally identical
graphs and use the same timing loop: warm up, grow a batch until one batch
exceeds 10ms so clock resolution is not a meaningful error term, then run batches
until ≥5 rounds and ≥0.5s total. Reported figures are the **best per-op time
across rounds**, which is the most noise-resistant summary; `ns_per_op` in the
raw JSONL is the mean.

`run.sh` first runs both harnesses in `-verify` mode and diffs the output. This
guards the whole comparison against the failure mode that matters most here: a
graph whose nodes are not *necessary* stabilizes almost instantly and would look
fast rather than broken. All eight observed values match exactly, so both sides
are demonstrably doing the same work.

Two deliberate configuration choices:

- go-incr is measured with `NewSequentialIdentifierProvier`. The default
  `crypto/rand` identifier provider is measured separately as
  `wide/construct_ident_default/*`; it costs roughly 6–9% on construction and
  nothing on updates.
- OCaml `max_height_allowed` and go-incr `OptGraphMaxHeight` are both raised to
  fit the deep cases (the OCaml default is 128, go-incr's is 256).

## Results

Ryzen/WSL2, 8 cores; Go 1.26.2, OCaml 5.1.1, incremental v0.17. Best-of times.
Three independent runs agreed within a few percent.

| case | go-incr | ocaml | ratio |
|---|---:|---:|---:|
| wide/construct/1024 | 1.82ms | 967.7us | 1.9x |
| wide/update_one/1024 | 956ns | 266ns | 3.6x |
| wide/update_all/1024 | 270.0us | 129.7us | 2.1x |
| wide/update_same/1024 | 866ns | 79ns | **10.9x** |
| wide/construct/16384 | 32.69ms | 15.74ms | 2.1x |
| wide/update_one/16384 | 1.4us | 328ns | 4.2x |
| wide/update_all/16384 | 4.99ms | 2.16ms | 2.3x |
| wide/update_same/16384 | 1.2us | 75ns | **15.6x** |
| deep/construct/1 | 2.0us | 2.3us | *0.86x (go-incr wins)* |
| deep/update_one/1 | 213ns | 91ns | 2.3x |
| deep/construct/128 | 78.8us | 36.1us | 2.2x |
| deep/update_one/128 | 8.3us | 2.0us | 4.2x |
| deep/construct/2048 | 2.18ms | 622.0us | 3.5x |
| deep/update_one/2048 | 133.0us | 32.8us | 4.1x |
| bind/swap_chain/64 | 48.3us | 13.2us | 3.7x |
| bind/swap_chain/512 | 373.9us | 135.5us | 2.8x |
| bind/wide_swap/256 | 433.5us | 160.8us | 2.7x |
| bind/wide_swap/4096 | 7.39ms | 3.63ms | 2.0x |
| bind/wide_construct/4096 | 15.06ms | 7.72ms | 1.9x |

### Summary

`incremental` is faster on essentially every case, by **~2x on construction and
bulk recomputation** and **~4x on incremental updates**. go-incr wins exactly one
case — creating a trivial graph — where its cheaper setup shows through before
per-node costs dominate. Notably the ratios are *flat* in graph size: go-incr's
asymptotics match, so this is a constant-factor gap, not an algorithmic one.

Decomposing the deep-chain numbers separates fixed from marginal cost:

|  | fixed per stabilization | marginal per node |
|---|---:|---:|
| go-incr | 213ns | **64.8ns** |
| incremental | 91ns | **16.0ns** |

(marginal = (depth-2048 time − depth-1 time) / 2047)

So the headline is ~65ns vs ~16ns to recompute one node.

### Where go-incr's time goes

CPU profile of `deep/update_one/2048` (`-cpuprofile`):

```
18.87%  recomputeHeapList.push
16.98%  Graph.recompute
13.21%  recomputeHeap.removeMinUnsafe
11.32%  recomputeHeapList.pop
 9.43%  mapIncr.Stabilize        <-- the only actual user computation
 7.55%  recomputeHeap.addNodeUnsafe
```

**Roughly 55% of the time is recompute-heap bookkeeping; under 10% is the user's
own map function.** The heap operations are already O(1) with the right data
structure, so the cost is not algorithmic. A line-level profile of `push` shows
where it actually goes:

```
50ms   v.Node().nextInRecomputeHeap = nil
40ms   v.Node().previousInRecomputeHeap = nil
```

90 of those 100ms are two statements that do nothing but reach node metadata.
Every node is held as an `INode` interface, so each field access is a
non-inlinable virtual `Node()` call followed by a pointer hop to a separately
allocated struct — and the hot loop does this several times per node. OCaml's
nodes are records accessed by direct field offset, with the recompute heap
bucketed by height in flat arrays.

That points at the structural cause of the ~4x, and at the shape of any fix:
reduce indirection per node visit (cache the `*Node` alongside the interface in
heap entries, or store node metadata inline in heap slots) rather than
micro-optimizing the heap algorithm.

### One semantic difference, not just a speed difference

`wide/update_same` sets a var to the value it already holds. `incremental`
applies a physical-equality cutoff by default and drops the update in ~75ns
regardless of graph size. go-incr has no default cutoff, so it propagates the
whole way and costs the same as a real update — hence 11–16x, growing with graph
size.

This is a design choice rather than a bug: go-incr offers `Cutoff`/`Cutoff2`
nodes to opt in explicitly. But for workloads where writes are frequently
no-ops — a common pattern in practice — the OCaml default avoids a lot of work
that go-incr users must remember to suppress by hand. Worth considering whether
an opt-in `OptGraphDefaultCutoff` or a cutoff-by-default `Var` variant would be
a good addition.

## Layout

- `go/main.go` — Go harness. `-filter`, `-verify`, `-cpuprofile`.
- `ocaml/bench.ml` — OCaml harness. First arg is a filter, or `-verify`.
- `compare.py` — joins the two JSONL outputs into the table above.
- `run.sh` — builds, verifies equivalence, runs both, prints comparison.
