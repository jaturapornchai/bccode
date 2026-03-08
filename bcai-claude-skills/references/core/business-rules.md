# Business Rules — Checklist และ Common Mistakes

## ขั้นตอนบังคับก่อน Implement Feature ใหม่

### Step 1: เข้าใจ Requirements
ก่อนเขียนโค้ดต้องตอบคำถามเหล่านี้ให้ได้:
- Feature นี้เกี่ยวข้องกับ document type ไหน? (PO/SO/PI/SI/SA/TR)
- มี state transitions ที่ต้องจัดการไหม?
- ต้องมี approval ไหม? ใครเป็น approver?
- กระทบตารางไหนบ้าง? stock เปลี่ยนไหม?
- ต้องส่ง notification ไหม? (LINE OA)
- มี permission ที่ต้องตรวจสอบไหม?

### Step 2: ตรวจสอบ API ที่มีอยู่
- ใช้ MCP tool ดูว่ามี endpoint แล้วไหม:
  - General: `GET /goapi/mcp/tools` (business data tools)
  - Dev: `GET /goapi/mcp/dev/tools` (ครบทุก tools รวม database, API catalog)
- ถ้าไม่มี → เขียน `prompts/api_requests/{feature}.md` ก่อน ห้ามสร้าง API เอง

### Step 3: หา Feature ที่ใกล้เคียงใน Codebase
- Flutter: หา bloc ที่คล้ายกันใน `bcaiaccount/lib/bloc/` อ่านก่อน pattern match
- Go: หา handler ที่คล้ายกันใน `backend/handlers/` อ่านก่อน pattern match

---

## Validation Rules

### Frontend (ก่อนส่ง API)
- วันที่เอกสารต้องไม่เป็นอนาคต (เกิน 1 วัน)
- ต้องมีรายการสินค้าอย่างน้อย 1 รายการ
- ทุกรายการ: จำนวน > 0, ราคา >= 0
- ยอดรวมต้องตรงกับผลรวมรายการ (คำนวณฝั่ง client แล้วยืนยัน)
- ตรวจสอบ state ก่อนแสดงปุ่ม/action

```dart
// State-based button visibility
bool canEdit = doc.docstatus == 'draft' || doc.docstatus == 'rejected';
bool canSubmit = doc.docstatus == 'draft';
bool canApprove = doc.docstatus == 'pending' && currentUser.isApprover;
bool canCancel = !['completed', 'cancelled'].contains(doc.docstatus);
```

### Backend (ทุก handler ต้อง validate)
- `shopid` ต้องไม่ว่าง
- State transition ต้องถูกต้อง (ดู document-flows.md)
- user มีสิทธิ์เข้าถึง shop นี้ไหม?
- document ที่จะแก้มีอยู่จริงไหม?
- numeric precision: ราคาใช้ทศนิยม 2 ตำแหน่ง, จำนวนใช้ 3 ตำแหน่ง

---

## Common Mistakes — ต้องระวัง

### Mistake 1: State transition ผิด
```go
// ผิด — ข้าม pending
if doc.Status == "draft" { doc.Status = "approved" }

// ถูก
if doc.Status != "pending" {
    return c.JSON(400, map[string]interface{}{
        "message": "เอกสารต้องอยู่ในสถานะ pending จึงจะอนุมัติได้",
    })
}
```

### Mistake 2: ลืม Transaction เมื่อ update หลายตาราง
```go
// ผิด — ถ้า stock update สำเร็จแต่ doc update ล้มเหลว = ข้อมูลไม่ sync
UpdateStock(db, doc.Items)
UpdateDocument(db, doc)

// ถูก
tx, _ := db.Begin()
defer func() { if r := recover(); r != nil { tx.Rollback() } }()
UpdateStock(tx, doc.Items)
UpdateDocument(tx, doc)
tx.Commit()
```

### Mistake 3: แสดงปุ่มโดยไม่ check permission
```dart
// ผิด
ElevatedButton(onPressed: onApprove, child: Text('อนุมัติ'))

// ถูก
if (currentUser.hasPermission('po_approve'))
  ElevatedButton(onPressed: onApprove, child: Text('อนุมัติ'))
```

### Mistake 4: ลืมส่ง shopid
```dart
// ผิด
await dio.post('/goapi/api/po/list', data: {'limit': 20});

// ถูก
await dio.post('/goapi/api/po/list', data: {
  'shopid': shopId,
  'limit': 20,
});
```

### Mistake 5: สร้าง Dart model โดยไม่ verify API response ก่อน
ถูกต้อง: ใช้ MCP tool `get_table_sample` หรือ `execute_query` ดู data จริงก่อนสร้าง model

### Mistake 6: Frontend สร้าง docno เอง
ถูกต้อง: docno สร้างโดย backend เท่านั้น — frontend ไม่ต้องส่ง docno ตอน insert

---

## Pre-Commit Checklist

### Business Logic
- [ ] Multi-tenant: shopid ทุกที่?
- [ ] State transitions ถูกต้องตาม document-flows.md?
- [ ] Permission check ครบทุก critical operation?
- [ ] Validation ทั้ง client-side และ server-side?
- [ ] Notification ส่งถูก trigger?
- [ ] Audit log บันทึกการเปลี่ยน state สำคัญ?

### Cross-System
- [ ] Frontend เรียก API endpoint ถูก path?
- [ ] Request/Response field names ตรงกัน?
- [ ] Error messages ที่ backend ส่ง — frontend handle ครบ?

### Testing
- [ ] Happy path ผ่าน
- [ ] Validation error แสดงถูก
- [ ] State transitions ทุก case
- [ ] Edge cases: ข้อมูลว่าง, จำนวน 0
