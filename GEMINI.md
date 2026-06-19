# BC Ai Account Gemini / Antigravity Entry

> ⚠️ **DEPRECATED (2026-06-19): Gemini / Antigravity (agy) ถอดออกจากทีม AI แล้ว.**
> ทีมปัจจุบัน = **Claude (Claude Code, หัวหน้า/คนคุม) + GLM 5.2 Think (ที่ปรึกษาหลัก, ไทย/เอเซีย) + Codex (สร้างรูป webp + code advisor รอง)**.
> งาน UX/UI ตอนนี้ Claude implement เอง โดยปรึกษา GLM สำหรับมุมคนไทย/เอเซีย. ดู [`.agents/rules/ai-team-governance.md`](.agents/rules/ai-team-governance.md).
> ไฟล์นี้คงไว้เป็น entry-routing เผื่อ Gemini ถูกนำกลับเข้าทีมในอนาคต — ปัจจุบันไม่ใช้ใน automated workflow.

Read in this order:
1. `AGENTS.md`
2. `.agents/rules/bc-account-core-rules.md`
3. `.agents/skills/bc-central-rules/SKILL.md` when working on rules, skills, wiki/LLM knowledge, runtime, deploy, database, storage, or cross-agent instructions
4. `.agents/wiki/llm-index.md` when working on reusable knowledge, source routing, prompts, handoffs, or agent context
5. `AI_INDEX.md`
6. Only the files routed by the current task

`AGENTS.md` is the source of truth for project rules, development environment, DEV deployment, database location, cross-agent compatibility, security, frontend, backend, and verification requirements.

Keep context small. Do not open generated docs, lockfiles, screenshots, manuals, build outputs, runtime logs, or full large files unless explicitly required.

For normal coding work, prefer focused inspection, the smallest safe patch, and focused verification. Keep any reusable rules, skills, prompts, handoffs, checklists, or workflows portable across Claude Code, Codex/GPT-5.5, and Google Antigravity/Gemini.
