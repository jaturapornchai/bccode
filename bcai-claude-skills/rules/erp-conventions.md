# ERP Conventions — BC AI Cloud

## Platform
BC AI Cloud = ระบบ ERP/Accounting บน Cloud สำหรับร้านค้า

## Database Architecture
| Database | บทบาท | ใช้ทำอะไร |
|----------|--------|-----------|
| **MongoDB** | Main data store | Source of truth — documents, transactions, masters |
| **PostgreSQL** | Relational processing | Joins, aggregations, reports |
| **ClickHouse** | OLAP/Analytics | Time-series, BI dashboards, analytics |

## Key Concepts

### shop_id
- ทุก request ต้องมี `shop_id` (identifier ของร้านค้า)
- Format: string (เช่น `"2jgDFkVsFdah2JnSMC89rM2eBMy"`)
- ทุก query ต้อง filter by shop_id

### Document Types
- Sale Invoice, Sale Return, Purchase, Purchase Return
- Stock Transfer, Stock Adjustment, Stock Balance
- Quotation, Credit Note, Debit Note

### Master Data
- Products (สินค้า) — barcode, name, price, unit
- Units (หน่วยนับ) — unit code, names (multi-language)
- Customers (ลูกค้า) — customer code, name, contact
- Suppliers (ผู้จำหน่าย) — supplier code, name, contact

## Port Standards
| Service | Port | Note |
|---------|------|------|
| MainAPI (gateway) | **8888** | จุดเข้าเดียว |
| MongoDB | 27017 | Native (ไม่อยู่ Docker) |
| PostgreSQL | 5432 | Native (ไม่อยู่ Docker) |
| Redis | 6379 | Docker internal |
| ClickHouse HTTP | 8123 | Docker internal |
| Kafka | 9092 | Docker internal |

## Domain
- VPS dev: `api.bcaicloud.com` → Hetzner
- ห้ามใช้ IP ตรง ให้ใช้ domain เสมอ

## Multi-language
- ระบบรองรับ multi-language (TH, EN, etc.)
- ใช้ `assets/language/languages.json` สำหรับ i18n
- Names ที่มีหลายภาษาจะเก็บเป็น `map[string]string` (`{"th": "...", "en": "..."}`)

## Backend-Frontend Coordination (สำคัญมาก)

**2 project ทำงานประสานกัน:**
| Project | Path | Tech |
|---------|------|------|
| **Backend** | `D:\bcdev\backend` | Go (Echo v4) |
| **Frontend** | `D:\bcdev\bcaiaccount` | Flutter (BLoC + Dio) |

### กฏการประสาน
1. **โครงสร้างข้อมูลต้องตรงกัน** — Backend response fields ต้องตรงกับ Frontend model
2. **เปลี่ยน Backend → ต้องบอก Frontend** — ถ้าแก้ response format ต้องสร้าง prompt ใน `prompts/api_requests/`
3. **เปลี่ยน Frontend → ต้องตรวจ Backend** — ใช้ MCP tools (`get_api_spec`, `get_model_schema`) ตรวจ API ก่อนแก้
4. **JSON field names ต้องตรง** — Backend ใช้ `json:"field_name"` (snake_case), Frontend model ต้อง map ตรงกัน
5. **Null safety** — Backend field ที่เป็น `omitempty` อาจไม่ส่ง → Frontend ต้อง handle nullable

### ข้อควรระวัง
- แก้ Backend response → **ต้องตรวจว่า Frontend ยังใช้ได้**
- เพิ่ม field ใหม่ → ปลอดภัย (Frontend เก่ายัง work)
- ลบ/เปลี่ยนชื่อ field → **อันตราย** (Frontend จะพัง) → ต้องแจ้ง + แก้ Frontend ด้วย
- เปลี่ยน type ของ field (เช่น `int` → `string`) → **อันตราย** → ต้องแก้ทั้ง 2 ฝั่ง
