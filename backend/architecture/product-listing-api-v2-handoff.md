# Handoff — Product Listing API v2 (เฟส 1 เสร็จ → ทำต่อเฟสถัดไป)

วันที่ 2026-09-03 · ผู้รับงานต่อ: AI/นักพัฒนาคนถัดไป · เอกสารนี้ต้องอ่านคู่กับ `product-listing-api-v2.md` (contract = source of truth)

## 0. กฎที่ห้ามละเมิด (สรุปจาก AGENTS.md + contract)

1. **ห้ามเอ่ยชื่อ marketplace/vendor ใด ๆ** ในโค้ด คอมเมนต์ ชื่อฟิลด์ ข้อความ error เอกสาร — ใช้คำว่า "ช่องทางขายออนไลน์ / รายการบนช่องทาง" (ค่า enum `channel` ใน DB เป็นข้อมูล ไม่ใช่ข้อความในโค้ด)
2. **ชั้นบัญชี read-only ผ่าน API v2** — payload มีฟิลด์บัญชี (contract §3) → `422 ACCOUNTING_FIELD_READONLY` ระบุทุกฟิลด์ ห้าม drop เงียบ
3. **เขียน Mongo ด้วย `$set` เฉพาะ path ที่ whitelist + filter `__v` + `$inc __v`** — ห้ามใช้ repository เดิมที่เขียนทับทั้ง doc (`internal/product/product/repositories/product_mongo_repository.go` `UpdateInCompany`)
4. ข้อความถึงผู้ใช้เป็นภาษาไทยที่คนอายุ 40+ อ่านแล้วรู้ว่าต้องทำอะไร (ดู §7 ของ contract)
5. ห้าม commit/push เอง (ลุงจืดสั่งเท่านั้น) · ห้ามแตะ 192.168.2.202 / DigitalOcean · ทดสอบกับ local stack เท่านั้น
6. ห้ามเก็บ secret ใน Mongo/โค้ด · รูปเก็บ MinIO + ต้องมี thumb คู่เสมอ (`urithumb`)
7. เปลี่ยนแปลงเล็กสุดที่แก้ปัญหา (patch > refactor) · ทำตาม style ไฟล์ข้างเคียง

## 1. สถานะ repo

- branch `dev` · checkpoint ล่าสุดที่ push แล้ว = `5ee63a38` (ก่อนเริ่มเฟส 1)
- **งานเฟส 1 ยังไม่ commit** (20 ไฟล์): แก้ `backend/internal/goapi/bootstrap.go`, `internal/product/product/models/product.go`, `internal/product/productbarcode/models/product_barcode.go`, `backend/architecture/product-listing-api-v2.md`; ไฟล์ใหม่ `internal/goapi/handlers/product_v2_*.go` (13 ไฟล์), `internal/product/product/models/product_listing.go` (+test), `internal/product/productbarcode/models/product_barcode_listing_test.go`
- ถ้าต้องย้อน: `git stash` หรือ `git checkout 5ee63a38 -- <path>` (ไฟล์ใหม่ลบทิ้ง)

## 2. สิ่งที่ทำเสร็จแล้ว (เฟส 1)

