# Database Schema — PostgreSQL Business Tables

## Architecture

- **Multi-tenant**: Each shop has a separate PostgreSQL database named by `shopid`
- **Connection**: `mypg.PgSqlFastConnect(shopID)` — connects directly to that shop's database
- **Every table has `shopid`** even within separate databases (for safety + query compatibility)

---

## Core Tables

### `doc` — All Document Types

All business documents stored in a single table, differentiated by `transflag`.

| Column | Type | Description |
|--------|------|-------------|
| `shopid` | varchar | Shop ID (multi-tenant key) |
| `docno` | varchar | Document number (composite PK with transflag) |
| `transflag` | int | Document type (see TransFlag table below) |
| `docdatetime` | timestamptz | Document datetime |
| `perioddatetime` | timestamptz | Accounting period date |
| `custcode` | varchar | Customer/supplier code |
| `totalamount` | numeric | Total before tax |
| `vatamount` | numeric | VAT amount |
| `grandtotal` | numeric | Grand total |
| `roundamount` | numeric | Rounding amount |
| `discountamount` | numeric | Discount |
| `paytype` | int | Payment type (1=Cash, 2=Credit) |
| `paycashamount` | numeric | Cash received |
| `paycashchange` | numeric | Change given |
| `branchid` | varchar | Branch ID |
| `salechannelcode` | varchar | Sales channel |
| `iscancel` | boolean | Cancelled |
| `isdelete` | boolean | Soft deleted |
| `cancelreason` | varchar | Cancellation reason |
| `taxdocno` | varchar | Tax invoice number |
| `guidfixed` | varchar | Document GUID |
| `creator_code` | varchar | Creator code |
| `creator_name` | varchar | Creator name |
| `created_at` | timestamptz | Created datetime |
| `currency` | varchar | Base currency (for accounting) |
| `doccurrency` | varchar | Document currency |
| `exchangerate` | numeric | Exchange rate |

**TransFlag — Document Types:**

| transflag | Type |
|-----------|------|
| 1 | Quotation |
| 2 | Sales Order |
| 3 | Sales Invoice |
| 8 | Sales Credit Note |
| 12 | Purchase Invoice |
| 16 | Purchase Return |
| 20 | Purchase Order |
| 21 | Purchase Requisition |
| 22 | Request for Quotation |
| 30 | Stock Adjustment |
| 31 | Stock Transfer |

**Example query:**
```sql
SELECT docno, docdatetime, custcode, grandtotal
FROM doc
WHERE shopid = $1
  AND transflag = $2
  AND iscancel = false
  AND (isdelete = false OR isdelete IS NULL)
  AND docdatetime >= $3
ORDER BY docdatetime DESC
LIMIT $4 OFFSET $5
```

---

### `docdetail` — Document Line Items

| Column | Type | Description |
|--------|------|-------------|
| `shopid` | varchar | Shop ID |
| `docno` | varchar | Document number (FK -> doc.docno) |
| `transflag` | int | Document type (FK -> doc.transflag) |
| `linenumber` | int | Line number |
| `barcode` | varchar | Barcode |
| `itemcode` | varchar | Item code |
| `unitcode` | varchar | Unit code |
| `totalqty` | numeric | Quantity |
| `price` | numeric | Unit price |
| `sumamount` | numeric | Line total |
| `discountamount` | numeric | Discount |
| `calcflag` | numeric | +1 (stock in) / -1 (stock out) |
| `iscalcstock` | smallint | 1=calculate stock, 0=skip |
| `unitstand` | numeric | Unit standard ratio |
| `unitdivide` | numeric | Unit divisor |
| `whcode` | varchar | Warehouse code |
| `locationcode` | varchar | Location code |
| `towhcode` | varchar | Destination warehouse (for transfers) |
| `tolocationcode` | varchar | Destination location |
| `description` | text | Line description |

