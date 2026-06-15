# AI Team Governance — Multi-Agent (BC Account)

> ธรรมนูญทีม AI สำหรับ `D:\bccode`. ตั้งโดยลุงจืด 2026-06-04.
> Claude = หัวหน้า + ผู้ตรวจงาน · Codex = เขียนโค้ด · Gemini/Antigravity = ออกแบบ UX/UI.
> ทุก agent รัน **เหมาจ่าย (subscription/OAuth) เท่านั้น — ห้าม per-token API key**.
> นี่เป็น **soft role division**: เป็น default routing ไม่ใช่ role-lock — agent ใดทำ layer อื่นได้เมื่องาน end-to-end ต้องการ.

## 0. โครงทีม
| Agent | บทบาท | หน้าที่ |
|---|---|---|
| **Claude Code** | 👑 หัวหน้า + 🔍 ผู้ตรวจ | วางแผน, แตกงาน, มอบหมาย, ตรวจรับ, merge. ไม่เขียน production code เอง ยกเว้น glue สั้น ๆ |
| **Codex** | 🛠️ วิศวกรโค้ด | เขียน/แก้ backend + frontend implementation ตามสเปก |
| **Gemini CLI / Antigravity (agy)** | 🎨 นักออกแบบ UX/UI | ออกแบบ layout/flow/หน้าตา ให้สวย+ใช้ง่าย+เข้าใจง่าย |

## 1. การมอบหมาย (Delegation)
| ลักษณะงาน | Worker | โมเดล / คำสั่ง |
|---|---|---|
| Backend logic / API / DB / refactor ใหญ่ / debug ยาก | Codex | `gpt-5.5` |
| Frontend implementation / integration | Codex | `gpt-5.5` |
| งานย่อย / ขนาน / boilerplate / แก้เล็ก | Codex | `gpt-5.4-mini` |
| ออกแบบ UX/UI (automated) | Gemini | `dispatch-design.sh` (plan/read-only) |
| ออกแบบ UX/UI (iterate) | agy | **interactive** — ลุงจืดเปิด `agy` ในเทอร์มินัลเอง |
| วางแผน / ตรวจงาน / merge | Claude | — |

หลัง Gemini/agy ออกแบบเสร็จ → ส่งสเปก UI ให้ **Codex** implement เป็นโค้ดจริง.

## 2. คำสั่งจริง (verified 2026-06-04)
เรียกผ่าน Bash tool (git-bash). ใช้ helper ใน `.agents/orchestration/`:
```bash
# Codex — เขียนโค้ด (workspace-write sandbox)
bash .agents/orchestration/dispatch-codex.sh        "<5-section spec>"   # gpt-5.5      + medium (complex)
bash .agents/orchestration/dispatch-codex.sh --mini "<5-section spec>"   # gpt-5.4-mini + low (fast/parallel)
bash .agents/orchestration/dispatch-codex.sh --deep "<5-section spec>"   # gpt-5.5      + high (architecture)
# Gemini — ออกแบบ UX/UI (read-only)
bash .agents/orchestration/dispatch-design.sh "<design brief>"
```
ดิบ (ถ้าไม่ใช้ helper):
```bash
codex exec --skip-git-repo-check --ephemeral -m gpt-5.5 -s workspace-write -c model_reasoning_effort=medium -c 'mcp_servers={}' "<spec>"
GEMINI_CLI_TRUST_WORKSPACE=true gemini -p "<brief>" --approval-mode plan -o text
```
รันขนานแบบแยก worktree (กัน merge ชน): `git worktree add ../wt-x -b feat/x` แล้ว `cd` เข้าไปรัน worker ใน worktree.

