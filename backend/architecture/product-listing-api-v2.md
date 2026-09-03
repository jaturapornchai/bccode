# Product Listing API v2 (`/goapi/product/v2`)

สถานะ: **draft contract** (2026-09-03) — ยังไม่ implement; ออกแบบให้ตรงกับ MongoModel project "BC Ai Account" diagram "ข้อมูลหลัก" (collection `product`, `productbarcode`, `channel_shop`, `channel_listing`, `channel_category_map`, `channel_brand_map`)

## Objective

ระบบมีผู้ใช้ 2 กลุ่มบนข้อมูลสินค้าชุดเดียว

| กลุ่ม | ต้องการ | ชั้นข้อมูล |
|---|---|---|
| นักบัญชี | รหัส ชื่อ หน่วยนับ ประเภท หมวด/กลุ่ม ภาษี → ออกเอกสารได้ทันที | **ชั้นบัญชี** (`product` core, `productbarcode` core) |
| เจ้าของกิจการ | สี/ขนาด รูป น้ำหนัก-ขนาดกล่อง ราคาขายส่ง รายละเอียดยาว ลงขายหลายร้านหลายช่องทาง | **ชั้นลงขาย** (`product.listing`, `productbarcode.listing`, `package*`, media) + **ชั้นช่องทาง** (`channel_*`) |

API ชุดนี้ให้บริการ**ชั้นลงขายและชั้นช่องทางเท่านั้น** รูปแบบ resource (item / model / tier / price / stock / media / lookup / shop / listing / sync) ออกแบบให้ใกล้เคียง API ของช่องทางขายออนไลน์ทั่วไป เพื่อให้ผู้เขียน connector ทำ mapping ได้ตรงไปตรงมา

หลักการบังคับ

1. **ชั้นบัญชีเป็น read-only ผ่าน API ชุดนี้** — payload ที่มีฟิลด์ชั้นบัญชี (ดู §3) ต้องถูกปฏิเสธด้วย `422 ACCOUNTING_FIELD_READONLY` พร้อมรายชื่อฟิลด์ ห้าม silently drop
2. **เขียนแบบ `$set` เฉพาะฟิลด์ + ตรวจ `__v`** — ห้ามใช้ repository ที่เขียนทับทั้งเอกสาร (`product_mongo_repository.go` `UpdateInCompany` เขียน full doc) เพราะ worker ซิงค์จะทับข้อมูลที่นักบัญชีเพิ่งแก้
3. **ตัวเลือกที่ขายได้ (model) = 1 เอกสาร `productbarcode`** — API ไม่สร้าง collection ตัวเลือกใหม่ barcode ยังเป็น stock key เดิม (stock spine อ่าน `barcode + whcode` ระดับ holding, `stockprocess.go`)
4. **สิ่งที่ช่องทางเป็นเจ้าของ** (รหัสรายการ/ตัวเลือกบนช่องทาง หมวด คุณลักษณะ แบรนด์ ผลซิงค์) อยู่ใน `channel_listing` เท่านั้น ไม่เขียนกลับเข้า `product`
5. **ไม่มี token ของร้านใน MongoDB** — `channel_shop.auth.credentialref` ชี้ไป secret store (วิธีเก็บรอตัดสิน G3)

## 1. Conventions

| เรื่อง | ค่า |
|---|---|
| Base path | `/goapi/product/v2` (อยู่ใต้ `authGroup` เดิมของ goapi) |
| Auth | `Authorization: Bearer <session token>` ตาม goapi เดิม; scope `holdingcode` + `businesscode` มาจาก `UserInfo` ของ session (บริษัทที่ผู้ใช้เลือก) ถ้าส่ง `holdingcode`/`businesscode` มาใน query/body ต้องตรงกับ session ไม่งั้น `403 FORBIDDEN`; ไม่ได้เลือกบริษัท → `409 COMPANY_REQUIRED` |
| Content type | `application/json; charset=utf-8` ยกเว้น upload = `multipart/form-data` |
| Response สำเร็จ | `{"success": true, "data": …}` |
| Response ผิดพลาด | `{"success": false, "code": "SNAKE_CASE", "message": "ข้อความไทยที่บอกว่าต้องทำอะไรต่อ", "fields": [{"field": "listing.title", "code": "TOO_LONG", "message": "…"}]}` — `fields` ใส่เฉพาะ validation |
| รหัสอ้างอิง | ใช้ business key เสมอ: `itemcode` (= `product.code`), `barcode` (= `productbarcode.barcode`), `shopid`, `channel` — ไม่ใช้ `_id`/`guidfixed` ใน API |
| เงิน | string ทศนิยม เช่น `"199.00"` (backend เก็บ Decimal128) |
| เวลา | ISO 8601 UTC เช่น `"2026-09-03T04:00:00Z"` |
| น้ำหนัก/ขนาด | กิโลกรัม / เซนติเมตร (Number) |
| Pagination | `offset` + `limit` (ค่าเริ่มต้น 0 / 50, สูงสุด 200) ตอบ `{"items": [], "total": n, "offset": 0, "limit": 50}` — เลือก offset ตาม `reports.go` เดิม จอ list ของระบบเป็นตารางเลขหน้า ไม่ใช่ infinite scroll |
| Idempotency | endpoint ที่สร้าง (`model/add`, `listing/create`, `sync/push`) รับ header `Idempotency-Key` (UUID) เก็บ 24 ชม. ซ้ำ → ตอบผลเดิม |
| Concurrency | ทุก update ส่ง `__v` ของเอกสารที่อ่านมา ไม่ตรง → `409 VERSION_CONFLICT` |
| ค่า enum | ตัวพิมพ์ใหญ่ตาม model (`NEW|USED`, `NORMAL|UNLIST`, `ok|error|pending` สำหรับ syncstate) |

