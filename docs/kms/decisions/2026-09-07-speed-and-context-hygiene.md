---
date: 2026-09-07
status: accepted
tags: [bc-account, performance, context-efficiency, prompt-caching, multi-ai]
---

# กฎปฏิบัติการ: ความเร็วสูงสุด + ประหยัด Context และ Token

## Context

ในการพัฒนาซอฟต์แวร์ด้วย AI Agents (Claude Code, Gemini/Antigravity, Codex) บน codebase ขนาดใหญ่ การสูญเสีย token และการบวมของ Context Window เกิดจาก:
1. การเปิดอ่านไฟล์ขนาดยักษ์ (10,000+ บรรทัด) ทั้งไฟล์
2. การพ่น output จากคำสั่ง shell / test runner / git ยาวเป็นพันบรรทัด
3. การ rewrite โค้ดยาวๆ แทนการ patch เฉพาะส่วน
4. การค้นหาข้อมูลกว้างขวางใน main conversation thread

คำสั่งลุงจืด: "ทำเลย" เพื่อให้ระบบทำงานเร็วขึ้นแบบไม่เปลือง context และ token

## Decision

บังคับใช้ 5 กฎหลักในการทำงานของ AI ทุกตัว:
1. **Surgical Read**: อ่านไฟล์ด้วย StartLine/EndLine เสมอ และใช้ `docs/reference/CODE-MAP.md` นำทางไฟล์ขนาดเกิน 1,000 บรรทัด
2. **Surgical Patch**: แก้ไขเฉพาะบล็อกที่เปลี่ยนด้วย `replace_file_content` / targeted diff หลีกเลี่ยงการเขียนไฟล์ใหม่ทั้งก้อน
3. **Command Output Hygiene**:
   - รันเทสต์เฉพาะไฟล์ที่แตะต้อง (`npm test -- <path>.test.ts`)
   - คำสั่ง git ใช้ `git status -s` และ `git diff --stat`
   - คุมความยาว output ด้วย filter/head ไม่ให้เกิน 30-50 บรรทัด
4. **Context Isolation**: ใช้ Subagent ทำงานสำรวจค้นคว้าที่ต้องอ่านหลายไฟล์ แล้วส่งกลับมาเฉพาะผลลัพธ์สรุปสั้น
5. **Prompt Caching Friendly**: คงที่คำสั่งหลักใน `AGENTS.md` และ `docs/` เพื่อให้ติด prompt cache ประหยัด token 90% และประมวลผลเร็วขึ้น 2-4 เท่า

## Consequences

- ✅ การตอบสนองของ AI เร็วขึ้นทันที (latency ต่ำลงมาก)
- ✅ ไม่เปลือง token ทั้งฝั่ง Prompt และ Output
- ✅ ยืดอายุการสนทนาในแต่ละ session ได้ยาวนานขึ้นอย่างน้อย 5-10 เท่าโดยไม่ต้อง reset
