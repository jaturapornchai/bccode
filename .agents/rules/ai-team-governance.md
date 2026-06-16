# AI Team Governance — Multi-Agent (BC Account)

> ธรรมนูญทีม AI สำหรับ `D:\bccode`. ตั้งโดยลุงจืด.
> **ZCode** = หัวหน้า + ผู้ตรวจงาน + full-stack worker (GLM-5.2 powered) · **Codex (think)** = ผู้ช่วย/advisor (gpt-5.5 think) · Gemini/Antigravity = ออกแบบ UX/UI.
> ทุก agent รัน **เหมาจ่าย (subscription/OAuth) เท่านั้น — ห้าม per-token API key**.
> นี่เป็น **soft role division**: เป็น default routing ไม่ใช่ role-lock — agent ใดทำ layer อื่นได้เมื่องาน end-to-end ต้องการ.

## 0. โครงทีม
| Agent | บทบาท | หน้าที่ |
|---|---|---|
| **ZCode** | 👑 หัวหน้า + 🔍 ผู้ตรวจ + 🛠️ full-stack | วางแผน, แตกงาน, มอบหมาย, ตรวจรับ, merge. **เขียน production code เองได้ทุก layer** (backend/frontend/glue) — ไม่จำกัดแค่ glue อีกต่อไป |
| **Codex (think)** | 🧠 ผู้ช่วย/advisor | ให้คำแนะนำเชิงลึก (gpt-5.5 think) สำหรับงานยาก/สำคัญ ผ่าน `tools/ai/codex-advisor.ps1`. ทำงานเป็น advisor เท่านั้น — ไม่ได้เป็น worker หลัก (ZCode synthesize แล้วแก้ไฟล์เอง) |
| **Gemini CLI / Antigravity (agy)** | 🎨 นักออกแบบ UX/UI | ออกแบบ layout/flow/หน้าตา ให้สวย+ใช้ง่าย+เข้าใจง่าย |

## 1. การมอบหมาย (Delegation)
| ลักษณะงาน | Worker | โมเดล / คำสั่ง |
|---|---|---|
| Backend logic / API / DB / refactor ใหญ่ / debug ยาก | ZCode | GLM-5.2 (default) — เรียก Codex advisor เมื่อยาก (§11) |
| Frontend implementation / integration | ZCode | GLM-5.2 (default) — เรียก Codex advisor เมื่อยาก (§11) |
| งานย่อย / boilerplate / แก้เล็ก | ZCode | GLM-5.2 |
| คำแนะนำเชิงลึก / ตรวจสอบ blind spot (advisor) | Codex (think) | `gpt-5.5` ผ่าน `tools/ai/codex-advisor.ps1` |
| ออกแบบ UX/UI (automated) | Gemini | `dispatch-design.sh` (plan/read-only) |
| ออกแบบ UX/UI (iterate) | agy | **interactive** — ลุงจืดเปิด `agy` ในเทอร์มินัลเอง |
| วางแผน / ตรวจงาน / merge | ZCode | — |

หลัง Gemini/agy ออกแบบเสร็จ → ส่งสเปก UI ให้ **ZCode** implement เป็นโค้ดจริง (เรียก Codex advisor ถ้า implementation ยาก).

## 2. คำสั่งจริง (verified 2026-06-04)
เรียกผ่าน Bash tool (git-bash). ใช้ helper ใน `.agents/orchestration/`:
```bash
# Codex — advisor (gpt-5.5 think, advisory only — ZCode applies patches itself)
bash .agents/orchestration/dispatch-codex.sh --deep "<5-section spec: ask for analysis/plan/review, not file edits>"
# Gemini — ออกแบบ UX/UI (read-only)
bash .agents/orchestration/dispatch-design.sh "<design brief>"
# Codex advisor helper (direct, cleaner wrapper for advisory use)
powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\codex-advisor.ps1 -Mode all -Prompt "<advisory question>"
```
ดิบ (ถ้าไม่ใช้ helper):
```bash
codex exec --skip-git-repo-check --ephemeral -m gpt-5.5 -s workspace-write -c model_reasoning_effort=high -c 'mcp_servers={}' "<spec>"
GEMINI_CLI_TRUST_WORKSPACE=true gemini -p "<brief>" --approval-mode plan -o text
```
รันขนานแบบแยก worktree (กัน merge ชน): `git worktree add ../wt-x -b feat/x` แล้ว `cd` เข้าไปรัน worker ใน worktree.

