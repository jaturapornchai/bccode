---
name: audit-mongomodel-sync
description: Pull the live BC Ai Account design from MongoModel MCP, compare its revision, diagrams, collections, fields, indexes, relations, and workflows with the applicable requirements and current implementation, and report changes required in both BC code and MongoModel. Use for D:\bccode tasks involving MongoDB schema, data models, technical workflows, model-backed APIs/events, when Jead says MongoModel was edited, or when asked what must also be updated in MongoModel. Do not use for unrelated UI, copy, or infrastructure work that cannot affect a model or workflow.
---

# Audit MongoModel Sync

Follow `D:\bccode\AGENTS.md`, and use `D:\bccode\docs\kms\00-source-router.md` only to locate implementation evidence. Business-rule documentation in `docs/` was removed on 2026-09-03 and its replacement is pending, so there is no business-rule Source of Truth yet. Never infer authority from recency, from a higher MongoModel revision, or from existing code.

## Fixed Inputs

- BC repository: `D:\bccode`
- MongoModel project: `BC Ai Account`
- MCP dependency: `mongomodel` at `http://localhost:3100/mcp`
- MongoModel implementation repository: `D:\mongomodel`

Do not recreate or rely on `D:\bccode\datamodel` or `D:\bccode\datamodelwikillm`. Read diagrams, workflows, and wiki output live through MCP.

## MCP Access

Prefer the registered `mongomodel` MCP tools. If the current Codex session has not loaded that dependency but the loopback service is running, use the read-only fallback:

```powershell
& "docs/skills/audit-mongomodel-sync/scripts/invoke-mongomodel-read.ps1" `
  -Tool list_projects `
  -ArgumentsJson '{}'
```

The fallback still reads the live MCP endpoint and supports every read tool this workflow uses (`list_projects`, `list_diagrams`, `list_workflows`, `get_diagram`, `get_workflow`, `get_shared_brain`, `list_revisions`, `check_descriptions`, `lint_model`, `lint_workflows`, `generate_code`, `get_project_context`, `wait_for_project_change`). It must never read `D:\mongomodel\data\projects.json` as a substitute and refuses mutation tools.

## Required Workflow

1. Follow `D:\bccode\AGENTS.md`.
2. Apply the revision gate only when the request depends on or can affect MongoModel:
   - Call `list_projects` once, select the exact project `BC Ai Account`, and record `projectRev`.
   - Reuse evidence only when the current Codex thread still contains the exact `projectRev`, affected artifact name or ID, and sufficient detail level.
   - If `projectRev` is unchanged and every affected artifact was already inspected at the required detail, do not call `get_diagram`, `get_workflow`, or `get_project_context` again.
   - Invalidate reuse for a new Codex session, changed revision, missing evidence, newly affected artifact, or insufficient detail.
   - Do not create a persistent local schema copy or treat cached evidence as a Source of Truth.
3. Determine the affected diagrams, collections, and workflows from the request, applicable docs, and exact source evidence.
4. Pull only evidence that cannot be reused:
   - Use `list_diagrams` or `list_workflows` only when the exact name or ID is unknown.
   - Call `get_diagram` with `detail=outline` first.
   - Request `detail=summary` filtered to affected collections only when fields, indexes, or relations are required. Request `detail=full` only when exact raw structure is required.
   - Call `get_workflow` only for affected workflows, beginning with `detail=outline`; request more detail only when required.
   - Use `get_shared_brain`, `list_revisions`, or `get_project_context` only when that specific evidence is required.
5. Do not call `wait_for_project_change` during ordinary tasks. Use it only when Jead explicitly asks Codex to monitor or wait for MongoModel changes.
6. Call `get_project_context` at most once and only for a project-wide audit, an unresolved cross-diagram dependency, or a revision change whose affected scope cannot be identified through narrow calls.
7. Inspect the exact BC source, tests, schema, configuration, API/event contracts, and runtime evidence for the same scope.
8. Compare both directions. Never assume MongoModel and source are synchronized automatically.

## Comparison Checklist

Check every applicable item:

- Collection name, ownership, tenant scope, and lifecycle.
- Field name, BSON/API type, required/optional state, default, children, and description.
- Business keys, composite key groups, exact indexes, uniqueness, sparsity, and field order.
- Relations, source and target fields, cardinality, and cross-diagram references.
- Workflow steps, branches, states, failure paths, and `dataAccess` collection/field/operation references.
- API, event, Kafka, cache, PostgreSQL, and ClickHouse contracts coupled to the model.
- Tests, validation, serialization, and exact accounting-number representation when applicable.

## Classify Every Difference

Place each difference in exactly one group:

1. **Docs decision needed** — a business requirement in the affected scope is missing, unclear, or conflicting. There is no business-rule Source of Truth in `docs/` right now, so stop and ask Jead instead of guessing (`AGENTS.md` rule 1).
2. **BC implementation change** — source, tests, or runtime differ from the exact implementation evidence located through `D:\bccode\docs\kms\00-source-router.md`, or from a business rule Jead has confirmed.
3. **MongoModel content change** — live MongoModel does not represent an applicable confirmed requirement, or a confirmed Data Model/Technical Workflow change has not been recorded there.
4. **MongoModel MCP product change** — the MCP/UI/tool behavior cannot represent, validate, retrieve, or safely update the required model. Follow the maintenance route in `D:\bccode\docs\kms\00-source-router.md`.
5. **Aligned** — no change is required for the inspected scope.

Always include the MongoModel change section, even when the result is `ไม่มีรายการต้องปรับ`.

## Applying Authorized Changes

- For audit or review requests, remain read-only and report the exact required changes.
- For an implementation request, get Jead's confirmation for every unresolved business rule first, then apply the narrowest change and verify it (`AGENTS.md`).
- Before any MongoModel mutation, reread the relevant diagram or workflow and use its current revision/expected revision when the tool supports it.
- Never bulk-replace a diagram when a narrow field, collection, relation, or workflow update is sufficient.
- Never write generated output over BC source blindly. Use `generate_code` only as comparison evidence, then apply a reviewed patch to the exact source.

## Verification

After a MongoModel content change:

1. Call `check_descriptions` for the affected diagram.
2. Call `lint_model` and, when workflows are affected, `lint_workflows`.
3. Call `list_projects` to record the new `projectRev`, then reload only the affected diagram/workflow.
4. Reload `get_project_context` only when the mutation was project-wide or changed cross-diagram structure that narrow reads cannot verify.
5. Report the before and after `projectRev` values.

After a BC implementation change, run the narrowest source/runtime tests that prove the affected contract. If both sides changed, verify both sides before claiming alignment.

## Output Contract

Return a concise Thai report containing:

- `MONGOMODEL SNAPSHOT` — project, revision, inspected diagrams/workflows.
- `DOCS GAP/CONFLICT` — exact file and unresolved point, or `ไม่มี`.
- `BC IMPLEMENTATION` — required code/schema/test/runtime changes, or `ไม่มี`.
- `MONGOMODEL CONTENT` — required diagram/workflow changes, or `ไม่มี`.
- `MONGOMODEL MCP` — required MCP/UI/tooling changes, or `ไม่มี`.
- `VERIFICATION` — commands/tool calls and observed evidence.

If MCP is unavailable, report the exact connection/tool failure and stop model-affecting work. Do not fall back to a deleted local wiki or infer the model from the junction.
