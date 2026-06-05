# PostgreSQL Product Schema

ข้อมูลนี้ตรวจจาก PostgreSQL จริงของ DEV runtime เมื่อวันที่ 2026-06-05 โดย query เฉพาะ metadata จาก `information_schema.columns` ไม่ได้ดึงข้อมูลสินค้าออกมา

## ตาราง `public.product`

ตาราง `public.product` เป็น PostgreSQL projection/read model สำหรับรายการสินค้าแม่ ใช้แสดงข้อมูลสินค้าและยอดคงเหลือระดับสินค้า ไม่ใช่ตารางสำหรับบาร์โค้ด/SKU ย่อย

พบ table นี้ใน 2 database และ schema เหมือนกัน

| ลำดับ | Field | Type | Null | Default | คำอธิบาย |
|---:|---|---|---|---|---|
| 1 | `id` | `integer` | ไม่ได้ | `nextval('product_id_seq')` | เลขลำดับภายใน PostgreSQL |
| 2 | `itemcode` | `text` | ได้ | - | รหัสสินค้า |
| 3 | `name0` | `text` | ไม่ได้ | - | ชื่อสินค้า |
| 4 | `unitcode` | `text` | ไม่ได้ | - | รหัสหน่วยนับหลัก |
| 5 | `unitname` | `text` | ไม่ได้ | - | ชื่อหน่วยนับหลัก |
| 6 | `balanceqty` | `numeric` | ได้ | `0` | ยอดคงเหลือรวมของสินค้า |
| 7 | `balanceqtyword` | `text` | ได้ | `''` | ยอดคงเหลือแบบแปลงเป็นหน่วยแพ็ก เช่น ลัง/โหล/ชิ้น |
| 8 | `pendingrecvqty` | `numeric` | ได้ | `0` | จำนวนรอรับเข้า |
| 9 | `pendingrecvqtyword` | `text` | ได้ | `''` | จำนวนรอรับเข้าแบบข้อความหน่วยแพ็ก |
| 10 | `pendingsendqty` | `numeric` | ได้ | `0` | จำนวนรอส่งออก |
| 11 | `pendingsendqtyword` | `text` | ได้ | `''` | จำนวนรอส่งออกแบบข้อความหน่วยแพ็ก |

## ข้อสังเกต

- `product` ตอนนี้มีเฉพาะข้อมูลหลักและยอดคงเหลือ
- ไม่มี field `categorycode`, `vattype`, `costprice`
- หน้าสินค้าที่ต้องแสดงยอดคงเหลือควรใช้ `balanceqty` หรือ `balanceqtyword`
- ข้อมูล SKU, barcode, ราคา, และหน่วยย่อยอยู่ที่ `productbarcode`
- ข้อมูลเต็มสำหรับ create/edit สินค้ายังต้องอ่านจาก MongoDB operational data ตามกฎระบบ
- Query ที่อ้าง `public.product` ต้องป้องกันกรณี tenant บางรายยังไม่มี table นี้ เพราะตรวจพบว่าหลาย database มี `productbarcode` แต่ยังไม่มี `product`
