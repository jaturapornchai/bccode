# ธุรกรรมซื้อ–ขาย (Transaction: Sales & Purchase) — เส้นทางจาก Mongo → Kafka → PG projection
> ตรวจล่าสุด: 2026-09-07 (commit d93a210d) — ผู้เขียน: AI reader; ทุกข้อเท็จจริงอ้าง path:line · ตรวจซ้ำโดย fact-checker

## 1. ภาพรวมเส้นทางข้อมูล (ตามโค้ดจริง)
1. **HTTP module** ทำงานเมื่อ `DEV_API_MODE` = "" หรือ "2" (backend/main.go:258) → เขียนเอกสารลง MongoDB collection ของแต่ละ module (ตาราง §2)
2. หลัง Mongo write สำเร็จ service ยิง Kafka แบบ **fire-and-forget `go func`** และแค่ `fmt.Printf` เมื่อ publish พลาด (backend/internal/transaction/saleinvoice/services/saleinvoice_service.go:387-397, backend/internal/transaction/purchaseorder/services/purchaseorder_http_service.go:273-282)
3. **goapi consumer** เริ่มที่ `handlers.StartConsumers()` (backend/internal/goapi/bootstrap.go:157-160, gate ด้วย env `ENABLE_KAFKA=="true"`; env ตัวนี้ถูก set จาก key `enablekafka` ใน bootstrap.json → backend/internal/goapi/setupconfig/loader.go:78,293; local container มี `/app/bootstrap.json` ค่า `"enablekafka": "true"` — docker exec mainapi read-only) → เขียนตาราง PG per-holding `doc / docdetail / docref / docpayment` (backend/internal/goapi/mypg/insert_doc.go:77-83,171-173,239-242,323-329)
4. หลัง insert → `docwaitprocess` + `stockwaitprocess` (backend/internal/goapi/mypg/doc.go:47-71) → `ProcessDocumentStatusAsync` (backend/internal/goapi/handlers/kafka/utils.go:501-505 → `ProcessDocumentStatusByShop` เรียกเฉพาะ `ProcessDocPurchaseBatch` backend/internal/goapi/process/process-doc/process-doc-helper.go:19-20) จึงประมวลผลเฉพาะสถานะใบสั่งซื้อ transflag=6 (§8)
5. **Legacy consumers** (`backend/internal/transaction/transactionconsumer/**`, 164 ไฟล์ — git ls-files) ทำงานเฉพาะ `DEV_API_MODE` = "" หรือ "1" (backend/main.go:653-726) เขียนตารางชุด `*transaction` คนละชุดกับ goapi (§9) — เครื่อง local รัน `DEV_API_MODE=2` (docker exec mainapi env) จึง **ไม่ได้รัน**

## 2. Inventory sub-package (git ls-files backend/internal/transaction | awk -F/ '{print $4}' | sort | uniq -c)
รวม 54 รายการ (นับรวม `models`, `repositories`, `transactionconsumer`); ในบทความนี้เจาะฝั่งซื้อ–ขาย–รับ/จ่ายชำระ (module เช็ค/เงินฝาก/มัดจำ/สต็อก อยู่บทอื่น). path ย่อในตารางด้านล่าง (เช่น `saleinvoice_http.go`) อยู่ใต้ `backend/internal/transaction/<module>/` และ `main.go` = `backend/main.go`. ทุก module ใช้ route ชุดเดียวกัน: `POST …/bulk`, `GET …`, `GET …/list`, `POST …`, `GET …/:id`, `GET …/code/:code`, `PUT …/:id`, `DELETE …/:id`, `DELETE …` (ตัวอย่าง backend/internal/transaction/saleorder/saleorder_http.go:47-56). Path จริง = `config.PathPrefix()` + path และมี alias `/v1/...` อัตโนมัติ (backend/pkg/microservice/microservice.go:115, backend/pkg/microservice/microservice_http.go:22-37)