## 3. เหมาจ่าย — ห้าม per-token (บังคับ)
- ZCode → subscription (GLM-5.2 powered CLI) · Codex → `codex login` แบบ ChatGPT sign-in · agy/Gemini → Google OAuth.
- ❌ ห้ามตั้ง `OPENAI_API_KEY` / `GEMINI_API_KEY` / `GOOGLE_API_KEY` / `ZAI_API_KEY` เป็น per-token billing (= จ่ายตาม token). helper จะ **ABORT** ถ้าเจอ `OPENAI_API_KEY`.
- โดน rate limit / โควต้าหมด → **หยุดรอ refresh หรือสลับ `gpt-5.4-mini`** ห้าม fallback ไป API key.
- (ฝั่ง OpenAI dashboard) ตั้ง hard spending limit = $0 — ลุงจืดต้องตั้งเอง (นอกเครื่อง).

## 4. Handoff Protocol — สเปกต้องครบ 5 หัวข้อ
ทุก prompt ที่ส่ง Codex/Gemini ต้องมี (กัน hallucination + หลุดสเปก):
1. **GOAL** — เป้าหมาย 1-2 ประโยค
2. **CONTEXT** — ไฟล์/โมดูลที่เกี่ยว (`path:line`)
3. **DECISIONS** — สิ่งที่ตัดสินใจแล้ว (stack, ห้ามเพิ่ม dep ฯลฯ)
4. **CONSTRAINTS** — กฎ (snake_case, holdingcode tenant, MongoDB-first, security)
5. **VERIFICATION** — คำสั่งที่ต้องผ่าน (`cd frontend; npm run typecheck` / touched-package `go build` / `/healthz`)

ข้ามสายงานยาว ใช้ `.agents/handoffs/` (template มีอยู่แล้ว) เป็น context contract.

## 5. Audit Checklist — หัวหน้าตรวจหลังลูกน้องส่ง (ห้ามเชื่อทันที)
- [ ] ตรงสเปก (GOAL/CONSTRAINTS ครบ)
- [ ] **รัน VERIFICATION จริงเอง** — ไม่เชื่อคำลูกน้อง (กฎ VERIFY BEFORE DONE)
- [ ] ไม่มี hallucination (เรียก API/func/file ที่ไม่มีจริง — grep ตรวจ)
- [ ] ปลอดภัย (ไม่มี secret hardcode, ไม่มี string-concat SQL, จัดการ error)
- [ ] multi-tenant: แยก `holdingcode` ถูก ไม่รั่วข้าม tenant
- [ ] อ่านง่าย/ไม่ over-engineer (YAGNI)
- [ ] (งาน UI) สวย + ใช้ง่าย + เข้าใจง่าย ตามโจทย์
ไม่ผ่าน → เขียน feedback ชัด ส่งกลับ agent เดิมแก้ (อย่าแก้เอง).
งานสำคัญ/ใหญ่/เสี่ยง → เสริมด้วย **Cross-AI Review** (Codex advisor หา bug/security, Gemini ดู UX/UI) ตาม §10 แล้ว ZCode สังเคราะห์.

## 6. Workflow มาตรฐาน
`PLAN` (ZCode แตกงาน เลือก agent) → `DELEGATE` (ส่ง 5-section spec, ขนานได้ผ่าน worktree) → `AUDIT` (ตรวจ + รัน test จริง) → `FIX` (ส่ง feedback กลับ) → `INTEGRATE` (merge → สรุปไทยให้ลุงจืด).

