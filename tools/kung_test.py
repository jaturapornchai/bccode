#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Random test runner for น้องกุ้ง agent.
Runs N questions in parallel, collects timing + tools + HTML format check + quality metrics.
Thai-safe (unlike bash curl which mangles UTF-8 on Windows Git Bash).
"""
import json
import sys
import time
import argparse
import urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

# Force UTF-8 stdout on Windows — use reconfigure (Python 3.7+) to avoid
# wrapping that breaks at GC time when threads are still running.
if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")

SHOP = "3AEz8tu22GHPpAZ0XhwPFM4fjY9"
URL = "http://localhost:8888/goapi/api/v1/chatbot/chat-agent-v2-sync"

# Diverse question set covering: product code, product name, customer, supplier,
# internal policy (KB), market comparison (web), sales aggregation, vague/edge cases
QUESTIONS = [
    # Product code / barcode / SKU lookups
    "ROF002 ราคาเท่าไหร่",
    "TIL001 มีในสต็อกไหม",
    "บาร์โค้ด 8850007001001 คือสินค้าอะไร",
    # Product name / brand / category searches
    "หาสินค้า TOA ทั้งหมด",
    "กระเบื้องลอนคู่ มียี่ห้อไหนบ้างในร้าน",
    "สีทาบ้านมีขายไหม",
    # Customer / debtor
    "ลูกค้าชื่อกระเบื้องทอง มีไหม",
    "รายชื่อลูกค้า 5 รายแรก",
    # Supplier / creditor
    "ซัพพลายเออร์ปูนซีเมนต์มีใครบ้าง",
    "เจ้าหนี้รายใหญ่ที่สุด 3 อันดับ",
    # Internal docs (KB)
    "วิธีลางานพนักงานทำยังไง",
    "กระเบื้องระเบิด สาเหตุและวิธีป้องกัน",
    # Aggregation / analytics
    "สินค้าราคาแพงที่สุด 5 อันดับ",
    "มีสินค้าทั้งหมดกี่รายการในร้าน",
    "ยอดขายเดือนนี้เท่าไหร่",
    # Market comparison (web)
    "ปูนซีเมนต์ตราอินทรี ราคาตลาดเท่าไหร่เทียบกับในร้าน",
    "TOA สีเบอร์ไหนขายดีในตลาด",
    # Edge cases
    "มีลูกหนี้ค้างชำระกี่ราย",
    "สวัสดีครับ",
    "ช่วยแนะนำอะไรได้บ้าง",
]


def ask(idx: int, q: str, out_dir: Path) -> dict:
    body = json.dumps({
        "holding_code": SHOP,
        "question": q,
        "session_id": f"pytest-{int(time.time())}-{idx}",
        "output_format": "html",
    }).encode("utf-8")
    req = urllib.request.Request(
        URL,
        data=body,
        headers={"Content-Type": "application/json; charset=utf-8"},
        method="POST",
    )
    t0 = time.time()
    try:
        with urllib.request.urlopen(req, timeout=300) as r:
            raw = r.read()
            dt = time.time() - t0
            data = json.loads(raw.decode("utf-8"))
    except Exception as e:
        dt = time.time() - t0
        return {"idx": idx, "q": q, "error": str(e), "dt": dt}
    (out_dir / f"resp-{idx}.json").write_bytes(raw)
    return {"idx": idx, "q": q, "dt": dt, "data": data}


def analyze(r: dict) -> dict:
    if "error" in r:
        return {**r, "verdict": "ERROR"}
    d = (r["data"].get("data") or {})
    ans = d.get("answer", "") or ""
    tools = d.get("tools_used") or []
    iters = d.get("iterations", 0)
    model = r["data"].get("token_usage", {}).get("model", "?")
    tnames = [t.get("tool", "?") for t in tools]

    has_html = any(tag in ans for tag in ("<h2", "<table", "<p>", "<ul>"))
    md_leak = ans.lstrip().startswith("##") or "\n## " in ans or "\n| --" in ans

    # Heuristic quality signals.
    # A response with tools=0 + "ไม่พบ" is suspicious, but NOT if the answer is well-structured
    # (has all 3 sections) and is long enough to look like a real attempt. The KB pre-fetch
    # mechanism provides context even without explicit tool calls, so a tools=0 answer can
    # be legitimate when the pre-fetch chunks gave the model everything it needed.
    has_sections = all(k in ans for k in ("แผนการหาคำตอบ", "บทวิเคราะห์"))
    no_tool_fail = (
        len(tools) == 0
        and ("ไม่พบ" in ans or "ไม่มี" in ans)
        and not has_sections  # if sections are present + "ไม่พบ", it's a valid negative answer
        and len(ans) < 500    # too short to be a real 3-section answer
    )

    verdict = "PASS"
    if not has_html:
        verdict = "FAIL-HTML"
    elif md_leak:
        verdict = "FAIL-MDLEAK"
    elif no_tool_fail:
        verdict = "FAIL-HALLUC"
    elif not has_sections:
        verdict = "WARN-STRUCT"

    return {
        "idx": r["idx"],
        "q": r["q"],
        "dt": r["dt"],
        "iters": iters,
        "tools": tnames,
        "tool_count": len(tnames),
        "model": model,
        "ans_len": len(ans),
        "verdict": verdict,
    }


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--out", default="C:/Users/jatur/AppData/Local/Temp/kung-test")
    ap.add_argument("--parallel", type=int, default=3)
    ap.add_argument("--only", type=int, default=None, help="run only index N")
    args = ap.parse_args()

    out = Path(args.out)
    out.mkdir(parents=True, exist_ok=True)

    qs = [(i, q) for i, q in enumerate(QUESTIONS)]
    if args.only is not None:
        qs = [qs[args.only]]

    print(f"Running {len(qs)} questions with parallelism={args.parallel}...")
    results = []
    with ThreadPoolExecutor(max_workers=args.parallel) as ex:
        futures = {ex.submit(ask, i, q, out): (i, q) for i, q in qs}
        for fut in as_completed(futures):
            r = fut.result()
            results.append(r)
            i = r["idx"]
            tag = "OK" if "error" not in r else "ERR"
            print(f"  [{i}] {tag} {r['dt']:.1f}s — {r['q'][:60]}")

    results.sort(key=lambda r: r["idx"])
    analyses = [analyze(r) for r in results]

    print()
    print(f"{'#':<3} {'verdict':<14} {'time':>6} {'iter':>5} {'tools':>6} {'question':<40} {'tool-names'}")
    print("-" * 130)
    for a in analyses:
        print(f"{a['idx']:<3} {a['verdict']:<14} {a['dt']:>6.1f} {str(a.get('iters','?')):>5} "
              f"{a.get('tool_count',0):>6} {a['q'][:38]:<40} {','.join(a.get('tools', []))}")

    print()
    pass_cnt = sum(1 for a in analyses if a["verdict"] == "PASS")
    print(f"Pass: {pass_cnt}/{len(analyses)}")
    print(f"Avg time: {sum(a['dt'] for a in analyses)/len(analyses):.1f}s")
    print(f"Total tools called: {sum(a.get('tool_count',0) for a in analyses)}")
    print(f"Avg tools/q: {sum(a.get('tool_count',0) for a in analyses)/len(analyses):.1f}")
    # Save summary
    (out / "summary.json").write_text(json.dumps(analyses, ensure_ascii=False, indent=2), encoding="utf-8")
    print(f"\nDetails saved to {out}")


if __name__ == "__main__":
    main()