| module | หน้าที่ | route prefix | Mongo collection | MODULE_NAME (prefix เลขที่) | สถานะ | อ้างอิง |
|---|---|---|---|---|---|---|
| saleinvoice (12 ไฟล์) | ขายสินค้า / POS | `/transaction/sale-invoice` (+ `/auto-coupon`, `/last-pos-docno`, `/guidpos/:code`, `/recalpoint/:code`, `/export`) | `transactionsaleinvoice` | `SI`, TRANS_FLAG 44 | LIVE | saleinvoice_http.go:81-95; models/saleinvoice.go:11; services/saleinvoice_service.go:53-54; main.go:430 |
| saleinvoicereturn | รับคืนสินค้า | `/transaction/sale-invoice-return` | `transactionsaleinvoicereturn` | `ST` | LIVE | saleinvoicereturn_http.go:56-66; models/saleinvoicereturn.go:10; services/saleinvoicereturn_service.go:50; main.go:431 |
| saleorder | ใบสั่งขาย | `/transaction/sale-order` | `transactionsaleorder` | `SO` | LIVE | saleorder_http.go:47-56; models/saleorder.go:10; services/saleorder_http_service.go:39; main.go:445 |
| quotation | ใบเสนอราคา | `/transaction/quotation` | `transactionquotation` | `QT` | LIVE (HTTP) / ไม่มี goapi consumer | quotation_http.go:47-56; models/quotation.go:10; services/quotation_http_service.go:39; main.go:444 |
| pickandpack | จัดสินค้า/แพ็ก | `/transaction/pickandpack` (+approve/closejob/confirm ฯลฯ) | `transactionpickandpack` | `PP` | LIVE / ไม่มี goapi consumer | pickandpack_http.go:86-109; models/pickandpack.go:11; services/pickandpack_http_service.go:47; main.go:448 |
| saleinvoicebomprice | ราคาชุด BOM ของ SI (อ่านอย่างเดียว) | `/transaction/sale-invoice-price` | `transactionsaleinvoicebomprices` | – | LIVE (GET only) | saleinvoicebomprice_http.go:39-42; models/saleinvoicebomprice.go:9; main.go:545 |
| saledebitnote | เพิ่มหนี้ขาย | `/transaction/bank/saledebitnote` | `transactionsaledebitnote` | `SA` | **DEAD** — ไม่มี `saledebitnote.New…Http` ใน main.go (grep เจอเฉพาะ legacy consumer import :173 และ migration mode 3 :615) | saledebitnote_http.go:47-56; models/saledebitnote.go:10; services MODULE_NAME :39 |
| purchaseorder | ใบสั่งซื้อ (มี validator + multi-currency + WHT rate check) | `/transaction/purchase-order` | `transactionpurchaseorder` | `PO` | LIVE | purchaseorder_http.go:85-94; models/purchaseorder.go:10; services/purchaseorder_http_service.go:42,222-284 (currencyRepo :49); validators/po_validator.go:61-63; main.go:441 |
| purchase | ซื้อ/รับสินค้าเต็มใบ | `/transaction/purchase` | `transactionpurchase` | `PU` | LIVE | purchase_http.go:51-60; models/purchase.go:10; services/purchase_service.go:46; main.go:428 |
| purchasepartial | รับสินค้าบางส่วน | `/transaction/purchasepartial` | `transactionpurchasepartial` (ตัวแปรชื่อ `saleorderCollectionName` — copy-paste; quotation.go:10 และ pickandpack.go:11 ก็ใช้ชื่อตัวแปรนี้) | `PP` (ชนกับ pickandpack) | LIVE | purchasepartial_http.go:47-56; models/purchasepartial.go:10; services/purchasepartial_http_service.go:39; main.go:446 |
| purchasereturn | ส่งคืนสินค้า | `/transaction/purchase-return` | `transactionpurchasereturn` | `PT` | LIVE | purchasereturn_http.go:50-59; models/purchasereturn.go:10; services/purchasereturn_service.go:46; main.go:429 |
| purchaserequisition | ใบขอซื้อ | `/transaction/purchase-requisition` | `transactionpurchaserequisition` | `PR` | LIVE | purchaserequisition_http.go:71-79; models/purchaserequisition.go:10; services/purchaserequisition_http_service.go:40; main.go:442 |
| rfq | สืบราคา | `/transaction/rfq` | `transactionrequestforquotation` | `RFQ` | LIVE | rfq_http.go:72-80; models/rfq.go:10; services/rfq_http_service.go:40; main.go:443 |
| purchasedebitnote | เพิ่มหนี้ซื้อ | `/transaction/bank/purchasedebitnote` | `transactionpurchasedebitnote` | `DN` | **DEAD** — ไม่มี `purchasedebitnote.New…Http` ใน main.go (grep เจอเฉพาะ legacy consumer import :166 และ migration mode 3 :607) | purchasedebitnote_http.go:47-56; models/purchasedebitnote.go:10; services MODULE_NAME :39 |
| paid | รับชำระหนี้ (ลูกหนี้) | `/transaction/paid` | `transactionpaid` | `EE` | LIVE / ไม่มี goapi consumer | paid/debtor_payment_http.go:46-55; paid/models/debtor_payment.go:11; paid/services/debtor_payment_http_service.go:39; main.go:437 |
| pay | จ่ายชำระหนี้ (เจ้าหนี้) | `/transaction/pay` | `transactionpay` | `DE` | LIVE / ไม่มี goapi consumer | pay/creditor_payment_http.go:47-56; pay/models/creditor_payment.go:11; pay/services/creditor_payment_http_service.go:25; main.go:438 |
| payment / paymentdetail | PG model+usecase (ไม่มี HTTP) ตาราง `paymenttransaction` / `paymenttransactiondetail` | – | – | – | STUBBED/legacy — ใช้เฉพาะ migration mode 3 (main.go:595-596) และ legacy SI consumer delete (transactionconsumer/saleinvoice/saleinvoice_transaction_consumer.go:212) | payment/models/payment.go:40; paymentdetail/models/payment_detail.go:36 |
| documentformate | ตั้งค่ารูปแบบเลขเอกสาร (doccode, module, dateformate, docformat, isautoformat…) | `/transaction/document-formate` (+`/default`) | `documentformate` | – | LIVE (main.go:512) แต่ **ไม่มี service ธุรกรรมใดเรียกใช้** (grep documentformate ใน backend/internal/transaction เจอเฉพาะ pkg ตัวเอง) | documentformate_http.go:43-53; models/documentformate.go:10,12-27 |
| smltransaction | เก็บเอกสารดิบจากระบบ SML (collection `smltransactions`) | `POST /sml-transaction`, `POST /sml-transaction/bulk`, `DELETE /sml-transaction` (GET ถูก comment ออก) | `smltransactions` | – | LIVE (main.go:410) | smltransaction_http.go:44-48; models/smltransaction.go:21 |
| models (37 ไฟล์) / repositories | shared header/detail struct + PG legacy model + Redis cache เลขที่ | – | – | – | LIVE | §3, §5 |