## 2. Permissions

ชื่อสิทธิ์ต่อไปนี้เป็น**ข้อเสนอ** (รอลุงจืดยืนยัน G9) ใช้โมเดล `rolepermission.permissions []string` เดิม

| สิทธิ์ | ให้ทำอะไร | endpoint |
|---|---|---|
| `/product` (เดิม) | อ่าน/แก้ชั้นบัญชี — **ไม่ผ่าน API v2** | — |
| `/product/listing` | แก้ชั้นลงขายของ product/productbarcode | `item/*`, `tier/*`, `model/*`, `price/*`, `media/*` |
| `/marketplace/shop` | ตั้งค่าร้าน คลัง ระดับราคา (เจ้าของกิจการ/ADMIN) | `shop/*` |
| `/marketplace/listing` | สร้าง/เผยแพร่/ดึงรายการบนช่องทาง | `listing/*`, `sync/*`, `stock/*`, `lookup/*` |

ทุกสิทธิ์ตรวจที่ backend (middleware + service) ไม่ใช่แค่ซ่อนปุ่มหน้าจอ

## 3. Field ownership (whitelist)

**ชั้นบัญชี — read-only ใน v2** (ส่งมา → 422)

- `product`: `code, names, guidfixed, unitcode, unitguid, unitnames, unitconversions, standvalue, dividevalue, condition, itemtype, materialtype, groupcode, groupnames, categorycode, categorynames, classcode, brandcode, brandnames, design/model/pattern/grade/groupsubone/groupsubtwo (code+names), bom, isusesubbarcodes, refbarcodes, orderpoint, minpoint, maxpoint, isactive, isdeleted, audit fields`
- `productbarcode`: `barcode, itemcode, guidfixed, itemunitcode, itemunitguid, itemunitnames, names, prices, standvalue, dividevalue, condition, ismainbarcode, isdeleted, audit fields`

**ชั้นลงขาย — เขียนได้ผ่าน v2**

- `product`: `description, imageuri, imageurithumb, images[], videos[], packageweight, packagelength, packagewidth, packageheight, listing.*`
- `productbarcode`: `imageuri, imageurithumb, images[], videos[], packageweight, packagelength, packagewidth, packageheight, listing.*`
- `channel_shop.*`, `channel_listing.*`, `channel_category_map.*`, `channel_brand_map.*` ทั้งเอกสาร (ยกเว้น audit fields ที่ backend ตั้งเอง)

หมายเหตุ `productbarcode.prices` เป็นของนักบัญชี ราคาที่ส่งขึ้นช่องทาง = `channel_listing.models[].price` (override ต่อร้าน) ถ้าว่างใช้ `prices[keynumber = ระดับราคาของ channel_shop.salechannelcode]`

## 4. Endpoint list

### 4.1 Item (สินค้า = `product`)

| Method | Path | สิทธิ์ | คำอธิบาย |
|---|---|---|---|
| POST | `item/get` | `/product/listing` | อ่านสินค้า 1 ตัว ทั้ง 2 ชั้น (ชั้นบัญชีส่งกลับแบบอ่านอย่างเดียว) + ตัวเลือกทั้งหมด |
| POST | `item/list` | `/product/listing` | รายการสินค้า กรอง `q`, `categorycode`, `hastiers`, `readiness` (`accounting|listing|channel`) |
| POST | `item/update-listing` | `/product/listing` | แก้ชั้นลงขายระดับสินค้า (`listing.*`, `package*`, `description`) |
| POST | `item/readiness` | `/product/listing` | ตรวจว่าสินค้าพร้อมลงขายช่องทางไหน คืนรายการที่ยังขาดเป็นภาษาไทย |

### 4.2 Tier variation (ชั้นตัวเลือก)

| Method | Path | สิทธิ์ | คำอธิบาย |
|---|---|---|---|
| POST | `tier/init` | `/product/listing` | กำหนดชั้นตัวเลือก (≤2 ชั้น) พร้อมสร้าง/จับคู่ productbarcode ให้ทุกชุดผสม |
| POST | `tier/update` | `/product/listing` | เพิ่ม/ลบ/เรียงตัวเลือก (ลบได้เฉพาะตัวเลือกที่ยังไม่มี model เผยแพร่) |

### 4.3 Model (ตัวเลือกที่ขายได้ = `productbarcode`)

