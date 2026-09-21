# GL API / MCP tokens

## Objective

เชื่อมโปรแกรมภายนอกหรือ MCP client เข้าระบบห้องบัญชีด้วย credential แยกจาก session ผู้ใช้ โดยผู้ดูแล OWNER/ADMIN ที่ยัง active เท่านั้นจัดการ token ได้ บังคับทั้งหน้าจอและ backend ผู้ใช้ทั่วไปเรียก API ตรงก็ไม่ได้

## Workflow

1. เลือก Holding → ตั้งค่าระบบและการเข้าถึง → การเชื่อมต่อภายนอก → จัดการ API / MCP token (`/mcp-tokens`) เมนูอยู่ต่อจากหมวดสิทธิ์ ไม่ต้องเลือกบริษัทก่อน ปุ่มเข้าแสดงเสมอโดยไม่เรียก API เพื่อซ่อนเมนู; หน้าจอและ backend ตรวจ OWNER/ADMIN ก่อนแสดงข้อมูลหรืออนุญาตให้จัดการ
2. กด **เพิ่ม MCP token** หรือ **เพิ่ม API token** เพื่อเปิดฟอร์มที่เลือกประเภทไว้แล้ว ระบุชื่อ แล้วเลือก `readonly` (ค่าเริ่มต้น) หรือ `readwrite` เลือกบริษัทที่อนุญาตได้ 1–100 บริษัท และอายุ 1–365 วัน (ค่าเริ่มต้น 90 วัน) หน้าจอแยกรายการ/รายละเอียดที่ปรับความกว้างได้ และพับคำแนะนำเชื่อมต่อไว้ใต้ฟอร์ม
3. คัดลอก token จากผลสร้างครั้งแรก ระบบไม่แสดงค่าลับซ้ำ ไม่เก็บใน browser storage และฐานข้อมูลเก็บเฉพาะ SHA-256 ของ token ที่สุ่มด้วย `crypto/rand`
4. ตั้ง MCP client เป็น Streamable HTTP ที่ `https://account.bcaicloud.com/mcp/gl` พร้อม header `Authorization: Bearer <token>`
5. เรียก initialize, tools/list และ tools/call; แต่ละ request ตรวจวันหมดอายุ การเพิกถอน สมาชิก/บทบาทผู้สร้าง บริษัทและสาขาที่ยัง active จาก PostgreSQL ใหม่
6. กด **ลบ token** ที่หน้ารายละเอียดแล้วยืนยัน ระบบเรียก revoke เดิมเพื่อหยุดการใช้งานและนำออกจากรายการหน้าจอ โดยคง metadata/audit ในฐานข้อมูล คำขอใหม่จะถูกปฏิเสธ งานที่เริ่มตรวจสิทธิ์ผ่านไปแล้วอาจดำเนินต่อจนจบ

## API / MCP separation

- API token ขึ้นต้น `bcaiapi_` ใช้เฉพาะ REST `/api/integration/gl/...` เช่น GET `/accounts`, GET `/journals/{id}`, GET `/reports/ledger` และ POST `/command`
- MCP token ขึ้นต้น `bcaimcp_` ใช้เฉพาะ `/mcp/gl`
- Backend ตรวจทั้ง prefix และ `kind` ที่เก็บใน PostgreSQL; สลับ prefix หรือใช้ผิด endpoint ถูกปฏิเสธ ไม่มี token ที่ใช้ได้ทั้งสองช่องทาง
- ทั้งสองชนิดแยก readonly/readwrite และการเพิกถอนเป็นราย token; session login ไม่ใช่ integration token

## Permissions and tools

| Mode | Tools | Behavior |
|---|---|---|
| readonly | gl_list, gl_get, gl_report | อ่าน master, รายวัน, ประวัติผลตรวจ และรายงาน หลายบริษัทในคำขอเดียวได้ |
| readwrite | เครื่องมืออ่านทั้งหมด + gl_command | เลือกบริษัทได้ครั้งละหนึ่งบริษัทเท่านั้น รวมคำขออ่าน; เขียนผ่าน service GL เดิม |

REST API ตรวจ readonly ก่อนเรียก POST /command และคืน HTTP 403 ส่วน `tools/list` ไม่แสดง gl_command ให้ readonly และ `tools/call` ตรวจซ้ำก่อนเข้า ledger แม้ client ระบุชื่อเอง Token เปลี่ยน Holding ไม่ได้ เลือกได้เฉพาะบริษัทใน allow-list ที่ยัง active; readonly อ่านพร้อมกันได้สูงสุด 50 บริษัทต่อคำขอ ส่วน readwrite ระบุหลายบริษัทจะถูกปฏิเสธก่อนเริ่มงาน แม้เป็นคำขออ่าน สิทธิ์ผูกกับผู้สร้าง และหยุดใช้ได้เมื่อผู้สร้างไม่เป็น OWNER/ADMIN แล้ว