## 3. Shared models (backend/internal/transaction/models)
- `TransactionHeader` (transaction.go:11-169): `docno`, `docdatetime`, `docdatelocal/doctimelocal` (:14-16 comment ระบุใช้สร้าง docno ตาม timezone ของ frontend), `transflag` (:21), `docreferences []TransactionDocRef` (:24), `custcode` (:36), ยอด `totalamount` (:49) ฯลฯ, `paymentdetail` (:59), creator/updater (:123-131), multi-currency `doccurrency/doccurrencysymbol/exchangerate` + ยอด `*doc` (:142-157), `whtentries` (:162), `creditdays/duedate` (:167-168)
- `Transaction = TransactionHeader + Details *[]Detail` (:213-216); `Detail` (:299-…) มี `linenumber`, `docref`, `calcflag`, `barcode`, ราคา/จำนวน/คลัง
- `TransactionDocRef` (:272-276) = `DocIdentity` + `docno` + `docdatetime` — **ไม่มี transflag ของเอกสารที่อ้าง** (ผลต่อ docref §7)
- `TransactionMoney/DetailMoney` (:218-223) ใช้กับ module รับ/จ่ายเงิน
- ตัวอย่าง embed: `SaleInvoice{ PartitionIdentity, ... }` (saleinvoice/models/saleinvoice.go:13-67); `SaleInvoiceInfo = DocIdentity + SaleInvoice` (:73-76) → `SaleInvoiceData = HoldingCodeentity + SaleInvoiceInfo` (:82-85) → `SaleInvoiceDoc = _id + SaleInvoiceData + ActivityDoc` (:87-91)
- PG legacy model (ใช้โดย transactionconsumer เท่านั้น): `saleinvoicetransaction(+detail)` (transaction_saleinvoice_postgres.go:60-64), `saleordertransaction`, `purchasetransaction`, `purchaseordertransaction`, `purchasereturntransaction`, `rfqtransaction`, `purchaserequisitiontransaction`, `paidtransaction`, `paytransaction`, `debtorpaymenttransaction`, `creditorpaymenttransaction`, `stocktransaction(+detail)` (stock_transaction_postgres.go:49,89), `debtortransaction` (debtor_transaction_postgres.go:36), `creditortransaction` (creditor_transaction_postgres.go:36)