## 3. เหมาจ่าย — ห้าม per-token (บังคับ)
- Claude → Claude Max 20x · Codex → `codex login` แบบ ChatGPT sign-in · agy/Gemini → Google OAuth.
- ❌ ห้ามตั้ง `OPENAI_API_KEY` / `GEMINI_API_KEY` / `GOOGLE_API_KEY` (= จ่ายตาม token). helper จะ **ABORT** ถ้าเจอ.
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
งานสำคัญ/ใหญ่/เสี่ยง → เสริมด้วย **Cross-AI Review** (Codex หา bug/security, Gemini ดู UX/UI) ตาม §10 แล้ว Claude สังเคราะห์.

## 6. Workflow มาตรฐาน
`PLAN` (Claude แตกงาน เลือก agent) → `DELEGATE` (ส่ง 5-section spec, ขนานได้ผ่าน worktree) → `AUDIT` (ตรวจ + รัน test จริง) → `FIX` (ส่ง feedback กลับ) → `INTEGRATE` (merge → สรุปไทยให้ลุงจืด).

## 7. Known limitations / สถานะ (verified 2026-06-04)
- **agy headless เสียบน Windows**: ติดตั้ง v1.0.5 + auth Google ผ่าน + generate ได้จริง แต่ `-p` ไม่ flush response ออก stdout (typing-animation ต้อง TTY จริง). → automated design ใช้ `gemini` (engine Gemini เดียวกัน), agy ใช้ interactive.
- **Gemini ต้อง trust**: ใส่ `GEMINI_CLI_TRUST_WORKSPACE=true` เสมอ (helper ใส่ให้แล้ว).
- **Codex config**: `~/.codex/config.toml` ตั้ง `sandbox_mode="danger-full-access"` สำหรับ interactive ของลุงจืด — automated dispatch ผ่าน helper force `workspace-write` เสมอ (ปลอดภัยกว่า ไม่แตะ config).
- สื่อสารกับลุงจืด = ภาษาไทยเสมอ. prompt ถึง worker = ชัด/กระชับ, symbol/path/command เป๊ะ.

## 8. งานสร้างภาพ — Codex สร้างเท่านั้น, ห้าม SVG/vector (set 2026-06-15)
- ภาพ visual ทุกชนิดในโปรเจค (page background, header, banner, illustration ประกอบ, ภาพตัวอย่าง) → **Codex สร้างเท่านั้น** ผ่าน `codex exec -m gpt-5.5 -c model_reasoning_effort=high`. Claude / Gemini / agy **ห้าม generate ไฟล์ภาพเอง** — สั่ง/มอบ Codex แล้วนำ asset ที่อนุมัติแล้วไปต่อ UI.
- **format: raster จริงเท่านั้น** — `.webp` (page background ตามกฏ Semantic Page Backgrounds), `.png` / `.jpg` ได้สำหรับงานภาพอื่น. ❌ **ห้าม `.svg` / vector / inline `<svg>` เป็น asset ภาพทุกกรณี** (ไอคอน UI จาก icon library เช่น `lucide-react` ไม่นับ — นั่นคือ component ไม่ใช่ไฟล์ภาพ).
- **brief สั้น ปล่อย Codex ออกแบบ**: ส่งแค่ concept + ข้อความที่ต้องการ (เน้นไทย ตัวใหญ่ อ่านง่าย) + "ใช้เครื่องมือสร้างภาพจริง (image_gen) ทำให้สวย/เหมือนจริงที่สุด". ❌ ห้าม over-specify hex/layout/shape ละเอียด — Codex จะ fallback ไปวาด Pillow flat แทน image_gen.
- **Claude verify เสมอ**: Read ดูภาพจริง (เหมือนจริง + ข้อความถูก + ขึ้นจอ + กว้าง ≥1200px + embed responsive) ก่อนบอกเสร็จ. Claude ไม่วาดเอง.
- ข้อยกเว้น: inline viz ในแชท (`show_widget`) ใช้ SVG/HTML ได้ — กฏนี้คุมเฉพาะ **ไฟล์ภาพ asset ในโปรเจค**.

