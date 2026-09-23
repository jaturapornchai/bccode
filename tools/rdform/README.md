# rdform — แบบฟอร์มกรมสรรพากรที่ backend กรอกเอง

ต้นฉบับแบบฟอร์ม = PDF ทางการของกรมสรรพากรที่ลุงจืดวางไว้ใน `mydocs/sample/` (read-only — เครื่องมือนี้แค่อ่าน/คัดลอก)
ทุกไฟล์เป็น PDF กรอกได้ (AcroForm) จึงมีพิกัดทุกช่องอยู่แล้ว งานของเราคือบอกว่า "ช่องหมายเลขไหนคือข้อมูลอะไร"

```
mydocs/sample/<แบบ>.pdf ──rdform_widgets.py──▶ ภาพมีหมายเลขช่อง (ไว้ดู)          
tools/rdform/specs/<code>.json (คนเขียน: ช่องไหน = key อะไร)
        └──rdform_build.py──▶ backend/internal/rdform/assets/<code>.pdf + specs/<code>.json (generated, embed ใน Go)
```

## คำสั่ง

```bash
py tools/rdform/rdform_widgets.py "mydocs/sample/ภงด.53-แบบ.pdf" /tmp/rdf/pnd53 2   # → /tmp/rdf/pnd53-p1.png (+widgets.json)
py tools/rdform/rdform_build.py --check pnd53      # ตรวจ spec อย่างเดียว
py tools/rdform/rdform_build.py pnd53 pnd53_attach # ตรวจ + เขียนไฟล์ให้ backend
```

ต้องมี `pypdf`, `pypdfium2`, `Pillow` (`py -m pip install --user pypdf pypdfium2 pillow`)

## รูปแบบ spec (`tools/rdform/specs/<code>.json`)

```json
{
 "code": "pnd53",                       // [a-z0-9_] — แบบ = pnd53, ใบแนบ = pnd53_attach
 "source": "ภงด.53-แบบ.pdf",            // ชื่อไฟล์ใน mydocs/sample
 "title": "ภ.ง.ด.53 แบบยื่นรายการภาษีเงินได้หัก ณ ที่จ่าย ...",
 "pages": [1],                          // หน้าที่เป็นแบบฟอร์มจริง (ไม่เอาหน้าคำชี้แจง)
 "fields": [
  {"key": "tax_id", "w": 1, "type": "taxid", "label": "เลขประจำตัวผู้เสียภาษีอากร ...", "group": "payer"},
  {"key": "month", "type": "choice", "label": "เดือนที่จ่ายเงินได้", "group": "period",
   "options": [{"value": "1", "w": 33, "label": "(1) มกราคม"}]}
 ],
 "table": {                             // มีเฉพาะใบแนบที่เป็นตารางรายการซ้ำ (ไม่มีก็ไม่ต้องใส่)
  "key": "payees", "label": "...",
  "columns": [{"key": "seq", "type": "int", "label": "ลำดับที่"}],
  "rows": [{"seq": 5, "tax_id": 6}]      // หนึ่ง object ต่อหนึ่งแถวบนกระดาษ: column key → หมายเลขช่อง
 }
}
```

- `w` = หมายเลขช่องบนภาพจาก `rdform_widgets.py` (ตัวเลขแดง = ช่องข้อความ, น้ำเงิน = ช่องติ๊ก)
- `label` = ข้อความตามที่พิมพ์บนแบบ (ภาษาไทย) — ใช้เป็นป้ายบนจอแก้ไข
- `group` = หมวดบนจอแก้ไข (`payer`, `address`, `period`, `filing`, `attachment`, `summary`, `signature`, `sheet`, หรือหมวดเฉพาะแบบ)
- ช่องหนึ่งใช้ได้ครั้งเดียว; ช่องที่ไม่รู้ความหมายแน่ชัด **ห้ามเดา** ปล่อยว่างแล้วรายงาน (build จะเตือน `unused widgets`)

### type

| type | ใช้กับ | backend วาดอย่างไร |
|---|---|---|
| `text` | ข้อความทั่วไป | บรรทัดเดียว ย่อฟอนต์ให้พอดีช่อง ชิดตาม Q ของช่อง |
| `digits` | ช่องตัวเลขเป็นตาราง (สาขา 5 หลัก, รหัสไปรษณีย์) | ทีละหลักลงช่องที่ตรวจเจอ |
| `taxid` | เลขประจำตัวผู้เสียภาษี 13 หลัก | ทีละหลักลง 13 ช่อง |
| `money` | จำนวนเงิน | บาทชิดขวาก่อนเส้นแบ่ง สตางค์ในช่องสตางค์ (ถ้าไม่มีเส้นแบ่ง = `1,234.56` ชิดขวา) |
| `int` | จำนวนนับ (ราย, แผ่น, ลำดับ) | ตัวเลข ชิดตาม Q |
| `check` | ช่องติ๊กเดี่ยว | ✓ เมื่อค่าเป็นจริง |
| `choice` | กลุ่มช่องติ๊กที่เลือกได้ข้อเดียว (radio) | ✓ ที่ตัวเลือกที่ตรงค่า |

## key มาตรฐาน (ใช้ชื่อเดียวกันทุกแบบ — backend เติมให้อัตโนมัติ)