## 4. ตาราง transflag (ที่มาจากโค้ด)
| transflag | เอกสาร | ที่ประกาศ |
|---|---|---|
| 6 | ใบสั่งซื้อ (PO) | backend/internal/goapi/handlers/kafka/constants.go:112 |
| 12 | ซื้อ/รับสินค้า (purchase) | constants.go:111 |
| 16 | ส่งคืนสินค้า (purchase return) | constants.go:114; legacy transactionconsumer/usecases/transaction_purchasereturn_phaser.go:35 |
| 21 | ใบขอซื้อ (PR) | constants.go:115 |
| 22 | สืบราคา (RFQ) | constants.go:116; legacy transactionconsumer/rfq/rfq_transaction_phaser.go:97 |
| 36 | ใบสั่งขาย (SO) | constants.go:110 |
| 44 | ขายสินค้า (SI) | constants.go:108; saleinvoice_service.go:54 |
| 48 | รับคืนสินค้า (SI return) | constants.go:109; legacy transactionconsumer/usecases/transaction_saleinvoicereturn_phaser.go:33 |
| 310 | รับสินค้าบางส่วน (purchase partial) | constants.go:113 |
| 54 / 56 / 58 / 60 / 66 / 68 / 72 | ยอดยกมา / เบิก / คืน / รับ / ปรับเพิ่ม / ปรับลด / โอนคลัง | constants.go:117-123 |
| 99 | ลูกหนี้อื่น (receivableother) | receivableother/services/receivableother_http_service.go:40 |
| ไม่มีค่า | quotation, paid, pay, pickandpack, debit note | ไม่พบ const ใน goapi/handlers/kafka/constants.go และ service |

- **ใครตั้ง transflag ลง Mongo:** มีเพียง saleinvoice ที่ set `dataDoc.TransFlag = TRANS_FLAG` (saleinvoice_service.go:356, 573, 994); module อื่นไม่ set (grep `TransFlag =` ใน services ของ saleorder/purchase/purchaseorder/purchasepartial/purchasereturn/purchaserequisition/rfq/quotation/saleinvoicereturn/paid/pay/pickandpack ไม่พบ) → เอกสาร `transactionpurchase` ใน Mongo local มี `transflag: 0` ทั้ง 6 ใบ ขณะที่ `transactionsaleinvoice` มี 44 (mongosh appdb, holding `test`)
- goapi consumer **บังคับค่าเอง** ตอนแปลงเป็น process model (purchase.go:257 `TRANS_FLAG_PURCHASE`, sale_invoice.go:241; module อื่นเช่นกัน — sale_return.go:198, sale_order.go:199, purchase_order.go:332, purchase_requisition.go:208, rfq.go:208, purchase_partial.go:232) → PG `doc` มี transflag 12/44 ถูกต้องแม้ Mongo เป็น 0
- ชุดที่ process ต้นทุนสต็อก: `TransFlagsToProcess = {54,12,310,48,60,58,66,44,16,56,68,72}` (backend/internal/goapi/myglobal/global.go:32); ยอดค้างรับ = PO 6 ที่ยังไม่ปิด − รับแล้ว 12/310 ที่อ้าง PO ผ่าน docref (process-stock/process-product-balance-update.go:206-226); ยอดค้างส่ง = SO 36 − ส่งแล้ว 44 (:268-285)

## 5. การออกเลขที่เอกสาร (document numbering)
- สูตร `MODULE_NAME + YYYYMMDD + %05d` เช่น `SI2026071900001` (saleinvoice_service.go:135-138, 140-165; purchaseorder_http_service.go:175-183, 185-217); ตัวเลขล่าสุดเก็บ Redis key `<holding>:<prefix>:doc` (repositories/cache_repository.go:25-48) ถ้า cache ว่างจึง `FindLastDocNo` จาก Mongo แล้วเช็คซ้ำด้วย `FindByDocIndentityGuid` → error "DocNo is exists" (saleinvoice_service.go:160-170, 336-346)
- แหล่งวันที่ต่างกัน 2 แบบ: **`DocDateLocal` ที่ frontend ส่ง** ใช้กับ PO/PR/RFQ (purchaseorder_http_service.go:175, purchaserequisition_http_service.go:83, rfq_http_service.go:83) ส่วน **`DocDatetime` (server)** ใช้กับ SI/ST/SO/QT/PU/PT/PP/paid/pay (saleinvoice_service.go:135; saleorder_http_service.go:79; purchase_service.go:92; pay/creditor_payment_http_service.go:74 ฯลฯ)
- SI แบบ POS: `IsPOS=true` ต้องส่ง docno มาเอง ไม่ gen (saleinvoice_service.go:312-333); `TaxDocNo` ว่าง → ใช้ docno (:358-360)
- cache เลขที่ถูก Save **ใน goroutine หลังตอบ client** (saleinvoice_service.go:394-396; PO save ซ้ำ 2 ครั้ง :271,280) → ยิง create พร้อมกันหลาย request อาจชนเลข (มีแค่ check "DocNo is exists" กันไว้ ไม่มี unique index ที่ตรวจในบทนี้ — ยังไม่ตรวจ index)
- Prefix ชนกันข้าม module: `PP` = purchasepartial (purchasepartial_http_service.go:39) และ pickandpack (pickandpack_http_service.go:47); `PO` = purchaseorder (:42) และ banktransferrecord (banktransferrecord_http_service.go:39) — คนละ collection จึงไม่ error แต่เลขที่ไม่ unique ทั้งระบบ
- `documentformate` มี field `docformat/isautoformat/dateformate` แต่ไม่มี service ธุรกรรมใดอ่านค่า (§2) → การตั้งค่าบนจอนี้ **ไม่มีผลกับเลขที่จริง**

