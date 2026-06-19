# AI Team Governance — Multi-Agent (BC Account)

> ธรรมนูญทีม AI สำหรับ `D:\bccode`. ตั้งโดยลุงจืด (restructured 2026-06-19).
> **Claude (Claude Code)** = 👑 หัวหน้า/คนคุม + ผู้ตรวจ + full-stack worker · **GLM 5.2 Think (1m)** = 🧠 ที่ปรึกษาหลัก (เก่งคนเอเซีย/ระบบงานไทย + code) · **Codex (gpt-5.5 think)** = 🎨 สร้างรูป (webp/png/jpg, ห้าม SVG) + 🧠 code advisor รอง.
> **Gemini/Antigravity ถอดออกจากทีมแล้ว** (set 2026-06-19) — Claude implement UX/UI เอง, ปรึกษา GLM สำหรับมุมคนไทย/เอเซีย.
> ทุก advisor รัน **เหมาจ่าย (subscription/OAuth) เท่านั้น — ห้าม per-token API key**.
> นี่เป็น **soft role division**: เป็น default routing ไม่ใช่ role-lock — Claude ทำได้ทุก layer เมื่องาน end-to-end ต้องการ.

## 0. โครงทีม
| Agent | บทบาท | หน้าที่ |
|---|---|---|
| **Claude (Claude Code)** | 👑 หัวหน้า + 🔍 ผู้ตรวจ + 🛠️ full-stack worker | วางแผน, เขียน production code ทุก layer (backend/frontend/glue), แก้ diff, รัน VERIFICATION จริง, ตัดสินใจสุดท้าย, **รับผิดชอบผลทั้งหมด**, สรุปไทยให้ลุงจืด |
| **GLM 5.2 Think (1m)** | 🧠 ที่ปรึกษาหลัก (primary advisor) | ให้คำแนะนำเชิงลึก: planning, code, review, bug/security/weakness, blind-spot. **เก่งเป็นพิเศษเรื่องคนเอเซีย/ระบบงานไทย** — ธุรกิจไทย บัญชีไทย ERP/POS ไทย ภาษี/เอกสารไทย ค้าปลีก/ร้านอาหาร Thai UX. ผ่าน `tools/ai/glm52-think-planner.ps1` (skill `glm52-planner`). Advisory only — Claude synthesize แล้วแก้ไฟล์เอง |
| **Codex (gpt-5.5 think)** | 🎨 สร้างรูป + 🧠 code advisor รอง | (1) สร้างไฟล์ภาพ raster ทั้งหมด — `.webp`/`.png`/`.jpg` **ห้าม `.svg`/vector** (§8). (2) advisor รอง: หา bug/security/edge-case/logic/blind-spot ผ่าน `tools/ai/codex-advisor.ps1` (skill `codex-advisor`) หรือ `codex exec review`. Advisory only — Claude synthesize เอง |

## 1. การมอบหมาย (Delegation)
| ลักษณะงาน | Worker | ที่ปรึกษา / คำสั่ง |
|---|---|---|
| Backend logic / API / DB / refactor ใหญ่ / debug ยาก | Claude | ปรึกษา GLM (หลัก) / Codex (รอง) เมื่อยาก (§11) |
| Frontend implementation / integration / UX/UI | Claude | ปรึกษา GLM เรื่อง Thai UX/flow เมื่อต้องการ (§11) |
| งานย่อย / boilerplate / แก้เล็ก | Claude | ทำเอง ไม่ต้องปรึกษา |
| คำแนะนำเชิงลึก ไทย/เอเซีย / business / code / blind-spot | GLM 5.2 (หลัก) | `tools/ai/glm52-think-planner.ps1` |
| code review รอง / bug / security / edge-case | Codex (รอง) | `tools/ai/codex-advisor.ps1` หรือ `codex exec review` |
| สร้างไฟล์ภาพ (background/banner/illustration) | Codex | `codex exec -m gpt-5.5 ... high` (§8) |
| วางแผน / ตรวจงาน / verify / merge / สรุป | Claude | — |

