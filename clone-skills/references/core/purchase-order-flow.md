# Purchase Order (PO) Flow

## What is a Purchase Order?

A PO documents the intent to buy goods from a supplier — specifying items, quantities, and prices.

**A PO has no effect on stock or AP** — it is only a commitment. Stock increases only after a Goods Receipt is created.

---

## Main Workflow — From Order to Payment

```
 1. Create PO
    "What to buy, from whom, how much"
         |
         v
 2. Submit for Approval
         |
    +----+----+
    v         v
 Approved   Rejected
    |         |
    |         v
    |    Edit -> Resubmit
    |
    v
 3. Goods Receipt (Purchase Receive)
    "Goods received and inspected"
    -> Stock increases
    -> Average cost recalculated
         |
         v
 4. AP Accrual
    "Record liability to supplier"
    -> AP balance increases
         |
         v
 5. Creditor Payment
    "Pay per credit terms"
    -> AP balance decreases
```

### Impact Summary per Step

| Step | Stock | AP | Cash |
|------|-------|----|------|
| 1. Create PO | No change | No change | No change |
| 2. Approve PO | No change | No change | No change |
| 3. Goods Receipt | **Increase** | No change | No change |
| 4. AP Accrual | No change | **Increase** | No change |
| 5. Payment | No change | **Decrease** | **Decrease** |

---

## Related Documents

```
Purchase Order (PO)       TransFlag=6    <- Commitment only, no effect
     |
     +---> Goods Receipt (PP)  TransFlag=310  <- Stock increases
     |         |
     |         +---> AP Accrual (AP) TransFlag=12  <- Record liability
     |
     +---> Direct Purchase (PU) TransFlag=12  <- Buy + receive immediately

Purchase Return (PT)      TransFlag=16  <- Return to supplier, stock decreases
```

---

## PO Statuses

### Approval Status

| Status | Badge | Meaning |
|--------|-------|---------|
| Not submitted | (none) | Created, not yet sent for approval |
| Auto-approved | `auto_approved` | No approval rule or creator has sufficient authority |
| Pending | `pending` (yellow) | Awaiting approver action |
| Approved | `approved` (green) | Can create Goods Receipt |
| Rejected | `rejected` (red) | Returned — can edit and resubmit |

### Receiving Status

| Status | Badge | Meaning |
|--------|-------|---------|
| Not received | (none) | No Goods Receipt yet |
| Fully received | Green | All items received per PO |
| Partially received | Yellow | Some items/quantities received |
| Over-received | Yellow | Received more than ordered |
| Manually closed | Purple | Remaining items will not be received (with reason) |

### Allowed Actions per Status

| Status | Edit | Delete | Print | Submit |
|--------|------|--------|-------|--------|
| Not submitted | Yes | Yes | No | Yes |
| Auto (unreferenced) | Yes | Yes | Yes | -- |
| Auto (referenced) | No | No | Yes | -- |
| Pending | No | No | No | No (withdraw first) |
| Approved | No | No | Yes | -- |
| Rejected | Yes | No | No | Yes (resubmit) |

**Rule:** If any approval level has been completed (`currentApprovedLevel > 0`), editing is blocked.

---

## PO Data Structure

### Header

| Field | Example | Notes |
|-------|---------|-------|
| Doc number | PO2026030800001 | Auto-generated: PO + date + running |
| Date | 2026-03-08 | Document creation date |
| Supplier | ABC Co., Ltd. | Selected from Creditor list |
| Purchase type | General | Determines approval rule |
| VAT | Include VAT 7% | 0=None, 1=Included, 2=Excluded, 3=Exempt |
| Credit term | 30 days | Payment due date |
| Currency | THB / USD | Multi-currency + exchange rate |
| Withholding tax | 3% | Supported rates: 0.5, 1, 2, 3, 5, 10, 15% |

### Line Items