## 6. Kafka topics ที่ publish (ฝั่ง HTTP module)
- ทุก module มี config 6 topic: `when-<module>-created|updated|deleted` + `-bulk-*` เช่น saleinvoice/config/saleinvoice_messagequeue_config.go:4-9, saleorder …:4-9, purchaseorder …:4-9, purchasepartial, purchasereturn, purchaserequisition, rfq, quotation (`when-quotation-*`), pickandpack (`when-pickandpack-*`); ชื่อพิเศษ: paid = `when-debtor-payment-*` (paid/config/debtor_payment_message_queue_config.go:4-9), pay = `when-creditor-payment-*` (pay/config/creditor_payment_message_queue_config.go:4-9)
- Publish ทำใน `go func` หลัง Mongo write ทุก module (`svc.repoMq.Create/Update/Delete` — saleorder_http_service.go:155,192,220; quotation :155; paid :151; pay :152; rfq :152; purchaserequisition :162) → client ได้ 200 ก่อน Kafka ยืนยัน; ถ้า broker ล่ม เอกสารอยู่ใน Mongo แต่ไม่ถึง PG และไม่มี retry/outbox ในโฟลเดอร์นี้
- ในโค้ด การ `CreateTopicR` (5 partition, retention 7 วัน) สำหรับ topic ธุรกรรมมีเฉพาะใน **legacy consumer** (transactionconsumer/saleinvoice/saleinvoice_transaction_consumer.go:71-76) ซึ่ง mode 2 ไม่รัน — topic บน local จึงเกิดจาก **broker auto-create** (`KAFKA_AUTO_CREATE_TOPICS_ENABLE=true` ใน env ของ container kafka; `kafka-topics --describe` ได้ PartitionCount 1 ไม่ใช่ 5) เมื่อ producer/consumer แตะ topic ครั้งแรก; local มี `created/updated/deleted` ของ saleinvoice, saleinvoicereturn, saleorder, purchase, purchaseorder, purchasepartial, purchasereturn (3 topic/ครอบครัว) และ purchaserequisition, rfq (6 topic รวม bulk) — ไม่มี `when-quotation-*`, `when-debtor-payment-*`, `when-creditor-payment-*`, `when-pickandpack-*` (kafka-topics --list)

## 7. goapi consumer → PostgreSQL (doc / docdetail / docref / docpayment)
- **Entry จริง** = `handlers.StartConsumers` (backend/internal/goapi/handlers/kafka.go:26-1009) group id = `<base>-<KafkaConsumerGroupVersion>` (kafka.go:19-23) → local มี `goapi-saleinvoice-consumer-v1`, `goapi-purchaseorder-consumer-v1` ฯลฯ (kafka-consumer-groups --list). ส่วน `kafka.StartConsumers` ใน handlers/kafka/manager.go:10-49 เป็นสำเนาเก่าที่เรียกได้จาก `StartKafkaConsumersWithActualImplementation` (kafka_bridge.go:584-586) ซึ่ง **ไม่มีใครเรียก** → DEAD