| Method | Path | สิทธิ์ | คำอธิบาย |
|---|---|---|---|
| POST | `model/list` | `/product/listing` | ตัวเลือกทั้งหมดของสินค้า พร้อม tierindex ราคา รูป สถานะขายออนไลน์ |
| POST | `model/add` | `/product/listing` + `/product` (สร้าง barcode) | สร้าง productbarcode ใหม่จาก tierindex (ต้องมีสิทธิ์สร้างบาร์โค้ดด้วย เพราะสร้างเอกสารชั้นบัญชี) |
| POST | `model/update` | `/product/listing` | แก้ `listing.*`, `package*`, รูป ของตัวเลือก |
| POST | `model/delete` | `/product/listing` | ปิดลงขาย (`listing.isforsale=false`) **ไม่ลบบาร์โค้ด** (บาร์โค้ดมีประวัติสต๊อก) |

### 4.4 Price / Stock

| Method | Path | สิทธิ์ | คำอธิบาย |
|---|---|---|---|
| POST | `price/update` | `/product/listing` | ตั้งราคาต่อร้านของหลาย model ในครั้งเดียว (`channel_listing.models[].price`) |
| POST | `stock/get` | `/marketplace/listing` | ยอดที่จะส่ง = คงเหลือคลังที่จับคู่ − `stockbuffer` แยกตาม model/คลัง (คำนวณ ไม่เก็บ) |
| POST | `stock/update` | `/marketplace/listing` | สั่งส่งสต๊อกขึ้นช่องทางทันที (คิวงาน) — ค่าที่ส่งคำนวณเอง ห้ามผู้ใช้กำหนดตัวเลขตรง |

### 4.5 Media

| Method | Path | สิทธิ์ | คำอธิบาย |
|---|---|---|---|
| POST | `media/upload` | `/product/listing` | multipart รูป/วิดีโอ → MinIO + สร้าง thumb อัตโนมัติ (ต่อยอด `/goapi/image/upload` เดิม) คืน `uri`, `urithumb`, มิติ, ขนาดไฟล์ |
| POST | `media/attach` | `/product/listing` | ผูก uri ที่อัปโหลดแล้วเข้า `product.images[]`/`productbarcode.images[]`/`tiers.options[].imageuri`/`sizechart` พร้อมลำดับ |
| POST | `media/detach` | `/product/listing` | ถอดรูป (ไม่ลบไฟล์ทันที ลบเมื่อไม่มีใครอ้าง) |

### 4.6 Lookup (ข้อมูลอ้างอิงของช่องทาง อ่านจาก cache)

| Method | Path | สิทธิ์ | คำอธิบาย |
|---|---|---|---|
| POST | `lookup/category-tree` | `/marketplace/listing` | ต้นไม้หมวดของช่องทาง (cache ต่อ channel+region) |
| POST | `lookup/attribute-tree` | `/marketplace/listing` | คุณลักษณะที่หมวดปลายสุดบังคับ/เลือกได้ |
| POST | `lookup/brand-list` | `/marketplace/listing` | แบรนด์ในหมวด (paging) |
| POST | `lookup/category-map/set` | `/marketplace/listing` | จับคู่ `categorycode` ของเรา → หมวดช่องทาง (`channel_category_map`) |
| POST | `lookup/brand-map/set` | `/marketplace/listing` | จับคู่ `brandcode` ของเรา → แบรนด์ช่องทาง (`channel_brand_map`) |

### 4.7 Shop

| Method | Path | สิทธิ์ | คำอธิบาย |
|---|---|---|---|
| POST | `shop/list` | `/marketplace/shop` | ร้านที่เชื่อมของบริษัทนี้ + สถานะสิทธิ์ (หมดอายุเมื่อไหร่) |
| POST | `shop/connect-start` | `/marketplace/shop` | คืน URL ให้เจ้าของไปกดอนุญาตที่ช่องทาง |
| POST | `shop/connect-callback` | `/marketplace/shop` | รับ code จากช่องทาง → เก็บ credential ใน secret store → สร้าง `channel_shop` |
| POST | `shop/update` | `/marketplace/shop` | ตั้ง `salechannelcode`, `branchcode`, `stockbuffer`, `isactive` |
| POST | `shop/warehouse-map` | `/marketplace/shop` | จับคู่คลังของช่องทาง ↔ `warehouse.code` |
| POST | `shop/refresh-limits` | `/marketplace/shop` | ดึง `limits` + `logisticchannels` จากช่องทางมาพักใหม่ |

### 4.8 Listing (รายการบนช่องทาง = `channel_listing`)