## 7. Known limitations / สถานะ (verified 2026-06-04)
- **agy headless เสียบน Windows**: ติดตั้ง v1.0.5 + auth Google ผ่าน + generate ได้จริง แต่ `-p` ไม่ flush response ออก stdout (typing-animation ต้อง TTY จริง). → automated design ใช้ `gemini` (engine Gemini เดียวกัน), agy ใช้ interactive.
- **Gemini ต้อง trust**: ใส่ `GEMINI_CLI_TRUST_WORKSPACE=true` เสมอ (helper ใส่ให้แล้ว).
- **Codex config**: `~/.codex/config.toml` ตั้ง `sandbox_mode="danger-full-access"` สำหรับ interactive ของลุงจืด — automated dispatch ผ่าน helper force `workspace-write` เสมอ (ปลอดภัยกว่า ไม่แตะ config).
- **ZCode (GLM-5.2 powered)**: เป็น active agent หลักตามคำสั่งลุงจืด (set 2026-06-17) — ทำหน้าที่ leader + auditor + full-stack worker เอง. Codex เป็น advisor (gpt-5.5 think) เท่านั้นในทีมนี้. Claude ถูกถอดออกจากทีมแล้ว (เลิกใช้).
- สื่อสารกับลุงจืด = ภาษาไทยเสมอ. prompt ถึง worker = ชัด/กระชับ, symbol/path/command เป๊ะ.

## 8. งานสร้างภาพ — Codex สร้างเท่านั้น, ห้าม SVG/vector (set 2026-06-15)
- ภาพ visual ทุกชนิดในโปรเจค (page background, header, banner, illustration ประกอบ, ภาพตัวอย่าง) → **Codex สร้างเท่านั้น** ผ่าน `codex exec -m gpt-5.5 -c model_reasoning_effort=high`. ZCode / Gemini / agy **ห้าม generate ไฟล์ภาพเอง** — สั่ง/มอบ Codex แล้วนำ asset ที่อนุมัติแล้วไปต่อ UI.
- **format: raster จริงเท่านั้น** — `.webp` (page background ตามกฏ Semantic Page Backgrounds), `.png` / `.jpg` ได้สำหรับงานภาพอื่น. ❌ **ห้าม `.svg` / vector / inline `<svg>` เป็น asset ภาพทุกกรณี** (ไอคอน UI จาก icon library เช่น `lucide-react` ไม่นับ — นั่นคือ component ไม่ใช่ไฟล์ภาพ).
- **brief สั้น ปล่อย Codex ออกแบบ**: ส่งแค่ concept + ข้อความที่ต้องการ (เน้นไทย ตัวใหญ่ อ่านง่าย) + "ใช้เครื่องมือสร้างภาพจริง (image_gen) ทำให้สวย/เหมือนจริงที่สุด". ❌ ห้าม over-specify hex/layout/shape ละเอียด — Codex จะ fallback ไปวาด Pillow flat แทน image_gen.
- **ZCode verify เสมอ**: Read ดูภาพจริง (เหมือนจริง + ข้อความถูก + ขึ้นจอ + กว้าง ≥1200px + embed responsive) ก่อนบอกเสร็จ. ZCode ไม่วาดเอง.
- ข้อยกเว้น: inline viz ในแชท (`show_widget`) ใช้ SVG/HTML ได้ — กฏนี้คุมเฉพาะ **ไฟล์ภาพ asset ในโปรเจค**.

