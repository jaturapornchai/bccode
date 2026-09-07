---
date: 2026-09-03
status: proposed
tags: [bc-account, architecture, product, plan, workflow-output]
---

# แผน (proposed) สินค้า 2 กลุ่มผู้ใช้ — ผลจาก plan workflow wf_037d9857-748

> เอกสารนี้เป็น output ของ workflow (11 agents) ยังไม่ผ่านการตัดสินใจของลุงจืด; Claude verify แล้วเฉพาะ ม.86/4 และ ม.87 (rd.go.th 5208/5209) ข้ออ้างกฎหมาย/บัญชีอื่นยังต้อง re-verify กับ primary source ก่อนใช้ (ดูหัวข้อ critique) และ**ห้ามใช้ชื่อ vendor ใน code/field/docs ของ repo** (แผนนี้เอ่ยชื่อเพื่อการวิจัยเท่านั้น)
> คะแนน judge: {"1. บัญชีเป็นแกน–ขายออนไลน์เป็นชั้นเสริม (Ledger-Core / Listing-Overlay)":7,"2. SKU-Atom — บาร์โค้ด = หน่วยขาย (SKU), สินค้า = แม่แบบบัญชี":6,"3. Progressive Profile — สินค้า 1 เรคคอร์ด 3 ระดับความพร้อม + productlisting":7.5,"A — บัญชีเป็นแกน–ขายออนไลน์เป็นชั้นเสริม (Ledger-Core / Listing-Overlay)":6,"B — SKU-Atom (บาร์โค้ด = หน่วยขาย, สินค้า = แม่แบบบัญชี)":7,"C — Progressive Profile (L1→L2→L3) + productlisting เป็นเรคคอร์ดลูก":8,"บัญชีเป็นแกน–ขายออนไลน์เป็นชั้นเสริม (Ledger-Core / Listing-Overlay)":7.5,"SKU-Atom — บาร์โค้ด = หน่วยขาย (SKU), สินค้า = แม่แบบบัญชี":5.5,"Progressive Profile — 3 ระดับความพร้อม + productlisting เป็นเรคคอร์ดลูก":7}

# แผนสุดท้าย: สินค้า/บริการ BC Ai Account
ฐาน = Progressive Profile (คะแนนรวม 22.5 สูงสุด) + graft จาก Ledger-Core (20.5) และ SKU-Atom (18.5)

## 1. หลักการออกแบบ
1. **ระบบเดียว doc เดียว** — `product` 1 doc/บริษัท (holdingcode+businesscode+code) เป็นความจริงเดียว ไม่มี collection คู่ขนานของข้อมูลบัญชี
2. **2 ชั้นข้อมูล** — ชั้นบัญชี (L1, 6 ช่อง, ออกเอกสารได้ทันที) และชั้นเสริม (L2 ขาย-คลัง, L3 ตลาด) — ชั้นเสริมไม่ครบ**ไม่บล็อก**เอกสาร แค่แสดงป้าย "ยังขาด"
3. **เจ้าของ field ชัด** — ภาษี/ประเภท/หน่วย/กลุ่ม/WHT อยู่ที่ `product` เท่านั้น เป็นของนักบัญชี; SKU **ห้าม override** (ภาษีต่างกัน = คนละสินค้า); รูป/ตัวเลือก/น้ำหนัก/listing เป็นของเจ้าของ
4. **variant = แถว `productbarcode`** (1 แถว = 1 หน่วยขาย/SKU); **barcode ยังเป็น stock key** (ProductBarcodePg PK `product_barcode.go:373-376`, stockprocess by barcode) — ไม่ re-key ledger; เพิ่ม itemcode+dimensionkey เป็น attribute ไว้ roll-up
5. **เขียนผ่าน endpoint แยกชั้น** แบบ whitelist (pattern CoreOnly `product_barcode_request.go:38-54`) + `$set` เฉพาะ field + revision check; listing/sync state อยู่นอก product doc (`productlisting`); publish ผ่าน outbox

## 2. โมเดลข้อมูล
| collection | field group | ชั้น | บังคับ | เจ้าของ |
|---|---|---|---|---|
| product | code (A-Z0-9- ≤35), names[] (≤256/ภาษา), itemtype, unitcode+unitconversions[], vattype (0 คิด/1 ยกเว้น), groupcode(→GL), **whttype ใหม่** | บัญชี | ใช่ (whttype เฉพาะบริการ) | นักบัญชี |
| product | materialtype, orderpoint/min/max, suppliers/manufacturers[], bom[] | ขาย-คลัง | ไม่ | นักบัญชี/คลัง |
| product | description (≤10,000), images[]{xorder,uri,thumburi,width,height,bytes}, videos, brand/model/… triplets, packageweight/l/w/h, **variantaxes[] ใหม่** (≤3 แกน, option มี code/image/hex), condition, preorderdays | ตลาด | ไม่ | เจ้าของ |
| product | **defaultbarcode, isdisabled+reason ใหม่** | ระบบ | — | backend |
| productbarcode | itemcode (FK picker), barcode, names, itemunitcode, prices[]{keynumber,price}, images | บัญชี/ขาย | ใช่ | นักบัญชี (keynumber 1) |
| productbarcode | **variant{axisvalues,dimensionkey}, sellersku (^[A-Z0-9_-]{1,50}$ unique/บริษัท), gtin+gtinstatus HAS/NONE, skupackage* override, channelprices[]{channelcode,price,promoprice,from,to}, isdefault, isdisabled** | ตลาด | ไม่ | เจ้าของ |
| units | + **unececode** (UNECE Rec 20), isservicedefault; ตัด unitname1..4 (`unit.go:13`); companyguids→businesscodes | บัญชี | unececode เฉพาะ e-Tax | นักบัญชี |
| salechannel | + channeltype, accountid, warehousemap[]{whcode,channellocationid}; price→keynumber ต้องถูกอ่านจริง (`salechannel.go:16` ไม่มีใครอ่าน) | ตลาด | — | เจ้าของ |
| **productlisting ใหม่** (ลูก: product×channel×account) | title/description override, categoryid, brandid, attributes[], images (อ้าง image id), status DRAFT→READY→PENDING→LIVE/REJECTED/UNLIST/DELETED, validation{errors_th}, skus[]{barcode,dimensionkey,marketmodelid,shopsku,customprice,stocks[]{whcode,sellableqty},sync*} — ย้าย MarketplaceProductMap/SKUMap (`product_barcode.go:450-535`) มาใช้ | ตลาด | — | เจ้าของ |
| organizationcompanies | + vatregistered, legaltype, costmethod FIFO/AVG, priceincludesvat, default units (ปัจจุบันมีแค่ TaxID `company.go:17`) + ตาราง vatrate มีวันมีผล (system) | นโยบาย | — | นักบัญชี |

