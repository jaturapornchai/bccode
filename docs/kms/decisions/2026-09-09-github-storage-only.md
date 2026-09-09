---
date: 2026-09-09
status: accepted  # proposed | accepted | deprecated | superseded
tags: [bc-account, ci, tooling, github, decision-by-owner]
---

# GitHub เป็นที่เก็บโค้ดอย่างเดียว — ยกเลิก CI ทั้งหมด ย้ายการตรวจมาที่เครื่อง

## Context

- `.github/workflows/ci.yml` มี 5 jobs (backend-test, frontend-test, code-map-check, backend-outbox-integration, backend-projection-kafka-integration) แต่ **ไม่ได้รันเลยตั้งแต่ 2026-09-02**
- สาเหตุยืนยันจาก GitHub เอง (ไม่ใช่การเดา): `gh api repos/jaturapornchai/bccode/check-runs/101924873905/annotations` ตอบว่า

  > "The job was not started because your account is locked due to a billing issue."

  ทุก run หลัง 2026-09-02 จบด้วย `failure` ภายใน 3-5 วินาที โดย job ไม่มี step แม้แต่ step เดียว (`gh api .../actions/runs/34182717816/jobs` → `steps: 0` ทั้ง 4 job)
- **repo นี้เป็น public** (`gh repo view --json visibility` → `PUBLIC`) และ GitHub Actions บน public repo + standard runner **ฟรีไม่จำกัด ไม่นับนาที** — ค่าใช้จ่ายที่ทำให้บัญชีถูกล็อกจึงมาจากที่อื่นในบัญชี ไม่ใช่จาก repo นี้
- ลุงจืดตัดสินใจ 2026-09-09: **"ต้องการให้ github เก็บ code อย่างเดียว ไม่ต้องการจ่าย github แล้ว"**

## Decision

1. **ลบ `.github/workflows/ci.yml` ทิ้งถาวร** (โฟลเดอร์ root `.github/` หายไปทั้งโฟลเดอร์) — GitHub เหลือหน้าที่เดียวคือ remote `origin` สำหรับเก็บ/แชร์โค้ด
2. **ย้ายคำสั่งทั้งหมดมาที่ `tools/verify.sh`** — คัดลอกจาก workflow เดิม**แบบคำต่อคำ** เพื่อให้ coverage เท่าเดิม:

   | target | เท่ากับ job เดิม | หมายเหตุ |
   |---|---|---|
   | `codemap` | `code-map-check` | เรียก `tools/gen-code-map.ps1 -Check` |
   | `frontend` | `frontend-test` (ครึ่งแรก) | lint + typecheck + vitest |
   | `frontend-build` | `frontend-test` (ครึ่งหลัง) | `next build` แยกออกมาเพราะช้า |
   | `backend` | `backend-test` | `golang:1.26` + quarantine list |
   | `outbox` | `backend-outbox-integration` | mongo rs0 + postgres:17-alpine |
   | `projection` | `backend-projection-kafka-integration` | `backend/.ci/projection.compose.yml` |

   ทางลัด: `npm run verify` (= `fast` = codemap + frontend) ก่อน push, `npm run verify:all` ก่อน deploy
3. **`.githooks/pre-commit` คือการตรวจอัตโนมัติเดียวที่เหลือ** และตรวจแค่ `docs/reference/CODE-MAP.md` — ทุก clone ต้อง `npm run hooks:install` เอง
4. กู้ workflow เดิมได้เสมอ: `git show 1799b069:.github/workflows/ci.yml > .github/workflows/ci.yml`

## Consequences

**ผลดี**
- ไม่มี run สีแดงติดทุก commit ใน GitHub อีกต่อไป และไม่มีเอกสารไหนหลอกว่า "CI จะจับให้"
- ชุดตรวจเดียวกันรันได้บนเครื่อง dev ทันที ไม่ต้องรอ runner และไม่ผูกกับสถานะบัญชี GitHub
- ไม่มีค่าใช้จ่าย GitHub ผูกกับ repo นี้เลย

**ผลเสีย / ความเสี่ยง (ยอมรับแล้ว)**
- **ไม่มีอะไรบังคับให้ตรวจ** — ถ้าไม่มีคนพิมพ์ `npm run verify` ก็ไม่มีใครรู้ว่าพัง; PR จากคนนอกจะไม่ถูกตรวจอัตโนมัติเลย
- ผลตรวจไม่ถูกเก็บเป็นหลักฐานกลาง (ไม่มี artifact/summary ในระบบ) ต้องแปะผลเองตอนรายงาน
- พิสูจน์ได้ทันทีว่าความเสี่ยงนี้เป็นจริง: รัน `sh tools/verify.sh frontend` ครั้งแรกหลังย้าย **ไม่ผ่าน** — eslint มี 2 error ค้างอยู่ที่ HEAD (`frontend/src/app/menu/dashboard-home.tsx:194` เรียก `useMemo` แบบมีเงื่อนไข, `frontend/src/app/menu/manage-shortcuts-screen.tsx:133` เรียก `Date.now()` ระหว่าง render) ซึ่งหลุดเข้ามาช่วงที่ CI ตายพอดี
- `backend/.github/workflows/*.yaml` (9 ไฟล์ยุค repo เก่า) ยังอยู่ — ไม่เคยถูก GitHub รันเพราะไม่ได้อยู่ root แต่ยังไม่ลบเพราะอยู่นอกขอบเขตงานนี้ (ต้องถามลุงจืดก่อน)

## Alternatives considered

| ทางเลือก | ทำไมไม่เลือก |
|---|---|
| เก็บ `ci.yml` ไว้แต่เปลี่ยนเป็น `workflow_dispatch` อย่างเดียว | บัญชีถูกล็อกอยู่ดี กดเองก็ไม่รัน และไฟล์ที่ค้างไว้ทำให้เข้าใจผิดว่ายังมี CI |
| ปลดล็อก billing แล้วใช้ Actions ฟรีต่อ (public repo ไม่เสียเงิน) | เป็นเรื่องบัญชีส่วนตัวของลุงจืด — ลุงจืดสั่งชัดว่าไม่ต้องการยุ่งกับ GitHub billing อีก |
| ตั้ง self-hosted runner บน `.202` | ยังต้องผ่าน GitHub Actions ซึ่งถูกล็อกที่ระดับบัญชี จึงไม่แก้ปัญหา |
| ย้าย CI ไป GitLab/Woodpecker/Drone ที่ self-host | เพิ่มระบบใหม่ให้ดูแลโดยไม่มีคนขอ — ถ้าภายหลังต้องการค่อยพิจารณา |

## Reference

- `tools/verify.sh` — ตัวรันจริง
- `.githooks/pre-commit`, `tools/install-hooks.mjs` — การตรวจอัตโนมัติที่เหลือ
- `docs/kms/11-testing-quality.md` §5 — ตารางเทียบ target ↔ job เดิม
- `docs/kms/15-known-issues.md` KI-23 — ประวัติการล็อก billing (ปิดเรื่องด้วย ADR นี้)