## 9. Dual-track — ZCode + Codex advisor เขียน/วางแผนขนาน แล้ว ZCode ตัดสินใจรวม (set 2026-06-15, updated 2026-06-17)
สำหรับงาน coding **สำคัญ / ยาก / มีหลายวิธีทำ** (design, refactor ใหญ่, แก้ bug ยาก, algorithm, implementation ที่มี trade-off):
1. **ZCode คิด+เขียน solution ของตัวเอง** พร้อมกับ **ถาม Codex advisor ทำโจทย์เดียวกัน** (spec 5-section เดียวกัน, ผ่าน `codex-advisor.ps1` — รันขนานได้) → ได้ 2 มุมมองอิสระ.
2. **ZCode เปรียบเทียบ 2 ผล แล้วตัดสินใจ**: เลือกอันที่ดีกว่า / รวมจุดเด่นทั้งคู่ / สังเคราะห์เป็นอันใหม่. **ZCode เป็นผู้ตัดสินสุดท้าย + รับผิดชอบผลลัพธ์** (รัน VERIFICATION จริงก่อนเสร็จเสมอ ตาม §5).
3. **บอกลุงจืดสั้นๆ** ว่าเลือก/รวมยังไง + เพราะอะไร (เอาจุดเด่นของ Codex advisor ตรงไหน, Codex เก่งกว่าตรงไหน) — ไม่ทิ้ง 2 ผลดิบให้เทียบเอง.
4. **cost-aware (Pareto)**: งานเล็ก / ตรงไปตรงมา / 1 วิธีชัดเจน → **ไม่ต้อง dual** (ZCode ทำเอง). dual เปลือง ~2 เท่า ใช้เฉพาะงานที่คุ้มได้ 2 มุมมอง.

## 10. Cross-AI Review — Codex advisor + Gemini ช่วยตรวจ (set 2026-06-16, updated 2026-06-17)
นอกจาก ZCode audit (§5), งาน **สำคัญ / ใหญ่ / เสี่ยง / ก่อน commit-push ใหญ่** ให้ขอ review หลายมุมจาก AI อื่นด้วย แล้ว ZCode สังเคราะห์ + ตัดสิน:
1. **Codex advisor review** = หา bug / security / logic / edge case / correctness:
   - diff ล่าสุด: `codex exec review` หรือ `codex review`
   - งานใหญ่/เจาะจง: `powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\ai\codex-advisor.ps1 -Mode review -Prompt "review <scope>: หา bug, security, edge case, logic ผิด — ตอบเป็น list พร้อม file:line"`
2. **Gemini review** = UX/UI / อ่านง่าย / เข้าใจง่าย / consistency / accessibility (read-only):
   - `bash .agents/orchestration/dispatch-design.sh "review <ไฟล์/จอ>: ชี้จุด UX/UI ที่ควรปรับให้ใช้ง่าย+เข้าใจง่าย+สวยขึ้น เป็น list"`
3. **ZCode สังเคราะห์ + ตัดสิน (รับผิดชอบสุดท้าย)** — รวมผล review ทั้ง 3 ฝั่ง (ZCode+Codex+Gemini), คัดอันจริง / ตัด false positive, **รัน VERIFICATION จริงเองเสมอ** (Codex/Gemini review = ความเห็นเพิ่ม ไม่แทน ZCode verify ตาม IRON RULE 2), แก้เท่าที่ควร แล้วกลั่นเป็นสรุปไทยให้ลุงจืด (ไม่ทิ้ง raw log)
4. **cost-aware (Pareto)**: งานเล็ก / 1-2 บรรทัด / trivial → **ZCode ตรวจพอ ไม่ต้อง cross-review**. ใช้ cross-review เฉพาะงานที่คุ้ม (เปลือง token+latency เพิ่ม) — เกณฑ์เดียวกับ §9
5. **R rules**: การ review เป็น read-only (R2) เรียกได้เลย; ถ้า "แก้ตาม review" แตะ R0/R1 ยัง flag ลุงจืดก่อนตามปกติ (IRON RULE 5)