## 2. คำสั่งจริง
เรียก advisor ผ่าน Bash tool (git-bash) หรือ PowerShell helper:
```bash
# GLM 5.2 Think — ที่ปรึกษาหลัก (code + ไทย/เอเซีย)
powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\glm52-think-planner.ps1 -Mode all -Prompt "<advisory question + evidence ที่อ่านแล้ว>"
# Codex — code advisor รอง (bug/security/blind-spot)
powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\codex-advisor.ps1 -Mode review -Prompt "<review scope: หา bug/security/edge-case>"
# Codex — สร้างไฟล์ภาพ (§8)
codex exec --skip-git-repo-check -m gpt-5.5 -s workspace-write -c model_reasoning_effort=high "<brief สั้น: concept + ข้อความไทย + 'ใช้ image_gen สร้างจริง .webp'>"
```
รันงานขนานแบบแยก worktree (กัน merge ชน): `git worktree add ../wt-x -b feat/x` แล้ว `cd` เข้าไปทำใน worktree.

## 3. เหมาจ่าย — ห้าม per-token (บังคับ)
- GLM → Z.AI subscription (`ZAI_API_KEY` แบบ subscription/coding-plan ไม่ใช่ per-token billing) · Codex → `codex login` แบบ ChatGPT sign-in.
- ❌ ห้ามตั้ง `OPENAI_API_KEY` / `ZAI_API_KEY` / `GEMINI_API_KEY` / `GOOGLE_API_KEY` เป็น per-token billing. helper จะ **ABORT** ถ้าเจอ `OPENAI_API_KEY`.
- โดน rate limit / โควต้าหมด → **หยุดรอ refresh หรือสลับ `gpt-5.4-mini`** (Codex) ห้าม fallback ไป API key. GLM unavailable → แจ้ง fallback แล้ว Claude ทำต่อเดี่ยวเมื่อปลอดภัย.
- ❌ ห้ามเขียน `ZAI_API_KEY` / `OPENAI_API_KEY` ลง repo/docs/rules/prompts/logs/commits/PR เด็ดขาด.

## 4. Handoff Protocol — สเปกที่ส่ง advisor ต้องครบ 5 หัวข้อ
ทุก prompt ที่ส่ง GLM/Codex ต้องมี (กัน hallucination + หลุดสเปก):
1. **GOAL** — เป้าหมาย 1-2 ประโยค
2. **CONTEXT** — ไฟล์/โมดูลที่เกี่ยว (`path:line`) + evidence ที่ Claude อ่านแล้ว
3. **DECISIONS** — สิ่งที่ตัดสินใจแล้ว (stack, ห้ามเพิ่ม dep ฯลฯ)
4. **CONSTRAINTS** — กฎ (snake_case→ no-underscore, holdingcode tenant, MongoDB-first, decimal, security)
5. **VERIFICATION** — คำสั่งที่ต้องผ่าน (`cd frontend; npm run typecheck` / touched-package `go build` / `/healthz`)

ข้ามสายงานยาว ใช้ `.agents/handoffs/` (template มีอยู่แล้ว) เป็น context contract.

## 5. Audit Checklist — Claude ตรวจหลังรับคำแนะนำ advisor (ห้ามเชื่อทันที)
- [ ] ตรงสเปก (GOAL/CONSTRAINTS ครบ)
- [ ] **รัน VERIFICATION จริงเอง** — ไม่เชื่อคำ advisor (กฎ VERIFY BEFORE DONE)
- [ ] ไม่มี hallucination (เรียก API/func/file ที่ไม่มีจริง — grep ตรวจ)
- [ ] ปลอดภัย (ไม่มี secret hardcode, ไม่มี string-concat SQL, จัดการ error)
- [ ] multi-tenant: แยก `holdingcode` ถูก ไม่รั่วข้าม tenant
- [ ] decimal: เงิน/ราคา/ต้นทุน ไม่ใช้ float (ERP Accounting Decimal Iron Rule)
- [ ] อ่านง่าย/ไม่ over-engineer (YAGNI)
- [ ] (งาน UI) สวย + ใช้ง่าย + เข้าใจง่าย (Premium UX/UI Standard) + responsive 375px
GLM/Codex = ความเห็นเพิ่ม ไม่แทน Claude verify. Claude คัด false positive แล้วแก้เอง.