**ตัด/ย้ายจาก product** (`product.go`): dimensions[] :22 → variantaxes; refbarcodes :90 + isusesubbarcodes :89 → ลบ; manufacturerguid/code/names :19-21 → ลบ (ซ้ำ manufacturers[]); qty :103 → ลบ; marketplaceproducts :98 → productlisting; ฟิลด์ร้านอาหาร → ซ่อนตาม business type ไม่แตะ
**ตัดจาก productbarcode** (ตายอยู่แล้วเพราะ CoreOnly ทิ้ง): classification 10 triplets :18-50, itemtype/vattype/taxtype/materialtype :81-84, marketplaceproducts :79, refbarcodes/bom :183-185, dividevalue/standvalue :59,70-71 (ratio อยู่ที่ product.unitconversions) — ทำ Phase 4
**ราคา**: อยู่ที่ productbarcode เท่านั้น (product.go:24 bson:"-"); **รูป**: ไฟล์ใน MinIO + thumb, Mongo เก็บแค่ uri/thumburi/มิติ
**ตัดทิ้ง**: atlas productcolors/productsizes/productvariantmatrices/productchannelprices + Go color/option/optionpattern (ไม่มี consumer `mongodb_atlas.go:188-197`) → ใช้ `dimension` module เป็น "ชุดตัวเลือกมาตรฐาน" (คง aliases/hex ตอนย้าย)

## 3. หน้าจอ 2 กลุ่ม
**นักบัญชี**
- **"เพิ่มสินค้า/บริการ" 1 ฟอร์ม 1 save**: ชื่อ* | ประเภท* radio (สินค้า/บริการ/ไม่ตัดสต๊อก) | หน่วยนับ* picker + "เพิ่มหน่วยใหม่" inline (ชื่อ/รหัส UNECE) | VAT* (คิด VAT/ยกเว้น; default บริษัท; ซ่อนทั้งช่องถ้าไม่จด) | กลุ่ม (default "ทั่วไป") | ราคาขาย (→prices[1] ของ defaultbarcode) | หัก ณ ที่จ่าย (โชว์เฉพาะบริการ) | รหัส auto (แก้ได้ก่อนมีเอกสาร) | บาร์โค้ด auto (= รหัส หรือปุ่ม EAN-13) | section พับ "หน่วยซื้อ/ขายเพิ่ม" (โหล/ลัง) — CTA เดียว "บันทึกและใช้งานได้ทันที"; backend `POST /product/quick` สร้าง product + productbarcode(isdefault) ใน Mongo txn (rs0 verified `docker-compose.local.yml:13`)
- **จอเอกสารขาย** `/transaction/sale-invoice` — **ยังไม่มีใน Next.js** (เมนูชี้แต่ไม่มีโฟลเดอร์) → สร้างขั้นต่ำ: ลูกค้า, ตาราง picker (ชื่อ/รหัส/สแกน → เติม barcode/itemcode/unit/price/vattype ลง Detail `transaction.go:305-341`), "+ สินค้าใหม่" = modal ฟอร์มเดียวกัน, "บันทึกและพิมพ์"
- `/productbarcode`: itemcode → product picker (แทน free-text `barcode-form.tsx:455-470`), เพิ่มช่องราคา, ลบ orphan picker (`product-screen.tsx:470-484`)

**เจ้าของ** — จอสินค้าเดียวกัน section "ข้อมูลเสริมสินค้า" (แทน 11 tab `product-screen.tsx:322-336`): รูป (MinIO ≤9, 1:1, ≥600px, เตือนไทยต่อรูป) | รายละเอียดขาย/ยี่ห้อ ("ไม่มียี่ห้อ" ได้) | น้ำหนัก-ขนาด | ตัวเลือก: เลือกแกน ≤3 → "สร้างรายการย่อย" gen แถว productbarcode (sellersku ตาม skucoderule, GTIN/ไม่มี, ราคา-โปรต่อช่องทาง, รูปต่อค่าแกนแรก) | ป้าย "ยังขาด: …" คลิก focus; จอ **"ลงขายออนไลน์"** (`/productlisting`): checklist ไทยต่อช่องทาง, bulk หลายสินค้า, export Excel mass-upload (Phase 3 ก่อน API)

**เมนู** 5 กลุ่ม → 3: **สินค้า (บัญชี)** [สินค้า/บริการ, หน่วยนับ, กลุ่มสินค้า, บาร์โค้ด/ป้ายราคา, สินค้าชุด] | **คลังและการผลิต** [คลัง, BOM] | **ขายออนไลน์ (เจ้าของร้าน)** [ลงขายออนไลน์, ชุดตัวเลือกมาตรฐาน, รายละเอียดประกอบสินค้า (ยุบ 10 จอเป็น 1), ช่องทางขาย, ร้านค้าออนไลน์]; ลบเมนู สี/ไซซ์/ตารางตัวเลือก/ราคาตามช่องทาง

## 4. สิทธิ์
- key ใหม่ในโมเดลเดิม (`role_permission.go:25-40`): `product` (L1), `product-sale-stock` (L2), `product-online` (L3), `product-listing`
- **ชุด "บัญชี"**: product:*, product-sale-stock:*, product-unit:*, product-group:*, barcode:*, sale:*, product-online (อ่าน) — **ชุด "ขายออนไลน์"**: product (อ่าน), product-online:*, product-listing:*, barcode:update (field ตลาดเท่านั้น), product-dimension:*, sale-channel:*
- **backend** (ปัจจุบัน `/product` ไม่มี middleware `product_http.go:64-69`): middleware อ่าน /me + endpoint แยก `PUT /product/:id` (L1) / `/sale-stock` / `/online` / `/product/barcode/:id/online` — whitelist ตัด field นอกชั้นทิ้ง, `$set` เฉพาะ field (repo Update เป็น full-doc `product_mongo_repository.go:152` → overwrite ข้ามคน), revision precondition, 403 ข้อความไทยระบุช่อง, contract test ต่อ whitelist; ล็อก code/unitcode/itemtype หลังมีเอกสาร (ADMIN+เหตุผล, audit)

