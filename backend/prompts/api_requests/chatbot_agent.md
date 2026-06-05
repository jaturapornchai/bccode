# Chatbot Agent — AI Chat ที่ใช้ Tools ดึงข้อมูลจริง

## บริบท

Backend เพิ่ม endpoint ใหม่ `POST /goapi/api/v1/chatbot/chat-agent` — เป็น AI chatbot ที่ใช้ ReAct pattern (เรียก tools ซ้ำๆ จนได้คำตอบ)

**ต่างจาก chat-gemini อย่างไร:**
- `chat-gemini` (เดิม): AI สร้าง SQL → query DB → แปลงเป็นคำตอบ → ตอบได้แค่สินค้า/สต็อก
- `chat-agent` (ใหม่): AI เรียก MCP tools อัตโนมัติ → ตอบได้ทุกอย่าง (ยอดขาย, ลูกค้า, สต็อก, กำไร, dashboard)

**ให้ frontend ใช้ `chat-agent` แทน `chat-gemini`** สำหรับ chatbot หลัก

---

## API Endpoint

### `POST /goapi/api/v1/chatbot/chat-agent`

**Request:**
```json
{
  "holdingcode": "2jgDFkVsFdah2JnSMC89rM2eBMy",
  "question": "ยอดขายวันนี้เท่าไหร่"
}
```

**Response (สำเร็จ):**
```json
{
  "success": true,
  "message": "ตอบสำเร็จ",
  "data": {
    "answer": "ยอดขายวันนี้ (2026-03-02) รวม 15,234.50 บาท จาก 23 ออเดอร์\n\nรายละเอียด:\n- สินค้าขายดี: สว่านไฟฟ้า MAKITA (3 ชิ้น)\n- ช่วงเวลาขายดี: 10:00-12:00\n- เทียบเมื่อวาน: เพิ่มขึ้น 12.5%",
    "html": "",
    "toolsused": [
      {
        "tool": "getdailysales",
        "params": {"date": "2026-03-02", "holdingcode": "..."},
        "durationms": 245
      },
      {
        "tool": "gettopsellingproducts",
        "params": {"fromdate": "2026-03-02", "todate": "2026-03-02", "limit": 5, "holdingcode": "..."},
        "durationms": 189
      }
    ],
    "iterations": 3
  },
  "token_usage": {
    "prompt_tokens": 3500,
    "completion_tokens": 200,
    "total_tokens": 3700,
    "cost_usd": 0,
    "cost_thb": 0,
    "model": "meta-llama/llama-4-maverick-17b-128e-instruct"
  },
  "suggestedquestions": null,
  "timestamp": "2026-03-02T10:30:00Z"
}
```

**Response (error):**
```json
{
  "success": false,
  "message": "Agent ทำงานไม่สำเร็จ",
  "error": "AI ตอบไม่ได้: ทุก provider ใช้ไม่ได้: ...",
  "timestamp": "2026-03-02T10:30:00Z"
}
```

---

## Fields สำคัญ

| Field | Type | คำอธิบาย |
|-------|------|----------|
| `data.answer` | string | คำตอบเป็นข้อความ (ภาษาไทย) — แสดงให้ user เลย |
| `data.toolsused` | array | รายการ tools ที่ AI เรียกใช้ (สำหรับ debug/transparency) |
| `data.toolsused[].tool` | string | ชื่อ tool ที่ใช้ |
| `data.toolsused[].params` | object | parameters ที่ส่งให้ tool |
| `data.toolsused[].error` | string | error ถ้า tool fail (ไม่มีถ้าสำเร็จ) |
| `data.toolsused[].durationms` | number | เวลาที่ tool ใช้ (ms) |
| `data.iterations` | number | จำนวนรอบที่ AI คิด (1 = ตอบเลย, 2+ = เรียก tools ก่อนตอบ) |
| `token_usage` | object | จำนวน tokens ที่ใช้ (รวมทุก iteration) |

---

## สิ่งที่ต้องทำใน Frontend

### 1. เปลี่ยน Chatbot ให้เรียก `chat-agent` แทน `chat-gemini`

```
เดิม:  POST /goapi/api/v1/chatbot/chat-gemini
ใหม่:  POST /goapi/api/v1/chatbot/chat-agent
```

**Request เปลี่ยน:**
- เดิม: `{"holdingcode": "...", "question": "...", "functionname": "product"}`
- ใหม่: `{"holdingcode": "...", "question": "..."}` (ไม่ต้องส่ง `functionname` แล้ว)

**Response เปลี่ยน:**
- เดิม: `data.answer` + `data.html`
- ใหม่: `data.answer` + `data.toolsused` + `data.iterations` (ไม่มี `html` ส่วนใหญ่)

### 2. แสดง `data.answer` เป็น Text

`data.answer` เป็น plain text ภาษาไทย — แสดงใน chat bubble ได้เลย

ตัวอย่างคำตอบ:
- "ยอดขายวันนี้ 15,234.50 บาท จาก 23 ออเดอร์"
- "สินค้าขายดี 5 อันดับแรกเดือนนี้: 1. สว่าน MAKITA (120 ชิ้น) ..."
- "ขออภัย ระบบไม่สามารถเชื่อมต่อฐานข้อมูลได้ในขณะนี้"

