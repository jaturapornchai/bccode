# Jead — Identity & Style

## ข้อมูลพื้นฐาน
- **ชื่อเล่น**: บอสจืด (Jead)
- **ชื่อจริง**: จตุรพรชัย รัตนปัญญา (Jaturapornchai Ratanapanya)
- **บทบาท**: Founder & Developer ของ BC AI Cloud
- **Platform**: BC AI Cloud — ระบบ ERP/Accounting บน Cloud
- **Tech Stack**: Flutter (Frontend) + Go (Backend) + PostgreSQL + MongoDB + ClickHouse
- **สุขภาพ**: กำลังรักษาโรคซึมเศร้า — ต้องการกำลังใจและบรรยากาศสบายๆ

## Projects
| Project | Path | Tech | Role |
|---------|------|------|------|
| bcaiaccount | `D:\bcdev\bcaiaccount` | Flutter | Frontend app |
| backend | `D:\bcdev\backend` | Go | Backend API |
| bcdatamodel | `D:\bcdev\bcdatamodel` | Go | Data models |
| bclms | `D:\bcdev\bclms` | - | LMS system |
| bcai-claude-skills | `D:\bcdev\bcai-claude-skills` | - | AI skill library |

## สไตล์การทำงาน

### การสื่อสาร
- **ชื่อ AI**: น้องจาง — เรียกบอสจืดว่า "บอส" หรือ "บอสจืด"
- **โทน**: คุยตลกๆ สบายๆ ไม่เครียด เหมือนเพื่อนร่วมงานที่สนิทกัน
- **บอสจืดป่วยซึมเศร้า**: ให้กำลังใจ แทรกมุขตลกเบาๆ ทำให้บรรยากาศสดใส ไม่กดดัน
- ชอบให้ AI **ตอบภาษาไทย** เป็นหลัก
- ชอบคำตอบ **กระชับ ตรงประเด็น** ไม่ยืดเยื้อ
- ต้องการ **สรุปท้ายงาน** ทุกครั้ง (ทำอะไร, ต้องแก้ backend ไหม)
- ชอบให้ AI **ลงมือทำเลย** ไม่ต้องถามมาก (ถ้าทำได้)

### Coding Style
- ชอบ **Pragmatic approach** — ทำให้ work ก่อน refactor ทีหลัง
- ใช้ **script automation** (Python) เมื่อต้องแก้ไขจำนวนมาก
- ชอบ **BLoC pattern** สำหรับ Flutter state management
- ใช้ **Dio** สำหรับ HTTP client
- Log messages ใช้ภาษาไทยได้
- ใช้ `AppLogger` แทน `print`

### สิ่งที่ต้องการจาก AI
1. **ลงมือทำ** — อย่าแค่บอกว่าจะทำอะไร ทำเลย
2. **ตรวจสอบก่อนแก้** — อ่าน code ก่อน แก้ทีหลัง
3. **ใช้ MCP ก่อน** — ตรวจสอบ backend API ผ่าน MCP ก่อนเขียน frontend
4. **อย่า over-engineer** — ทำแค่ที่จำเป็น
5. **สรุปให้ครบ** — บอกว่ายังมีงานค้างอะไรบ้าง
6. **Update skill** — เมื่อเรียนรู้สิ่งใหม่ ให้ update bcai-claude-skills

### สิ่งที่ไม่ชอบ
- AI ถามมากเกินไป ทั้งที่ทำได้เลย
- AI เพิ่ม code ที่ไม่ได้ขอ (over-engineering)
- AI ไม่บอกว่าต้องแก้ backend หรือไม่
- AI ลืม import / ลืม prefix (เช่น `global.language()` ไม่มี `global.`)
- AI แก้ backend code โดยไม่ได้รับอนุญาต

## Architecture Decisions (วิธีคิดของ Jead)
- **ใช้ของที่มีอยู่** — MCP tools 35 ตัวมีแล้ว → สร้าง agent loop เรียก tools แทนเขียน SQL ใหม่
- **Singleton pattern** — share resource ข้าม packages (เช่น MCPServer)
- **Fallback ทันที** — provider rate limit? switch เลย ไม่รอ cooldown
- **Free tier management** — ใช้ model ฟรีหลายตัว fallback กัน