จำนวนเงินใช้ decimal strings คำสั่งเขียนยังต้องผ่าน double-entry, lock, version และ requestid/idempotency เดิม การนำเข้าไม่ทำให้ผ่านรายการอัตโนมัติ GL audit actor มี suffix `:mcp:<token-id>` หรือ `:api:<token-id>` เพื่อย้อนกลับถึง credential โดยไม่เก็บค่าลับ

## API and dependencies

- Session admin API: GET/POST `/mcp-tokens`, POST `/mcp-tokens/{id}/revoke` ผ่าน BFF `/api/mcp-tokens`
- Create: `{ "name": "Accounting assistant", "kind": "mcp", "mode": "readonly", "companyCodes": ["C01", "C02"], "expiresAt": "2026-12-20T00:00:00Z" }` วันต้องอยู่ในอนาคตไม่เกินหนึ่งปีจากเวลาสร้าง
- MCP endpoint: `/mcp/gl` ใช้เฉพาะ MCP bearer token ไม่รับ browser session หรือ token ใน query string
- Token metadata: id, name, kind, mode, holdingCode, companyCodes, companyCode (legacy), branchCode (legacy), createdBy, createdAt, expiresAt, revokedAt, lastUsedAt
- ตาราง `mcp_access_tokens`, `mcp_token_companies` (many-to-many) และ `mcp_token_audit` อยู่ในฐานกลาง PostgreSQL `bcai_projection` ร่วมกับ users/membership/บริษัท สร้างโดย backend หลังตรวจ admin ผ่าน และกรอง Holding ทุกคำขอ ไม่มี SQL ที่ต้องรันด้วยมือ; ข้อมูลรายวันยังอยู่ฐานของ Holding
- Audit บันทึก create/revoke/use โดยไม่บันทึก secret; รายการ token ไม่ส่ง hash กลับ
- ใช้ Go stdlib, Echo, PostgreSQL และ Next.js เดิม ไม่มี MongoDB, ClickHouse, Kafka หรือ Redis เพิ่ม ไม่มี MCP SDK dependency เพิ่ม

## Config

- `BCAI_LOCAL_BACKEND_URL`: ที่อยู่ backend สำหรับ Next.js proxy เดิม เช่น `http://mainapi:8888` ใน Docker (port 8888)
- `BCAI_MCP_ALLOWED_ORIGINS`: backend allowlist ของ browser Origin คั่นด้วย comma ค่าเริ่มต้น `https://account.bcaicloud.com`; client ที่ไม่มี Origin ใช้งานได้ ส่วน Origin อื่นคืน 403
- PostgreSQL ใช้ connection configuration เดิมของระบบ ไม่มี secret ใหม่ใน source
- จำกัด request MCP 2 MiB และ timeout backend 35 วินาที / proxy 40 วินาที; metadata list ล่าสุดไม่เกิน 500 รายการ

## Usage example

สำหรับ API ใช้ token อีกตัว เช่น `curl https://account.bcaicloud.com/api/integration/gl/accounts -H "Authorization: Bearer $API_TOKEN" -H "X-BC-Company-Code: C01"`

ตัวอย่างสำหรับ client ที่ตั้ง header เองได้ ใช้ shell environment ที่ตั้ง `MCP_TOKEN` ไว้แล้ว ห้ามวาง token ใน URL หรือ commit ลงไฟล์:

```bash
curl https://account.bcaicloud.com/mcp/gl \
  -H "Authorization: Bearer $MCP_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"accounting-client","version":"1.0"}}}'

curl https://account.bcaicloud.com/mcp/gl \
  -H "Authorization: Bearer $MCP_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'MCP-Protocol-Version: 2025-11-25' \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"gl_list","arguments":{"companyCode":"C01","resource":"accounts","query":{"limit":"20"}}}}'
```

## Multi-company reads

- API ใช้ `X-BC-Company-Code: C01` สำหรับบริษัทเดียว หรือ `X-BC-Company-Codes: C01,C02` สำหรับ readonly หลายบริษัท ห้ามส่งทั้งสอง header พร้อมกัน
- MCP ใช้ `companyCode: "C01"` หรือ `companyCodes: ["C01", "C02"]` ใน arguments ห้ามใช้ทั้งสองพร้อมกัน
- ทั้งสอง mode เลือกบริษัทใน allow-list ของ token ได้หลายแห่ง แต่ **readwrite ทำคำขอได้บริษัทเดียวเสมอ** และไม่รับ batch หลายบริษัท
- ตรวจสิทธิ์ทุกบริษัทก่อนเริ่มอ่าน หากมีบริษัทนอกสิทธิ์ ปฏิเสธทั้งคำขอ ผลหลายบริษัทอยู่ใน `data.companies[]` แต่ละรายการมี `companyCode` และ `response`; ถ้าบางบริษัทอ่านผิดพลาด `success=false` และ MCP `isError=true` พร้อมผลแยกบริษัท
- ไม่รวมยอดต่างบริษัท/สกุลเงินให้อัตโนมัติ แต่ละบริษัทใช้ snapshot ของตัวเอง ไม่ใช่ snapshot เดียวกันทั้ง Holding
- Token ใหม่ใช้ได้ทุกสาขาในบริษัทที่เลือก; token เดิมคงขอบเขตบริษัท/สาขาเดิม การสร้าง join table/backfill ไม่ขยายสิทธิ์เดิม
- หากมีบริษัทที่อนุญาตเพียงแห่งเดียว สามารถเว้น company selector ได้เพื่อรองรับ client/token เดิม; หากมีหลายแห่งต้องระบุ