| ส่วน | ไฟล์ | หมายเหตุ |
|---|---|---|
| Model ชั้นลงขาย | `product_listing.go` | `ProductListing{Title,Description,Condition,Preorder*,PurchaseLimit*,Wholesale[],SizeChart*,Tiers[]}`; `ListingPrice` = Decimal128 ที่รับ JSON number/string ตอบ string; ทุกฟิลด์ใหม่ `omitempty` เพื่อไม่ให้ save เดิมลบทิ้ง |
| Model เพิ่ม | `product.go`, `product_barcode.go` | `imageurithumb`, `images[].urithumb`, `videos[].durationsec/sizebytes` (product เท่านั้น), `productbarcode.listing{tierindex []int, gtin, isforsale}` |
| Endpoint | `product_v2_get.go`, `product_v2_update.go` | `POST /goapi/product/v2/item/get`, `POST /goapi/product/v2/item/update-listing` (ลงทะเบียน `bootstrap.go:417-418`) |
| Endpoint | `product_v2_readiness.go` | `POST /goapi/product/v2/item/readiness` (§5.3, bootstrap บรรทัดถัดจาก update-listing) — ฟังก์ชัน readiness กลางอยู่ไฟล์นี้ (`productV2ListingReadiness` ถูกย้ายมาจาก update); กลุ่ม `channel` ถูก omit จาก response เมื่อไม่ส่ง channel/shopid |
| Endpoint | `product_v2_tier.go`, `product_v2_tier_validate.go` | `POST /goapi/product/v2/tier/init`, `POST /goapi/product/v2/tier/update` (bootstrap.go:420-421) — รับสถานะสุดท้ายทั้งชุด (tiers+models), เขียน `products.listing.tiers` + `productbarcodes` ใน transaction เดียว (standalone Mongo → เขียนเรียงลำดับพร้อม log เตือน), บาร์โค้ดที่หลุดชุดผสมถูกปิดขายไม่ถูกลบ |
| Whitelist | `product_v2_whitelist.go` | รายชื่อฟิลด์บัญชีจริงจาก product.go + audit; key อื่นที่ไม่รู้จัก → `UNKNOWN_FIELD`; `packageweight` ระดับบน → `INVALID_FIELD` (ต้องส่งผ่าน `package{}`) |
| Validators | `product_v2_validate.go` | E01 E02 E05 E07 E08 E20 E24 + URI/รูป/วิดีโอ/package bounds; ค่า default ใน `productV2DefaultLimits` (`product_v2_types.go`) |
| Permission | `product_v2_permission.go` | key `/product/listing` (const เดียว); โหลด shopusers → role_permission union; `*` หรือ ADMIN/OWNER ไม่มี record ของตัวเอง → ผ่าน; USER ต้องมี key จริง |
| Types/error | `product_v2_types.go` | envelope `{success,data}` / `{success:false,code,message,fields[]}`; body limit 1 MiB → 413 |
| Tests | `product_v2_*_test.go`, `product_listing_test.go`, `product_barcode_listing_test.go` | ~43 test (whitelist ทุกฟิลด์บัญชี, $set builder, validators, __v conflict, permission, JSON/BSON ของราคา) |

พฤติกรรมสำคัญของ `update-listing`
- decode body เป็น map ก่อน (ตรวจ key ต้องห้าม) แล้วค่อย decode typed ด้วย `DisallowUnknownFields`
- `listing` เขียนเป็น sub-document ทั้งก้อนที่ merge กับค่าเดิม (ไม่ dotted ใต้ `listing` เพื่อเลี่ยง "cannot create field in null parent"); `listing.tiers` แก้ผ่าน endpoint นี้ไม่ได้ (→ 422 READONLY รอ `tier/init`)
- filter `{holdingcode, businesscode, code, deletedat:nil, __v}` (`__v`=0 ยอมรับ doc ที่ไม่มี `__v`); matched 0 → นับ doc แยกเป็น 404 vs 409
- `item/get` คืน `accounting{...readonly:true}`, `listing` (เติม preorder/purchaselimit/wholesale/tiers ว่างให้เสมอ, condition ว่าง → NEW), `package`, `media`, `models[]`, `__v`