## 6. Workflow มาตรฐาน
`PLAN` (Claude แตกงาน) → `ADVISE` (ปรึกษา GLM/Codex เมื่อคุ้ม, §11) → `IMPLEMENT` (Claude เขียนเอง, ขนานได้ผ่าน worktree) → `AUDIT` (ตรวจ + รัน test จริง, §5) → `INTEGRATE` (merge → สรุปไทยให้ลุงจืด).

## 7. Known limitations / สถานะ (verified 2026-06-19)
- **Claude (Claude Code) = active agent หลัก** ตามคำสั่งลุงจืด (set 2026-06-19) — leader + auditor + full-stack worker เอง. รับผิดชอบ source/runtime verification + final answer ทั้งหมด.
- **GLM 5.2 Think (1m)** = ที่ปรึกษาหลัก ผ่าน `tools/ai/glm52-think-planner.ps1` (ต้อง `ZAI_API_KEY` subscription). ถ้า key หาย/quota หมด → แจ้ง fallback แล้ว Claude ทำต่อเดี่ยว.
- **Codex (gpt-5.5 think)** = สร้างไฟล์ภาพ + code advisor รอง. `~/.codex/config.toml` interactive ของลุงจืดเป็น `danger-full-access` แต่ image/advisor dispatch ผ่าน helper force `workspace-write` เสมอ.
- **Gemini/Antigravity (agy) ถอดออกจากทีมแล้ว** (set 2026-06-19) — ไม่ใช้ใน automated workflow อีก. งาน UX/UI Claude implement เอง (ปรึกษา GLM เรื่องมุมคนไทย/เอเซีย). `dispatch-design.sh` ถูกลบ.
- สื่อสารกับลุงจืด = ภาษาไทยเสมอ. prompt ถึง advisor = ชัด/กระชับ, symbol/path/command เป๊ะ.

## 8. งานสร้างภาพ — Codex สร้างเท่านั้น, ห้าม SVG/vector (set 2026-06-15, reaffirmed 2026-06-19)
- ภาพ visual ทุกชนิดในโปรเจค (page background, header, banner, illustration ประกอบ, ภาพตัวอย่าง) → **Codex สร้างเท่านั้น** ผ่าน `codex exec -m gpt-5.5 -c model_reasoning_effort=high`. Claude / GLM **ห้าม generate ไฟล์ภาพเอง** — สั่ง/มอบ Codex แล้วนำ asset ที่อนุมัติแล้วไปต่อ UI.
- **format: raster จริงเท่านั้น** — `.webp` (page background ตามกฏ Semantic Page Backgrounds), `.png` / `.jpg` ได้สำหรับงานภาพอื่น. ❌ **ห้าม `.svg` / vector / inline `<svg>` เป็น asset ภาพทุกกรณี** (ไอคอน UI จาก icon library เช่น `lucide-react` ไม่นับ — นั่นคือ component ไม่ใช่ไฟล์ภาพ).
- **brief สั้น ปล่อย Codex ออกแบบ**: ส่งแค่ concept + ข้อความที่ต้องการ (เน้นไทย ตัวใหญ่ อ่านง่าย) + "ใช้เครื่องมือสร้างภาพจริง (image_gen) ทำให้สวย/เหมือนจริงที่สุด". ❌ ห้าม over-specify hex/layout/shape ละเอียด — Codex จะ fallback ไปวาด Pillow flat แทน image_gen.
- **Claude verify เสมอ**: Read ดูภาพจริง (เหมือนจริง + ข้อความถูก + ขึ้นจอ + กว้าง ≥1200px + embed responsive) ก่อนบอกเสร็จ. Claude ไม่วาดเอง.
- ข้อยกเว้น: inline viz ในแชท (`show_widget`) ใช้ SVG/HTML ได้ — กฏนี้คุมเฉพาะ **ไฟล์ภาพ asset ในโปรเจค**.

