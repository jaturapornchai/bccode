---
date: 2026-09-20
status: accepted  # proposed | accepted | deprecated | superseded
tags: [bc-account, docs, process, multi-ai, skills]
supersedes: 2026-09-07-consolidate-docs-for-multi-ai.md (เฉพาะส่วนตำแหน่ง skill)
---

# กฎและ skill ของ AI ทุกตัวอยู่ที่เดียว: `AGENTS.md` + `.agents/skills/`

## Context

ลุงจืดใช้ AI หลายตัวใน repo นี้ (Claude Code/Desktop, Codex, Gemini CLI, Antigravity, ZCode) และสั่ง 2026-09-20 ว่า "ต้องการให้รวมกฏ skill ฯลฯ ไว้ที่เดียวกัน แล้วให้ ai ทุกตัวมาใช้ จะได้ตรวจง่าย แก้ไขง่าย"

สภาพก่อนแก้ (ตรวจ 2026-09-20):
- กฎ: `AGENTS.md` เป็นตัวจริงอยู่แล้ว — Codex อ่านตรง, Claude ผ่าน `CLAUDE.md` (`@AGENTS.md`), ZCode อ่าน `AGENTS.md` (ส่วน `CLAUDE.md` ใช้แค่ตอน migrate ครั้งแรก), Antigravity อ่าน `AGENTS.md` ได้ตั้งแต่ IDE 1.20.5 — **Gemini CLI (0.44.1 บนเครื่องนี้) ไม่อ่าน** เพราะค่าเริ่มต้นหา `GEMINI.md` เท่านั้น
- Skill: อยู่ `docs/skills/` ตาม ADR 2026-09-07 — **ไม่มี AI ตัวไหน auto-discover** ทุกตัวต้องพึ่งกฎ "เปิดอ่านเอง" ซึ่งพลาดง่าย; `~/.zcode/cli/config.json` ยังจำ path `D:/bccode/.agents/skills/flutter-expert/SKILL.md` ไว้ = หลักฐานว่า ZCode สแกน `.agents/skills/` ของ repo นี้จริง
- ตำแหน่งที่แต่ละเครื่องมือค้นหา skill ระดับโปรเจ็กต์ (ตรวจจากเอกสารทางการ 2026-09-20): Codex `.agents/skills/` · Gemini CLI `.gemini/skills/` หรือ alias `.agents/skills/` · Antigravity `.agents/skills/` · ZCode `.agents/skills/` (จาก config บนเครื่อง) · **Claude Code `.claude/skills/` เท่านั้น** (ไม่รองรับ `.agents/skills/` — issue anthropics/claude-code#31005)

## Decision

| สิ่งที่ | ตัวจริง | เครื่องมือแต่ละตัวมาอ่าน |
|---|---|---|
| กฎโปรเจ็กต์ | `AGENTS.md` | Codex/ZCode/Antigravity ตรง · Claude `CLAUDE.md` → `@AGENTS.md` · Gemini CLI `.gemini/settings.json` → `{"context":{"fileName":["AGENTS.md","GEMINI.md"]}}` |
| Skill | `.agents/skills/<name>/SKILL.md` (`git mv` จาก `docs/skills/`) | Codex/Gemini CLI/Antigravity/ZCode สแกนเอง · Claude Code: `npm run ai:link` สร้าง junction `.claude/skills` → `.agents/skills` (`tools/ai-link.mjs`; `.claude/` gitignored จึงรันครั้งเดียวต่อ clone) |
| ฐานความรู้ | `docs/kms/` (คงเดิม) | ทุกตัวอ่านตามผัง `docs/README.md` |
| ข้อกำหนดของลุงจืด | `mydocs/` (คงเดิม, read-only) | — |

- ห้ามสร้างไฟล์กฎต่อเครื่องมือ (`GEMINI.md`, `.agents/rules/*`, `.cursorrules`, `.codex/AGENTS.md`) — ถ้าจำเป็นทำเป็น pointer มาที่ `AGENTS.md` เท่านั้น
- ห้ามคัดลอก skill ไปที่อื่น; แก้ที่ `.agents/skills/` แล้ว commit พร้อมงาน (กฎ Mandatory Skill Upgrade เดิม)
- path อ้างอิง `docs/skills/` → `.agents/skills/` แก้ทั้ง repo ยกเว้น ADR 2026-09-07 (เก็บเป็นประวัติ ตั้ง status superseded)

## Alternatives

- **คง `docs/skills/` + junction ไปทุกเครื่องมือ** — ปฏิเสธ: junction ไม่ถูก commit → clone ใหม่/เครื่องอื่นไม่มี, Codex/Gemini/ZCode จะไม่เห็น skill
- **สร้าง `.agents/rules/` ให้ Antigravity** — ไม่จำเป็น: อ่าน `AGENTS.md` ที่ root ได้แล้ว และจะกลายเป็นสำเนากฎที่ drift
- **symlink `.claude/skills` แบบ commit ลง git** — ปฏิเสธ: symlink บน Windows ต้องเปิด Developer Mode/`core.symlinks` ทุกเครื่อง; junction ผ่านสคริปต์ปลอดภัยกว่า
- **รวมกฎ global 3 ไฟล์ (`~/.claude/CLAUDE.md`, `~/.codex/AGENTS.md`, `~/.gemini/GEMINI.md`)** — นอกขอบเขต repo; พบว่า drift แล้ว (Codex global บอกห้ามใช้ DeepSeek ขณะที่ Claude global บอกให้ใช้) → เสนอลุงจืดแยกต่างหาก

## Consequences

- ✅ แก้กฎที่เดียว (`AGENTS.md`) แก้ skill ที่เดียว (`.agents/skills/`) — 4 ใน 5 เครื่องมือ auto-discover, Claude ผ่าน junction
- ⚠️ clone ใหม่ต้องรัน `npm run ai:link` (และ `npm run hooks:install`) เอง ไม่งั้น Claude Code จะไม่เห็น skill — แต่กฎ "เปิดอ่าน SKILL.md จาก disk ก่อนงาน" ยังคุ้มครองอยู่
- ⚠️ Gemini CLI โหลด `.agents/skills/` เฉพาะ workspace ที่ `/trust` แล้ว
- ⚠️ ADR/เอกสารเก่าที่กล่าวถึง `docs/skills/` ถูกแก้ path แล้ว (ยกเว้น ADR 2026-09-07 ที่เก็บไว้เป็นประวัติ)