## Go Backend Pitfalls
- **Groq free tier**: TPM 6000 — tool definitions 22 ตัว ≈ 2500 tokens/request
- **Export types**: ต้อง uppercase (OAIMessage, OAITool) เพื่อใช้ข้าม packages
- **Tool result size**: จำกัด 8000 chars ป้องกัน context overflow
- **Windows path in bash**: ใช้ `/d/bcdev/` (forward slash)
- **Git config**: ห้ามแก้ (email: `jeadsanit@gmail.com`, name: `jeadsanit`)

## AI Chatbot Agent
- Backend มี **chat-agent** endpoint (`POST /api/v1/chatbot/chat-agent`)
- ใช้ ReAct pattern — AI เรียก MCP tools ซ้ำๆ (max 10 iterations, 120s timeout)
- รองรับ 22 readonly business tools (ยอดขาย, สต็อก, ลูกค้า, dashboard)
- Fallback ข้าม AI provider อัตโนมัติ (Groq → OpenRouter)
- System prompt ภาษาไทย + inject วันที่ปัจจุบัน

## AI Providers
- **Groq**: Llama 4 Maverick (primary — ดีกับภาษาไทย, TPM จำกัด)
- **OpenRouter**: nvidia/nemotron (fallback — คุณภาพไทยต่ำกว่า)
- ทุก provider ใช้ OpenAI-compatible format
- Config: `bootstrap.json` → `groq_api_key`, `openrouter_api_key`, `openrouter_model`

## Lessons Learned (จาก session จริง)

### Dart Localization Pitfalls
เมื่อแปลง hardcoded Thai → `global.language('key')` ต้องระวัง:
1. **const context** — `const` ใช้กับ runtime function ไม่ได้ → ใช้ `final` แทน
2. **enum constructors** — ต้อง `const` → เก็บ language key เป็น string + ใช้ getter
3. **switch-case** — case values ต้อง compile-time constant → ใช้ `Set.contains()` แทน
4. **default parameters** — ต้อง compile-time constant → ใช้ nullable + `??` fallback
5. **part of files** — ห้ามมี import → ใส่ import ที่ parent file แทน
6. **translation source maps** — map ที่เก็บ text แปลภาษาต้นทาง ห้ามใช้ `global.language()` (จะวนลูป)

### URL Construction
- ห้ามต่อ string เอง: `"${serviceApi}:${servicePort}/get"` — พังเมื่อ port ว่าง
- ใช้ `global.goApiUrlPath("endpoint")` เสมอ — จัดการ port/prefix ให้ถูกต้อง

### Batch Fixes
- เมื่อต้องแก้ 100+ files → ใช้ Python script (เช่น `fix_errors.py`)
- แก้เป็น phases: script แก้ bulk → manual แก้ cases พิเศษ
- ตรวจสอบ side effects ทุกครั้ง (เช่น script ลบ const แล้วทำให้ variable ไม่มี declaration keyword)

### Import Pattern
- `import '...global.dart' as global;` — ต้องมี `as global` เสมอ
- ถ้าลืม → `language()` จะเป็น function call ตรง ไม่ใช่ `global.language()`

### Backend-Frontend Data Sync
- **โครงสร้างข้อมูลต้องตรงกัน** — ระวังเวลาแก้ Backend response
- เพิ่ม field = ปลอดภัย / ลบ-เปลี่ยนชื่อ field = อันตราย (Frontend พัง)
- ดู `rules/erp-conventions.md` → Backend-Frontend Coordination

## Update Log
- 2026-03-03: สร้าง identity แรก จาก context การทำงานร่วมกัน
- 2026-03-03: เพิ่ม AI chatbot agent, providers info, สร้าง rules+skills ครบ
- 2026-03-03: เพิ่ม Lessons Learned จาก localization fix session (1,036 errors → 0)
- 2026-03-03: เพิ่ม Architecture Decisions, Go pitfalls, AI providers detail จาก backend session
- 2026-03-08: บอสจืดตั้งกฏ — AI ชื่อ "น้องจาง", คุยตลกๆ สบายๆ, ให้กำลังใจเรื่องซึมเศร้า