| topic | callback (kafka.go) | handler จริง | transflag ที่บังคับ |
|---|---|---|---|
| when-saleinvoice-* | :41-45 `CallSaleInvoiceConsumer` → kafka_bridge.go:195-196 | sale_invoice.go:40-136 `ProcessSaleInvoiceDocument` | 44 |
| when-saleinvoicereturn-* | :79-84 | sale_return.go:39-134 | 48 |
| when-saleorder-* | :269-274 | sale_order.go:41-135 | 36 |
| when-purchase-* | :309-314 | purchase.go:44-150 | 12 |
| when-purchaseorder-* | :349-354 | purchase_order.go:63-228 (ใช้ Tx + history) | 6 |
| when-purchaserequisition-* (+bulk) | :387-443 | purchase_requisition.go:36-103 | 21 |
| when-rfq-* (+bulk) | :455-511 | rfq.go:36-103 | 22 |
| when-purchasepartial-* | :523-528 | purchase_partial.go:39-129 | 310 |
| when-purchasereturn-* | :561-566 | purchase_return.go:40-129 | 16 |
| quotation / paid / pay / pickandpack / debit note | **ไม่มี** (grep `when-quotation|debtor-payment|creditor-payment|pickandpack` ใน kafka.go ไม่พบ) | – | – |

ขั้นตอนมาตรฐาน (ตัวอย่าง SI, sale_invoice.go):
1. decode + ตรวจ `HoldingCode/DocNo` (:52-60) → `PgSqlFastConnect(holdingCode)` เลือก DB ตาม holding (:70)
2. **upsert = DELETE แล้ว INSERT ใหม่**: `DeleteDocPgSql` ลบ `doc`, `docdetail` (ตาม businesscode+docno+transflag) และ `docpayment` (mypg/doc.go:24-44) — **ไม่ลบ `docref`** (การลบ docref มีที่เดียวคือ rebuild ทั้งชุดตาม transflag ใน process/build/build-doc.go:98 — ต้องใช้ `rg --no-ignore` เพราะ `.ignore` ตัด `**/build/`; ไม่มีใน path per-document) และ `docref` มีแค่ `id SERIAL PRIMARY KEY` + index ธรรมดา ไม่มี unique (process/build/create-database.go:1280-1309) → update ซ้ำจะสะสมแถว docref ซ้ำ (ยังไม่ทดสอบ runtime เพราะ local docref=0)
3. map header → `DocStruct/DocPaymentStruct` (`myglobal.MapDocStructFromMongo`) และ docref จาก header `DocReferences` ด้วย `MapDocRefStruct` ที่ใส่ `DocRefNoTransFlag: 0 // TODO` (myglobal/map-data-from-mongo.go:205-212); `setDocumentCompany` ต้องมี businesscode ไม่งั้น error (utils.go:166, 288-290)
4. COPY (`BulkInsertWithCopy` insert_doc.go:146,187) ลง `doc` (columns :77-83), `docref` (:171-173), `docpayment` (:239-242; ข้าม transflag 6/36/72 :227-230; ถ้า transflag 0 ใช้ของ header :217-224), แล้ว `docdetail` (:323-329) ผ่าน `InsertDocumentToPostgreSQL/InsertDocDetailToPostgreSQL` (utils.go:286-306, 339-350)
5. `InsertDocumentToClickHouse` = **stub return nil** (utils.go:375-378) และ `SoftDeleteDocClickHouse` ว่าง (:495-498) — สอดคล้องกับ ClickHouse ที่ถอดออก 2026-09-06
6. `ProcessDocumentStockCalculation` (utils.go:161) → error แค่ log (sale_invoice.go:118-123)
7. `AddToDocWaitProcessQueues` insert `docwaitprocess(docno, transflag)` + `stockwaitprocess(itemcode)` จาก docdetail (mypg/doc.go:47-71) → `ProcessDocumentStatusAsync` (ชื่อ Async แต่เรียก **synchronous** ไม่มี `go`; utils.go:501-505)

