---
name: bc-deploy
description: Prepare to deploy BC Account
disable-model-invocation: true
---

Prepare BC Account for deployment:

1. Read `D:\bccode\.agents\rules\bc-account-core-rules.md`; it is the canonical deploy/runtime/storage/secret/version source.
2. Use `D:\bccode\AI_INDEX.md` to route to exact deploy scripts, Docker files, frontend build config, backend config, and tests.
3. `deploy dev` means backend and frontend together unless Jead explicitly approves a partial emergency action.
4. Keep deploy instructions portable across Claude Code, Codex/GPT-5.5, and Google Antigravity/Gemini: plain Markdown, explicit paths, explicit commands, source-backed verification.
5. Use env/secret variable names only. Never write server passwords, MongoDB Atlas connection strings with credentials, Cloudflare tokens, API keys, or bearer tokens into docs, logs, commits, PRs, or chat summaries.
6. Verify the relevant tests/lint/build/migrations, endpoint health, API version compatibility, and rollback path before claiming deploy readiness.
7. Summarize changed artifacts, verification, public endpoints, compatibility impact, and rollback notes.

IMPORTANT: This skill must only be invoked via /bc-deploy — do not auto-invoke.