## 5. กฎหมาย/บัญชีไทย
| ข้อกำหนด | แหล่ง (verified) | field |
|---|---|---|
| ชื่อ ชนิด ประเภท ปริมาณ มูลค่า VAT ต่อรายการ; อย่างย่อต้องราคารวม VAT | ม.86/4, 86/6 https://www.rd.go.th/5208.html | names ≥1, ชื่อพิมพ์ = names+ตัวเลือก+กลุ่ม, Detail qty/price/vattype, priceincludesvat |
| ชื่ออังกฤษได้ | ป.85/2542 https://www.rd.go.th/3569.html | names[] หลายภาษา |
| รหัสแทนชื่อเฉพาะเครื่องบันทึกเงินสด + ตารางรหัสทั้งระบบ | ประกาศ VAT 73 https://www.rd.go.th/3389.html | code ล็อก, isdisabled แทนลบ, รายงานตารางรหัส |
| รายงานสินค้ารายสถานประกอบการ, 3 วันทำการ, เก็บ 5 ปี, เฉพาะผู้ขายสินค้า | ม.87, 87/3 https://www.rd.go.th/5209.html https://www.rd.go.th/5205.html | itemtype, ledger มี branch, ห้ามลบ movement |
| ลงตามปริมาณจริง+เลขที่ใบสำคัญ; ค้าปลีกลงเป็นกลุ่มได้ | ประกาศ 89 ข้อ 9 https://www.rd.go.th/3374.html | movement.docno; ม.87 group by itemcode |
| บุคคลธรรมดาตรวจนับ 30 มิ.ย./31 ธ.ค. | ประกาศ 104 https://www.rd.go.th/3289.html | company.legaltype, จอตรวจนับ |
| บัญชีสินค้ามี ชื่อ ชนิด จำนวน หน่วยนับ; เอกสารมีหน่วย/ราคาต่อหน่วย/รวม | พ.ร.บ.บัญชี 2543 https://www.dbd.go.th/data-storage/attachment/cf553ff61a275899be24bbc4.pdf ; ประกาศ 2544 ข้อ 6(7),10 https://www.dip.go.th/uploadcontent/boom_LAW/GNB_022.pdf | unitcode required, itemunitcode |
| VAT ปกติ/0%/ยกเว้น | ม.80, 80/1, 81 https://www.rd.go.th/5206.html | vattype 0/1 + 0% override ที่เอกสาร |
| 7% ถึง 30 ก.ย. 2570 | พ.ร.ฎ.807 https://www.rd.go.th/21221.html | ตาราง vatrate มีวันมีผล (ไม่อยู่บน item) |
| WHT บริการ 3%/เช่า 5%/ขนส่ง 1%; ภ.ง.ด.3/53 ตามผู้รับ | ท.ป.4/2528 https://www.rd.go.th/3479.html ; https://www.rd.go.th/fileadmin/download/insight_pasi/wht_3_53_030260.pdf | whttype (บริการ); แบบ 3/53 จาก legaltype คู่ค้า |
| e-Tax: Name ≤256 บังคับ, ID ≤35, BilledQuantity บังคับ, unitCode Rec20, TypeCode VAT/FRE | ETDA ขมธอ.3-2560 v2.0 https://www.etda.or.th/getattachment/43f4a6d7-946e-4fc3-b5d4-e3c64f9d197a/20250515_ETDA-Rec-3-2560_English-V03.pdf.aspx ; XSD https://github.com/ETDA/XMLValidation | code regex, units.unececode, gtin→GlobalID |
| ต้นทุน FIFO/ถัวเฉลี่ย (ห้าม LIFO), LCNRV, จัดประเภทสินค้าคงเหลือ | TAS 2 https://eservice.tfac.or.th/get_file/index.php?file=TAS_2_revised_2568.pdf ; NPAEs บทที่ 8 https://acpro-std.tfac.or.th/test_std/uploads/files/TFRS%20for%20NPAEs_Revise%202565R1.pdf | company.costmethod, materialtype, GL ที่ product group |

**Unverified (ห้ามเขียนลง docs จนกว่าจะเปิดต้นฉบับ)**: ประกาศ 89 ข้อ 10-14 (มาตรฐานซอฟต์แวร์/audit trail), ประกาศ VAT 39, ขมธอ.3 ฉบับไทย + code list 5153, ตารางหน่วยไทย→UNECE (B.14 xlsx), TAS 2 ฉบับ 2568 จริง (ไฟล์ระบุประกาศ 34/2562), e-Receipt 2565, LINE MyShop API ทั้งหมด, ค่า taxtype ปัจจุบัน

## 6. เฟส
**Phase 1 — นักบัญชี 1 ฟอร์ม** (checkpoint ก่อน)
งาน: POST /product/quick (txn + PublishOrOutbox `organization/events/events.go:33`), whttype, units.unececode, company settings, vatrate table, ฟอร์มด่วน + inline unit, จอ sale-invoice ขั้นต่ำ + proxy `/api/transaction`, barcode picker + ช่องราคา, middleware + endpoint L1, seed หน่วยไทย+Rec20
ไฟล์: `product_http_service.go`, `product_http.go`, `productbarcode_http_service.go`, `company.go`, `unit.go`, `product-screen.tsx`, `barcode-form.tsx`, `menu-data.ts`, `permission-actions.ts`, ใหม่ `frontend/src/app/transaction/sale-invoice/`
AC: Given บริษัทจด VAT / When กรอก ชื่อ+ประเภท+หน่วย แล้วบันทึก / Then appdb มี product 1 + productbarcode isdefault 1 ใน txn เดียว, ไม่มี orphan · Given สินค้า L1 / When เลือกในบรรทัดเอกสาร / Then Detail มี barcode/itemcode/unitcode/vattype/itemtype ครบ · Given user ชุด "ขายออนไลน์" / When PUT vattype / Then 403 ไทย ค่าเดิมคง
UAT: Create→Mongo→Update→Mongo→Delete(=isdisabled)→Mongo ทีละ step + ออกใบกำกับ 1 ใบ

**Phase 2 — ขาย-คลัง + variant**
งาน: section L2/L3, variantaxes + dimension master ขยาย, generator แถว productbarcode (sellersku/gtin/channelprices), รูป MinIO thumb+มิติ, endpoint /sale-stock /online whitelist, ป้าย "ยังขาด" (คำนวณตอนอ่าน), เอกสารโอนสต๊อก default→variant, แก้ barcode unique ระดับ holding → บริษัท (`productbarcode_mongo_repository.go:161-177`), Excel import variant, ยุบ 10 descriptor, ลบ atlas 4 + Go orphan
AC: Given สินค้า L1 มีสต๊อก 200 / When สร้างตัวเลือก สี×ไซซ์ 6 ชุด / Then ได้ productbarcode 6 แถว ครบ dimensionkey/sellersku unique, default barcode ยังไม่ปิดจนกว่าโอนสต๊อก · Given อัปโหลดรูป / Then MinIO มี object+thumb, Mongo เก็บแค่ uri/มิติ
UAT: CRUD SKU + Mongo + MinIO object ทีละ step; ตรวจว่าเอกสารเก่าไม่เปลี่ยน

**Phase 3 — listing/sync**
งาน: productlisting + state machine + validator config ต่อร้าน/หมวด (Shopee get_item_limit, Lazada GetCategoryAttributes, TikTok Get Attributes), export Excel mass-upload, salechannel.warehousemap, เพิ่ม businesscode/whcode ใน marketplacestockbalances (`goapi/inventory/database.go:117-129`), sellable = onhand − reserved − buffer, รายงาน ม.87/ตรวจนับ/ตารางรหัส/e-Tax XML; API sync จริง = R0 ขออนุมัติต่อช่องทาง
AC: Given ชื่อ 12 ตัว / When ตรวจก่อนลงขาย TikTok / Then error ไทย "ชื่อสั้นกว่า 25 ตัวอักษร" และ Shopee ผ่าน · Given LIVE / When แก้ sellersku / Then ถูกปฏิเสธ
UAT: CRUD listing + Mongo; worker เขียนเฉพาะ productlisting ไม่แตะ product