| Method | Path | สิทธิ์ | คำอธิบาย |
|---|---|---|---|
| POST | `listing/create` | `/marketplace/listing` | สร้างร่างจาก item (เติมค่าเริ่มต้นจาก category/brand map) state = `draft` |
| POST | `listing/get` | `/marketplace/listing` | อ่านรายการ 1 ตัว (รวม models, sync, violation) |
| POST | `listing/list` | `/marketplace/listing` | รายการตาม `channel`, `shopid`, `sync.state`, `itemstatus` |
| POST | `listing/update` | `/marketplace/listing` | แก้หมวด/คุณลักษณะ/แบรนด์/ขนส่ง/รูปแบบรายละเอียด/กำหนดเผยแพร่ |
| POST | `listing/validate` | `/marketplace/listing` | ตรวจตาม §5 + `limits` ของร้าน ไม่เขียนอะไร คืน errors ไทย |
| POST | `listing/publish` | `/marketplace/listing` | validate → `requestedstatus=NORMAL` → เข้าคิว push |
| POST | `listing/unlist` | `/marketplace/listing` | `requestedstatus=UNLIST` → เข้าคิว push |
| POST | `listing/delete` | `/marketplace/listing` | ลบรายการบนช่องทาง (ถ้าเผยแพร่แล้วต้อง unlist ก่อน) + soft delete เอกสาร |

### 4.9 Sync

| Method | Path | สิทธิ์ | คำอธิบาย |
|---|---|---|---|
| POST | `sync/push` | `/marketplace/listing` | ส่งเนื้อหา/ราคา/สต๊อกของรายการที่ระบุ (ข้ามถ้า `contenthash` เดิม) |
| POST | `sync/pull` | `/marketplace/listing` | อ่านสถานะ/สต๊อก/สถิติ/การละเมิดกลับมาเก็บ `platform`, `stats`, `violation` |
| POST | `sync/status` | `/marketplace/listing` | ความคืบหน้าคิว + ข้อผิดพลาดล่าสุด (ไทย) |

ทุก endpoint เป็น `POST` + JSON body ตามธรรมเนียม goapi เดิม (`/api/product/barcode`, `/image/*`)

## 5. Endpoint details

### 5.1 `item/get`

Request
```json
{ "itemcode": "SHIRT-001" }
```
Response
```json
{
  "success": true,
  "data": {
    "accounting": {
      "code": "SHIRT-001",
      "names": [{ "code": "th", "name": "เสื้อยืดคอกลม" }],
      "unitcode": "PCS",
      "itemtype": 0,
      "categorycode": "APPAREL",
      "brandcode": "",
      "isactive": true,
      "readonly": true
    },
    "listing": {
      "title": "เสื้อยืดคอกลม ผ้าฝ้าย 100%",
      "description": "…",
      "condition": "NEW",
      "preorder": { "ispreorder": false },
      "purchaselimit": { "min": 1, "max": 10 },
      "wholesale": [{ "mincount": 10, "maxcount": 49, "unitprice": "179.00" }],
      "sizechart": { "uri": "/goapi/s3/file/…", "urithumb": "…" },
      "tiers": [
        { "xorder": 0, "name": "สี", "options": [
          { "xorder": 0, "name": "ดำ", "imageuri": "…", "imageurithumb": "…" },
          { "xorder": 1, "name": "ขาว", "imageuri": "…", "imageurithumb": "…" } ] },
        { "xorder": 1, "name": "ขนาด", "options": [
          { "xorder": 0, "name": "M" }, { "xorder": 1, "name": "L" } ] }
      ]
    },
    "package": { "weight": 0.25, "length": 30, "width": 20, "height": 3 },
    "media": {
      "imageuri": "…", "imageurithumb": "…",
      "images": [{ "xorder": 0, "uri": "…", "urithumb": "…" }],
      "videos": []
    },
    "models": [
      { "barcode": "8850000000011", "tierindex": [0, 0], "itemunitcode": "PCS",
        "price": "199.00", "gtin": "8850000000011", "isforsale": true,
        "imageurithumb": "…", "package": null }
    ],
    "__v": 7
  }
}
```

### 5.2 `item/update-listing`

Request (ส่งเฉพาะฟิลด์ที่แก้ ใช้ `$set` ตาม path)
```json
{
  "itemcode": "SHIRT-001",
  "__v": 7,
  "listing": { "title": "เสื้อยืดคอกลม ผ้าฝ้าย 100% นุ่มใส่สบาย", "condition": "NEW" },
  "package": { "weight": 0.25, "length": 30, "width": 20, "height": 3 }
}
```
Response `{"success": true, "data": {"itemcode": "SHIRT-001", "__v": 8, "readiness": {…}}}`

Validation: E01 E02 E05 E06 E07 E08 E20 E24 — ถ้า body มี `names`, `unitcode`, `code` ฯลฯ → `422 ACCOUNTING_FIELD_READONLY` `fields: [{"field": "unitcode", …}]`

### 5.3 `item/readiness`

Request `{ "itemcode": "SHIRT-001", "channel": "shopee", "shopid": "123456" }` (`channel`/`shopid` ไม่ใส่ = ตรวจเฉพาะกฎกลาง)

Response
```json
{ "success": true, "data": {
  "accounting": { "ready": true, "missing": [] },
  "listing": { "ready": false, "missing": [
    { "field": "package.weight", "message": "ยังไม่ได้กรอกน้ำหนักสินค้า (กิโลกรัม)" },
    { "field": "media.images", "message": "ต้องมีรูปสินค้าอย่างน้อย 1 รูป" } ] },
  "channel": { "ready": false, "missing": [
    { "field": "categoryid", "message": "ยังไม่ได้เลือกหมวดหมู่ของช่องทางสำหรับหมวด APPAREL" } ] }
} }
```

