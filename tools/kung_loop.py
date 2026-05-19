#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Multi-round test runner that aggregates variance across N rounds.

Produces a single summary table showing per-question mean/stddev of time, iterations,
tools count, and pass rate — so you can spot variance-driven failures vs. systemic issues.
"""
import sys
import json
import time
import argparse
import statistics
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor, as_completed

if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")  # type: ignore[attr-defined]
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")  # type: ignore[attr-defined]

# Import from kung_test
sys.path.insert(0, str(Path(__file__).parent))
from kung_test import QUESTIONS, ask, analyze  # noqa


def run_round(round_idx: int, out_base: Path, parallel: int) -> list:
    round_dir = out_base / f"round-{round_idx}"
    round_dir.mkdir(parents=True, exist_ok=True)
    print(f"\n=== Round {round_idx} ({len(QUESTIONS)} questions, parallel={parallel}) ===")
    results = []
    with ThreadPoolExecutor(max_workers=parallel) as ex:
        futs = {ex.submit(ask, i, q, round_dir): (i, q) for i, q in enumerate(QUESTIONS)}
        for fut in as_completed(futs):
            r = fut.result()
            results.append(r)
            tag = "OK" if "error" not in r else "ERR"
            print(f"  [{r['idx']}] {tag} {r['dt']:.1f}s — {r['q'][:50]}")
    results.sort(key=lambda r: r["idx"])
    return [analyze(r) for r in results]


def aggregate(rounds: list[list[dict]]) -> None:
    n_q = len(rounds[0])
    n_r = len(rounds)
    print(f"\n=== AGGREGATE over {n_r} rounds × {n_q} questions ===\n")

    # Header
    print(f"{'#':<3} {'question':<38} {'pass':>5} {'time(μ±σ)':>14} {'iter(μ±σ)':>12} {'tools(μ±σ)':>12} {'verdicts'}")
    print("-" * 130)

    totals = {"pass": 0, "times": [], "iters": [], "tools": [], "verdicts": {}}

    for qi in range(n_q):
        times, iters, tools, verdicts, passes = [], [], [], [], 0
        q_name = rounds[0][qi]["q"]
        for r in rounds:
            a = r[qi]
            if a["verdict"] == "PASS":
                passes += 1
                totals["pass"] += 1
            verdicts.append(a["verdict"])
            totals["verdicts"][a["verdict"]] = totals["verdicts"].get(a["verdict"], 0) + 1
            if "dt" in a:
                times.append(a["dt"])
                totals["times"].append(a["dt"])
            if a.get("iters"):
                iters.append(a["iters"])
                totals["iters"].append(a["iters"])
            if a.get("tool_count") is not None:
                tools.append(a["tool_count"])
                totals["tools"].append(a["tool_count"])

        def fmt(vs):
            if not vs:
                return "?"
            if len(vs) == 1:
                return f"{vs[0]:.1f}"
            return f"{statistics.mean(vs):.1f}±{statistics.stdev(vs):.1f}"

        vsum = ",".join(v[:4] for v in verdicts)
        print(f"{qi:<3} {q_name[:36]:<38} {passes}/{n_r:<3} {fmt(times):>14} {fmt(iters):>12} {fmt(tools):>12} {vsum}")

    print()
    total = n_q * n_r
    pass_rate = totals["pass"] / total * 100
    print(f"Overall pass rate: {totals['pass']}/{total} ({pass_rate:.0f}%)")
    if totals["times"]:
        print(f"Overall time: mean={statistics.mean(totals['times']):.1f}s "
              f"median={statistics.median(totals['times']):.1f}s "
              f"max={max(totals['times']):.1f}s")
    if totals["iters"]:
        print(f"Overall iter: mean={statistics.mean(totals['iters']):.1f} max={max(totals['iters'])}")
    if totals["tools"]:
        print(f"Overall tools/q: mean={statistics.mean(totals['tools']):.1f} max={max(totals['tools'])}")
    print(f"Verdict counts: {totals['verdicts']}")

    # Surface slowest / most-iterated / failing questions
    print("\nSlowest questions (by mean time):")
    by_time = sorted(
        range(n_q),
        key=lambda qi: -statistics.mean([r[qi]["dt"] for r in rounds if "dt" in r[qi]] or [0]),
    )[:5]
    for qi in by_time:
        q = rounds[0][qi]["q"]
        avg = statistics.mean([r[qi]["dt"] for r in rounds if "dt" in r[qi]])
        print(f"  [{qi}] {avg:.1f}s — {q}")

    print("\nMost iteration (by mean iter count):")
    by_iter = sorted(
        range(n_q),
        key=lambda qi: -statistics.mean([r[qi].get("iters", 0) for r in rounds]),
    )[:5]
    for qi in by_iter:
        q = rounds[0][qi]["q"]
        avg = statistics.mean([r[qi].get("iters", 0) for r in rounds])
        print(f"  [{qi}] {avg:.1f} iter — {q}")

    print("\nFailing questions (verdict != PASS in any round):")
    for qi in range(n_q):
        bad = [r[qi]["verdict"] for r in rounds if r[qi]["verdict"] != "PASS"]
        if bad:
            q = rounds[0][qi]["q"]
            print(f"  [{qi}] {q} -> {bad}")


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--rounds", type=int, default=3)
    ap.add_argument("--parallel", type=int, default=3)
    ap.add_argument("--out", default="C:/Users/jatur/AppData/Local/Temp/kung-loop")
    ap.add_argument("--tag", default=None, help="tag to append to out dir")
    args = ap.parse_args()

    out_base = Path(args.out)
    if args.tag:
        out_base = out_base / args.tag
    out_base.mkdir(parents=True, exist_ok=True)

    t0 = time.time()
    rounds = []
    for r in range(1, args.rounds + 1):
        rounds.append(run_round(r, out_base, args.parallel))

    aggregate(rounds)
    (out_base / "rounds.json").write_text(
        json.dumps(rounds, ensure_ascii=False, indent=2), encoding="utf-8"
    )
    print(f"\nTotal wall time: {time.time()-t0:.1f}s")
    print(f"Saved to {out_base}")


if __name__ == "__main__":
    main()
