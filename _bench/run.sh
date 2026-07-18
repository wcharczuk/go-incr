#!/usr/bin/env bash
# Runs both benchmark harnesses and prints the comparison table.
#
# Requires an opam switch with `incremental` installed; see README.md.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
out="${OUT_DIR:-$(mktemp -d)}"
switch="${OPAM_SWITCH:-bench}"

export PATH="$HOME/.local/bin:$PATH"
eval "$(opam env --switch="$switch")"

echo "building..." >&2
go build -o "$out/gobench" "$here/go"
(cd "$here/ocaml" && dune build ./bench.exe)
ocamlbench="$here/ocaml/_build/default/bench.exe"

# Cross-check that both harnesses build equivalent graphs before trusting any
# timings: a graph with no necessary nodes would stabilize instantly and look
# fast rather than wrong.
echo "verifying graph equivalence..." >&2
"$out/gobench" -verify > "$out/go.verify"
"$ocamlbench" -verify > "$out/ocaml.verify"
if ! diff -u "$out/go.verify" "$out/ocaml.verify"; then
  echo "FAIL: harnesses disagree on computed values; timings are not comparable" >&2
  exit 1
fi
echo "ok: both harnesses agree on all observed values" >&2

echo "benchmarking go-incr..." >&2
"$out/gobench" > "$out/go.jsonl"
echo "benchmarking ocaml incremental..." >&2
"$ocamlbench" > "$out/ocaml.jsonl"

echo
python3 "$here/compare.py" "$out/go.jsonl" "$out/ocaml.jsonl"
echo
echo "raw results in $out" >&2