### 5.4 `tier/init`

Request
```json
{
  "itemcode": "SHIRT-001",
  "__v": 8,
  "tiers": [
    { "name": "สี", "options": [{ "name": "ดำ", "imageuri": "…" }, { "name": "ขาว", "imageuri": "…" }] },
    { "name": "ขนาด", "options": [{ "name": "M" }, { "name": "L" }] }
  ],
  "models": [
    { "tierindex": [0, 0], "barcode": "8850000000011" },
    { "tierindex": [0, 1], "barcode": "8850000000028" },
    { "tierindex": [1, 0], "barcode": "", "generate": true },
    { "tierindex": [1, 1], "barcode": "", "generate": true }
  ],
  "itemunitcode": "PCS"
}
```
พฤติกรรม

- ทำใน MongoDB transaction: `$set product.listing.tiers` + สร้าง/อัปเดต `productbarcode.listing.tierindex` ทุกชุดผสม
- `barcode` ที่ระบุต้องมีอยู่และ `itemcode` ตรงกัน (E10); `generate=true` → สร้างบาร์โค้ดใหม่ (รูปแบบเลขตาม `barcode-form.tsx` เดิม, `ismainbarcode=false`, `itemunitcode` ตามที่ส่ง, `names` = ชื่อสินค้า + ชื่อตัวเลือก) — ต้องมีสิทธิ์ `/product` เพราะสร้างเอกสารชั้นบัญชี
- ถ้าสินค้ามี `channel_listing` ที่ `sync.state=published` และ `platform.haspromotion=true` → `409 LISTING_LOCKED_BY_PROMOTION` (E17)

Response `{"success": true, "data": {"itemcode": "…", "__v": 9, "models": [{"tierindex": [1,0], "barcode": "SHIRT-001-03", "created": true}, …]}}`

Validation: E03 E04 E09 E10 E11 E12

### 5.5 `model/update`

Request
```json
{
  "barcode": "8850000000011",
  "__v": 3,
  "listing": { "gtin": "8850000000011", "isforsale": true },
  "package": { "weight": 0.3, "length": 30, "width": 20, "height": 3 },
  "images": [{ "xorder": 0, "uri": "…", "urithumb": "…" }]
}
```
Validation: E05 E06 E13 E14 — ห้ามส่ง `prices`, `itemunitcode`, `names` (422)

### 5.6 `price/update`

Request
```json
{
  "channel": "shopee", "shopid": "123456",
  "items": [
    { "barcode": "8850000000011", "price": "199.00" },
    { "barcode": "8850000000028", "price": "199.00" },
    { "barcode": "SHIRT-001-03", "price": null }
  ]
}
```
`price: null` = ล้าง override กลับไปใช้ราคาจาก `prices[]` ตามระดับราคาของร้าน

Response `{"success": true, "data": {"updated": 3, "queued": true, "listingid": "…"}}`

Validation: E15 E16 E17 — ราคาต้องอยู่ใน `channel_shop.limits.pricemin..pricemax`

### 5.7 `stock/get`

Request `{ "channel": "shopee", "shopid": "123456", "barcodes": ["8850000000011"] }`

Response
```json
{ "success": true, "data": { "items": [
  { "barcode": "8850000000011", "stocks": [
    { "locationid": "TH-BKK-01", "warehousecode": "WH01", "onhand": 120, "buffer": 5, "sellable": 115 } ],
    "lastpushedstock": 110, "platformstock": 110, "reservedstock": 2 } ] } }
```
`sellable = max(onhand − buffer, 0)`; ยอดจอง (SO ค้างส่ง) ยังไม่หัก — รอตัดสิน G6

### 5.8 `media/upload`

multipart: `file` (jpg/png/webp ≤ 10 MB หรือ mp4 ≤ 30 MB), `kind` = `image|video|sizechart|desc`

Response
```json
{ "success": true, "data": {
  "uri": "/goapi/s3/file/product/2026/09/abc.webp",
  "urithumb": "/goapi/s3/file/product/2026/09/abc_thumb.webp",
  "width": 1200, "height": 1200, "bytes": 184320, "etag": "…" } }
```
กฎ: รูปสินค้าต้อง ≥ 500×500 และอัตราส่วน 1:1 (ระบบ crop กลางภาพให้ถ้าไม่ใช่) — ขนาดขั้นต่ำเป็นค่าของเราเอง (G8); วิดีโอเก็บ `durationsec`, `sizebytes` ไว้ตรวจกับช่องทางตอน publish

### 5.9 `shop/warehouse-map`