พฤติกรรมสำคัญของ `item/readiness` (เพิ่ม 2026-09-03, UAT ผ่านครบ + restore แล้ว)
- `accounting` = เปิดใช้งาน + มีชื่อ + มีหน่วยนับ (ชั้นบัญชีแก้ผ่าน v2 ไม่ได้ ข้อความจึงชี้ให้ไปแก้ที่หน้าสินค้า)
- `listing` = กฎกลางเดิม (ชื่อ/กล่อง/รูป) + ความยาวชื่อ E01 (reuse `validateListingTitle`) + E09 เต็ม: ต้องมี model เปิดขาย ≥1 และถ้ามี tiers ต้องมี model ครบทุกชุดผสม (เช็ค tierindex ในช่วง + นับชุดที่ขาด)
- `channel` = ส่ง `channel`+`shopid` คู่กันเท่านั้น (ส่งตัวเดียว → 422); ร้านไม่เจอ → 404 SHOP_NOT_FOUND; ตรวจ isactive, auth.tokenexpiresat, channel_category_map (ตาม categorycode สินค้า), brandmandatory → channel_brand_map
- ชื่อ collection จริง (ยังไม่มีเอกสารใน appdb — `shop/*` เฟส 2 จะเป็นคนสร้าง): `channel_shops`, `channel_category_maps`, `channel_brand_maps` (const ใน `product_v2_types.go`); ยังไม่มี index — เฟส 2 ต้องสร้างตาม MongoModel
- readiness ใน response ของ `update-listing` ยังเป็น core-only เหมือนเดิม (ไม่โหลด models เพิ่ม) — ตั้งใจให้ behavior เดิมไม่เปลี่ยน

## 3. วิธี build / test / UAT (ต้องทำทุกครั้งก่อนบอกเสร็จ)

เครื่อง Windows นี้ `go build ./...` พังเพราะ confluent-kafka (CGO) → ใช้ Docker:

```bash
cd /d/bccode/backend && MSYS_NO_PATHCONV=1 docker run --rm -v D:/bccode/backend:/src -w /src -e GOFLAGS=-mod=mod golang:1.26 sh -c "apt-get update -qq >/dev/null 2>&1 && apt-get install -y -qq librdkafka-dev >/dev/null 2>&1; go build ./... && echo BUILD_OK && go vet ./internal/goapi/handlers/ ./internal/product/... && echo VET_OK && go test ./internal/goapi/handlers/ ./internal/product/product/models/ ./internal/product/productbarcode/models/"
```

UAT local (ผลล่าสุด: ผ่านทุก step — `item/*` กับ `P1-004`, `tier/init`+`tier/update` กับ `P1-008` เมื่อ 2026-09-03; ทั้งสองตัว restore กลับสภาพเดิมแล้ว):
1. `cd D:\bccode\backend && docker compose -f docker-compose.yml -f docker-compose.local.yml build mainapi && docker compose -f docker-compose.yml -f docker-compose.local.yml up -d --no-deps mainapi` (port 8888)
2. token: `POST /demo-login` (whitelist ใน `cmd/app/main.go:138`) → ได้ user demo แล้ว `POST /select-holding` body `{"holdingcode":"demo","businesscode":"C01"}` ด้วย token เดิม (token ไม่เปลี่ยน ค่าบริษัทผูกกับ session ฝั่ง server)
3. `curl -X POST http://localhost:8888/goapi/product/v2/item/get -H "Authorization: Bearer <token>" -d '{"itemcode":"P1-004"}'` (ส่งไทยผ่านไฟล์ UTF-8 ไม่ใช่ inline ใน Git Bash)
4. หลังทุก write ตรวจ `mongosh` ใน container `mongodb` db `appdb` collection `products` ทันที: เปลี่ยนเฉพาะ path ที่ส่ง, `__v` +1, ฟิลด์บัญชีเดิม, ไม่มี doc ใหม่ · ทดสอบ negative 422/409/E01/PACKAGE_INCOMPLETE แล้ว DB ต้องไม่เปลี่ยน · จบแล้ว restore doc ด้วย `updateOne` ระบุ `holdingcode+businesscode+code` (ห้าม regex กว้าง)
- gofmt: ไฟล์เดิมบางไฟล์เป็น CRLF (`product.go`, `bootstrap.go`) — แก้ด้วย patch แบบรักษา EOL อย่า `gofmt -w` ทั้งไฟล์; ไฟล์ใหม่ทั้งหมดเป็น LF
- ถ้า shell หา `docker`/`docker-compose` ไม่เจอ ให้ `export PATH="$PATH:/c/Program Files/Docker/Docker/resources/bin:/c/Program Files/Docker/Docker/resources/cli-plugins"` ก่อน (ใน shell นี้ใช้ `docker-compose` ไม่ใช่ `docker compose`)