| Field | Example | Notes |
|-------|---------|-------|
| Item code | P001 | Search by barcode or code |
| Item name | A4 Paper | Multi-language |
| Quantity | 100 | Must be > 0 |
| Unit | Ream | Multi-unit supported |
| Unit price | 120.00 | Must be >= 0 |
| Discount | 10%+5% | Stacked discounts |
| Line total | 10,260.00 | Auto-calculated |
| Warehouse | WH01 | Receiving warehouse |

---

## Validation

### Frontend (before save)
- Supplier required
- At least 1 line item
- Quantity > 0, Price >= 0
- Grand total > 0

### Backend (3-layer)
1. **Format check** — required fields not empty
2. **Business logic** — date, supplier, VAT type, no NaN/Infinity
3. **Auto-fix** — exchange rate <= 0 corrected to 1.0

### Permission check
- `pending` or `approved` -> editing blocked (must reject/withdraw first)
- Any approval level completed -> editing absolutely blocked

---

## System Internals (Developer Reference)

### Document Number (DocNo)
```
PO + YYYYMMDD + 00001 (5-digit running)
Example: PO2026030800001
```
- Running number stored in Redis cache
- Fallback: query MongoDB for latest number

### Data Flow — Where Data Lives

The system uses 3 databases, each with a specific role:

```
 User saves PO
      |
      v
 +-----------------------------------------------------------+
 |  MongoDB (Primary)                                         |
 |  Collection: transactionPurchaseOrder                      |
 |  -> Stores complete PO (header + details + approval)       |
 |  -> Used for CRUD via API                                  |
 |  -> Frontend reads/writes here                             |
 +------------------------+----------------------------------+
                          | Kafka Message
                          | (when-purchaseorder-created/updated/deleted)
                          v
 +-----------------------------------------------------------+
 |  PostgreSQL (Secondary — queries/reports)                   |
 |  Table: purchase_order_transaction                          |
 |  Table: purchase_order_transaction_detail                   |
 |  -> Consumer converts MongoDB doc -> PG row                 |
 |  -> Used for JOIN, aggregate, reports                       |
 |  -> GoAPI fetches PO list from here (getdoc)                |
 +------------------------+----------------------------------+
                          | Async sync
                          v
 +-----------------------------------------------------------+
 |  ClickHouse (Analytics — dashboard/BI)                      |
 |  -> Aggregated data                                         |
 |  -> Purchase volume reports, supplier analysis              |
 |  -> MCP tools read from here (dashboard KPIs)               |
 +-----------------------------------------------------------+
```

### Database Usage Summary

| Operation | Database | Why |
|-----------|----------|-----|
| Create/Edit/Delete PO | **MongoDB** | Primary storage, flexible schema |
| List POs (getdoc) | **PostgreSQL** | Fast filter/sort/pagination |
| Get PO by GUID | **MongoDB** | Direct primary read |
| Purchase volume reports | **ClickHouse** | Fast aggregation |
| MCP tools (dashboard) | **ClickHouse** | Optimized for analytics |
| Approval status | **MongoDB** (GoAPI) | Separate collection |

### Kafka Topics

| Topic | Trigger |
|-------|---------|
| `when-purchaseorder-created` | PO created |
| `when-purchaseorder-updated` | PO updated |
| `when-purchaseorder-deleted` | PO deleted |
| `when-purchaseorder-bulk-created` | Bulk create |
| `when-purchaseorder-bulk-updated` | Bulk update |
| `when-purchaseorder-bulk-deleted` | Bulk delete |

### Flow When PO is Referenced (Goods Receipt)