**Phase 4** — ตัด dead field productbarcode/Pg/ClickHouse/Kafka/import + MongoModel diagram + skill

## 7. คำถามที่ต้องตัดสิน
**ต้องตัดสินก่อนเริ่ม**
1. รหัสสินค้า auto? — เสนอ: auto prefix กลุ่ม+running, A-Z0-9- ≤35, ล็อกหลังมีเอกสาร (ADMIN+เหตุผล)
2. ราคารวม VAT flag ระดับไหน — เสนอ: ต่อ price level (keynumber) ไม่ใช่บริษัทอย่างเดียว
3. ขอบเขตจอเอกสาร Phase 1 = sale-invoice + พิมพ์ PDF อย่างเดียว? — เสนอ: ใช่
4. ราคาเป็นของใคร — เสนอ: keynumber 1 บัญชี, channelprices เจ้าของ
5. settings บริษัทวางที่ organizationcompanies? — ต้องชี้ collection
6. default barcode เมื่อสร้างตัวเลือก — เสนอ: ห้ามปิดถ้ามีสต๊อก ต้องออกเอกสารโอนก่อน

**ควรตัดสิน**: costmethod FIFO/ถัวเฉลี่ย (เสนอ ถัวเฉลี่ย moving, cost pool ต่อ itemcode) · GL mapping ที่ product group 5 บัญชี · ยุบ 10 descriptor เป็น 1 จอ · วิธี detect "มีเอกสารแล้ว" (flag จาก transaction consumer) · หน่วย default ชิ้น/งาน · Flutter/POS ยังอ่าน field ที่จะตัดไหม

**ปรับภายหลังได้**: คำไทย SKU/variant ("หน่วยขาย"/"ตัวเลือก") · ช่องทางแรก Phase 3 · description 10,000 · ชื่อชุดสิทธิ์ · ม.87 group ระดับ product/variant default

## 8. ความเสี่ยงและสิ่งที่ตัดทิ้ง
**ความเสี่ยง**: full-doc `$set` เขียนทับข้ามคน → endpoint แยก+revision; ไม่มี outbox บน product (dual-write `product_http.go:50`) → ใช้ PublishOrOutbox; ตัด field กระทบ Pg/ClickHouse/Kafka/import/types contract test → Phase 4 หลัง checkpoint; marketplace limit dynamic + LINE unverified → validator เป็น config; Phase 1 ใหญ่เพราะต้องสร้างจอเอกสาร → ล็อกขอบเขต
**ตัดทิ้งจาก design อื่น**: SKU-Atom re-key ledger เป็น itemcode+dimensionkey (rewrite stock calculator) · vatcategory 3 ค่า มี 0% บน item (0% เป็นเรื่องธุรกรรม) · diff-middleware บน PUT เดียว (ทดสอบยาก) · profile status persist แล้วใช้กรอง picker (drift → สินค้าหาย) · reuse taxtype เป็น WHT (ค่าเดิมไม่รู้) · marketplaceproducts ฝังใน product (worker ชนผู้ใช้) · promoprice ตัวเดียวต่อ SKU · atlas productcolors เป็นแหล่ง option · บังคับยี่ห้อใน L3

ข้อดี: นักบัญชีเหลือ 1 ฟอร์ม 1 save, ownership ชัด, ledger ไม่ต้อง rewrite | ข้อเสีย: Phase 1 ต้องสร้างจอเอกสารทั้งจอ + middleware ใหม่, productlisting เพิ่ม collection | คำแนะนำ: ตอบข้อ 7 กอง "ต้องตัดสินก่อน" แล้ว commit checkpoint ก่อนแตะ schema

