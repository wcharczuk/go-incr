#!/usr/bin/env python3
"""Interleaved A/B benchmark comparison.

Usage: ab.py ROUNDS "label=command" "label=command" [...]

Runs every command once per round, alternating, and reports the per-case minimum
across rounds for each label.

Interleaving matters: this machine's absolute throughput drifts substantially
between sessions (observed ~2x on allocation-heavy cases), so comparing numbers
captured minutes apart is meaningless. Alternating within one invocation exposes
both variants to the same conditions, and taking the min across rounds discards
rounds that lost to background noise.
"""
import json
import subprocess
import sys


def human(ns):
    if ns < 1_000:
        return f"{ns:.0f}ns"
    if ns < 1_000_000:
        return f"{ns/1_000:.1f}us"
    return f"{ns/1_000_000:.2f}ms"


def run(cmd):
    out = subprocess.run(cmd, shell=True, capture_output=True, text=True, check=True)
    res = {}
    for line in out.stdout.splitlines():
        if line.strip():
            r = json.loads(line)
            res[r["name"]] = r["min_ns"]
    return res


def main():
    rounds = int(sys.argv[1])
    specs = [a.split("=", 1) for a in sys.argv[2:]]
    best = {label: {} for label, _ in specs}

    for i in range(rounds):
        for label, cmd in specs:
            print(f"round {i+1}/{rounds}: {label}", file=sys.stderr)
            for name, ns in run(cmd).items():
                prev = best[label].get(name)
                if prev is None or ns < prev:
                    best[label][name] = ns

    labels = [l for l, _ in specs]
    base = labels[0]
    # Only report cases every variant produced, so the table never implies a
    # comparison that was not actually measured.
    names = [n for n in best[base] if all(n in best[l] for l in labels)]

    header = f"{'case':<38}" + "".join(f"{l:>11}" for l in labels)
    for l in labels[1:]:
        header += f"{'x' + l[:8]:>11}"
    print(header)
    print("-" * len(header))
    for name in names:
        row = f"{name:<38}" + "".join(f"{human(best[l][name]):>11}" for l in labels)
        for l in labels[1:]:
            row += f"{best[l][name] / best[base][name]:>10.2f}x"
        print(row)

    skipped = [n for n in best[base] if n not in names]
    if skipped:
        print(f"\nnot in all variants, omitted: {', '.join(skipped)}", file=sys.stderr)


if __name__ == "__main__":
    main()