## 9. Dual-track — Claude เขียนเอง + ถาม advisor ขนาน แล้ว Claude ตัดสินรวม (set 2026-06-19)
สำหรับงาน coding **สำคัญ / ยาก / มีหลายวิธีทำ** (design, refactor ใหญ่, แก้ bug ยาก, algorithm, implementation ที่มี trade-off):
1. **Claude คิด+เขียน solution ของตัวเอง** พร้อมกับ **ถาม GLM 5.2 (และ/หรือ Codex) ทำโจทย์เดียวกัน** (spec 5-section เดียวกัน, ผ่าน helper — รันขนานได้) → ได้หลายมุมมองอิสระ.
2. **Claude เปรียบเทียบผล แล้วตัดสินใจ**: เลือกอันที่ดีกว่า / รวมจุดเด่น / สังเคราะห์เป็นอันใหม่. **Claude เป็นผู้ตัดสินสุดท้าย + รับผิดชอบผลลัพธ์** (รัน VERIFICATION จริงก่อนเสร็จเสมอ ตาม §5).
3. **บอกลุงจืดสั้นๆ** ว่าเลือก/รวมยังไง + เพราะอะไร (เอาจุดเด่นของ advisor ตรงไหน) — ไม่ทิ้งผลดิบให้เทียบเอง.
4. **cost-aware (Pareto)**: งานเล็ก / ตรงไปตรงมา / 1 วิธีชัดเจน → **ไม่ต้อง dual** (Claude ทำเอง). dual เปลือง ~2 เท่า ใช้เฉพาะงานที่คุ้ม.

## 10. Cross-AI Review — GLM + Codex ช่วยตรวจ แล้ว Claude สังเคราะห์ (set 2026-06-19)
นอกจาก Claude audit (§5), งาน **สำคัญ / ใหญ่ / เสี่ยง / ก่อน commit-push ใหญ่** ให้ขอ review หลายมุมจาก advisor แล้ว Claude สังเคราะห์ + ตัดสิน:
1. **GLM 5.2 review** (หลัก) = หา bug / business-mismatch (มุมไทย/เอเซีย) / security / logic / edge case:
   - `powershell ... glm52-think-planner.ps1 -Mode review -Prompt "review <scope>: หา bug, business mismatch, security, edge case — list พร้อม file:line"`
   - หรือ `-Mode weakness` หาจุดอ่อน/สิ่งที่ Claude อาจพลาด.
2. **Codex review** (รอง) = bug / security / correctness / edge case มุม engineering:
   - diff ล่าสุด: `codex exec review` หรือ `codex review`
   - งานเจาะจง: `powershell ... codex-advisor.ps1 -Mode review -Prompt "review <scope> ..."`
3. **Claude สังเคราะห์ + ตัดสิน (รับผิดชอบสุดท้าย)** — รวมผล review ทั้ง 3 ฝั่ง (Claude+GLM+Codex), คัดอันจริง / ตัด false positive, **รัน VERIFICATION จริงเองเสมอ** (advisor review = ความเห็นเพิ่ม ไม่แทน Claude verify ตาม IRON RULE 2), แก้เท่าที่ควร แล้วกลั่นเป็นสรุปไทยให้ลุงจืด (ไม่ทิ้ง raw log).
4. **cost-aware (Pareto)**: งานเล็ก / 1-2 บรรทัด / trivial → **Claude ตรวจพอ ไม่ต้อง cross-review**.
5. **R rules**: review = read-only (R2) เรียกได้เลย; "แก้ตาม review" ถ้าแตะ R0/R1 ยัง flag ลุงจืดก่อน (IRON RULE 5).

