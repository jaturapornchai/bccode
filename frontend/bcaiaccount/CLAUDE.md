# BC AI Cloud — Flutter App Rules

Multi-platform Flutter app (Web, Windows, Android, iOS). Package: `smlaicloud`. State: BLoC. HTTP: Dio. Auth: Firebase + JWT.

## Critical Rules

### ภาษา + Mock Data
- **AI ตอบ user เป็นภาษาไทยเสมอ**
- **ห้าม mock / hardcode data** ใน production code (ลูกค้า, สินค้า, ราคา, ที่อยู่ ฯลฯ) — ใช้ API + Repository + BLoC เสมอ ถ้า API ยังไม่พร้อมให้แสดง Empty/Loading state

### Security: ห้าม Frontend เรียก AI API ตรง
- ห้าม hardcode AI API key (OpenRouter, Groq, DeepSeek, Gemini, OpenAI ฯลฯ) ใน Flutter
- ห้ามเรียก endpoint AI provider ตรง — ต้องผ่าน backend (`/goapi/api/v1/chatbot/*`)
- เหตุผล: key ใน frontend ถูกขโมยได้ง่าย (decompile/sniff)

### ห้าม Hardcode สี
- ทุกสี UI ต้องมาจาก `global.theme.*` — ห้าม `Colors.*` หรือ `Color(0xFF...)` ตรง
- รายละเอียด + mapping ดู `D:\bcdev\clone-skills\rules\coding-style.md`
- StatefulWidget ทุกตัวต้องใช้ `global.ThemeRefreshMixin`
- ยกเว้น: brand colors, login screens, semantic status

### Localization
- ทุก UI text ใช้ `global.language("key")` — ห้าม hardcode ภาษาไทย (ยกเว้น login flow ที่โหลด language ยังไม่ได้)
- Keys อยู่ใน `D:\bcdev\backend\assets\language\languages.tsv` (TSV: `key\tth\ten\tcn\tja\tkm\tko\tlo\tmy\tvi`)
- Yes: log/comment/exception message, month/day arrays, NumberToWord constants

### Backend Access (Cross-Project)
- AI **อ่าน + แก้ backend (`d:\bcdev\backend\`) ได้** — แต่ระวัง field/type ที่ frontend ใช้ร่วม
- ลำดับ: ใช้ MCP tools ก่อน (ดูด้านล่าง) → ถ้าไม่พอค่อย read source
- เพิ่ม field = ปลอดภัย, ลบ/เปลี่ยนชื่อ = ต้องแก้ทั้ง 2 ฝั่ง
- DB schema ต้องผ่าน backend เท่านั้น — ห้าม AI สร้าง/แก้ migration ตรง

### MCP-First (Dev Tools เท่านั้น)
- **Base:** `http://localhost:9090/goapi/mcp`
- General: `/mcp/sse` `/mcp/tools` (business data)
- Dev: `/mcp/dev/sse` `/mcp/dev/tools` (รวม `query_mongodb`, `execute_query`, `execute_pg_command` ฯลฯ)
- Auth: `X-API-Key: bc_live_{32}` (ขอจาก `POST /api/mcp/keys`)
- เรียก `GET /mcp/dev/tools` เพื่อดูรายการล่าสุดเสมอ — **อย่าจำ**
- Flutter app **ห้ามเรียก MCP ตรง** — ใช้ REST ผ่าน Dio เท่านั้น

### API Spec Prompt
- ถ้า API ที่ต้องการยังไม่มี → เขียน spec ที่ `prompts/api_requests/{feature}.md`
- หรือแก้ backend ตรงก็ได้ (ตามสะดวก)

## Backend URLs
- MainAPI = port **9090** (รวม GoAPI ใต้ prefix `/goapi`)
- Local: `http://localhost:9090/goapi/api/health`
- Config: `web/config.json` → `goapi_url`

## Flavors + Build
- `bcaidev` / `bcaiuat` / `bcaiprod`
- `flutter run -t lib/main_bcaidev.dart`
- `flutter build web -t lib/main_bcai{dev,uat,prod}.dart --release`
- `flutter build windows -t lib/main_bcaidev.dart --release`

## Code Conventions
- Logger: `AppLogger.{info,error,debug,warning}()` — ภาษาไทยใน log/comment ได้
- BLoC: events/states pattern
- HTTP: `api/client.dart` (Dio)
- Token: `global.appConfig.getString("token")`

## ทุกครั้งที่เสร็จงาน — สรุป Backend Impact
- **ไม่ต้องแก้** ถ้า frontend แก้ครบ + API/fields พร้อม
- **ต้องแก้** ถ้าต้องเพิ่ม endpoint/field/logic → เขียน prompt ใน `prompts/api_requests/` แล้วแจ้ง user
- **ห้ามจบงานโดยไม่บอก**

## Skill Library
อ่าน `D:\bcdev\clone-skills\` ทุก session: `identity.md`, `rules/*.md`, `references/{core,flutter,go}/`, `skills/*/SKILL.md`
