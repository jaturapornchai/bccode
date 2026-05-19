# Database Architecture

BC Ai Account Platform ใช้ 3 ฐานข้อมูลหลัก + Message Queue + Cache

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   MongoDB    │     │  PostgreSQL  │     │  ClickHouse  │
│  (Source of  │────▶│  (Relational │────▶│   (OLAP /    │
│   Truth)     │     │   per-shop)  │     │  Analytics)  │
└──────────────┘     └──────────────┘     └──────────────┘
       │                    │                     │
       │              ┌─────┴─────┐               │
       │              │   Kafka   │               │
       │              │  (Events) │               │
       │              └───────────┘               │
       │                                          │
       └──────────────┬───────────────────────────┘
                      │
                ┌─────┴─────┐
                │   Redis   │
                │  (Cache)  │
                └───────────┘
```

## MongoDB — Source of Truth

**บทบาท:** เก็บข้อมูลต้นฉบับ (master data + transactions) จาก POS/Frontend

**ลักษณะสำคัญ:**
- 1 database ต่อ 1 shop (ตั้งชื่อตาม shopid)
- ต้องมี Replica Set `rs0` (สำหรับ Change Streams)
- ลงบน host (native) — ไม่อยู่ใน Docker (ข้อมูลปลอดภัย)
- เข้าถึงจาก Docker ผ่าน `host.docker.internal`

**Collections หลัก (ต่อ shop):**

| Collection | เนื้อหา |
|-----------|---------|
| `transactionSaleInvoice` | ใบขาย (transflag=44) |
| `transactionSaleReturn` | ใบรับคืน (transflag=48) |
| `transactionPurchase` | ใบซื้อ (transflag=12) |
| `transactionPurchaseReturn` | ใบส่งคืน (transflag=16) |
| `transactionStockPickup` | ใบเบิก (transflag=56) |
| `transactionStockReturn` | ใบรับคืนเบิก (transflag=58) |
| `transactionStockReceive` | ใบรับ (transflag=60) |
| `transactionStockTransfer` | ใบโอน (transflag=72) |
| `transactionStockAdjust` | ปรับเพิ่ม (transflag=66) / ปรับลด (transflag=68) |
| `transactionOpeningBalance` | ยอดยกมา (transflag=54) |
| `productBarcode` | สินค้า + barcode |
| `debtorMaster` | ลูกหนี้/ลูกค้า |
| `creditorMaster` | เจ้าหนี้/ผู้จำหน่าย |
| `memberMaster` | สมาชิก |
| `warehouseMaster` | คลังสินค้า |
| `branchMaster` | สาขา |
| `departmentMaster` | แผนก |
| `unitMaster` | หน่วยนับ |
| `couponMaster` | คูปอง |
| `saleChannelMaster` | ช่องทางขาย |
| `employeeMaster` | พนักงาน |

## PostgreSQL — Relational per-shop

**บทบาท:** เก็บข้อมูลเชิง relational สำหรับ report + query + BI

**ลักษณะสำคัญ:**
- 1 database ต่อ 1 shop (ตั้งชื่อตาม shopid)
- ลงบน host (native) — ไม่อยู่ใน Docker
- ใช้ `*sql.DB` ผ่าน `PgSqlFastConnect(shopid)`
- มี approval workflow tables

**Tables หลัก (ต่อ shop):**

| Table | เนื้อหา |
|-------|---------|
| `doc` | Header ของเอกสาร (transflag ระบุประเภท) |
| `docdetail` | รายละเอียดสินค้าในเอกสาร |
| `docpayment` | การชำระเงิน |
| `docref` | เอกสารอ้างอิง |
| `productbarcode` | สินค้า + barcode + ราคา |
| `warehouses` | คลังสินค้า |
| `locations` | ตำแหน่งในคลัง |
| `debtors` | ลูกหนี้ |
| `creditors` | เจ้าหนี้ |
| `queues` | คิว approval workflow |

## ClickHouse — OLAP Analytics

**บทบาท:** วิเคราะห์ข้อมูล (Dashboard, Report, MCP Tools) — ข้อมูลเยอะ query เร็ว

**ลักษณะสำคัญ:**
- 1 database กลาง (ไม่แยกต่อ shop)
- Partition by `shopid` (multi-tenant isolation)
- ใช้ `ReplacingMergeTree` สำหรับ dedup (doc, productbarcode)
- ใช้ `MergeTree` สำหรับ append-only (docdetail, processstock*)
- ZSTD compression + LowCardinality optimization
- Bloom filter indexes สำหรับ barcode lookup

**ดูรายละเอียดทุกตารางที่ [clickhouse.md](clickhouse.md)**

## Kafka — Event Streaming

**บทบาท:** ส่งข้อมูลจาก MongoDB → PostgreSQL → ClickHouse ผ่าน events

**ลักษณะสำคัญ:**
- Broker: `kafka:29092` (internal Docker), `localhost:9092` (external)
- Auto-create topics
- Consumer groups ตามประเภทเอกสาร

**Topics หลัก:**

| Topic Pattern | เนื้อหา |
|--------------|---------|
| `{shopid}-sale-invoice` | ขาย |
| `{shopid}-sale-return` | รับคืน |
| `{shopid}-purchase` | ซื้อ |
| `{shopid}-purchase-return` | ส่งคืน |
| `{shopid}-stock-pickup` | เบิก |
| `{shopid}-stock-return` | รับคืนเบิก |
| `{shopid}-stock-receive` | รับ |
| `{shopid}-stock-transfer` | โอน |
| `{shopid}-stock-adjust` | ปรับยอด |
| `{shopid}-product-barcode` | ข้อมูลสินค้า |

## Redis — Cache

**บทบาท:** Cache สำหรับ session, token, และ temporary data

**ลักษณะสำคัญ:**
- Bind: `127.0.0.1:6379`
- ใช้สำหรับ MCP API key management
- Session cache สำหรับ user auth

## Data Flow

```
Frontend (POS/App)
    │
    ▼
MongoDB (Source of Truth) ──Kafka──▶ PostgreSQL (per-shop)
    │                                       │
    │                                       ▼
    └──────────────────────────────▶ ClickHouse (OLAP)
                                           │
                                           ▼
                                    MCP Tools / Dashboard
                                           │
                                           ▼
                                    AI (Claude Desktop)
```

1. **Frontend** → เขียนข้อมูลลง **MongoDB** (source of truth)
2. **Kafka Consumer** → อ่าน change events → เขียนลง **PostgreSQL** (per-shop relational)
3. **Stock Processor** → คำนวณ cost + stock balance → เขียนลง **ClickHouse** (OLAP)
4. **MCP Tools / Dashboard** → query จาก **ClickHouse** + **PostgreSQL**
5. **AI (Claude Desktop)** → เรียก **MCP Tools** ผ่าน SSE endpoint