ความต่างต่อ module ที่สำคัญ:
- **purchase (12)** สร้าง docref จาก `detail.DocRef` ระดับบรรทัด พร้อม `DocRefNoTransFlag: 6` (purchase.go:90-104) แล้ว enqueue PO ที่ถูกอ้างเข้า `docwaitprocess` transflag 6 (:137-140) → สถานะ PO อัปเดต
- **purchasepartial (310)** ใช้ header `DocReferences` ผ่าน `MapDocRefStruct` (DocRefNoTransFlag=0; purchase_partial.go:93-95) และ **ไม่ enqueue PO / ไม่เรียก ProcessDocumentStatusAsync** (enqueue เฉพาะเลขที่ตัวเอง transflag 310 :124 แล้ว return :128; ผู้เรียก `ProcessDocumentStatusAsync` มีแค่ purchase.go:146, sale_invoice.go:132, sale_order.go:131, sale_return.go:130) → รับบางส่วนแล้ว `isref/isclosed` ของ PO ไม่ถูกคำนวณจนกว่าจะมี event อื่น (อนุมานจากโค้ด; ยังไม่ทดสอบ runtime)
- **purchaseorder (6)** ทำใน transaction (`db.BeginTx` purchase_order.go:97 → commit :135), คืนค่า `isclosedmanual*` หลัง DELETE+INSERT (:139-146), บันทึก history `datahistory.SavePOHistory` (:198 ตอน create/update; :45 ตอน delete), ใส่ตาราง `queues` ผ่าน `AddDocToProcessQueue` (:155-157; queue_helper.go:15-67; ถ้าพลาดจึง fallback `AddToDocWaitProcessQueues` :156) และเรียก `ProcessDocumentStatusByDocNo` ตรง (:160)
- **delete ทุก module** = `DeleteDocumentFromDatabases` (ผู้เรียก: ทั้ง 9 handler ซื้อ–ขาย + stock_* ใน handlers/kafka) → soft delete `UPDATE doc SET isdelete=true` + insert docwaitprocess (utils.go:461-492) — `docdetail` ไม่ถูก flag
- **semantics consumer**: kafka-go `reader.ReadMessage` แบบมี GroupID, `StartOffset: FirstOffset`, `CommitInterval: projectionCommitInterval(topic)` (mykafkaconsumer/kafka_consumer.go:298, 422-434) → offset ถูก commit ตามรอบอ่าน ไม่ผูกกับผลของ handler; handler รันใน worker ต่อ holding พร้อม timeout 30s (:77, 191-240) และเมื่อ handler error/timeout จะ **แค่ log + นับ counter** (:217-220, 231-236) → ไม่มี DLQ/retry (ตรงกับที่ lead สรุป "DLQ log-only")

## 8. Post-processing: docwaitprocess / stockwaitprocess / queues
- `docwaitprocess` ถูกอ่านและลบ **เฉพาะ transflag=6** (`rg --no-ignore "FROM docwaitprocess"` ใน goapi เจอเฉพาะ process/process-doc/process-doc-purchase-batch.go:37,69,100,209; process-doc-purchase.go:160,195) → แถว 44/12/48/16/310/21/22/36 ที่ทุก consumer insert ไว้ **ไม่มีใครลบ** — local DB `test`: `docwaitprocess` 12=5, 44=3 ค้างอยู่ (psql read-only)
- PO batch: `isref` = มี docref ที่ `docnotransflag IN (310,12)` ชี้มา (batch.go:30-37); `iscomparedsuccess/isclosed` เทียบผลรวม docdetail ของ 12/310 กับ 6 (batch.go:50-90); จบด้วย `DELETE FROM docwaitprocess WHERE transflag = 6` (:100)
- `stockwaitprocess` ถูกอ่าน/ลบโดย process-stock/product-calc-cost.go:67, 292
- ตาราง `queues` + `WorkerManager` (bootstrap.go:136-141; workers/doc_processor.go:64-68 `OptimalWorkerCount(4,32,2.0)` ตาม CPU; poll `Sleep(100ms)` :239,252,272): dispatch ตาม transflag string (doc_processor.go:357-372) **ติดป้ายผิด** ("12"→SaleInvoice, "44"→Creditor, "48"→Customer, "54"→Debtor) และปลายทาง `process.ProcessPurchaseOrderStatus / ProcessSaleInvoiceStatus` เป็น TODO no-op (process/process_status.go:11-36); ผู้ enqueue มีแค่ PO consumer (purchase_order.go:155 → `INSERT INTO queues` mypostgres/queue.go:70; `rg --no-ignore` ไม่พบผู้เรียกอื่น) → local `queues=0`
- ปิด PO ด้วยมือ: `POST /goapi/api/purchase-order/manual-close` (bootstrap.go:506 → handlers/purchase_order_close.go:89-122); รายงานขาย transflag=44: `POST /goapi/api/report/sales/by-document|summary` (bootstrap.go:401-402; sales_report.go:177)

