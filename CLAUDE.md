# BC Ai Account — Agent Entry

> Identical entry for every agent (Claude Code / ZCode / Codex / any). `CLAUDE.md`, `GEMINI.md`, and `AGENTS.md` point to the same rule set. **Opus 4.8 is the orchestrator/lead** (set 2026-07-17 by Jead): it splits the task, dispatches to the advisor/subagent pool (GLM 5.2, Kimi K3 Max, DeepSeek, ChatGPT 5.6, Claude Fable, Claude Sonnet), evaluates the returned work, and re-assigns — while grounding decisions in real source/runtime evidence, applying, verifying, and owning the result. See AGENTS.md / core-rules "Multi-Model Orchestration".

Read in this order:
1. `AGENTS.md`
2. `.agents/rules/bc-account-core-rules.md`
3. `.agents/skills/bc-central-rules/SKILL.md` when working on rules, skills, wiki/LLM knowledge, runtime, deploy, database, storage, or cross-agent instructions
4. `.agents/wiki/llm-index.md` when working on reusable knowledge, source routing, prompts, handoffs, or agent context
5. `AI_INDEX.md`
6. Only the files routed by the current task

Keep context small. Do not open generated docs, lockfiles, screenshots, manuals, or full large files unless explicitly required.

For coding work, prefer focused inspection, the smallest safe patch, and focused verification.
