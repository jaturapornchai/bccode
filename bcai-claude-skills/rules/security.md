# Security Rules — กฏความปลอดภัย

## Config
- **bootstrap.json ONLY** — ห้ามใช้ .env, .env.*, godotenv.Load()
- `bootstrap.json` อยู่ใน `.gitignore` (ห้าม commit)
- Credentials ทั้งหมดอยู่ใน bootstrap.json

## Docker
- เฉพาะ **port 8888** ที่เปิดออกนอก (`0.0.0.0`)
- Infrastructure ทุกตัว bind `127.0.0.1` (internal only)
- MongoDB/PostgreSQL = native บน host (ไม่อยู่ใน Docker)
- Docker containers เข้าถึง host DB ผ่าน `host.docker.internal`

## Git
- ห้าม commit: `bootstrap.json`, `.env`, credentials, API keys
- ห้าม force push to main/master
- ใช้ `dev` branch สำหรับ development

## API Keys
- MCP API keys ใช้ prefix `bc_live_` หรือ `bc_test_`
- Keys เก็บใน MongoDB (encrypted)
- Permission levels: readonly, developer (*), custom

## Code
- ห้าม panic — ใช้ return error
- ห้ามใช้ raw SQL โดยไม่ parameterize (ป้องกัน SQL injection)
- ห้ามเก็บ password เป็น plain text
- ห้ามใช้ --no-verify หรือ --force โดยไม่ได้รับอนุญาต