**Example query (doc JOIN docdetail):**
```sql
SELECT d.docno, d.docdatetime, dd.barcode, dd.itemcode, dd.totalqty, dd.price
FROM docdetail dd
INNER JOIN doc d ON dd.docno = d.docno AND dd.transflag = d.transflag
WHERE dd.barcode = $1
  AND d.transflag = $2
  AND d.iscancel = false
  AND (d.isdelete = false OR d.isdelete IS NULL)
ORDER BY d.docdatetime DESC
```

---

### `product` — Products

| Column | Type | Description |
|--------|------|-------------|
| `shopid` | varchar | Shop ID |
| `itemcode` | varchar | Item code (composite PK with shopid) |
| `name0` | varchar | Product name (primary language) |
| `name1` | varchar | Product name (English) |
| `groupcode` | varchar | Product group code |
| `unitcode` | varchar | Base unit |
| `isdelete` | boolean | Soft deleted |
| `isinactive` | boolean | Inactive |

---

### `productbarcode` — Barcodes

| Column | Type | Description |
|--------|------|-------------|
| `shopid` | varchar | Shop ID |
| `barcode` | varchar | Barcode (composite PK with shopid) |
| `itemcode` | varchar | Item code (FK -> product.itemcode) |
| `unitcode` | varchar | Unit |
| `unitstand` | numeric | Unit standard ratio |
| `unitdivide` | numeric | Unit divisor |
| `price1` | numeric | Selling price 1 |
| `price2` | numeric | Selling price 2 |
| `price3` | numeric | Selling price 3 |
| `barcodemain` | varchar | Main barcode for the product |

---

### `stockbalance` — Stock Balance

| Column | Type | Description |
|--------|------|-------------|
| `shopid` | varchar | Shop ID |
| `itemcode` | varchar | Item code |
| `whcode` | varchar | Warehouse code |
| `locationcode` | varchar | Location code |
| `balance` | numeric | Balance (in standard unit) |

> Some systems calculate balance directly from `docdetail` instead of using `stockbalance`.

---

### `stockcard` — Stock Movement History

| Column | Type | Description |
|--------|------|-------------|
| `shopid` | varchar | Shop ID |
| `docno` | varchar | Document number |
| `transflag` | int | Document type |
| `docdatetime` | timestamptz | Datetime |
| `itemcode` | varchar | Item code |
| `barcode` | varchar | Barcode |
| `whcode` | varchar | Warehouse code |
| `locationcode` | varchar | Location code |
| `qty` | numeric | Quantity (+receive, -issue) |
| `balance` | numeric | Balance after transaction |
| `price` | numeric | Cost price |

---

### `customer` — Customers

| Column | Type | Description |
|--------|------|-------------|
| `shopid` | varchar | Shop ID |
| `custcode` | varchar | Customer code (composite PK with shopid) |
| `name0` | varchar | Customer name |
| `name1` | varchar | Customer name (EN) |
| `address` | text | Address |
| `tel` | varchar | Phone |
| `taxid` | varchar | Tax ID |
| `creditlimit` | numeric | Credit limit |
| `creditterm` | int | Credit term (days) |
| `isdelete` | boolean | Soft deleted |

---

### `supplier` — Suppliers/Creditors

Same structure as `customer` but table name is `supplier`.

---

## MongoDB Collections

MongoDB stores denormalized data for read-heavy operations.

| Collection | Data |
|-----------|------|
| `productbarcode` | Products + barcodes + prices + all units |
| `customer` | Customer data with balance |
| `supplier` | Supplier/creditor data |
| `employee` | Employees |
| `warehouse` | Warehouses + locations |
| `trans` | Denormalized transaction data |

Every MongoDB document has `shopid` and `guidfixed` fields.

---

## ClickHouse Tables (Analytics)

| Table | Data |
|-------|------|
| `sales_logs` | All transaction records for analytics |
| `stock_logs` | Stock movement records |

Used for **read-only analytics** only — never write from regular handlers.

---

## Tips

- Always use `AND (isdelete = false OR isdelete IS NULL)` — not just `AND isdelete = false`
- `docno + transflag` is the composite key for doc and docdetail
- Stock balance = `SUM((totalqty * calcflag) * unitstand / unitdivide)` from docdetail where `iscalcstock = 1`
