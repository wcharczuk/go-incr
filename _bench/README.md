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
fast rather than broken. All eight observed values match exactly.

That checks the answers, not the work. For the work, `-stats` compares how many nodes each
shape recomputes; those counts agree except in three understood places, which
`ALGORITHMS.md` documents. `run.sh` does not gate on them.

Two deliberate configuration choices:

- go-incr is measured with a sequential identifier provider. The default
  `crypto/rand` identifier provider is measured separately as
  `wide/construct_ident_default/*`; it costs roughly 6–9% on construction and
  nothing on updates.
- OCaml `max_height_allowed` and go-incr `OptGraphMaxHeight` are both raised to
  fit the deep cases (the OCaml default is 128, go-incr's is 256).

## Results

The current standing, and the analysis behind it, live in `ALGORITHMS.md` under "Current
standing". They are deliberately not duplicated here: this file used to carry its own copy of
the results table and it went badly stale, describing a state of the library that has not been
true for some time.

The short version, as of the last run: go-incr is ahead on construction, on updating every
input, and on swapping a large bind subgraph; behind by 1.2 - 1.45x on propagating a single
input change and on swapping a small bind subgraph; and behind by 4 - 6x on writing a `Var`
the value it already holds, which is a difference in defaults that `VarEqual` closes.

Two cautions, both learned the hard way and both in `ALGORITHMS.md` at more length. Compare
interleaved and never across sessions -- the same unmodified binary drifts by up to 2x between
runs on this machine. And establish the noise floor before believing a delta: two builds of
identical source differ by 1-2%, and `wide/update_all` at large sizes needs a distribution
rather than a single minimum.

## Layout

- `go/main.go` — Go harness. `-filter`, `-verify`, `-stats`, `-cpuprofile`, `-memprofile`.
  `-stats` reports how many nodes each shape recomputes, which is the check that the two
  libraries are doing the same work rather than merely producing the same answers.
- `ocaml/bench.ml` — OCaml harness. First arg is a filter, or `-verify`.
- `compare.py` — joins the two JSONL outputs into a comparison table.
- `ab.py` — interleaved A/B of two or more Go harness binaries, taking per-case minimums.
  This is the right tool for measuring a change to go-incr; see "Measuring changes" above.
- `run.sh` — builds, verifies equivalence, runs both, prints comparison.
