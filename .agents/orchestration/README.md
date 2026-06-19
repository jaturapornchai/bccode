# AI Team Orchestration (advisor helpers)

Claude (Claude Code) is the leader/auditor + full-stack worker. Claude calls these helpers to get **advisory** input from GLM 5.2 Think (primary) and Codex (secondary), then synthesizes, applies patches itself, and runs real VERIFICATION.
Full governance: [`.agents/rules/ai-team-governance.md`](../rules/ai-team-governance.md).

## Helpers
| Helper | Advisor | Role | Auth |
|---|---|---|---|
| `tools/ai/glm52-think-planner.ps1` | GLM 5.2 Think (1m) | **primary advisor** — plan/code/review/weakness; Asian/Thai business + accounting + ERP/POS + Thai UX | Z.AI subscription (`ZAI_API_KEY`) |
| `tools/ai/codex-advisor.ps1` | Codex (gpt-5.5 think) | **secondary advisor** — bug/security/edge-case (Mode all/code/review/weakness) | ChatGPT subscription |
| `dispatch-codex.sh [--mini] "<spec>"` | Codex | advisory / deep analysis (think) — Claude applies patches itself | ChatGPT subscription |
| `codex exec -m gpt-5.5 ... high "<brief>"` | Codex | **image generation** — `.webp`/`.png`/`.jpg` only, never `.svg`/vector (governance §8) | ChatGPT subscription |

- Advisors are **advisory only**: they must NOT edit files; Claude reads evidence, synthesizes the best of both models, applies patches, and verifies.
- `dispatch-codex.sh`: `--mini` → `gpt-5.4-mini`+low reasoning (fast/parallel/small); default → `gpt-5.5`+medium (complex); `--deep` → `gpt-5.5`+high (hardest/architecture). Tuned for speed (clears MCP + caps reasoning + `--ephemeral`). Run `health-check.sh` before a big session.
- Every `<spec>` must carry the 5 Handoff sections: GOAL, CONTEXT, DECISIONS, CONSTRAINTS, VERIFICATION.
- **Gemini/Antigravity (`agy`) removed from the team (2026-06-19).** `dispatch-design.sh` deleted. UX/UI is implemented by Claude; ask GLM 5.2 for the Asian/Thai-user perspective.

## Usage from Claude (via Bash / PowerShell tool)
```bash
# GLM 5.2 Think — primary advisor (code + Asian/Thai business)
powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\glm52-think-planner.ps1 -Mode all -Prompt "GOAL: ... CONTEXT: ... DECISIONS: ... CONSTRAINTS: ... VERIFICATION: ..."
# Codex — secondary advisor (bug/security/edge-case)
powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\codex-advisor.ps1 -Mode review -Prompt "review <scope>: find bug, security, edge case — list with file:line"
# Codex — deep analysis via dispatch
bash .agents/orchestration/dispatch-codex.sh "GOAL: ... CONTEXT: ... DECISIONS: ... CONSTRAINTS: ... VERIFICATION: ..."
# Codex — generate an image asset (.webp)
codex exec --skip-git-repo-check -m gpt-5.5 -s workspace-write -c model_reasoning_effort=high "<short brief: concept + Thai text + use image_gen, output .webp>"
```

## Hard rules
- Subscription/OAuth only. Scripts ABORT if `OPENAI_API_KEY` (per-token billing) is set. `ZAI_API_KEY` must be a subscription/coding-plan key, never per-token, and never written into repo/docs/logs.
- Codex runs `--sandbox workspace-write` (edits repo, not full system) — but advisor output is advisory; **Claude applies patches itself**.
- After an advisor returns, **Claude MUST run the VERIFICATION command itself** — never trust the advisor's word (IRON RULE: VERIFY BEFORE DONE).
- Image assets: Codex only, raster only (`.webp`/`.png`/`.jpg`), **never SVG/vector** (governance §8).