## 11. GLM ↔ Codex Bidirectional Advisory — เอาความรู้ 2 model มารวมกัน (set 2026-06-17)
งาน BC Account ที่ **สำคัญ / ยาก / มีหลายวิธีทำ / เสี่ยง** (planning, implementation, code review, debug, UX/UI, weakness check) ต้องเอาความรู้ของ **2 model** มารวมกัน แล้วเลือก/รวมเอาที่ดีที่สุด:

**สองทิศทาง (bidirectional) — เลือกตามว่าใครเป็น active agent ตอนนั้น:**

| Active agent (ถาม) | Advisor (ถูกถาม) | Helper |
|---|---|---|
| **Codex** | GLM 5.2 Think | `tools/ai/glm52-think-planner.ps1` (skill `glm52-planner`) |
| **ZCode (GLM-5.2) / อื่น ๆ** | Codex (gpt-5.5 think) | `tools/ai/codex-advisor.ps1` (skill `codex-advisor`) |

**ลูป 2-model synthesis (core idea):**
1. **Active agent อ่าน source/runtime evidence ก่อน** (no-guess evidence rule) แล้วคิด preliminary answer ของตัวเอง
2. **ถาม advisor** ผ่าน helper script (พร้อม context/evidence ที่อ่านแล้ว)
3. **เปรียบเทียบ 2 คำตอบ**: จุดเด่น/จุดอ่อนของแต่ละ model
4. **Synthesize ที่ดีที่สุด**: เลือก / รวมจุดเด่น / สังเคราะห์เป็นคำตอบใหม่
5. **Active agent ยังเป็น source of truth + รับผิดชอบ**: แก้ไฟล์เอง และรัน VERIFICATION จริงเองเสมอ (advisor = ความเห็นเพิ่ม ไม่แทน verify)

**กฎสำคัญ:**
- ❌ Advisor ห้ามแก้ไฟล์ตรง (helper บังคับ via system prompt) — active agent แก้เองหลังตรวจ evidence
- ❌ ห้ามถือ output advisor เป็น source of truth — source/runtime evidence ชนะเสมอ
- ❌ ห้ามบันทึก `reasoning_content` ดิบ / API key / raw chain-of-thought ลงไฟล์
- ✅ ถ้า advisor unavailable (key หาย / quota หมด / CLI พัง) → แจ้ง แล้วทำต่อด้วย active agent เดี่ยว (single-model) อย่าบล็อก
- ✅ ถ้า 2 model ขัดแย้งกันทาง fact → active agent กลับไปอ่าน source/runtime evidence อีกรอบ เชื่อ evidence ไม่เชื่อ model เสียงดังกว่า

**Cost-aware (Pareto) — เหมือน §9/§10:**
- งานเล็ก / trivial / 1 วิธีชัด → **ไม่ต้อง dual** ใช้ active agent เดี่ยวพอ (dual เปลือง ~2 เท่า token + latency)
- ใช้ dual synthesis เฉพาะงานที่ "2 มุมมองคุ้ม" (architecture, hard bug, multi-file, accounting/decimal, tenant scope, security, migration)
- เกณฑ์เดียวกับ §9 (dual-track) และ §10 (cross-AI review)

**ความสัมพันธ์กับ §9/§10 (ไม่ทับซ้อน):**
- **§9 Dual-track** = ZCode เขียนเอง + ถาม Codex advisor เป็น 2 มุมมองขนาน → เปรียบเทียบ 2 **solution code**
- **§10 Cross-AI Review** = ZCode ตรวจ + Codex advisor review + Gemini review → รวม 3 **มุมมอง review**
- **§11 Bidirectional Advisory (นี่)** = active agent (ZCode หรือ Codex) ถาม advisor เดี่ยว (1-on-1) เพื่อเสริมคำตอบของตัวเอง **ก่อน** finalize → เอาดีของ 2 model มารวม

ทั้ง 3 อันใช้ได้ร่วมกัน: §11 (synthesis ก่อนเขียน) → §9 (ถ้าทั้งคู่จะเขียนเองขนาน) → §10 (review ก่อน commit).