### 3. แสดง Tools Used (optional — สำหรับ transparency)

แสดงส่วน "AI ใช้ข้อมูลจาก:" ใต้คำตอบ:
```
AI ใช้ข้อมูลจาก:
  📊 ยอดขายรายวัน (245ms)
  🏆 สินค้าขายดี (189ms)
```

**Mapping ชื่อ tool → label ภาษาไทย:**

| tool name | label |
|-----------|-------|
| `searchproducts` | ค้นหาสินค้า |
| `getdailysales` | ยอดขายรายวัน |
| `getsalesbydaterange` | ยอดขายตามช่วงเวลา |
| `gettopsellingproducts` | สินค้าขายดี |
| `getsalesbyseller` | ยอดขายตามพนักงาน |
| `getmonthlysummary` | สรุปรายเดือน |
| `getdashboardkpis` | KPI Dashboard |
| `getbusinesshealth` | สุขภาพธุรกิจ |
| `getprofitanalysis` | วิเคราะห์กำไร |
| `getaccountsreceivable` | ลูกหนี้การค้า |
| `getaccountspayable` | เจ้าหนี้การค้า |
| `getcashflow` | กระแสเงินสด |
| `getinventoryvalue` | มูลค่าสินค้าคงเหลือ |
| `getlowstockalerts` | สินค้าใกล้หมด |
| `getdeadstock` | สินค้าค้างสต็อก |
| `getinventoryturnover` | อัตราหมุนเวียนสินค้า |
| `gettopcustomers` | ลูกค้ารายใหญ่ |
| `getcustomergrowth` | การเติบโตลูกค้า |
| `getcustomersegments` | กลุ่มลูกค้า |
| `getyoycomparison` | เปรียบเทียบปีต่อปี |
| `getmomcomparison` | เปรียบเทียบเดือนต่อเดือน |
| `listunits` | หน่วยนับสินค้า |

### 4. Loading State

Agent อาจใช้เวลา 5-30 วินาที (เรียก tools หลายรอบ) — ต้องมี loading ที่ดี:

```
🤔 กำลังคิด...        → ส่ง request
🔍 กำลังดึงข้อมูล...   → รอ response (อาจนานถ้าเรียกหลาย tools)
✅ ได้คำตอบแล้ว       → แสดงคำตอบ
```

**Timeout:** Backend มี timeout 120 วินาที — frontend ควรตั้ง timeout ไว้ 130 วินาที

### 5. Suggested Questions (ตัวอย่างคำถาม)

แสดงตัวอย่างคำถามให้ user กดได้:

```
💡 ลองถาม:
• ยอดขายวันนี้เท่าไหร่
• สินค้าขายดี 10 อันดับเดือนนี้
• สินค้าไหนใกล้หมดสต็อก
• เปรียบเทียบยอดขายเดือนนี้กับเดือนที่แล้ว
• ลูกค้ารายใหญ่ 5 อันดับแรก
• กำไรเดือนนี้เท่าไหร่
• สุขภาพธุรกิจเป็นอย่างไร
```

### 6. Error Handling

| เงื่อนไข | แสดง |
|----------|------|
| `success: false` | แสดง `error` ใน chat bubble สีแดง |
| Network timeout (>130s) | "ขออภัย ระบบใช้เวลานานเกินไป กรุณาลองใหม่" |
| HTTP 500 | "เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง" |

---

## ตัวอย่างคำถามที่ Agent ตอบได้ (22 tools)

| หมวด | ตัวอย่างคำถาม |
|------|-------------|
| **ยอดขาย** | "ยอดขายวันนี้", "ยอดขายสัปดาห์นี้", "ยอดขายเดือน มี.ค." |
| **สินค้า** | "ค้นหาสินค้า MAKITA", "สินค้าขายดี 10 อันดับ" |
| **เปรียบเทียบ** | "เทียบยอดเดือนนี้กับเดือนก่อน", "เทียบปีนี้กับปีที่แล้ว" |
| **สต็อก** | "สินค้าใกล้หมด", "สินค้าค้างสต็อกเกิน 90 วัน", "มูลค่าสินค้าคงเหลือ" |
| **ลูกค้า** | "ลูกค้ารายใหญ่", "ลูกค้าใหม่เดือนนี้", "วิเคราะห์กลุ่มลูกค้า" |
| **การเงิน** | "กำไรเดือนนี้", "กระแสเงินสด", "ลูกหนี้ค้างชำระ" |
| **Dashboard** | "สุขภาพธุรกิจ", "KPI เดือนนี้" |

---

## สรุปสิ่งที่ต้องทำ

1. **เปลี่ยน endpoint** จาก `chat-gemini` → `chat-agent` (ลบ `functionname`)
2. **แสดง `data.answer`** เป็น text ใน chat bubble
3. **แสดง toolsused** (optional) เป็น transparency section ใต้คำตอบ
4. **Loading state** ที่ดี (อาจ 5-30 วินาที)
5. **ตัวอย่างคำถาม** ให้ user กดได้
6. **Error handling** สำหรับ timeout และ error