## 11. Claude ↔ Advisor Synthesis — เอาความรู้หลาย model มารวมกัน (set 2026-06-19)
งาน BC Account ที่ **สำคัญ / ยาก / มีหลายวิธีทำ / เสี่ยง** (planning, implementation, code review, debug, UX/UI, weakness check) ให้ Claude เอาความรู้ของ advisor มาเสริม แล้วเลือก/รวมเอาที่ดีที่สุด:

| Active agent (ถาม) | Advisor (ถูกถาม) | เมื่อไหร่ | Helper |
|---|---|---|---|
| **Claude** | **GLM 5.2 Think** (หลัก) | เกือบทุกงานที่คุ้ม โดยเฉพาะมุมไทย/เอเซีย/business/บัญชี/POS + code | `tools/ai/glm52-think-planner.ps1` (skill `glm52-planner`) |
| **Claude** | **Codex** (รอง) | code review / bug / security / edge-case มุม engineering | `tools/ai/codex-advisor.ps1` (skill `codex-advisor`) |

**ลูป synthesis (core idea):**
1. **Claude อ่าน source/runtime evidence ก่อน** (No-Guess Evidence Rule) แล้วคิด preliminary answer ของตัวเอง.
2. **ถาม advisor** ผ่าน helper (พร้อม context/evidence ที่อ่านแล้ว) — GLM ก่อน (หลัก), เสริม Codex เมื่อเป็นเรื่อง code/security ลึก.
3. **เปรียบเทียบคำตอบ**: จุดเด่น/จุดอ่อนของแต่ละ model.
4. **Synthesize ที่ดีที่สุด**: เลือก / รวมจุดเด่น / สังเคราะห์เป็นคำตอบใหม่.
5. **Claude ยังเป็น source of truth + รับผิดชอบ**: แก้ไฟล์เอง และรัน VERIFICATION จริงเองเสมอ (advisor = ความเห็นเพิ่ม ไม่แทน verify).

**กฎสำคัญ:**
- ❌ Advisor ห้ามแก้ไฟล์ตรง (helper บังคับ via system prompt) — Claude แก้เองหลังตรวจ evidence.
- ❌ ห้ามถือ output advisor เป็น source of truth — source/runtime evidence ชนะเสมอ.
- ❌ ห้ามบันทึก `reasoning_content` ดิบ / API key / raw chain-of-thought ลงไฟล์.
- ✅ ถ้า advisor unavailable (key หาย / quota หมด / CLI พัง) → แจ้ง แล้วทำต่อด้วย Claude เดี่ยว (single-model) อย่าบล็อก.
- ✅ ถ้า model ขัดแย้งกันทาง fact → Claude กลับไปอ่าน source/runtime evidence อีกรอบ เชื่อ evidence ไม่เชื่อ model เสียงดังกว่า.

**Cost-aware (Pareto):**
- งานเล็ก / trivial / 1 วิธีชัด → **ไม่ต้องถาม advisor** ใช้ Claude เดี่ยวพอ.
- ใช้ synthesis เฉพาะงานที่ "มุมมองเพิ่มคุ้ม" (architecture, hard bug, multi-file, accounting/decimal, tenant scope, security, migration, Thai business flow).

**ความสัมพันธ์ §9/§10/§11:**
- **§9 Dual-track** = Claude เขียนเอง + ถาม advisor ขนาน → เปรียบเทียบ **solution code**.
- **§10 Cross-AI Review** = Claude ตรวจ + GLM review + Codex review → รวม **มุมมอง review** ก่อน commit.
- **§11 Synthesis (นี่)** = Claude ถาม advisor เพื่อเสริมคำตอบ **ก่อน** finalize → เอาดีของหลาย model มารวม.

ทั้ง 3 ใช้ร่วมกันได้: §11 (synthesis ก่อนเขียน) → §9 (ถ้าเขียนเองขนาน) → §10 (review ก่อน commit).