```
 PO Approved
      |
      v
 Create Purchase Receive referencing PO
      |
      +---> MongoDB: transactionPurchasePartial (stores receipt)
      |
      +---> Kafka: when-purchasereceive-created
      |         |
      |         v
      |    PostgreSQL: purchase_receive_transaction (sync)
      |         |
      |         v
      |    PostgreSQL: docdetail (for cost calculation)
      |         |
      |         v
      |    PostgreSQL: stockwaitprocess (calculation queue)
      |         |
      |         v
      |    PostgreSQL: processstockcost (cost calculation result)
      |         |
      |         v
      |    PostgreSQL: productbarcode (update stock + averagecost)
      |         |
      |         v
      |    ClickHouse: stock data (async sync for reports)
      |
      +---> MongoDB: PO.isref = true (mark PO as referenced)
      |
      +---> Create AP Accrual Receive
                |
                +---> MongoDB: transactionAccrualReceive
                +---> Kafka -> PostgreSQL: ap_purchasereceive_transaction
                +---> PostgreSQL: creditor balance increases
```

---

## Database Schema — Complete Fields (for Reports/BI)

### MongoDB Collection: `transactionPurchaseOrder`

Stores complete PO as a single JSON document — uses `Transaction` model from Go backend.

### PostgreSQL: `purchase_order_transaction` (Header)

Embeds `TransactionPG` base struct + additional fields.

| Column | Type | Description |
|--------|------|-------------|
| `shopid` | TEXT | Shop ID |
| `parid` | TEXT | Partition ID |
| `guidfixed` | TEXT | Fixed GUID (PK) |
| `transflag` | INT2 | Document type (6=PO) |
| `docno` | TEXT | Document number |
| `docdate` | TIMESTAMPTZ | Document date |
| `docreftype` | INT2 | Reference document type |
| `docrefno` | TEXT | Reference document number |
| `docrefdate` | TIMESTAMPTZ | Reference document date |
| `branchcode` | TEXT | Branch code |
| `branchnames` | JSONB | Branch names (multi-lang) |
| `description` | TEXT | Description/remarks |
| `taxdocno` | TEXT | Tax document number |
| `taxdocdate` | TIMESTAMPTZ | Tax document date |
| `creditorcode` | TEXT | **Creditor/supplier code** |
| `creditornames` | JSONB | **Creditor names (multi-lang)** |
| `iscancel` | BOOLEAN | Cancelled |
| `isbom` | BOOLEAN | Is BOM |
| `status` | INT2 | Document status |
| `vattype` | INT2 | VAT type (0/1/2/3) |
| `vatrate` | FLOAT8 | VAT rate % |
| `totalvalue` | FLOAT8 | Total value |
| `discountword` | TEXT | Discount text (e.g. "10%+5%") |
| `totaldiscount` | FLOAT8 | Total discount |
| `deliveryamount` | FLOAT8 | Delivery charge |
| `totalbeforevat` | FLOAT8 | Total before VAT |
| `totalvatvalue` | FLOAT8 | VAT amount |
| `totalexceptvat` | FLOAT8 | Total VAT-exempt |
| `totalaftervat` | FLOAT8 | Total after VAT |
| `totalamount` | FLOAT8 | **Grand total** |
| `guidref` | TEXT | Reference GUID |
| `guidpos` | TEXT | POS GUID |
| `devicename` | TEXT | Device name |
| `inquirytype` | INT | Inquiry type |
| `ismanualamount` | BOOLEAN | Manual amount entry |
| `pointdiscountamount` | FLOAT8 | Points discount |
| `paypointamount` | FLOAT8 | Points paid |
| `alcoholamount` | FLOAT8 | Alcohol value |
| `otheramount` | FLOAT8 | Other value |
| `drinkamount` | FLOAT8 | Beverage value |
| `foodamount` | FLOAT8 | Food value |

### PostgreSQL: `purchase_order_transaction_detail` (Detail)

Embeds `TransactionDetailPG` base struct.