| key | ความหมาย |
|---|---|
| `tax_id`, `branch_no`, `name` | เลขผู้เสียภาษี / สาขาที่ / ชื่อ ของ **ผู้ยื่นแบบ** (ผู้หักภาษี / ผู้ประกอบการ / บริษัท) |
| `addr_building`, `addr_room`, `addr_floor`, `addr_village`, `addr_no`, `addr_moo`, `addr_soi`, `addr_junction`, `addr_road`, `addr_subdistrict`, `addr_district`, `addr_province`, `addr_postcode`, `phone` | ที่อยู่ผู้ยื่นแบบ (แยก = `addr_junction`) |
| `month` (choice value `1`–`12`), `year_be` | เดือน/ปี พ.ศ. ของงวดภาษี |
| `filing_type` (choice `normal` / `additional`), `additional_no` | ยื่นปกติ / ยื่นเพิ่มเติมครั้งที่ |
| `signer_name`, `signer_position`, `sign_day`, `sign_month`, `sign_year_be` | ผู้ลงนาม + วันที่ยื่น |
| `sheet_no`, `sheet_total` | ใบแนบ แผ่นที่ / ในจำนวน ... แผ่น |
| `page_total_*` | ยอดรวมท้ายใบแนบแต่ละแผ่น (backend รวมให้) |
| `attach_payees`, `attach_sheets` | จำนวนราย/แผ่นของใบแนบที่ระบุบนแบบ |

ตารางรายผู้มีเงินได้ (ใบแนบ ภ.ง.ด.): `seq`, `tax_id`, `branch_no`, `name`, `address1`, `address2`, และรายการเงินได้ `l1_date`, `l1_income_type`, `l1_rate`, `l1_amount`, `l1_tax`, `l1_condition` (`l2_…`, `l3_…` สำหรับบรรทัดถัดไปของผู้รับรายเดียวกัน)

key อื่นเฉพาะแบบ ตั้งเป็นอังกฤษ snake_case ที่อ่านแล้วรู้ความหมาย (เช่น `sales_amount`, `output_tax`, `input_tax`)
ช่องหมายเลขข้อบนแบบ (ข้อ 1., 2., …) ให้ใส่เลขข้อไว้ใน `label` ด้วย

ตัวอย่างอ้างอิงที่ทำเสร็จแล้ว: `specs/pnd53.json` (แบบ) และ `specs/pnd53_attach.json` (ใบแนบแบบตาราง)

## สิ่งที่ตัว build/ตัวกรอกทำให้เอง (ไม่ต้องเขียนในสเปก)

- **ช่องเงินแบบบาท|สตางค์**: build วัดเส้นแบ่งบนแบบจริงแล้วเขียนลงสเปกที่ generate — `split` (x ของเส้นแบ่งบาท/สตางค์), `baht_end` (ขอบขวาช่องบาทเมื่อมีขีด `-` คั่น), `cells` + `satang: 2` (แบบที่พิมพ์ตัวเลขช่องละหลัก) — ตัวกรอกวางบาทชิดขวาก่อนเส้น สตางค์ในช่องสตางค์
- **คู่ช่อง `X_baht` + `X_satang`** (แบบที่พิมพ์เป็น 2 ช่องกรอกแยก): จอเห็นเป็นยอดเงินเดียว `X` (`rdform.Spec.MoneyPairs`) ตัวกรอกแยกบาท/สตางค์ให้
- **คอลัมน์ `X_d1`..`X_dN`** ของตารางใบแนบ (เลขที่พิมพ์ช่องละหลักเป็นคนละช่องกรอก เช่น สาขาที่ของใบแนบ ภ.พ.30): จอเห็นเป็นช่องเดียว `X` (`rdform.Table.DigitGroups`)
- **ใบแนบ 2 แบบ** (`rdform.SchemaFor`): สเปกมี `table` = `rows` (ระบบแบ่งแผ่นละ `len(table.rows)` แถว เติม `seq`/`sheet_no`/`sheet_total`/`page_total_*` เอง); ไม่มี `table` = `sheets` (หนึ่งแผ่นต่อหนึ่งรายการ เช่น ใบแนบ ภ.ธ.40 รายสถานประกอบการ)
- ค่าที่ต้องคำนวณตามสูตรบนแบบ (ยอดรวม/ภาษีสุทธิ) อยู่ใน Go `backend/internal/goapi/handlers/tax_form_compute.go` ไม่ใช่ในสเปก — เพิ่มแบบใหม่ที่มีสูตร ต้องเพิ่ม computer + unit test ที่นั่น และเพิ่มแบบใน `taxFormOrder`/`taxForms` ของ `tax_form.go` + แถว `tax_form_title_<code>` ใน `languages.tsv`

แบบที่ทำครบแล้ว (สเปกใน `specs/`): ภ.ง.ด.2, 2ก, 3, 53 (+ใบแนบ), ภ.พ.30 (+ใบแนบรายสาขา), ภ.พ.36, ภ.ธ.40 (+ใบแนบ), ภ.ง.ด.50, 51, 93, 94 — ภ.ง.ด.1/1ก ไม่ทำ (ระบบเงินเดือนอยู่นอกขอบเขต), 50 ทวิ ใช้ `backend/internal/whtcert` (ADR `docs/kms/decisions/2026-09-23-rd-tax-forms-engine.md`)