Request
```json
{ "channel": "shopee", "shopid": "123456", "__v": 2,
  "warehouses": [
    { "locationid": "TH-BKK-01", "warehousecode": "WH01", "isdefault": true },
    { "locationid": "TH-CNX-01", "warehousecode": "WH02" } ] }
```
Validation: `warehousecode` ต้องมีใน `warehouse` ของบริษัท; ร้านที่ `ismultiwarehouse=false` ให้มีได้ 1 รายการ (E22)

### 5.10 `listing/create`

Request `{ "channel": "shopee", "shopid": "123456", "itemcode": "SHIRT-001" }`

พฤติกรรม: สร้าง `channel_listing` state `draft`; เติม `categoryid/categorypath/attributes` จาก `channel_category_map[categorycode]`, `brandid/brandname` จาก `channel_brand_map[brandcode]` (ไม่มี → 0 / "No Brand"); สร้าง `models[]` จาก productbarcode ที่ `itemcode` ตรงและ `listing.isforsale=true`; `tiers[]` คัดลอกโครงจาก `product.listing.tiers`

Response = `listing/get`

Validation: E18 (ซ้ำ) E09

### 5.11 `listing/validate`

Request `{ "channel": "shopee", "shopid": "123456", "itemcode": "SHIRT-001" }`

Response
```json
{ "success": true, "data": { "ready": false, "errors": [
  { "code": "E01", "field": "listing.title", "message": "ชื่อสินค้ายาว 130 ตัวอักษร เกินที่ร้านนี้รับได้ 120 ตัวอักษร กรุณาย่อชื่อ" },
  { "code": "E19", "field": "attributes", "message": "หมวดหมู่นี้บังคับกรอก 'วัสดุ' และ 'ประเทศผู้ผลิต'" } ],
  "warnings": [ { "code": "W01", "field": "listing.description", "message": "รายละเอียดสั้นกว่า 100 ตัวอักษร ช่องทางอาจจัดอันดับต่ำ" } ] } }
```

### 5.12 `listing/publish`

Request `{ "channel": "shopee", "shopid": "123456", "itemcode": "SHIRT-001", "__v": 4, "scheduledpublishtime": null }`

พฤติกรรม: รัน `listing/validate` → ถ้าผ่าน `requestedstatus=NORMAL`, `sync.state=queued`, ใส่คิว push (outbox); ผลจริงมาทาง `sync/status`/`sync/pull` — API นี้ไม่รอช่องทางตอบ

Response `{"success": true, "data": {"sync": {"state": "queued"}, "jobid": "…"}}`

Validation: ทุกข้อใน §5 + E21 (`scheduledpublishtime` 1 ชม.–90 วัน)

### 5.13 `sync/pull`

Request `{ "channel": "shopee", "shopid": "123456", "itemcodes": ["SHIRT-001"] }`

พฤติกรรม: อ่านสถานะจากช่องทาง → เขียนเฉพาะ `channel_listing.platform`, `itemstatus`, `models[].platformstock/reservedstock`, `stats`, `violation`, `sync.lastpullat` — **ไม่แตะ `product`/`productbarcode`**

### 5.14 `sync/status`

Response
```json
{ "success": true, "data": { "jobs": [
  { "jobid": "…", "itemcode": "SHIRT-001", "state": "error",
    "lasterrorcode": "CHANNEL_ATTRIBUTE_MISSING",
    "lasterrorth": "ช่องทางแจ้งว่าขาดคุณลักษณะ 'วัสดุ' กรุณาเพิ่มในหน้ารายการลงขายแล้วกดเผยแพร่อีกครั้ง",
    "attempts": 2, "nextretryat": "2026-09-03T05:10:00Z" } ] } }
```

## 6. Validation rules