## 9. Dual-track — Codex + Claude เขียนโปรแกรมขนาน แล้ว Claude ตัดสินใจรวม (set 2026-06-15)
สำหรับงาน coding **สำคัญ / ยาก / มีหลายวิธีทำ** (design, refactor ใหญ่, แก้ bug ยาก, algorithm, implementation ที่มี trade-off):
1. **Claude คิด+เขียน solution ของตัวเอง** พร้อมกับ **มอบ Codex ทำโจทย์เดียวกัน** (spec 5-section เดียวกัน, รันขนาน — แยก `git worktree` ถ้าแตะไฟล์ชนกัน) → ได้ 2 มุมมองอิสระ.
2. **Claude เปรียบเทียบ 2 ผล แล้วตัดสินใจ**: เลือกอันที่ดีกว่า / รวมจุดเด่นทั้งคู่ / สังเคราะห์เป็นอันใหม่. **Claude เป็นผู้ตัดสินสุดท้าย + รับผิดชอบผลลัพธ์** (รัน VERIFICATION จริงก่อนเสร็จเสมอ ตาม §5).
3. **บอกลุงจืดสั้นๆ** ว่าเลือก/รวมยังไง + เพราะอะไร (แก้ของ Codex ตรงไหน, Codex เก่งกว่าตรงไหน) — ไม่ทิ้ง 2 ผลดิบให้เทียบเอง.
4. **cost-aware (Pareto)**: งานเล็ก / ตรงไปตรงมา / 1 วิธีชัดเจน → **ไม่ต้อง dual** (Claude ทำเอง หรือ Codex เดี่ยวตาม §1). dual เปลือง ~2 เท่า ใช้เฉพาะงานที่คุ้มได้ 2 มุมมอง.

## 10. Cross-AI Review — Codex + Gemini ช่วยตรวจ (set 2026-06-16)
นอกจาก Claude audit (§5), งาน **สำคัญ / ใหญ่ / เสี่ยง / ก่อน commit-push ใหญ่** ให้ขอ review หลายมุมจาก AI อื่นด้วย แล้ว Claude สังเคราะห์ + ตัดสิน:
1. **Codex review** = หา bug / security / logic / edge case / correctness:
   - diff ล่าสุด: `codex exec review` หรือ `codex review`
   - งานใหญ่/เจาะจง: `bash .agents/orchestration/dispatch-codex.sh --deep "review <scope>: หา bug, security, edge case, logic ผิด — ตอบเป็น list พร้อม file:line"`
2. **Gemini review** = UX/UI / อ่านง่าย / เข้าใจง่าย / consistency / accessibility (read-only):
   - `bash .agents/orchestration/dispatch-design.sh "review <ไฟล์/จอ>: ชี้จุด UX/UI ที่ควรปรับให้ใช้ง่าย+เข้าใจง่าย+สวยขึ้น เป็น list"`
3. **Claude สังเคราะห์ + ตัดสิน (รับผิดชอบสุดท้าย)** — รวมผล review ทั้ง 3 ฝั่ง (Claude+Codex+Gemini), คัดอันจริง / ตัด false positive, **รัน VERIFICATION จริงเองเสมอ** (Codex/Gemini review = ความเห็นเพิ่ม ไม่แทน Claude verify ตาม IRON RULE 2), แก้เท่าที่ควร แล้วกลั่นเป็นสรุปไทยให้ลุงจืด (ไม่ทิ้ง raw log)
4. **cost-aware (Pareto)**: งานเล็ก / 1-2 บรรทัด / trivial → **Claude ตรวจพอ ไม่ต้อง cross-review**. ใช้ cross-review เฉพาะงานที่คุ้ม (เปลือง token+latency เพิ่ม) — เกณฑ์เดียวกับ §9
5. **R rules**: การ review เป็น read-only (R2) เรียกได้เลย; ถ้า "แก้ตาม review" แตะ R0/R1 ยัง flag ลุงจืดก่อนตามปกติ (IRON RULE 5)