| Column | Type | Description |
|--------|------|-------------|
| `id` | SERIAL | PK (auto) |
| `shopid` | TEXT | Shop ID |
| `guidfixed` | TEXT | GUID (FK -> header) |
| `parid` | TEXT | Partition ID |
| `docno` | TEXT | Document number |
| `docdate` | TIMESTAMPTZ | Document date |
| `linenumber` | INT2 | Line number |
| `barcode` | TEXT | Barcode |
| `itemcode` | TEXT | Item code (via itemguid) |
| `itemguid` | TEXT | Item GUID |
| `itemnames` | JSONB | Item names (multi-lang) |
| `itemtype` | INT2 | Item type |
| `unitcode` | TEXT | Unit code |
| `unitnames` | JSONB | Unit names (multi-lang) |
| `qty` | FLOAT8 | Quantity |
| `price` | FLOAT8 | Unit price |
| `priceexcludevat` | FLOAT8 | Price excl. VAT |
| `discount` | TEXT | Discount text |
| `discountamount` | FLOAT8 | Discount amount |
| `sumamount` | FLOAT8 | **Line total** |
| `sumamountexcludevat` | FLOAT8 | Line total excl. VAT |
| `sumamountchoice` | FLOAT8 | Choice total |
| `totalvaluevat` | FLOAT8 | Line VAT amount |
| `vattype` | INT2 | Line VAT type |
| `taxtype` | INT2 | Other tax type |
| `vatcal` | INT2 | VAT calculation method |
| `whcode` | TEXT | Warehouse code |
| `whnames` | JSONB | Warehouse names (multi-lang) |
| `locationcode` | TEXT | Location code |
| `locationnames` | JSONB | Location names (multi-lang) |
| `standvalue` | FLOAT8 | Unit standard value (for multi-unit) |
| `dividevalue` | FLOAT8 | Unit divide value |
| `groupcode` | TEXT | Product group code |
| `groupnames` | JSONB | Group names (multi-lang) |
| `refguid` | TEXT | Reference GUID |
| `docref` | TEXT | Reference document number |
| `docrefdatetime` | TIMESTAMPTZ | Reference document date |
| `remark` | TEXT | Line remark |
| `ischoice` | INT2 | Is choice item |
| `foodtype` | INT2 | Food type |
| `whcodedestination` | TEXT | Destination warehouse code |
| `whcodedestinationnames` | JSONB | Destination warehouse names |
| `locationcodedestination` | TEXT | Destination location code |
| `locationdestination` | JSONB | Destination location names |

### PostgreSQL: Purchase Receive + AP Tables (same structure)

| Table | Purpose | Extra Fields |
|-------|---------|-------------|
| `purchasereceive_transaction` | Goods Receipt header | `creditorcode`, `creditornames` |
| `purchasereceive_transaction_detail` | Goods Receipt detail | (same as PO detail) |
| `ap_purchasereceive_transaction` | AP Accrual header | `creditorcode`, `creditornames` |
| `ap_purchasereceive_transaction_detail` | AP Accrual detail | (same as PO detail) |

### PostgreSQL (GoAPI): `docdetail` — Stock/Cost Calculation Input

| Column | Type | Description |
|--------|------|-------------|
| `id` | SERIAL | PK |
| `docdatetime` | TIMESTAMPTZ | Document datetime (sort key) |
| `docno` | TEXT | Document number |
| `docref` | TEXT | Reference number |
| `description` | TEXT | Description |
| `linenumber` | INT | Line number |
| `transflag` | INT | Document type |
| `calcflag` | INT | 1=stock increase, 2=stock decrease |
| `calcseq` | INT | Calculation sequence |
| `iscancel` | BOOLEAN | Cancelled (default false) |
| `itemcode` | TEXT | Main item code |
| `barcodemain` | TEXT | Main barcode |
| `barcode` | TEXT | Sub-barcode |
| `unitcode` | TEXT | Unit code |
| `whcode` | TEXT | Warehouse |
| `locationcode` | TEXT | Location |
| `totalqty` | NUMERIC(18,8) | Quantity |
| `unitstand` | NUMERIC(18,8) | Unit standard value |
| `unitdivide` | NUMERIC(18,8) | Unit divide value |
| `price` | NUMERIC(18,2) | Price |
| `priceexcludevat` | NUMERIC(18,2) | Price excl. VAT |
| `sumamount` | NUMERIC(18,2) | **Total amount (used for cost calculation)** |
| `isupdated` | BOOLEAN | Updated flag (default false) |
| `iscalcstock` | INT | Stock calculated (default 0) |
| `price_doc` | NUMERIC(18,2) | Price (document currency) |
| `sumamount_doc` | NUMERIC(18,2) | Amount (document currency) |
| `discountamount_doc` | NUMERIC(18,2) | Discount (document currency) |
| `priceexcludevat_doc` | NUMERIC(18,2) | Price excl. VAT (document currency) |
| `sumamountexcludevat_doc` | NUMERIC(18,2) | Amount excl. VAT (document currency) |
| `totalvaluevat_doc` | NUMERIC(18,2) | VAT amount (document currency) |

