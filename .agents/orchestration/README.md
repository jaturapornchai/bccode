# AI Team Orchestration (dispatch helpers)

ZCode (leader/auditor + full-stack worker) calls these to delegate work to worker agents, then audits the result.
Full governance: [`.agents/rules/ai-team-governance.md`](../rules/ai-team-governance.md).

## Helpers
| Script | Worker | Role | Auth |
|---|---|---|---|
| `dispatch-codex.sh [--mini] "<spec>"` | Codex | advisory / deep analysis (think) — ZCode applies patches itself | ChatGPT subscription |
| `dispatch-design.sh "<brief>"` | Gemini | design UX/UI (read-only spec) | Google OAuth |
| `tools/ai/codex-advisor.ps1` | Codex | cleaner advisory wrapper (Mode all/code/review/weakness) | ChatGPT subscription |

- `--mini` → `gpt-5.4-mini`+low reasoning (fast/parallel/small); default → `gpt-5.5`+medium (complex); `--deep` → `gpt-5.5`+high (hardest/architecture).
- Tuned for speed+reliability: clears MCP servers + caps reasoning + `--ephemeral` (≈6–20s/call vs minutes on config default `xhigh`). Run `health-check.sh` before a big session.
- Every `<spec>` must carry the 5 Handoff sections: GOAL, CONTEXT, DECISIONS, CONSTRAINTS, VERIFICATION.
- `agy` (Antigravity CLI) is installed + authed but its headless stdout is broken on Windows; use it
  INTERACTIVELY (`agy` in a terminal) for iterative design. Automated design uses `gemini` (same engine).

## Usage from ZCode (via Bash tool)
```bash
bash .agents/orchestration/dispatch-codex.sh "GOAL: ... CONTEXT: ... DECISIONS: ... CONSTRAINTS: ... VERIFICATION: ..."
bash .agents/orchestration/dispatch-codex.sh --mini "GOAL: rename field x->y in 3 files ..."
bash .agents/orchestration/dispatch-design.sh "Design the product-marketplace tab: who=SME owner, tone=premium/clean ..."
powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\codex-advisor.ps1 -Mode all -Prompt "review this approach and find blind spots"
```

## Hard rules
- Subscription/OAuth only. Scripts ABORT if `OPENAI_API_KEY` / `GEMINI_API_KEY` is set (per-token billing).
- Codex runs `--sandbox workspace-write` (edits repo, not full system) — but in the ZCode-led team, Codex output is advisory; ZCode applies patches itself.
- After a worker returns, ZCode MUST run the VERIFICATION command itself — never trust the worker's word.
