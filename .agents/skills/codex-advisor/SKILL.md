---
name: codex-advisor
description: Use when GLM 5.2 Think or Claude needs Codex (gpt-5.5 think) advisory input for BC Account work — planning, code ideas, bug review, or weakness/blind-spot checks. The caller (GLM/Claude) asks Codex, takes Codex's answer, and synthesizes the best of both models. Helper only; the caller remains the source of truth and applies patches itself.
---

# Codex Advisor (GLM/Claude → ask Codex → synthesize)

This is the **reverse direction** of `glm52-planner`. Use it when the active agent
is **GLM 5.2 Think** (or Claude) and wants a second model's perspective from
**Codex (gpt-5.5 think)** before finalizing its own answer.

## When to use

Use as the secondary-advisor attempt for BC Account work where the active agent
wants Codex's perspective:

- The active agent is doing planning, implementation, code review, debugging,
  UX/UI review, recommendations, or weakness checks.
- A second model's view would reduce blind spots (dual-model synthesis).
- The work is **non-trivial** (architecture, hard bug, multi-file change,
  security-sensitive, accounting/decimal, tenant scope).

For trivial work (rename, comment, single-line fix), skip this helper — it costs
token + latency for little gain (same Pareto rule as `ai-team-governance.md` §9/§10).

## Two-model synthesis loop (the core idea)

```
active agent (GLM 5.2 or Claude)
        │
        │  1. read source/runtime evidence first (no guess)
        │  2. form own preliminary answer
        ▼
   codex-advisor.ps1  ──►  Codex (gpt-5.5 think) advisory answer
        │
        │  3. compare: Codex-strong points vs active-agent-strong points
        │  4. synthesize the best of both into the final answer
        │  5. active agent remains source of truth + applies patches + verifies
        ▼
   final answer / patch + real VERIFICATION
```

The active agent **never** delegates final responsibility to Codex. Codex is an
advisory input; the active agent synthesizes, decides, edits, and verifies.

## Supported helper roles (Mode)

- `plan` — safe sequence, evidence to inspect, risks, verification.
- `code` — implementation suggestions, target files/functions, smallest safe patch shape.
- `review` — bug/security/regression/missing-test findings from provided context.
- `weakness` — blind spots, missing evidence, unsafe assumptions, and checks to close gaps.
- `all` — combined planning, code, review, and weakness audit (default for general tasks).

## Source of truth

- The active agent (GLM/Claude) must read the active project source, local docs,
  tests, and runtime output first.
- Codex output is **advisory only**. It cannot establish APIs, schemas, business
  rules, or completion status.
- Codex must not edit files in this advisory call (the helper enforces it via the
  system prompt). The active agent applies patches after checking Codex advice
  against source evidence.
- The active agent remains responsible for analysis, implementation, verification,
  and final reporting.
- For accounting, ERP, security, migration, deployment, or production work, final
  decisions must be based on the active agent's inspected source/runtime evidence.

## Subscription-only (no per-token API key)

- Auth: `codex login` (ChatGPT sign-in). Subscription only.
- The helper ABORTs if `OPENAI_API_KEY` is set (same policy as `dispatch-codex.sh`).
- On rate-limit / quota exhaustion: stop, report, wait for refresh, or fall back
  to a single-model answer. Never switch to a per-token API key.

## Command

From `D:\bccode`:

```powershell
# Inline prompt, default model gpt-5.5 + reasoning high + mode all
powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\codex-advisor.ps1 -Prompt "review this approach and find blind spots"

# With a prompt file (longer prompts)
powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\codex-advisor.ps1 -PromptFile D:\path\to\prompt.txt

# With evidence context (active agent pastes the source it already read)
powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\codex-advisor.ps1 -Mode review -PromptFile D:\path\to\prompt.txt -ContextFile D:\path\to\evidence.txt

# Narrower / cheaper advisory
powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\codex-advisor.ps1 -Mode weakness -Reasoning medium -Prompt "what could I miss here"
```

Mode + reasoning combinations:

```powershell
# Plan only, faster
powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\codex-advisor.ps1 -Mode plan -Reasoning medium -Prompt "plan this task with risks and verification"

# Code suggestion, deep
powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\codex-advisor.ps1 -Mode code -Reasoning high -Prompt "suggest the smallest safe change after this evidence"
```

## Output handling

- Use Codex's visible output as advisory input only.
- Compare against the active agent's own preliminary answer.
- Synthesize the best of both into the final answer/patch.
- The active agent runs the real VERIFICATION itself (Codex advice does not
  replace `cd frontend; npm run typecheck`, touched-package `go build`, `/healthz`,
  or browser check).
- If Codex's answer and the active agent's answer disagree on a fact, the active
  agent re-checks source/runtime evidence and trusts the evidence — never the
  louder model.

## Failure / unavailability

- If `codex` CLI is missing or `codex login` is not authenticated, report that
  the Codex advisor is unavailable and continue with the active agent's own
  single-model answer. Do not block.
- If Codex times out or errors, retry once with `-Reasoning medium`. If still
  failing, fall back to single-model and report it.