## Critique (verify pass)
```json
{
"1_unbacked_claims": [
 {"claim":"คะแนน Progressive Profile 22.5 / Ledger-Core 20.5 / SKU-Atom 18.5","issue":"ไม่มี rubric/criteria/ผู้ให้คะแนนอ้างอิง — ตัวเลขลอย"},
 {"claim":"§1.4 'stockprocess by barcode' (ไม่ cite ไฟล์)","evidence_found":"backend/internal/stockprocess/models/stockprocess.go:15 Barcode, :17 WhCode — ไม่มี BusinessCode/BranchCode ใน struct → stock spine เป็น holding-scope (ยืนยันด้วย comment 'Stock still resolves barcode by Holding' productbarcode_mongo_repository.go:161-162) แผนไม่ระบุ dependency นี้"},
 {"claim":"§2 Phase2 'แก้ barcode unique ระดับ holding → บริษัท (productbarcode_mongo_repository.go:161-177)'","issue":"index นั้นชื่อ *_stock_guard และมี comment ว่าต้องคงไว้จน stock spine มี businesscode ทุก key/cache/projection — แผนตัดโดยไม่มีงาน 'เติม businesscode ลง stockprocess/Pg/ClickHouse' มาก่อน"},
 {"claim":"§3 'Mongo txn (rs0 verified docker-compose.local.yml:13)'","issue":"path จริง backend/docker-compose.local.yml:13 (verified); compose ของ .202/DO prod ไม่ได้ตรวจว่า --replSet"},
 {"claim":"§2 product.code regex 'A-Z0-9- ≤35' อ้าง ETDA","issue":"ETDA 8.20 ระบุ 'letters (A-Z) or numbers (0-9)' ไม่มีขีด (-) — ขีดเป็นของแผนเอง (barcode-form.tsx:420 อนุญาต - เฉพาะ barcode)"},
 {"claim":"§2 description ≤10,000","issue":"product.go:79 validate max=1500 ต้องแก้ — ไม่อยู่ในรายการ ตัด/ย้าย/แก้"},
 {"claim":"§2 'ฟิลด์ร้านอาหาร → ซ่อนตาม business type'","issue":"ไม่มี field business type ระดับบริษัทที่ cite (company.go:14-20 มีแค่ code/names/taxid/logouri/isactive)"},
 {"claim":"§3 'รหัส auto prefix กลุ่ม+running'","issue":"ไม่ cite running-number/sequence service ใน backend"},
 {"claim":"§3 'บาร์โค้ด auto = รหัส หรือปุ่ม EAN-13'","issue":"EAN-13 ที่ generate เอง (barcode-form.tsx:345-370 random 9 หลัก) ไม่ใช่ GS1-assigned; แผนไม่บอกว่าเป็น gtinstatus=HAS หรือ NONE"},
 {"claim":"§6 Phase2 'เอกสารโอนสต๊อก default→variant'","issue":"ไม่มี doc type สำหรับย้ายยอดข้าม barcode ของ item เดียวกัน (มีแค่ transaction/stocktransfer, stockadjustment ระดับคลัง) — ไม่ cite"},
 {"claim":"§6 Phase3 AC 'Shopee ผ่าน' ชื่อ 12 ตัว","issue":"get_item_limit เป็น sample (min 5) และ dynamic ต่อร้าน/หมวด — ไม่ใช่ค่า TH จริง"},
 {"claim":"§3 รูป '1:1, ≥600px'","issue":"มาจาก TikTok US policy (unverified TH); API TikTok ขั้นต่ำ 300px; Lazada ≤8 รูป ไม่ใช่ 9"},
 {"claim":"§2 sellersku '^[A-Z0-9_-]{1,50}$ unique/บริษัท'","issue":"verified แค่ TikTok ≤50 ไม่มีช่องว่าง; uppercase-only + uniqueness scope ของ Shopee/Lazada(‘same item’ vs ‘store’)/LINE unverified"},
 {"claim":"§2 units 'seed หน่วยไทย+Rec20'","issue":"ตาราง mapping ไทย→UNECE (B.14 xlsx) ยังไม่เปิด — seed จะเดา"},
 {"claim":"§7 'Flutter/POS ยังอ่าน field ที่จะตัดไหม'","issue":"ไม่มี Flutter code ใน repo นี้ (อยู่ D:\\bcdev) — ยังไม่ได้ grep"},
 {"claim":"§5 'ประกาศ 73 → isdisabled แทนลบ' และ 'ม.87 → ห้ามลบ movement'","issue":"เป็น inference ของแผน ไม่ใช่ข้อความในกฎหมาย (ประกาศ 89 ข้อ 10-14 เรื่อง audit trail ยังไม่อ่าน)"},
 {"claim":"§5 'ชื่อพิมพ์ = names+ตัวเลือก+กลุ่ม' = ชื่อ ชนิด ประเภท ตาม ม.86/4","issue":"interpretation; ไม่มีแหล่งที่บอกว่าประกอบชื่ออัตโนมัติแบบนี้ผ่าน"},
 {"claim":"§2 'MongoModel diagram product 97 / productbarcode 33 fields'","issue":"ไม่ cross-check กับ code (ProductBarcodeBase ~100 fields) — research gap ยังเปิด"},
 {"claim":"§4 'middleware อ่าน /me'","issue":"/me เป็น endpoint ฝั่ง client; backend ต้องคำนวณ union เอง (role_permission_http.go:131-164) — แผนไม่ระบุว่าจะ reuse function ไหน"}
],
"2_thai_legal_reverify": [
 {"statement":"รหัสสินค้าแทนชื่อได้เฉพาะใบกำกับอย่างย่อจากเครื่องบันทึกเงินสด + ต้องมีรหัสพร้อมคำแปลทั้งระบบ","check":"https://www.rd.go.th/3389.html ข้อ 2(3)(ซ), 3(2); และเครื่องบันทึกเงินสดต้องขออนุมัติ (ประกาศ VAT ฉบับที่ 46 — หา URL ต้นฉบับบน rd.go.th)"},
 {"statement":"ค้าปลีกลงรายงานสินค้าเป็นกลุ่มได้ (ใช้อ้าง roll-up variant→product)","check":"https://www.rd.go.th/3374.html ข้อ 9 เงื่อนไขเต็ม (ประเภทกิจการ/ต้องมีรายละเอียดรายวันหรือไม่); ข้อ 10-14 มาตรฐานโปรแกรม ชนิด ก/ข/ค/ง"},
 {"statement":"ลงรายงานสินค้าภายใน 3 วันทำการ; เก็บ 5 ปี; รายสถานประกอบการ; เฉพาะผู้ขายสินค้า","check":"https://www.rd.go.th/5209.html (ม.87 วรรคท้าย), https://www.rd.go.th/5205.html (ม.87/3)"},
 {"statement":"บุคคลธรรมดาไม่ต้องทำรายงานสินค้า แต่ตรวจนับ 30 มิ.ย./31 ธ.ค.","check":"https://www.rd.go.th/3289.html (ประกาศ 104) — ยืนยันยังมีผลและครอบคลุม ห้างหุ้นส่วนสามัญ/คณะบุคคล"},
 {"statement":"VAT 7% ถึง 30 ก.ย. 2570 (พ.ร.ฎ.807)","check":"https://www.rd.go.th/21221.html + อ่านตัวบท https://www.rd.go.th/fileadmin/user_upload/kormor/newlaw/dc807.pdf (สแกน ยังไม่ได้อ่าน)"},
 {"statement":"vattype มีแค่ 0 คิด/1 ยกเว้น; 0% เป็นเรื่องธุรกรรมเท่านั้น","check":"https://www.rd.go.th/5206.html ม.80/1 (1)-(6) — 0% มีกรณีนอกเหนือส่งออก (ขนส่งระหว่างประเทศ, ขายให้ UN/สถานทูต, ขายให้ผู้ประกอบการใน EPZ) ซึ่งผูกกับคู่ค้า/สินค้า ไม่ใช่เอกสารล้วน"},
 {"statement":"รายย่อย <1.8 ล้าน ไม่ต้องจด VAT (ฐานของ company.vatregistered)","check":"https://www.rd.go.th/5206.html ม.81/1 + พ.ร.ฎ.ฉบับที่ 432 (ยังไม่ verify ตัวเลข)"},
 {"statement":"WHT บริการ 3% / เช่า 5% / ขนส่ง 1%","check":"https://www.rd.go.th/3479.html — ตรวจ: ผู้จ่ายต้องเป็นนิติบุคคล (บุคคลธรรมดาผู้จ่ายไม่มีหน้าที่ ข้อ 12/1), เกณฑ์ 1,000 บาท, ค่าโฆษณา 2% (ข้อ 10), ค่าจ้างทำของ 3% (ข้อ 8), เบี้ยประกัน 1%, ขนส่งสาธารณะยกเว้น, ผู้รับต่างประเทศ (ภ.ง.ด.54) → whttype enum ยังไม่ครบ"},
 {"statement":"แบบ ภ.ง.ด.3/53 เลือกจาก legaltype ผู้รับ","check":"https://www.rd.go.th/fileadmin/download/insight_pasi/wht_3_53_030260.pdf ข้อ 1.1-1.2"},
 {"statement":"ม.86/4 ไม่บังคับรหัส/หน่วยนับ; ม.86/6 อย่างย่อต้องราคารวม VAT","check":"https://www.rd.go.th/5208.html; ข้อความอื่นตามประกาศ VAT ฉบับที่ 39 (เลขสาขา ฯลฯ — หา URL ต้นฉบับ)"},
 {"statement":"ชื่อสินค้าภาษาอังกฤษได้","check":"https://www.rd.go.th/3569.html (ป.85/2542)"},
 {"statement":"DBD: บัญชีสินค้าต้องมี ชื่อ ชนิด จำนวน หน่วยนับ; เอกสารต้องมี หน่วย/ราคาต่อหน่วย/รวม; ลงบัญชีสินค้าภายใน 15 วันนับแต่สิ้นเดือน (ข้อ 7(3) — แผนไม่กล่าวถึง)","check":"ราชกิจจานุเบกษา ประกาศกรมทะเบียนการค้า พ.ศ. 2544 (อ่านมาจาก https://www.dip.go.th/uploadcontent/boom_LAW/GNB_022.pdf + https://www.dbd.go.th/data-storage/attachment/fcbaf4217e9ec10119d1e10422.doc ยังไม่ใช่ต้นฉบับ)"},
 {"statement":"e-Tax: Name ≤256 บังคับ, ID ≤35, BilledQuantity บังคับ, unitCode Rec20, TypeCode VAT/FRE, ชื่อต้องมี name+type+category","check":"ขมธอ.3-2560 ฉบับไทย (etda.or.th) + https://github.com/ETDA/XMLValidation XSD + code list 5153 ที่ schemas.teda.th (EXC? รหัสอื่นที่ RD รับ) + ประกาศกรมสรรพากร 2565 https://www.rd.go.th/fileadmin/user_upload/kormor/newlaw/pae_TaxInvoiceA.pdf"},
 {"statement":"TAS 2: FIFO/ถัวเฉลี่ย ห้าม LIFO; moving average ใช้ได้","check":"https://eservice.tfac.or.th/get_file/index.php?file=TAS_2_revised_2568.pdf (ไฟล์ระบุ 34/2562 — หาฉบับ 2568 จริง) ย่อหน้า 23-27; NPAEs 8.9 ใช้คำ 'ถัวเฉลี่ย...แต่ละงวด' https://acpro-std.tfac.or.th/test_std/uploads/files/TFRS%20for%20NPAEs_Revise%202565R1.pdf — ยืนยันว่า moving average ผ่าน NPAEs"},
 {"statement":"กิจการไม่จด VAT ต้องรวม VAT ซื้อในต้นทุน (TAS 2 ย่อหน้า 11)","check":"ไฟล์ TAS 2 เดียวกัน ย่อหน้า 10-11 — แผน §2 costmethod ไม่กล่าวถึงเรื่องนี้"},
 {"statement":"ราคารวม VAT flag ต่อ price level (§7 ข้อ 2)","check":"ม.86/6 + ประกาศ 73 ข้อ 2(3) — กฎหมายพูดถึง 'ราคาที่แสดง' บนใบกำกับอย่างย่อ ไม่ใช่ price level; ต้องยืนยันว่า flag ต่อ level ไม่ขัด"}
],
"3_missing_topics": [
 "Lot/Serial/วันหมดอายุ: ไม่มีใน product/productbarcode/Detail (transaction.go:242 expireddate เป็นระดับเอกสาร) — กระทบ TAS 2 ย่อหน้า 23 ราคาเจาะจง, FEFO, สินค้าอาหาร/ยา, warranty serial",
 "Stock ต่อสาขา (ม.87 รายสถานประกอบการ): warehouse model ไม่มี branchcode (grep backend/internal/warehouse/models ว่าง) — ไม่มี mapping คลัง→สาขา→รายงาน; stockprocess ไม่มี businesscode/branchcode",
 "Stock ต่อบริษัท: stockprocess.go keyed holding+barcode+whcode ไม่มี businesscode — ขัด 'Company = boundary' และ Phase 3 marketplacestockbalances เพิ่ม businesscode/whcode ด้านเดียว",
 "หน่วยนับในรายงาน ม.87 'ปริมาณนับเป็น' = หน่วยเดียว: variant × unit pack (โหล/ลัง) ต้อง roll-up เป็นหน่วยฐานผ่าน product.unitconversions — ไม่มีกฎว่า stock card แสดงหน่วยไหน/ทศนิยม",
 "variant × unit: productbarcode ใช้เป็นทั้ง 'หน่วยขายเพิ่ม' และ 'variant' — ไม่มีกฎว่า SKU สี/ไซซ์ มีได้กี่ unit, isdefault ต่ออะไร, dimensionkey ของแถวหน่วยโหล",
 "GTIN vs barcode ภายใน: POS scan EAN จริงต้อง resolve ผ่าน gtin หรือ barcode? index/ClickHouse search (product_barcode.go:357-362) มีแค่ barcode; GS1 กำหนด GTIN ต่างกันต่อระดับบรรจุ (ชิ้น/โหล/ลัง)",
 "สินค้าชุด/บันเดิล: ลบ product.refbarcodes/isusesubbarcodes แต่คงเมนู 'สินค้าชุด' และ product-set-screen.tsx:417-483 ใช้ options.choices.refbarcode; ไม่แยก set-for-sale (itemtype 2, ราคาชุด, saleinvoicebomprice) vs BOM ผลิต vs marketplace combined_skus/wholesale",
 "materialtype (0 ทั่วไป/1 วัตถุดิบ/2 กึ่งสำเร็จรูป) ↔ NPAEs 8.19.2 (สำเร็จรูป/ระหว่างทำ/วัตถุดิบและวัสดุ) mapping ไม่กำหนด; GL 5 บัญชีต่อกลุ่ม (สินค้า/ต้นทุนขาย/ขาย/ลดมูลค่า/ขาดหาย) ยังเป็นคำถาม",
 "ต้นทุน: ไม่มี field cost/standard cost/last cost บน item, ไม่มี landed cost, ไม่มีนโยบายสต๊อกติดลบ, ไม่มี opening balance/ตรวจนับ (Phase 3 เท่านั้น), NRV write-down process, VAT ซื้อรวมต้นทุนเมื่อไม่จด VAT",
 "ราคา: ทศนิยม/rounding สตางค์, ราคารวม/ไม่รวม VAT ต่อ level, maxdiscount/discount ที่ productbarcode ทิ้งหรือไม่, promo ต่อช่องทางชนราคาสมาชิก keynumber 2, สกุลเงิน (currency module มี)",
 "WHT ฝั่งขาย vs ซื้อ: whttype บน product ใช้ตอน 'ซื้อบริการ' แต่ product master เป็นสินค้าที่ขาย — ต้องแยก item ซื้อ (masterexpense มีอยู่แล้ว?) กับ item ขาย; หน่วย default บริการเพื่อ BilledQuantity",
 "Inactive/ลบ: กฎเมื่อ isdisabled แล้วมีสต๊อกคงเหลือ/อยู่ใน BOM/listing LIVE; barcode ซ้ำเมื่อ re-enable; รายงานตารางรหัส (ประกาศ 73) ยังไม่มีสเปก",
 "SKU ซ้ำข้ามช่องทาง/ข้ามบริษัทใน holding: sellersku unique ต่อบริษัท แต่ barcode guard unique ต่อ holding; listing ต่อ account หลายร้านช่องเดียวกัน (Shopee 2 ร้าน) ใช้ sellersku เดียวกันได้ไหม",
 "Search/projection ของ field ใหม่ (variantaxes, whttype, unececode, channelprices) ลง Pg/ClickHouse/Kafka (Phase 4 ตัดอย่างเดียว ไม่พูดถึง add); ClickHouse search ต้องแสดงชื่อ variant",
 "Permission per branch/company (shopusers.AccessScopes user.go:116,141) กับ product L1/L2/L3 — ไม่กำหนดว่าชุดสิทธิ์ทำงานร่วม scope อย่างไร",
 "Audit log สำหรับการแก้ L1 หลังมีเอกสาร (ADMIN+เหตุผล) — ไม่มี collection/รูปแบบ",
 "Rename product.code ก่อนมีเอกสาร: barcode.itemcode immutable (CoreOnly) + Pg/ClickHouse/stockprocess คีย์ barcode/itemcode — cascade ไม่กำหนด",
 "Import/Export: template Excel variant, mass-upload Shopee/Lazada/TikTok format ต่างกัน — ไม่ระบุ format ต้นแบบ",
 "Marketplace reservation/oversell: ออเดอร์ที่ยังไม่ sync เป็น sale-invoice ต้องจอง stock ใน ledger ไหม (ม.87 3 วันทำการ)",
 "ของแถม/ตัวอย่าง/สินค้าฝากขาย/ค่ามัดจำ/ค่าบริการติดตั้ง — itemtype 3 not-stock ครอบคลุมหรือไม่",
 "Restaurant/eorder/pos consumers ของ field ที่จะซ่อน/ตัด (isalacarte, ordertypes, options) — eorder_http.go:160-172 อ่านอยู่ ไม่อยู่ในแผน"
],
"4_contradictions": [
 {"a":"§1.1 'ไม่มี collection คู่ขนานของข้อมูลบัญชี'","b":"§2 productbarcode ถือ names + prices keynumber 1 (ชั้นบัญชี) + vatrate table + productlisting — ข้อมูลบัญชีอยู่ 2 collection"},
 {"a":"§1.4 'barcode ยังเป็น stock key … ไม่ re-key'","b":"§6 Phase3 sellable คำนวณจาก marketplacestockbalances keyed itemcode+dimensionkey; §2 เพิ่ม businesscode/whcode ที่นั่นแต่ไม่แตะ stockprocess (holding+barcode) → 2 ledger ไม่ reconcile"},
 {"a":"§2 'แก้ barcode unique holding→บริษัท'","b":"repo comment :161-162 บอกว่า guard ต้องคงไว้จน stock spine มี businesscode — แผนไม่มีงานนั้นก่อน"},
 {"a":"§1.3 'SKU ห้าม override; ภาษี/หน่วย/ประเภท อยู่ที่ product ของนักบัญชี'","b":"§3 เจ้าของ 'สร้างรายการย่อย' gen แถว productbarcode ต้องเขียน names/itemunitcode/barcode (ชั้นบัญชี) แต่ §4 ชุด 'ขายออนไลน์' ได้แค่ product(อ่าน)+barcode:update(field ตลาด) — ไม่มี barcode:create"},
 {"a":"§2 ตัด atlas productchannelprices 'ใช้อันเดียว'","b":"§2 ยังมี 2 กลไก: salechannel.price→keynumber (ต้องถูกอ่านจริง) และ productbarcode.channelprices[] ต่อ SKU"},
 {"a":"§2 product ลบ refbarcodes/isusesubbarcodes","b":"§3 เมนูคง 'สินค้าชุด' + product-set-screen.tsx:417-483 อ้าง choice.refbarcode; bom[] อยู่ L2 แต่ set ขาย (itemtype 2) ไม่นิยาม"},
 {"a":"§2 productbarcode 'บังคับ ใช่' vattype ที่ product 'บังคับ'","b":"§3 ฟอร์ม 'ซ่อนทั้งช่อง VAT ถ้าไม่จด' — ค่าที่เก็บเมื่อซ่อน? และ 'whttype บังคับ (เฉพาะบริการ)' vs §3 'โชว์เฉพาะบริการ' ไม่ระบุ required"},
 {"a":"§2 code regex 'A-Z0-9- ≤35' อ้าง ETDA","b":"§5 ETDA 8.20 = A-Z/0-9 เท่านั้น"},
 {"a":"§2 'variant = แถว productbarcode 1 แถว = 1 SKU'","b":"§3 'หน่วยซื้อ/ขายเพิ่ม (โหล/ลัง)' ก็เป็นแถว productbarcode — 1 แถว = SKU หรือ = หน่วยขาย ไม่ชัด; dimensionkey ของแถวหน่วยเสริม"},
 {"a":"§5 'ม.87 group by itemcode' นำเสนอเป็น mapping ตัดสินแล้ว","b":"§7 'ม.87 group ระดับ product/variant default' ยังเป็นคำถาม"},
 {"a":"§3 'Delete(=isdisabled)' ใน UAT","b":"กฎ UAT (AGENTS.md) 'Delete → ตรวจว่าหายจริง' และ product delete endpoint ปัจจุบันลบจริง — ต้องนิยามใหม่ว่า UAT ตรวจอะไร"},
 {"a":"§2 productbarcode ตัด dividevalue/standvalue (Phase 4)","b":"ProductBarcodePg/StockTransactionDetail (stock_transaction_postgres.go:52-81) และ product Update 'unit snapshot resync' (product_http_service.go:383-404) ยังพึ่ง field เหล่านี้; Phase 1-3 ยังเขียนต่อ = สร้าง drift ก่อนตัด"},
 {"a":"§3 'บาร์โค้ด auto = รหัส'","b":"§2 gtin แยก + gtinstatus HAS/NONE + ปุ่ม EAN-13 generate — EAN ที่สร้างเองไปอยู่ช่องไหน และ POS scan ค้นช่องไหน"},
 {"a":"§1.2 'ชั้นเสริมไม่ครบ ไม่บล็อกเอกสาร'","b":"§6 Phase3 AC 'LIVE แล้วแก้ sellersku ถูกปฏิเสธ' + ล็อก code/unit/itemtype หลังมีเอกสาร — ชั้น L3 ย้อนมาบล็อกการแก้ L1/L2 โดยไม่มีกฎ precedence"}
]
}
```