### PostgreSQL (GoAPI): `processstockcost` — Cost Calculation Results

| Column | Type | Description |
|--------|------|-------------|
| `id` | SERIAL | PK |
| `docdatetime` | TIMESTAMPTZ | Document datetime |
| `docno` | TEXT | Document number |
| `docref` | TEXT | Reference number |
| `linenumber` | INT | Line number |
| `transflag` | INT | Document type |
| `itemcode` | TEXT | Item code |
| `barcode` | TEXT | Barcode |
| `unitcode` | TEXT | Standard unit |
| `whcode` | TEXT | Warehouse |
| `locationcode` | TEXT | Location |
| `totalqty` | NUMERIC(18,8) | Quantity changed |
| `unitstand` | NUMERIC(18,8) | Unit standard value |
| `unitdivide` | NUMERIC(18,8) | Unit divide value |
| `price` | NUMERIC(18,2) | Price |
| `averagecost` | NUMERIC(18,2) | **Average cost at that time** |
| `calcamount` | NUMERIC(18,2) | **Calculated amount** |
| `balanceqty` | NUMERIC(18,8) | **Balance quantity** |
| `balanceamount` | NUMERIC(18,2) | **Balance amount** |
| `unitcost` | NUMERIC(18,2) | Cost per unit |
| `guid` | TEXT | Reference GUID |

### ClickHouse: `docdetail` — Analytics Line Items

Engine: `MergeTree` PARTITION BY `shopid` ORDER BY `(shopid, docno, line_number)`

| Column | Type | Description |
|--------|------|-------------|
| `shopid` | String | Shop ID |
| `docno` | String | Document number |
| `docdatetime` | DateTime | Datetime |
| `perioddatetime` | DateTime | Period date |
| `line_number` | UInt32 | Line number |
| `transflag` | Int16 | Document type |
| `calcflag` | Int8 | Calculation flag |
| `calcseq` | Int32 | Calculation sequence |
| `guidfixed` | String | Document GUID |
| `guidpos` | String | POS GUID |
| `guidbranch` | String | Branch GUID |
| `branchid` | String | Branch ID |
| `itemcode` | Nullable(String) | Item code |
| `barcode` | String | Barcode |
| `barcodemain` | String | Main barcode |
| `itemname` | String | Item name |
| `itemnames` | String | Item names (JSON) |
| `unitcode` | String | Unit code |
| `unitstand` | Float64 | Standard value (default 1.0) |
| `unitdivide` | Float64 | Divide value (default 1.0) |
| `qty` | Float64 | Quantity |
| `price` | Float64 | Price |
| `discount` | String | Discount |
| `discountamount` | Float64 | Discount amount |
| `sumamount` | Float64 | Line total |
| `sumofcost` | Float64 | **Total cost** |
| `whcode` | String | Warehouse |
| `locationcode` | String | Location |
| `refguid` | String | Reference GUID |
| `ischoice` | Int32 | Choice item |
| `sumamountchoice` | Float64 | Choice amount |
| `isupdated` | Bool | Updated |
| `iscalcstock` | Int8 | Stock calculated |
| `price_doc` | Float64 | Price (document currency) |
| `sumamount_doc` | Float64 | Amount (document currency) |
| `discountamount_doc` | Float64 | Discount (document currency) |
| `priceexcludevat_doc` | Float64 | Price excl. VAT (document currency) |
| `sumamountexcludevat_doc` | Float64 | Amount excl. VAT (document currency) |
| `totalvaluevat_doc` | Float64 | VAT amount (document currency) |

