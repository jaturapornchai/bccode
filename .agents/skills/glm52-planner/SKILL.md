---
name: glm52-planner
description: Use when BC Account work needs GLM 5.2 Think as the PRIMARY advisor for Asian/Thai business/accounting/POS perspective, planning, implementation ideas, code review, or weakness/blind-spot checks. This is a helper only; Claude (Claude Code) remains the lead/worker and must still inspect source/runtime evidence itself.
---

# GLM 5.2 Helper (primary advisor)

Use this skill as the primary-advisor attempt for BC Account work that involves planning, implementation, code review, debugging, UX/UI review, recommendations, Thai business/accounting/POS workflow decisions, or weakness checks.

Claude (Claude Code) is the lead model that does the work: read source/runtime evidence, decide, edit, test, and report. GLM 5.2 Think contributes advisory opinions on code and especially where Asian-user / Thai business operations, Thai accounting software, Thai ERP, Thai tax/document workflows, retail, restaurant, or Thai POS context can improve Claude's decision. For engineering bug/security/edge-case angles, Claude may also ask the secondary advisor Codex (`codex-advisor`).

Supported helper roles:

- `plan`: safe sequence, evidence to inspect, risks, and verification.
- `code`: implementation suggestions, target files/functions, and smallest safe patch shape.
- `review`: bug/security/regression/missing-test findings from provided context.
- `weakness`: Codex blind spots, missing evidence, unsafe assumptions, and checks to close gaps.
- `all`: combined planning, code, review, and weakness audit.
- Prefer `all` for general Codex tasks unless a narrower mode is clearly sufficient.
- Use `-Prompt` or `-PromptFile` for agent calls. Do not rely on piped stdin through nested `powershell -File`.

## Source Of Truth

- Claude must read the active project source, local docs, tests, and runtime output first.
- GLM output is advisory only. It cannot establish APIs, schemas, business rules, or completion status.
- GLM must not edit files directly. Claude applies patches after checking against source evidence.
- Claude remains responsible for analysis, implementation, verification, and final reporting.
- Claude should explicitly consider GLM's Asian/Thai business/accounting/POS perspective, then accept, reject, or adapt it based on inspected evidence.
- For accounting, ERP, security, migration, deployment, or production work, final decisions must be based on Claude-inspected source/runtime evidence.

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
powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\glm52-think-planner.ps1 -Mode weakness -Prompt "what could Claude miss here"
```

## Output Handling

- Use the visible `content` only as planning advice.
- Do not expose or persist raw `reasoning_content`.
- The helper prints `reasoningchars` so Claude can verify thinking was used without revealing chain-of-thought.
- Merge useful GLM points into Claude's own plan, patch, review, or verification checklist only after checking them against local source/runtime evidence.

## Secondary advisor — codex-advisor

GLM 5.2 is the **primary** advisor. The **secondary** advisor is Codex
(`gpt-5.5 think`) via the sibling skill `codex-advisor` and helper
`tools/ai/codex-advisor.ps1` — use it for engineering bug/security/edge-case
angles. The synthesis loop is identical: Claude (active agent) reads evidence
first, asks the advisor(s), then synthesizes the best of all models and verifies
(see `ai-team-governance.md` §11).
