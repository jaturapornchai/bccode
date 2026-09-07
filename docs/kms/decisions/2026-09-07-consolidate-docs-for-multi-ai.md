---
date: 2026-09-07
status: accepted  # proposed | accepted | deprecated | superseded
tags: [bc-account, docs, process, multi-ai]
---

# รวมเอกสาร/skill/ความรู้ทั้งหมดไว้ใต้ `docs/` เพื่อให้ AI หลายตัวเข้าใจตรงกัน

## Context

ลุงจืดใช้ AI หลายตัวใน repo นี้ (Claude, Codex, Gemini) — ความรู้ที่อยู่ใน memory ส่วนตัวของ AI ตัวเดียว, Obsidian vault ของ Claude, หรือโฟลเดอร์เฉพาะเครื่องมือ (`.agents/skills`) ตัวอื่นมองไม่เห็น; skill ยังถูกแก้ด้วยมือได้ จึงต้องมี "ตัวจริง" ที่เดียวบน disk. คำสั่งวันนี้: "ผมสร้าง folder docs/skills… เก็บ skill ที่ผมใช้", "docs/kms เอาไว้เก็บฐานความรู้ต่างๆ จะได้ไม่ลืม", "ให้ทุกอย่างไปรวมใน docs เลย ai ตัวอื่นๆ จะได้เข้าใจตรงกัน", "bcdev ไม่ต้องสนใจใน project นี้"

## Decision

- `docs/skills/<name>/SKILL.md` = ที่เดียวของ skill ส่วนตัว (ย้ายจาก `.agents/skills/` ด้วย `git mv`; `.agents/` ลบ) — อ่านจาก disk ทุกครั้ง ห้ามใช้สำเนา/cache
- `docs/kms/` = ฐานความรู้ร่วม: บทความ 00–18 (15 บทความจากการอ่านโค้ดทั้ง repo ที่ commit `d93a210d` โดยนักอ่าน 15 ตัว + fact-checker 15 ตัวเปิดไฟล์ตรวจทุก citation แก้ 222 จุด; 3 บทความจาก memory/handoff), `decisions/` (ADR), `bugs/`, `snippets/`, `architecture/` (ย้ายจาก `backend/architecture/`)
- `docs/handoff/` (ย้าย `HANDOFF-*.md` จาก root), `docs/runbooks/` (ย้าย `RECOVERY-READINESS.md`), `docs/reference/CODE-MAP.md` (ย้ายจาก root; `tools/gen-code-map.ps1` ชี้ path ใหม่), `docs/kms/00-source-router.md` (เดิม `AI_INDEX.md`)
- `AGENTS.md` ข้อ 1–3 ชี้มาที่ `docs/` และมีกฎใหม่ "skill ส่วนตัวอยู่ที่ docs/skills และฐานความรู้อยู่ที่ docs/kms"; `CLAUDE.md` import `AGENTS.md` เหมือนเดิม
- ความรู้ที่โยงกับ `D:\bcdev` (project เก่า Flutter/pgvector) ไม่นำเข้า

## Alternatives

- เก็บ skill ที่ `.claude/skills` / `.agents/skills` ให้เครื่องมือ auto-discover — ปฏิเสธ: แต่ละ AI มองคนละที่ และลุงจืดตรวจ/แก้ยาก
- เก็บความรู้ใน memory ของ Claude + Obsidian ต่อ — ปฏิเสธ: AI ตัวอื่นไม่เห็น; Claude ยังจดใน memory ได้แต่ต้องสะท้อนลง `docs/kms` ด้วย

## Consequences

- ✅ ที่เดียวสำหรับทุก AI/คน; ทุกข้อเท็จจริงมี `path:line` + วันที่ตรวจ; ลุงจืดตรวจได้ใน repo
- ⚠️ Claude Code จะไม่ auto-load skill จาก `docs/skills/` — ต้องเปิดอ่านไฟล์ก่อนงาน UI/MongoModel ตามกฎ; docs จะเก่าลงถ้าโค้ดเปลี่ยนโดยไม่แก้ docs ใน commit เดียวกัน (กฎข้อ 3 ใน AGENTS.md ป้องกัน)
- ⚠️ package-level README (`backend/README.md`, `outbox/README.md` ฯลฯ) ยังอยู่ข้างโค้ดตามธรรมเนียม Go — ดัชนีใน `docs/kms/README.md` ชี้ไปหา