## Limitations

- เป็น token ที่ admin ออกให้และ client ตั้ง Bearer header เอง ไม่ได้ implement OAuth discovery/authorization server จึงยังไม่รองรับ client ที่บังคับ OAuth-only
- เปิด GL REST integration และ GL MCP เท่านั้น ไม่เปิด SQL โดยตรง และไม่เชื่อม fixed-asset MCP รุ่นเดิม
- Stateless JSON responses รองรับ protocol 2025-03-26, 2025-06-18 และ 2025-11-25; GET/DELETE คืน 405 ไม่มี SSE stream, session persistence หรือ background tasks
- ไม่มีแก้ kind/mode/scope ของ token เดิม ให้เพิกถอนแล้วสร้างใหม่; ไม่สามารถเรียกคืน secret ที่หายได้
- token มีสิทธิ์ตาม mode และผู้สร้าง จึงควรเลือก readonly สำหรับงานตรวจสอบ และแยก readwrite ตามการเชื่อมต่อที่ต้องเขียนจริง
- บัญชีผู้พัฒนาใช้งานทุกจอด้วย OWNER membership ปกติ ผูกกับ UID ที่ระบุไว้ ไม่ใช้ email จาก client หรือโปรไฟล์เป็นเงื่อนไขข้ามสิทธิ์ การเพิ่ม Holding ใหม่ต้องกำหนด membership ตามปกติ
- Google login บันทึก user และ Google subject ใน transaction เดียว ใช้ UID จริงจาก PostgreSQL สำหรับ session และ login ครั้งถัดไป; การซ่อมบัญชี Google ผู้พัฒนาเดิมจำกัดเฉพาะ UID ที่อนุมัติและข้อมูล Google ที่ backend ตรวจสอบแล้ว ไม่เชื่อมบัญชีทั่วไปจากอีเมลตรงกัน Session เก่าที่ใช้ UID ผิดต้องออกจากระบบและเข้า Google ใหม่หนึ่งครั้ง หน้าปฏิเสธสิทธิ์มีปุ่มโหลดใหม่สำหรับตรวจสิทธิ์ที่เปลี่ยนแล้ว

## Verification

- Unit tests ครอบคลุม readonly write denial ก่อนเข้า executor, fixed scope, malformed RPC, version, origin, bearer auth, ไม่มี token ใน URL และรักษา decimal strings/request IDs ที่เกิน JavaScript safe integer
- PostgreSQL integration ใช้ฐานทดสอบแยก: create/list/revoke handlers, non-admin/inactive admin 403, ข้ามบริษัท/สาขา, expiry boundary, revoked, demoted issuer, inactive company, wrong hash, secret ไม่อยู่ใน list และ revoke idempotency
- Frontend proxy tests ตรวจ upstream คงที่ ไม่ส่ง cookies ไม่ตาม redirect, body limit และไม่เปิดเผยรายละเอียด error
- Frontend build/typecheck ผ่าน; Vitest ทั้งชุด 699 tests ผ่าน; Playwright 5 cases บนหน้าจอจริงพร้อม HTTP fixtures ตรวจเมนู Holding โดยไม่เลือกบริษัท, สร้าง MCP/API พร้อมเลือกบริษัท, เพิกถอน, non-admin denied และล้าง secret เมื่อเปลี่ยน workspace ตรวจภาพ light/dark แล้ว
- Backend `go test` ใน mcptoken/mcpgateway/generalledger และ `go build ./...` ผ่าน รวม integration ที่ root รันซ้ำกับ PostgreSQL 18 แยก; ยังไม่ได้ใช้ token จริงของลูกค้าทดสอบผ่าน MCP client ภายนอก
- Production `r20260920-holding-tokens-1`: สำรองฐาน/config ก่อนสลับ mainapi/frontend healthy ไม่มี restart/error markers; หน้า `/mcp-tokens` คืน HTML 200, management ไม่มี session คืน 401, synthetic header token ผิดชนิดคืน 401 ทั้งสองทาง และ Origin ที่ไม่อนุญาตคืน 403 (เป็น smoke test ไม่ใช่การออก token จริง)

รูปแบบ transport อ้างอิง [MCP Streamable HTTP 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports), [lifecycle](https://modelcontextprotocol.io/specification/2025-11-25/basic/lifecycle) และ [tools](https://modelcontextprotocol.io/specification/2025-11-25/server/tools); ขอบเขต token แบบกำหนดเองตามข้อจำกัดข้างต้น