## Research gaps
- ขมธอ.3-2560 ฉบับภาษาไทย (PDF จาก etda.or.th/getattachment/6ed9ef37-...) ยังไม่ได้เปิด — ใช้ฉบับอังกฤษ V03 (2025) ของ ETDA + XSD/Schematron บน GitHub ETDA/XMLValidation แทน; code list B.11 (UN/EDIFACT 5153 ที่ schemas.teda.th) ยังไม่ได้ดาวน์โหลด จึงยืนยันได้เฉพาะตัวอย่าง VAT/FRE ในเอกสาร ยังไม่ยืนยันว่ามีรหัสอื่น (เช่น EXC) ที่ RD ยอมรับ
- ประกาศกรมทะเบียนการค้า พ.ศ. 2544 (บัญชีสินค้า/หน่วยนับ) อ่านจากฉบับรวมของกรมส่งเสริมอุตสาหกรรม (dip.go.th GNB_022.pdf) และเอกสารสรุปของ DBD (.doc) — ยังไม่ได้เปิดฉบับราชกิจจานุเบกษาต้นฉบับ (ข้อความตรงกันทั้ง 2 แหล่ง)
- แบบรายงานสินค้าและวัตถุดิบแนบท้ายประกาศ 89 เป็นภาพ gif 27 ชิ้นที่ตัดเป็นแถบ ประกอบภาพได้บางส่วน: ยืนยันหัวข้อ ชื่อสินค้า/วัตถุดิบ, ชนิด/ขนาด, ปริมาณนับเป็น, เลขที่ใบสำคัญ, วัน เดือน ปี, ปริมาณสินค้า, หมายเหตุ แต่ยังไม่เห็นคอลัมน์ย่อย รับ/จ่าย/คงเหลือ และช่องมูลค่าอย่างชัดเจน
- พ.ร.ฎ. ฉบับที่ 807 (ขยาย VAT 7% ถึง 30 ก.ย. 2570) — PDF เป็นภาพสแกน อ่านตัวบทไม่ได้ ยืนยันจากคำอธิบายบนหน้า 'กฎหมายออกใหม่' ของ rd.go.th เท่านั้น
- ไฟล์ TAS 2 จาก eservice.tfac.or.th ชื่อ TAS_2_revised_2568.pdf แต่หัวกระดาษระบุประกาศสภาวิชาชีพบัญชี ที่ 34/2562 — ย่อหน้า 9/23/25 ที่อ้างเป็นเนื้อหาเดิมของ IAS 2 (ไม่น่าเปลี่ยน) แต่ยังไม่ได้ยืนยันกับฉบับ ปรับปรุง 2568 จริง
- ประกาศ VAT ฉบับที่ 89 ข้อ 10-14 (มาตรฐานซอฟต์แวร์ ชนิด ก/ข/ค/ง ของกรมสรรพากร สำหรับโปรแกรมที่ลงรายงานภาษี/สินค้า) ยังไม่ได้อ่าน — อาจมีข้อกำหนดเพิ่มเรื่อง audit trail/การแก้ไขรายการที่กระทบการออกแบบ stock ledger
- เกณฑ์รายย่อยยกเว้น VAT ตาม ม.81/1 ปัจจุบัน (1.8 ล้านบาท/ปี ตาม พ.ร.ฎ.) ไม่ได้ verify เพราะไม่กระทบโครง item; ประกาศ VAT ฉบับที่ 39 (ข้อความอื่นในใบกำกับภาษี เช่น เลขสาขา, อัตราแลกเปลี่ยน) ไม่ได้อ่านเต็ม
- ไม่ได้ตรวจข้อกำหนด e-Receipt/ใบเสร็จรับเงินตามประกาศกรมสรรพากรปัจจุบัน (2565) แบบละเอียด นอกจากข้อ 4 เรื่องรูปแบบตามเว็บไซต์ RD; และไม่ได้ตรวจ requirement ของ marketplace (Shopee/Lazada) เพราะนอกขอบเขต task นี้
- ค่าที่เป็นไปได้ของ vattype/taxtype/itemtype ใน backend ปัจจุบันไม่พบ enum/const ใน Go (มีเฉพาะ int8 field) — ต้องดูจาก frontend หรือ docs.go ก่อนกำหนด mapping ไปยัง enum ใหม่ (ปกติ/0%/ยกเว้น, สินค้า/บริการ)
- LINE SHOPPING/MyShop official API reference could not be read (spec lives inside OA Plus console; Medium articles return HTTP 403 to fetch) — all LINE field names/limits are unverified and need confirmation from a MyShop-enabled account.
- Shopee Thailand-specific limits (title length, image size, model cap) are shop/category-dynamic; only the get_item_limit sample values (5–100 chars, 1–9 images, ≤2000-char description) were seen — real TH values must be fetched via get_item_limit with a live shop token. Seller Education Hub TH articles are login-gated.
- Shopee model_sku uniqueness per shop is not stated in the API docs (only in a mass-upload PDF snippet).
- Lazada SellerSku uniqueness scope is documented inconsistently ('same item' vs 'store'); Lazada TH Seller Center help pages redirect to login; per-category max variant attributes (maxItems) and total SKU cap ('Invalid variation' 209) numbers are not published — need GetCategoryAttributes for real TH categories.
- Lazada UpdateSellableQuantity/AdjustSellableQuantity request schema was only seen via an SDK mirror, not on open.lazada.com.
- TikTok Shop Thailand Seller University image/listing pages are JS-rendered; image policy (≥600 px, white background, 30-word description) was read from the US policy page and may differ slightly for TH.
- Shopee Global Product / TikTok Global Product (cross-border) flows were not examined — proposal assumes local Thai shops only.
- No official Thai legal/tax source was consulted in this task (product listing only); VAT/e-Tax implications of marketplace sales are out of scope here.
- Repo has no real marketplace API client (marketplace-screen.tsx is mock data) — sync semantics in the proposal are design intent, not existing behaviour.
- MongoModel diagram field counts (product 97 / productbarcode 33) not cross-checked against code (ProductBarcodeBase alone has ~100 fields) — verify with mongomodel MCP before using the diagram as the design baseline.
- Runtime data not inspected: how many productbarcode docs actually have Prices, RefBarcodes, MarketplaceProducts, or empty itemcode (orphans) in appdb — needed to judge which legacy fields are truly dead.
- No Next.js transaction (sale/purchase) line-entry screen was found under frontend/src/app (only menu routes /transaction/* in menu-data.ts); how a line resolves barcode→itemcode→unit in the current UI could not be verified — only the Go Detail model.
- Backend authorization: confirmed absent on product/barcode routes; not checked whether a global middleware enforces screen permissions elsewhere (grep found rolepermission only in main.go registration).
- Kafka consumer for productbarcode (productbarcode_consumer_service.go) and ClickHouse projection semantics of price keynumbers beyond 1/2/3 not read.
- Unit.CompanyGuids vs UI businesscodes mapping is done in the Next proxy/screen; backend-side handling not read in detail.
- Thai regulatory facts (VAT/e-Tax/GTIN) were not researched here — outside this read-only inventory; cite rd.go.th/etda.or.th before adding tax rules to the design.