| รหัส | กฎ | ที่ตรวจ |
|---|---|---|
| E01 | `listing.title` ความยาวอยู่ใน `limits.itemnamemin..itemnamemax` ของร้าน (ไม่มี limits → 5..120) ห้ามมีอักขระควบคุม | item/update-listing, listing/validate |
| E02 | `listing.description` (หรือ `description` ถ้าว่าง) อยู่ใน `limits.descriptionmin..descriptionmax`; `product.description` เองยังคง max 1500 ตาม `product.go` | เดียวกัน |
| E03 | `listing.tiers` ≤ 2 ชั้น; `xorder` ต่อเนื่อง 0..n | tier/* |
| E04 | จำนวนชุดผสม (ชั้น1 × ชั้น2) ≤ 50; ชื่อตัวเลือกไม่ซ้ำในชั้น; ความยาวชื่อชั้น/ตัวเลือก ≤ `limits.tiernamemax/tieroptionmax` | tier/* |
| E05 | `package.*` กรอกครบทั้ง weight+length+width+height หรือว่างทั้งหมด; ค่า > 0 | item/model update |
| E06 | น้ำหนัก/ขนาดของ model ถ้ากรอกจะแทนค่าของ product (ไม่ผสม) | model/update |
| E07 | `wholesale[]` ช่วงไม่ทับซ้อน, `mincount<maxcount`, `unitprice` ≤ ราคาปกติทุก model และ ≥ `limits.wholesalepctmin`% ของราคาปกติ | item/update-listing |
| E08 | `purchaselimit.min ≤ max`; `preorder.daystoship` อยู่ใน `channel_category_map.limits.daystoshipmin..max` ของหมวดที่จับคู่ | item/update-listing, listing/validate |
| E09 | สินค้าที่ไม่มี tiers ต้องมี model ที่ `isforsale=true` อย่างน้อย 1 และไม่เกิน 1 ต่อ listing; สินค้าที่มี tiers ทุกชุดผสมต้องมี model | tier/init, listing/create |
| E10 | ทุก `barcode` ใน models ต้องมี `productbarcode.itemcode == itemcode` | tier/init, model/add |
| E11 | `listing.tierindex` ยาวเท่าจำนวนชั้น ค่าอยู่ในช่วง options และไม่ซ้ำกันในสินค้าเดียวกัน | tier/*, model/add |
| E12 | รูปตัวเลือกใส่ได้เฉพาะชั้นแรก และต้องครบทุกตัวเลือกหรือไม่ใส่เลย | tier/* |
| E13 | `gtin` ว่างหรือเป็นเลข 8/12/13/14 หลักที่ผ่าน check digit (GS1 mod-10); ถ้า `channel_category_map.limits.gtinrule=Mandatory` ต้องมี | model/update, listing/validate |
| E14 | `images[]` ทุกตัวต้องมี `urithumb`; จำนวน ≤ `limits.imagecountmax` (ไม่มี → 9); `xorder` ไม่ซ้ำ | media/attach, listing/validate |
| E15 | ราคาเป็นทศนิยม ≥ 0 อยู่ใน `limits.pricemin..pricemax`; ทศนิยม ≤ 2 ตำแหน่ง | price/update |
| E16 | ถ้ามี `wholesale` ทุก model ในรายการเดียวกันต้องราคาเท่ากัน | price/update, listing/validate |
| E17 | ขณะ `platform.haspromotion=true` ห้ามแก้ title / ราคา / โครงตัวเลือก / หมวด → 409 | ทุก write ที่กระทบ |
| E18 | `channel_listing` ซ้ำ (channel+shopid+itemcode ที่ยังไม่ลบ) → 409 | listing/create |
| E19 | `attributes[]` ต้องครบทุกตัวที่ `mandatory=true` ของหมวด; ค่าเชิงปริมาณต้องมี `unit` ในรายการที่หมวดรับ; `valueid=0` ต้องมี `name` | listing/update, listing/validate |
| E20 | `listing.condition` ∈ NEW|USED; `descriptiontype=extended` ใช้ได้เฉพาะร้านที่ `limits` อนุญาต | item/update-listing, listing/update |
| E21 | `scheduledpublishtime` ว่าง หรืออยู่ระหว่าง now+1h ถึง now+90d และใช้ได้เฉพาะ state `draft`/`unlinked` | listing/publish |
| E22 | ร้าน `ismultiwarehouse=false` มี warehouse map ได้ 1 รายการ; ทุก `warehousecode` ต้องมีจริงในบริษัท | shop/warehouse-map |
| E23 | สต๊อกที่ส่ง = max(คงเหลือคลังที่จับคู่ − `stockbuffer`, 0) ≤ `limits.stockmax`; ผู้ใช้กำหนดตัวเลขตรงไม่ได้ | stock/update |
| E24 | payload มีฟิลด์ชั้นบัญชี → 422 `ACCOUNTING_FIELD_READONLY` ระบุทุกฟิลด์ | ทุก write ใน item/model |

## 7. Error codes

| HTTP | code | ข้อความไทย (ตัวอย่าง) |
|---|---|---|
| 401 | `UNAUTHORIZED` | กรุณาเข้าสู่ระบบใหม่ |
| 403 | `FORBIDDEN` | คุณไม่มีสิทธิ์แก้ข้อมูลส่วนนี้ ติดต่อผู้ดูแลเพื่อขอสิทธิ์ "ลงขายออนไลน์" |
| 409 | `COMPANY_REQUIRED` | กรุณาเลือกบริษัทก่อนใช้งาน |
| 404 | `ITEM_NOT_FOUND` | ไม่พบสินค้ารหัสนี้ในบริษัทที่เลือก |
| 404 | `MODEL_NOT_FOUND` | ไม่พบบาร์โค้ดนี้ |
| 404 | `SHOP_NOT_FOUND` | ไม่พบร้านค้าที่เชื่อมไว้ กรุณาเชื่อมร้านก่อน |
| 409 | `VERSION_CONFLICT` | มีคนแก้ข้อมูลนี้ก่อนหน้าคุณ กรุณาโหลดหน้าใหม่แล้วแก้อีกครั้ง |
| 409 | `LISTING_LOCKED_BY_PROMOTION` | สินค้านี้กำลังติดโปรโมชันบนช่องทาง แก้ชื่อ/ราคา/ตัวเลือกได้หลังโปรโมชันจบ |
| 409 | `LISTING_EXISTS` | สินค้านี้มีรายการลงขายในร้านนี้แล้ว |
| 409 | `LISTING_MUST_UNLIST_FIRST` | ต้องซ่อนรายการจากช่องทางก่อนลบ |
| 422 | `ACCOUNTING_FIELD_READONLY` | ฟิลด์ต่อไปนี้แก้ได้เฉพาะหน้าสินค้าโดยผู้มีสิทธิ์บัญชี: หน่วยนับ, ชื่อสินค้า |
| 422 | `VALIDATION_FAILED` | ข้อมูลยังไม่ครบ ดูรายการที่ต้องแก้ด้านล่าง (fields[]) |
| 422 | `TIER_TOO_MANY_COMBINATIONS` | ตัวเลือกรวมกันได้ 60 ชุด เกิน 50 ชุดที่ช่องทางรับ กรุณาลดตัวเลือก |
| 422 | `GTIN_INVALID` | เลข GTIN ไม่ถูกต้อง (ตรวจเลขหลักสุดท้ายไม่ผ่าน) ถ้าไม่มี GTIN ให้เว้นว่าง |
| 422 | `PACKAGE_INCOMPLETE` | กรอกน้ำหนักและขนาดกล่องให้ครบทั้ง 4 ช่อง หรือเว้นว่างทั้งหมด |
| 422 | `MEDIA_TOO_SMALL` | รูปเล็กกว่า 500×500 พิกเซล กรุณาใช้รูปที่ใหญ่กว่านี้ |
| 422 | `SHOP_TOKEN_EXPIRED` | สิทธิ์เชื่อมต่อร้านหมดอายุ กรุณากด "เชื่อมร้านใหม่" |
| 429 | `CHANNEL_RATE_LIMITED` | ช่องทางจำกัดจำนวนคำขอ ระบบจะลองใหม่ให้อัตโนมัติใน 1 นาที |
| 502 | `CHANNEL_ERROR` | ช่องทางตอบกลับผิดพลาด (รหัส …) ระบบบันทึกไว้แล้ว ลองใหม่ภายหลัง |

## 8. Phasing

| เฟส | ขอบเขต | ไม่ต่อช่องทางจริง? |
|---|---|---|
| 1 | `item/*`, `tier/*`, `model/*`, `media/*`, `price/update` (เก็บใน channel_listing แบบ draft), `listing/create|get|list|update|validate` — validator ใช้ค่า default เมื่อไม่มี `limits` | ใช่ ทดสอบด้วยหน้าจอ + Mongo ได้ครบ |
| 2 | `shop/*` (OAuth ต่อช่องทาง = R0 ขออนุมัติต่อช่องทาง), `listing/publish|unlist|delete`, `sync/push|status`, `stock/get|update`, outbox + worker | ต่อจริง 1 ช่องทางแรก |
| 3 | `lookup/*` cache + category/brand map UI, `sync/pull` (สถานะ/สถิติ/การละเมิด), มาตรการ rate limit/retry, `descriptiontype=extended` | ต่อจริง |
| 4 | ถอด Go `marketplaceproducts[]`, `marketplaceskumappings`, `sellersku`, `skupackage*` และ `tab-product-marketplace.tsx` (รอ G2) | — |

ทุกเฟส: UAT ตามกฎ AGENTS (CRUD + ตรวจ Mongo ทีละ step + MinIO มี object+thumb) และ contract test ต่อ whitelist §3 (ส่งฟิลด์บัญชีทุกตัวแล้วต้องได้ 422)

## 9. Open decisions (blocking implementation)

| # | คำถาม | ผลต่อ API |
|---|---|---|
| G2 | deprecate ฟิลด์ marketplace เดิมใน Go/frontend ได้ไหม | เฟส 4 |
| G3 | เก็บ credential ร้านที่ไหน | `shop/connect-callback` |
| G5 | 1 ร้าน = 1 บริษัท หรือผูกสาขา | key ของ `channel_shop` |
| G6 | หักยอดจอง SO ค้างส่งจากสต๊อกที่ส่งหรือไม่ | `stock/get` E23 |
| G7 | สินค้าหลายหน่วย (ขวด/แพ็ก/ลัง) = model หลายตัวใน item เดียว หรือแยก item | E09, `tier/init` |
| G9 | ชื่อสิทธิ์ตาม §2 | middleware |

## 10. Limitations

- ไม่ครอบคลุมคำสั่งซื้อ/การเงิน/ค่าธรรมเนียมของช่องทาง (คนละ API)
- ข้อจำกัดของช่องทาง (ความยาวชื่อ จำนวนรูป ราคา) เปลี่ยนได้ตามร้าน/หมวด → validator อ่านจาก `limits` ที่ดึงมาพัก ไม่ hard-code; ค่าที่ใช้ตอนไม่มี `limits` เป็นค่าประมาณของเราเอง
- ไม่รองรับสินค้าข้ามพรมแดน (cross-border) และรายละเอียดแบบผสมรูป (`extended`) ในเฟส 1
- ราคาบนช่องทางถือเป็นราคาที่ผู้ซื้อจ่าย การคำนวณ VAT อยู่ฝั่งเอกสารขายของเรา ไม่ได้อยู่ใน API นี้
