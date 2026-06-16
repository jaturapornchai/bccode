---
name: glm52-planner
description: Use when BC Account work needs secondary GLM 5.2 Think help for planning, implementation ideas, code review, or Codex weakness/blind-spot checks. This is a helper only; Codex must still inspect source/runtime evidence itself.
---

# GLM 5.2 Helper

Use this skill as the mandatory secondary-advisor attempt for BC Account work that involves planning, implementation, code review, debugging, UX/UI review, recommendations, or weakness checks.

Supported helper roles:

- `plan`: safe sequence, evidence to inspect, risks, and verification.
- `code`: implementation suggestions, target files/functions, and smallest safe patch shape.
- `review`: bug/security/regression/missing-test findings from provided context.
- `weakness`: Codex blind spots, missing evidence, unsafe assumptions, and checks to close gaps.
- `all`: combined planning, code, review, and weakness audit.
- Prefer `all` for general Codex tasks unless a narrower mode is clearly sufficient.
- Use `-Prompt` or `-PromptFile` for agent calls. Do not rely on piped stdin through nested `powershell -File`.

## Source Of Truth

- Codex must read the active project source, local docs, tests, and runtime output first.
- GLM output is advisory only. It cannot establish APIs, schemas, business rules, or completion status.
- GLM must not edit files directly. Codex applies patches after checking against source evidence.
- Codex remains responsible for analysis, implementation, verification, and final reporting.
- For accounting, ERP, security, migration, deployment, or production work, final decisions must be based on Codex-inspected source/runtime evidence.

## Secret Handling

- Never write the Z.AI API key into repo files, docs, rules, prompts, logs, commits, or PR text.
- The helper reads the key from `ZAI_API_KEY` by default.
- If the key is missing, continue with normal Codex planning and report that the GLM helper was unavailable.

## Command

From `D:\bccode`:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\glm52-think-planner.ps1 -Mode plan -Prompt "plan this task with risks and verification"
```

With a context file:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\glm52-think-planner.ps1 -Mode all -PromptFile D:\path\to\prompt.txt -ContextFile D:\path\to\evidence.txt
```

Mode examples:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\glm52-think-planner.ps1 -Mode code -Prompt "suggest code changes only after this evidence"
powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\glm52-think-planner.ps1 -Mode review -Prompt "review this diff and find blind spots"
powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\glm52-think-planner.ps1 -Mode weakness -Prompt "what could Codex miss here"
```

## Output Handling

- Use the visible `content` only as planning advice.
- Do not expose or persist raw `reasoning_content`.
- The helper prints `reasoningchars` so Codex can verify thinking was used without revealing chain-of-thought.
- Merge useful GLM points into Codex's own plan, patch, review, or verification checklist only after checking them against local source/runtime evidence.

## Reverse direction — codex-advisor

This skill is one half of the GLM ↔ Codex bidirectional advisory loop (see
`ai-team-governance.md` §11). The reverse direction — when GLM 5.2 (or Claude)
is the active agent and wants Codex's perspective — uses the sibling skill
`codex-advisor` and helper `tools/ai/codex-advisor.ps1`. The synthesis loop is
identical in both directions: active agent reads evidence first, asks the
advisor, then synthesizes the best of both models.
