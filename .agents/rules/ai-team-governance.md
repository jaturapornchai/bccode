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
4. **CONSTRAINTS** — กฎ (snake_case, holding_code tenant, MongoDB-first, security)
5. **VERIFICATION** — คำสั่งที่ต้องผ่าน (`cd frontend; npm run typecheck` / touched-package `go build` / `/healthz`)

ข้ามสายงานยาว ใช้ `.agents/handoffs/` (template มีอยู่แล้ว) เป็น context contract.

## 5. Audit Checklist — หัวหน้าตรวจหลังลูกน้องส่ง (ห้ามเชื่อทันที)
- [ ] ตรงสเปก (GOAL/CONSTRAINTS ครบ)
- [ ] **รัน VERIFICATION จริงเอง** — ไม่เชื่อคำลูกน้อง (กฎ VERIFY BEFORE DONE)
- [ ] ไม่มี hallucination (เรียก API/func/file ที่ไม่มีจริง — grep ตรวจ)
- [ ] ปลอดภัย (ไม่มี secret hardcode, ไม่มี string-concat SQL, จัดการ error)
- [ ] multi-tenant: แยก `holding_code` ถูก ไม่รั่วข้าม tenant
- [ ] อ่านง่าย/ไม่ over-engineer (YAGNI)
- [ ] (งาน UI) สวย + ใช้ง่าย + เข้าใจง่าย ตามโจทย์
ไม่ผ่าน → เขียน feedback ชัด ส่งกลับ agent เดิมแก้ (อย่าแก้เอง).

## 6. Workflow มาตรฐาน
`PLAN` (Claude แตกงาน เลือก agent) → `DELEGATE` (ส่ง 5-section spec, ขนานได้ผ่าน worktree) → `AUDIT` (ตรวจ + รัน test จริง) → `FIX` (ส่ง feedback กลับ) → `INTEGRATE` (merge → สรุปไทยให้ลุงจืด).

## 7. Known limitations / สถานะ (verified 2026-06-04)
- **agy headless เสียบน Windows**: ติดตั้ง v1.0.5 + auth Google ผ่าน + generate ได้จริง แต่ `-p` ไม่ flush response ออก stdout (typing-animation ต้อง TTY จริง). → automated design ใช้ `gemini` (engine Gemini เดียวกัน), agy ใช้ interactive.
- **Gemini ต้อง trust**: ใส่ `GEMINI_CLI_TRUST_WORKSPACE=true` เสมอ (helper ใส่ให้แล้ว).
- **Codex config**: `~/.codex/config.toml` ตั้ง `sandbox_mode="danger-full-access"` สำหรับ interactive ของลุงจืด — automated dispatch ผ่าน helper force `workspace-write` เสมอ (ปลอดภัยกว่า ไม่แตะ config).
- สื่อสารกับลุงจืด = ภาษาไทยเสมอ. prompt ถึง worker = ชัด/กระชับ, symbol/path/command เป๊ะ.