## 9. Legacy transactionconsumer (โหมด ""/"1")
- main.go:653-726 (block `devApiMode == "" || "1"`) register consumer ทุก module (SO/SI/ST/PO/PR/RFQ/PU/PT/receive/AP-AR payment/stock) group `TRANSACTION_CONSUMER_GROUP` default `transaction-consumer-group-01` (saleinvoice_transaction_consumer.go:67); SI consumer upsert `saleinvoicetransaction(+detail)` (:124) + `stocktransaction` (:137) + `debtortransaction` (:152) + payment (:170; ตาราง §3) และ delete รวม paymenttransaction (:191-212)
- ตารางชุดนี้สร้างเฉพาะ `DEV_API_MODE=3` (main.go:585-651; transactionconsumer/migration.go:11-24 — AutoMigrate เฉพาะ stock/creditor/debtor ส่วน purchase/SI ถูก comment ออก :17-23) — local `test` ไม่มี `saleinvoicetransaction` (psql: relation does not exist) ยืนยันว่าไม่เคยรัน
- transactionconsumer/transactionconsumer.go:100-150 comment การ CreateTopic/Consume ของ purchase, purchasereturn, saleinvoice ออกทั้งหมด เหลือ `ms.Consume` จริงเฉพาะ saleinvoicereturn (:154-159) → package นี้คือ **โค้ดคู่ขนานที่ไม่ได้ใช้ใน mode 2** แต่ยังคอมไพล์และถูก import (main.go:159-196)

## 10. หลักฐาน runtime (local, read-only)
- `docker exec mainapi env`: `DEV_API_MODE=2`, `KAFKA_SERVER_URL=kafka:29092`; `docker logs mainapi`: "GoAPI: 🚀 เริ่มต้น Kafka consumers..." (bootstrap.go:159) และ "เริ่มต้น Kafka consumers เรียบร้อย" (handlers/kafka.go:1008) เมื่อ 2026-09-05 23:27:40 → consumer goapi ทำงาน; `kafka-consumer-groups --list` มี `goapi-saleinvoice-consumer-v1`, `goapi-purchaseorder-consumer-v1` ฯลฯ ครบ 9 ครอบครัวซื้อ–ขาย และไม่มี `transaction-consumer-group-01`
- Mongo `appdb`: `transactionsaleinvoice` 6 (transflag 44 ทั้งหมด), `transactionpurchase` 6 (transflag 0 ทั้งหมด) — holding `test`; docno `SI2026071900001`–`SI2026071900006`; PG `test`: `doc` transflag 12=5, 44=5, `docdetail` 9, `docref` 0, `docpayment` 0, `queues` 0, `docwaitprocess` 12=5/44=3 → **SI ใน Mongo 6 แต่ PG 5** (1 ใบไม่ถูก project — สาเหตุยังไม่ตรวจ)
- Frontend: ไม่มีหน้า `frontend/src/app/transaction*` (ls) — มีแค่ dashboard เรียก `${mainApiUrl}<path>` โดย path = `/transaction/<module>/list` (frontend/src/app/menu/dashboard-home.tsx:33-36,107) และเมนู `frontend/src/lib/menu-data.ts` ที่ path ไม่ตรง API เช่น `/transaction/purchaseorder` (menu-data.ts:60) vs API `/transaction/purchase-order`

## ช่องว่าง / สิ่งที่ยังไม่ตรวจ
- ยังไม่ตรวจ middleware auth บน `/transaction/*` (พารามิเตอร์ `m ...echo.MiddlewareFunc` ใน microservice_http.go); ยังไม่ตรวจว่า bootstrap.json ใน repo/deploy อื่นตั้ง `enablekafka` เหมือน local container หรือไม่
- ยังไม่ตรวจ: route/หน้าที่ของ `smltransaction`, flow `pickandpack` (approve/closejob), `saleinvoicebomprice` consumer, legacy `purchaseorder` consumer/phaser, `payment` usecase, index unique ของ `docno` ใน Mongo
- ยังไม่ทดสอบ runtime: docref ซ้ำหลัง update, PO status หลังรับบางส่วน (310), เลขที่ PR/RFQ เมื่อ frontend ไม่ส่ง `docdatelocal` (getDocNoPrefix จะได้ prefix ไม่มีวันที่)
- สาเหตุ SI 6 ใบใน Mongo แต่ 5 ใน PG `test`
- คำถามถึงลุงจืด: (1) จะเก็บ legacy `transactionconsumer` (164 ไฟล์) ไว้หรือถอด? (2) quotation/paid/pay/pickandpack ตั้งใจให้ไม่มี PG projection หรือยังไม่ทำ? (3) prefix `PP`/`PO` ชนกันยอมรับได้ไหม? (4) แถว `docwaitprocess` ที่ไม่ใช่ 6 ควรมี job ลบหรือให้เลิก insert? (5) `saledebitnote/purchasedebitnote` ที่ไม่ register จะลบทิ้งหรือเปิดใช้?