## 4. งานถัดไป (เรียงตามลำดับที่ควรทำ)

> ผู้รับต่อ: zcode (ใช้ GLM เป็นผู้ช่วยคิดหลัก) — สำรวจ+ออกแบบ tier/* ไว้แล้วใน §4.1 ไม่ต้องเริ่มจากศูนย์
> ลุงจืดอนุมัติให้ใช้ Gemini 3.8 Flash เป็นที่ปรึกษาได้ด้วย: `PYTHONIOENCODING=utf-8 python ~/.claude/tools/openrouter-ask.py --model google/gemini-3.8-flash "คำถาม"` (key อยู่ที่ `~/.claude/.openrouter-key`; เครื่องนี้ไม่มี `py` launcher ให้ใช้ `python`)

1. ~~**`tier/init` + `tier/update`**~~ — **เสร็จแล้ว 2026-09-03** (โค้ด+เทส+UAT ผ่าน) ดูสรุปการตัดสินใจใน §4.1
2. **`model/list|update|delete`** (§4.3, E05 E06 E13 E14) — `model/delete` = `listing.isforsale=false` ห้ามลบ productbarcode
3. **legacy save ของหน้าสินค้าเดิม** (`internal/product/product/services/product_http_service.go:361` + persister full `$set`) — ต้องไม่ลบ `images[].urithumb`/`videos[].durationsec` และควร `$inc __v` เมื่อบันทึก (ตอนนี้แก้บัญชีระหว่าง get→update ของ v2 จะตรวจไม่เจอ) — R1 บอกลุงจืดก่อนแตะ service เดิม
4. **สิทธิ์ (G9)** — role_permission จริงใช้ menu id ไม่มี `/` (เช่น `product`); key `/product/listing` ยังไม่มีใครถือ → รอลุงจืดตั้งชื่อ แล้ว seed ใน `scripts/seed-demo.mjs` + หน้า permission set; เปลี่ยนที่ const `productV2PermissionKey` ที่เดียว
5. `media/upload|attach|detach` (§5.8) ต่อยอด `handlers/image_r2.go` + `image_thumbnail.go` (thumb อัตโนมัติมีอยู่แล้ว)
6. เฟส 2-3 (shop/listing/sync/lookup) ต้อง OAuth ต่อช่องทางจริง = R0 ขออนุมัติลุงจืดต่อช่องทาง

### 4.1 บันทึกออกแบบ `tier/init` + `tier/update` (**เขียนเสร็จแล้ว 2026-09-03** — ส่วนล่างคือบันทึกสำรวจเดิม เก็บไว้เป็นที่มาของการตัดสินใจ)

**สิ่งที่ส่งมอบจริง (ต่างจากบันทึกสำรวจตรงไหน):**

- **ตัดสินแล้ว: ทั้ง `init` และ `update` รับ "สถานะสุดท้ายทั้งชุด"** — ผู้เรียกส่ง `tiers[]` ใหม่ทั้งชุด พร้อม `models[]` ที่ระบุ `tierindex` คู่กับ `barcode` (หรือ `generate:true`) ทุกชุดผสม เซิร์ฟเวอร์จึง**ไม่ต้อง remap ด้วยชื่อ option** → ปัญหา "rename แล้ว mapping หลุด" หายไป และ**ไม่ต้องเพิ่ม option id ใน data model** (ปิดคำถามข้อสุดท้ายของ §5)
- ชุดผสมที่หายไปจากคำขอ (เช่น ลดจำนวนตัวเลือก) → บาร์โค้ดนั้นถูก `$set listing.tierindex=[] , listing.isforsale=false` ไม่ถูกลบ (มีประวัติสต๊อก)
- `tier/init` ต้องยังไม่มี `listing.tiers` (มีแล้ว → 409 `TIER_ALREADY_EXISTS`) · `tier/update` ต้องมีอยู่แล้ว (ไม่มี → 409 `TIER_NOT_INITIALIZED`)
- ตรวจ E17 ก่อนเขียนทุกครั้ง: มีรายการบนช่องทางที่ `sync.state:"published"` + `platform.haspromotion:true` → 409 `LISTING_LOCKED_BY_PROMOTION`
- สิทธิ์: ทุกคำขอต้องผ่าน `productV2PermissionKey`; ถ้ามี `generate:true` ต้องผ่าน `productV2CheckAnyPermission(ctx, userInfo, productV2ProductPermissionKeys)` (`{"/product","product"}` เผื่อ G9 ยังไม่สรุปชื่อ key)
- transaction: `productV2RunAtomic` ใช้ `session.WithTransaction`; ถ้าเซิร์ฟเวอร์เป็น standalone (ข้อความ "Transaction numbers are only allowed" ฯลฯ) จะ fallback เขียนแบบเรียงลำดับพร้อม `logger.Warn` — ไม่ใช่ล้มทั้งคำขอ
- บาร์โค้ดที่สร้างใหม่: prefix `200` + สุ่ม 9 หลัก + check digit, สุ่มไม่เกิน 5 รอบ (`crypto/rand`), เช็คซ้ำในบริษัทก่อน แล้วยัง insert ผ่าน unique index; ชนซ้ำ → 409 `BARCODE_GENERATE_FAILED` ให้ลองใหม่
- ยังคง**ไม่ copy `prices`** จากบาร์โค้ดหลัก (คำถามใน §5 ยังเปิด)
- เทส: `product_v2_tier_test.go` 13 เคส (validator ทุกกฎ, write plan, check digit เทียบเลขมาตรฐาน, payload scan, transaction detect) — ผ่านทั้งหมดใน Docker
- UAT ที่ทำจริงกับ `P1-008` (demo/C01): init 2 ตัวเลือก (ใช้บาร์โค้ดเดิม 1 + สร้างใหม่ 1) → update เป็น 2 ชั้น 4 ชุดผสม (สร้างเพิ่ม 2) → update ย่อกลับเหลือชั้นเดียว (2 ตัวที่หลุด `isforsale=false`) → negative: init ซ้ำ 409, `__v` ผิด 409, สินค้าไม่มี 404, ฟิลด์บัญชี 422, 3 ชั้น 422, models ไม่ครบ 422, บาร์โค้ดสินค้าอื่น 422, รูปไม่ครบ 422 (DB ไม่ขยับทุกเคส) → ลบบาร์โค้ดที่สร้าง 3 ตัวและ `$unset listing/__v` คืนสภาพเดิม (ยืนยันด้วย mongosh ทีละ step)


**ข้อเท็จจริงจากระบบ (verify แล้วกับ source/Mongo จริง):**
- รูปแบบบาร์โค้ด generate (`barcode-form.tsx:345-375` + `lib/product-barcode/utils.ts:465`): prefix `"200"` (ค่า default ของหน้าจอ) + random 9 หลัก (`digit % 10`) + EAN-13 check digit (น้ำหนัก 1,3 สลับ จากซ้าย) = 13 หลัก; frontend สุ่มสูงสุด 5 รอบแล้วเช็คซ้ำผ่าน listBarcodes
- `guidfixed` ของ productbarcode เป็น xid 20 ตัว — สร้างด้วย `xid.New()` (มีอยู่แล้วใน `internal/utils/random.go:55`)
- unique index จริงบน `productbarcodes`: `(holdingcode,businesscode,barcode)` unique → เลขชนกัน = duplicate key error ให้ตอบ 409 ให้ลองใหม่
- doc barcode จริง: `standvalue/dividevalue = Long(1)`, ไม่มี `ismainbarcode` บน doc เก่า (generate ต้องเซ็ต `false` ตาม contract), `condition:false`, `itemunitnames` มี
- E17: ดู collection `channel_listing` (MongoModel efaf857d): filter `{holdingcode,businesscode,itemcode,isdeleted:false, sync.state:"published", platform.haspromotion:true}` → 409 LISTING_LOCKED_BY_PROMOTION; ชื่อ collection จริงตั้งเป็น `channel_listings` ให้ตรง `channel_shops`/`channel_category_maps`/`channel_brand_maps` (const ใน `product_v2_types.go`)
- `GetAtlasConnection()` คืน `(*mongo.Client, *mongo.Database)` ใช้ `client.StartSession()` + `session.WithTransaction` ได้เลย (codebase ยังไม่เคยใช้ transaction — นี่จะเป็นที่แรก)
- readiness มี helper ให้ reuse: `productV2TierIndexInRange`, `productV2TierIndexKey`, `productV2ModelsReadiness` (`product_v2_readiness.go`)

**ดีไซน์ที่ตกลงไว้ (ผ่านรีวิว Gemini 3.8 Flash แล้ว):**
1. validate ทั้งหมดเป็น pure function ก่อนเข้า transaction: E03 (1–2 ชั้น, server กำหนด xorder ตามตำแหน่ง array เพราะ request §5.4 ไม่ส่ง xorder), E04 (ชุดผสม ≤50, ชื่อ option ไม่ซ้ำในชั้นหลัง TrimSpace, ยาวไม่เกิน default — เพิ่ม default tiername/tieroption max ใน `productV2DefaultLimits` แล้ว note ว่าเป็นค่าของเราเอง), E09 (models ครบทุกชุดผสมพอดี), E10 (barcode ที่ส่งมาต้องมีจริง + itemcode ตรง — query เดียวด้วย `$in`), E11 (tierindex ยาวเท่าจำนวนชั้น ในช่วง ไม่ซ้ำ), E12 (รูปเฉพาะชั้นแรก ครบทุกตัวหรือไม่ใส่เลย + ทุกรูปต้องมี `imageurithumb` ตามกฎ S3/thumb)
2. generate: สุ่ม candidate + เช็คซ้ำก่อน tx (5 รอบตาม frontend) แล้ว insert ใน tx — ถ้าชน unique index ใน tx → abort → 409 ให้ลองใหม่
3. ใน tx ห้ามมี side effect นอก memory (callback ของ WithTransaction ถูก retry ซ้ำได้) และทุก call ต้องใช้ `sc` (SessionContext) ไม่ใช่ parent ctx; รวม write ของ barcodes เป็น `BulkWrite` 1 roundtrip (Gemini เตือน N-roundtrips ใน tx เสี่ยง timeout)
4. update product: filter `{holdingcode,businesscode,code,deletedat:nil,__v}` + `$set listing.tiers` (ทั้งก้อน เหตุผลเดียวกับ update-listing) + `$inc __v`; matched 0 → abort → แยก 404/409
5. barcode เดิมที่จับคู่: `$set listing.tierindex + listing.isforsale=true` (สมมติ: เจ้าของจัดตัวเลือกเพื่อขาย — ถ้าไม่ตั้ง readiness E09 จะไม่ผ่านทันทีหลัง init) · barcode ใหม่: `ismainbarcode:false`, `itemunitcode` ตาม request, `names` = product.names แต่ละภาษา + " " + ชื่อ option ตาม tierindex, `standvalue/dividevalue=1`, guidfixed=xid, **ไม่ copy prices** (contract ไม่สั่ง — flag ให้ลุงจืดใน §5)
6. tier/update (contract ไม่มี §5.x — ต้องยืนยัน semantics กับลุงจืด): init ห้ามมี tiers เดิม / update ต้องมี tiers เดิม; remap tierindex ของ model เดิมด้วยชื่อ option (TrimSpace ทุกจุด); option ที่หายไปแต่มี model อยู่ใน channel_listing published → 422 ห้ามลบ; model ที่หลุดชุดผสม (ไม่ได้เผยแพร่) → `$set listing.isforsale=false` ไม่ลบ barcode · Gemini เตือน: rename option (แก้คำผิด) จะหลุด mapping ถ้า match ด้วยชื่อ — ทางแก้ระดับ spec คือส่ง id ประจำ option (ยกไปถามลุงจืด อย่าเดา)
7. สิทธิ์: ทุก tier/* ต้องผ่าน `productV2PermissionKey`; ถ้ามี `generate:true` ต้องผ่าน `productV2CheckPermission(ctx, userInfo, "product")` เพิ่ม (menu id จริงไม่มี slash ตาม G9 note)
8. UAT: ใช้ flow เดิม (§3) + เพิ่มขั้นตรวจ `productbarcodes` ทีละ step (doc ใหม่ครบ field, tierindex ตรง, isforsale) และทดสอบ rollback โดยยิง payload ที่ชน unique/ผิด __v แล้วดูว่า product.__v ไม่ขยับ

## 5. ข้อสมมติ / คำถามที่ยังเปิด (ห้ามเดา ถ้ากระทบให้ถามลุงจืด)

- G5: 1 ร้าน = 1 บริษัท (businesscode) — สมมติไว้ · G7: สินค้าหลายหน่วย (ขวด/แพ็ก/ลัง) = model หลายตัวใน item เดียว — สมมติไว้ · G6: สต๊อกที่ส่งยังไม่หักยอดจอง SO
- G3: credential ร้านเก็บที่ไหน (ยังไม่ทำ) · G2: deprecate `marketplaceproducts[]`/`marketplaceskumappings`/`sellersku`/`skupackage*` ใน Go + ลบ `frontend/src/app/menu/tab-product-marketplace.tsx` (รอยืนยัน)
- ค่า default เมื่อไม่มี `limits` ของร้าน: title 5–120, listing.description 0–2000, images ≤9, videos ≤1, wholesale ≤10, package ≤1000 kg/cm, daystoship ≥1 — ค่าของเราเอง ต้องยืนยันกับร้านจริงตอนเฟส 2
- products ใน appdb local ~40% ไม่มี `businesscode` → v2 มองไม่เห็น (พฤติกรรมเดียวกับ `FindByCodeInCompany` เดิม; ข้อมูล disposable ไม่ migrate)
- `unitprice` ทศนิยมยาวเก็บตามที่ส่ง (ยังไม่ปัด 2 ตำแหน่ง) — E15 จะบังคับใน `price/update`
- tier/*: barcode ที่ generate ไม่ copy `prices` จากบาร์โค้ดหลัก (contract §5.4 ไม่ได้สั่ง) — ถ้าอยากให้มีราคาทันทีต้องยืนยันกับลุงจืด
- ~~tier/update remap ด้วยชื่อ option~~ — **ปิดแล้ว 2026-09-03**: ทั้งสอง endpoint รับสถานะสุดท้ายทั้งชุด (ผู้เรียกส่ง barcode คู่ tierindex เอง) จึงไม่มีการ remap และไม่ต้องมี option id

## 6. Data model อ้างอิง

MongoModel MCP project "BC Ai Account" diagram "ข้อมูลหลัก" (id `efaf857d`, rev 3437): `product`, `productbarcode`, `channel_shop`, `channel_listing`, `channel_category_map`, `channel_brand_map` — lint 0 issue, คำอธิบายไทยครบ; เวลาเพิ่มฟิลด์ใน Go ให้ตรงกับไดอะแกรม (ไดอะแกรมเป็น SoT) · ADR: `D:\obsidian-vault\bc-dev\decisions\2026-09-03-product-two-layer-marketplace-model.md`