### ClickHouse: `processstockcost` — Analytics Cost Data

Engine: `MergeTree` PARTITION BY `(shopid, itemcode)` ORDER BY `(shopid, itemcode, docdatetime, whcode, locationcode)`
CODEC: ZSTD(1) all columns — bloom_filter on `docno`+`barcode`, minmax on `docdatetime`

| Column | Type | Description |
|--------|------|-------------|
| `shopid` | LowCardinality(String) | Shop ID |
| `itemcode` | LowCardinality(String) | Item code |
| `docdatetime` | DateTime | Datetime |
| `docno` | String | Document number |
| `docref` | String | Reference number |
| `linenumber` | UInt16 | Line number |
| `transflag` | UInt8 | Document type |
| `barcodemain` | String | Main barcode |
| `barcode` | String | Barcode |
| `unitcode` | LowCardinality(String) | Unit code |
| `whcode` | LowCardinality(String) | Warehouse |
| `locationcode` | LowCardinality(String) | Location |
| `totalqty` | Float64 | Quantity |
| `unitstand` | Float64 | Standard value |
| `unitdivide` | Float64 | Divide value |
| `price` | Float64 | Price |
| `averagecost` | Float64 | **Average cost** |
| `calcamount` | Float64 | **Calculated amount** |
| `balanceqty` | Float64 | **Balance quantity** |
| `balanceamount` | Float64 | **Balance amount** |
| `unitcost` | Float64 | Cost per unit |
| `guid` | String | GUID |
| `originalqty` | Int32 | Original quantity |
| `qty` | Int32 | Quantity |

---

### API Endpoints

| Action | Method | Path |
|--------|--------|------|
| Create PO | POST | `/transaction/purchase-order` |
| Update PO | PUT | `/transaction/purchase-order/:id` |
| Delete PO | DELETE | `/transaction/purchase-order/:id` |
| List POs | POST | `/goapi/getdoc` (system=purchase-order) |
| Submit approval | POST | `/goapi/api/approval/po-status/submit` |
| Withdraw approval | POST | `/goapi/api/approval/po-status/withdraw` |
| Manual close | POST | `/goapi/api/purchase-order/manual-close` |

### Key Code Files

**Backend (Go):**
| File | Purpose |
|------|---------|
| `internal/transaction/purchaseorder/purchaseorder_http.go` | All API routes |
| `internal/transaction/purchaseorder/services/purchaseorder_http_service.go` | Business logic + doc number generation |
| `internal/transaction/purchaseorder/validators/po_validator.go` | 3-layer validation |
| `internal/transaction/models/transaction.go` | Core data structure |
| `internal/transaction/purchasepartial/` | Goods Receipt |
| `internal/transaction/accrualreceive/` | AP Accrual |

**Frontend (Flutter):**
| File | Purpose |
|------|---------|
| `lib/screens/purchaseorder/purchaseorder_edit_screen.dart` | Create/edit screen |
| `lib/screens/purchaseorder/purchaseorder_list_screen.dart` | List screen |
| `lib/screens/purchaseorder/utils/po_workflow_manager.dart` | Flow management |
| `lib/screens/purchaseorder/utils/po_approval_helper.dart` | Status-based permissions |
| `lib/screens/purchaseorder/utils/po_save_helper.dart` | Data preparation before save |
| `lib/screens/purchaseorder/components/po_list_card.dart` | Card + status badges |
| `lib/bloc/trans/trans_bloc.dart` | State management (shared with other transactions) |
