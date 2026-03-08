# Workflow — Jead's Preferred Work Patterns

## Development Pipeline (ต้องทำครบทุกขั้น)
```
Plan → Implement → Build → Deploy → Test → Push
```

1. **Plan**: วางแผนก่อน (ถ้างานซับซ้อน) — ใช้ plan mode
2. **Implement**: ลงมือเขียน code ทันที (ไม่ถามมากเกินไป)
3. **Build**: `go build ./cmd/goapi/` หรือ flutter build — ต้องผ่าน
4. **Deploy**: `docker compose build mainapi && docker compose up -d`
5. **Test**: ทดสอบด้วย curl หรือ browser ทันที
6. **Push**: `git push` เมื่อ Jead สั่ง (ห้าม push เอง)

## Deploy Flow (Docker Desktop)
**Auto Deploy Rule (สำคัญมาก)**: ทุกครั้งที่มีการแก้ไข backend code → ต้อง deploy to Docker Desktop ใหม่เสมอ เพื่อให้ API ใหม่ทำงานได้ทันที
```bash
# 1. Build check
go build ./cmd/goapi/

# 2. Docker build
docker compose build mainapi

# 3. Deploy
docker compose up -d

# 4. Verify
curl http://localhost:8888/goapi/api/health
```

## Git Workflow
- **Branch**: ใช้ `dev` branch เป็นหลัก
- **Auto Push Rule (สำคัญมาก)**: ทุกครั้งที่มีการแก้ไข code → ต้อง commit + push to GitHub เสมอ เพื่อป้องกัน code หาย และสามารถเรียกกลับมาได้กรณี AI เข้าใจผิด หรือ user เข้าใจผิดทำให้ code พัง
- **Commit message**: ภาษาอังกฤษ กระชับ ตรงประเด็น
- **Git config**: ห้ามแก้ git config (user.email, user.name)

## Frontend-Backend Communication (สำคัญมาก)

**2 project ทำงานประสานกัน — ต้องระวังเรื่องโครงสร้างข้อมูล**

| Project | Path | AI Access |
|---------|------|-----------|
| Frontend | `D:\bcdev\bcaiaccount` | อ่าน+แก้ได้ |
| Backend | `D:\bcdev\backend` | อ่าน+แก้ได้ |

### กฏการประสานงาน:
1. **ตรวจสอบ data structure ก่อนเสมอ** — ใช้ MCP `get_model_schema`, `get_database_schema` เพื่อดู fields/types จริง
2. **ห้าม assume field names** — ต้อง verify กับ backend ผ่าน MCP หรืออ่าน source code
3. **JSON field names ต้องตรงกัน** — backend ใช้ snake_case (`trans_flag`) หรือ camelCase (`transFlag`) ต้องเช็คให้ตรง
4. **Null safety** — backend อาจ return null → frontend ต้อง handle ด้วย `?.` / `?? default`
5. **Type matching** — backend int64 → Dart int, backend float64 → Dart double, backend time.Time → Dart DateTime

### เมื่อทำ backend เสร็จ + frontend ต้องใช้:
1. สร้าง prompt spec ใน `prompts/api_requests/{feature}.md`
2. บอก Jead ว่า "ขอ prompt ไปแก้ frontend"
3. Prompt ต้องมี: endpoint, request, response, fields, error handling

### เมื่อทำ frontend + ต้องการ API ใหม่:
1. ใช้ MCP ค้นหาก่อนว่ามี API อยู่แล้วหรือไม่
2. ถ้าไม่มี → เขียน API Specification Prompt
3. สร้าง frontend code ไว้ก่อน (with placeholder/empty state)
4. แจ้ง Jead ว่ายังรอ backend

### ข้อควรระวังเรื่องโครงสร้างข้อมูล:
- **MongoDB → PostgreSQL sync** — data อาจมีชื่อ field ต่างกัน (mongo: `guidfixed`, pg: `id`)
- **ClickHouse** — ใช้สำหรับ analytics เท่านั้น อาจมี data delay
- **Enum values** — ต้องใช้ค่าเดียวกับ backend (ใช้ `/enum-list` ดูค่าจริง)
- **Date format** — backend ส่ง ISO 8601 (`2026-03-03T10:30:00Z`) → Dart parse ด้วย `DateTime.parse()`

## Task Completion Checklist
ทุกครั้งที่ทำงานเสร็จ ต้อง:
- [ ] Code compile/build ผ่าน
- [ ] Deploy สำเร็จ (ถ้า Jead ต้องการ)
- [ ] สรุปสิ่งที่ทำ
- [ ] บอกว่า frontend ต้องแก้อะไร (ถ้ามี)
- [ ] บอกว่ามี skill/rule ที่ควร update (ถ้ามี)

## สิ่งที่ห้ามทำ (เด็ดขาด)
- ห้าม force push / reset hard
- ห้าม deploy โดยไม่ build ก่อน
- ห้ามลบไฟล์โดยไม่ถาม (ยกเว้น Jead สั่ง)
- ห้ามแก้ docker-compose.yml ports โดยไม่บอก
- **ห้าม deploy ขึ้น VPS เด็ดขาด** — ยกเว้น Jead สั่งเป็นคำสั่งพิเศษ (deploy Docker Desktop เท่านั้น)

## Troubleshooting Patterns
เมื่อเจอปัญหา:
1. **Rate limit**: ใช้ fallback provider (อย่ารอ — switch เลย)
2. **Build error**: แก้ที่ต้นเหตุ (ห้ามข้าม lint/vet)
3. **Deploy fail**: ดู `docker compose logs mainapi`
4. **API error**: ทดสอบด้วย curl + ดู log
