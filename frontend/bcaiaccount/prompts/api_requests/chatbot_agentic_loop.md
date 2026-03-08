# API Request: Chatbot Agentic Loop (ReAct Pattern)

## สิ่งที่ต้องการ
ระบบ Chatbot ที่มีลักษณะเป็น "Agent" ที่สามารถวนลูปค้นหาข้อมูล เรียกใช้ Tools ต่างๆ ของระบบ และคิดวิเคราะห์อย่างต่อเนื่อง (Agentic Loop / ReAct Pattern) จนกว่าจะได้คำตอบสุดท้าย แทนที่การตอบเพียงครั้งเดียว (Single turn) แบบเดิม เพื่อให้ AI มีความอิสระในการใช้เครื่องมือ (Tools) หาข้อมูลก่อนที่จะตอบกลับผู้ใช้

## Endpoint ที่ต้องการ
- **Method:** POST
- **Path:** `/goapi/api/v1/chatbot/chat-agent` หรือรองรับเป็นแบบ SSE (Server-Sent Events) ที่ `/goapi/api/v1/chatbot/chat-agent/stream`
- **Auth:** Bearer token (JWT)

## พฤติกรรมที่คาดหวังจาก Backend
1. รับคำถามจากผู้ใช้ (Frontend)
2. Backend ส่งข้อความไปหา AI Provider (เช่น OpenAI, Gemini, Claude) พร้อมหน้าต่างระบุ "Tools" (Functions) ทั้งหมดที่มีให้ AI ใช้ เช่น:
   - `search_products(query)`
   - `get_daily_sales(...)`
   - `query_database(...)`
3. ถ้า AI ประเมินว่ายังไม่มีคำตอบที่เพียงพอ AI จะส่งคำสั่งเรียกใช้ Tool กลับมา 
4. Backend ทำการ Execute Tool นั้นและส่งผลลัพธ์ Tool Result กลับไปให้ AI
5. AI คิดต่อ (Thought process) และวนลูปข้อ 3-4 จนกว่าจะได้คำตอบสุดท้าย (Final Answer)
6. Backend ส่งคำตอบสุดท้ายแบบ Web/HTML (พร้อมข้อมูลตัวเลขต่างๆ) กลับไปให้ Frontend Flutter

*หากทำแบบ Streaming (SSE) สามารถส่งสถานะ (Thought/Action) ไปให้ Flutter แสดงคำว่า "กำลังระบุตัวเลือก...", "กำลังหาข้อมูลในฐานข้อมูล..." ระหว่างรอคำตอบได้*

## Request
```json
{
  "message": "วันนี้ขายอะไรได้บ้าง ลองดูข้อมูลเทียบกับเมื่อวาน แล้วสรุปสินค้าที่ขายดีที่สุด 3 อันดับแรก",
  "history": [
    {"role": "user", "content": "สวัสดี"},
    {"role": "assistant", "content": "สวัสดีค่ะ มีอะไรให้ช่วยไหมคะ"}
  ],
  "session_id": "optional-uuid"
}
```

## Expected Response (กรณีไม่ใช้งาน Stream)
ถ้าใช้ HTTP POST ปกติ รอจนตอบเสร็จ

```json
{
  "success": true,
  "data": {
    "reply_html": "<div><p>สรุปข้อมูลการขายวันนี้...</p>...</div>",
    "agent_steps": [
      "เรียกใช้ get_daily_sales(today)",
      "เรียกใช้ get_daily_sales(yesterday)",
      "เรียกใช้ get_top_selling_products(3)"
    ]
  }
}
```

## Expected Response (กรณีใช้งาน SSE - Server-Sent Events)
ถ้าใช้ SSE (แนะนำอย่างยิ่ง เพื่อให้ UX เหมือน Claude Desktop) จะส่ง Event แบ่งเป็นสถานะของ Agent ที่กำลังคิดและเรียกใช้เครื่องมือ

```text
// เมื่อเริ่มต้นคิดหรือเรียก Tool
event: status
data: {"action": "calling_tool", "tool_name": "get_daily_sales", "description": "กำลังดึงข้อมูลยอดขายประจำวัน..."}

// เมื่อกำลังวิเคราะห์ผลลัพธ์จาก Tool
event: status
data: {"action": "analyzing", "description": "กำลังวิเคราะห์ข้อมูล..."}

// กรณีดึง Tool สำเร็จ ได้ข้อมูลมาแล้ว
event: status
data: {"action": "tool_result", "tool_name": "get_daily_sales", "description": "ดึงข้อมูลยอดขายสำเร็จ (150 rows)"}

// เมื่อเริ่มส่งคำตอบสุดท้าย (Stream HTML/Markdown)
event: message
data: {"content": "<p>สรุปยอดขาย</p>"}

// เมื่อส่งคำตอบเสร็จสิ้น
event: done
data: {}
```

## Use Case ฝั่ง Frontend (การแสดงผลแบบ Claude Desktop)
- ผู้ใช้งานพิมพ์ถามคำถามที่มีความซับซ้อน เช่น "สินค้าไหนที่ใกล้จะหมดอายุสต๊อกและมียอดขายน้อยที่สุดให้จัดมา 5 รายการ?"
- Frontend จะเชื่อมต่อผ่าน SSE (`/goapi/api/v1/chatbot/chat-agent/stream`)
- **UI Steps:** ระหว่างรอคำตอบ Frontend จะแสดงรายการแบบ Dropdown หรือ List ย่อยๆ เช่น:
  - 🔄 *กำลังดึงข้อมูลคลังสินค้า (get_inventory_status)...*
  - ✅ *ดึงข้อมูลคลังสินค้าสำเร็จ*
  - 🔄 *กำลังดึงข้อมูลยอดขาย (get_daily_sales)...*
  - ✅ *ดึงข้อมูลยอดขายสำเร็จ*
- เมื่อ AI รวบรวมข้อมูลครบ Frontend จะเริ่ม Render ข้อความคำตอบ (Message) ที่ส่งมาแบบ Stream ให้ผู้ใช้อ่านเรียลไทม์

## Database / Tools ที่เกี่ยวข้อง (ตัวอย่างให้ Agent ใช้ฝั่ง Backend)
Backend ควรเปิด Tools ของ MCP หรือ Internal API ฟังก์ชันที่คล้ายกัน ให้ Agent. Tools ดังกล่าว ได้แก่:
- search_products
- get_daily_sales
- get_inventory_status
- ฯลฯ