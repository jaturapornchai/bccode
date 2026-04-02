# Product Cost Calculation

## Method: Moving Weighted Average

The system uses **weighted average cost** — every time goods are received, the system recalculates the average between existing and new inventory.

---

## How It Works

### Example: Buying Pens

| Event | Qty | Price/Unit | Value | Balance Qty | Balance Value | Avg Cost |
|-------|-----|-----------|-------|-------------|---------------|----------|
| Purchase 1 | +100 | 10 | 1,000 | 100 | 1,000 | **10.00** |
| Purchase 2 | +50 | 12 | 600 | 150 | 1,600 | **10.67** |
| Sale | -30 | (cost 10.67) | 320 | 120 | 1,280 | **10.67** |
| Purchase 3 | +80 | 11 | 880 | 200 | 2,160 | **10.80** |

### Formula

```
Average Cost = Balance Value / Balance Qty
```

**Goods In (Receipt):**
```
New Balance Value = Old Balance Value + Purchase Value (actual price before VAT)
New Balance Qty   = Old Balance Qty + Purchase Qty
New Average Cost  = New Balance Value / New Balance Qty
```

**Goods Out (Sale/Issue):**
```
Outgoing Value    = Current Average Cost x Sold Qty
New Balance Value = Old Balance Value - Outgoing Value
New Balance Qty   = Old Balance Qty - Sold Qty
New Average Cost  = New Balance Value / New Balance Qty
```

---

## Documents Affecting Cost

### Goods In (cost = actual purchase price)

| Document | TransFlag | Value Used |
|----------|-----------|-----------|
| Opening Balance | 54 | Carried-forward value |
| Purchase Invoice | 12 | Purchase price before VAT |
| Purchase Receive (partial) | 310 | Purchase price before VAT |
| Sales Return | 48 | Cost from original sales invoice |
| Finished Goods Receipt | 60 | Specified value |
| Issue Return | 58 | Cost from original issue |
| Stock Adjustment + | 66 | Specified value |

### Goods Out (cost = current average cost)

| Document | TransFlag | Value Used |
|----------|-----------|-----------|
| Sales Invoice | 44 | Average cost x qty |
| Purchase Return | 16 | Average cost x qty |
| Stock Pickup (Issue) | 56 | Average cost x qty |
| Stock Adjustment - | 68 | Average cost x qty |
| Stock Transfer | 72 | Average cost x qty |

### Cost Adjustment Only (no quantity change)

| Document | TransFlag | Effect |
|----------|-----------|--------|
| Cost Adjustment + | 866 | Balance value increases -> avg cost increases |
| Cost Adjustment - | 868 | Balance value decreases -> avg cost decreases |

---

## Calculation Flow

```
 1. Save document (purchase/sale/receipt/issue)
    -> Store in MongoDB
         |
         v
 2. Publish via Kafka
    -> Consumer writes to PostgreSQL
    -> Add itemcode to stockwaitprocess (calculation queue)
         |
         v
 3. Cost Calculation (Batch Processing)
    -> Fetch all transactions for the item, sorted by date + line number
    -> Process each transaction:
         |
         +-- If "in"  -> Use actual purchase price (SumAmount)
         +-- If "out" -> Use average cost x qty
         |
         +-- Calculate: balance qty + balance value
         +-- Calculate: new avg cost = balance value / balance qty
         |
         v
 4. Save Results
    -> processstockcost table (every transaction)
    -> Update productbarcode (latest average cost)
    -> Write to ClickHouse (for reports)
```

---

## Special Cases

### Zero Stock (balance qty = 0)
```
-> Force: balance value = 0, average cost = 0
-> Prevents orphan values
-> Next purchase -> cost starts fresh from new purchase price
```

### Returns
```
-> If reference document exists (DocRef) -> use cost from original document
-> If no DocRef -> use current average cost
```

### Warehouse Transfer
```
Source warehouse (out): average cost x qty
Destination warehouse (in): receives same cost (no recalculation)
```

### Multi-Unit Conversion
```
Product "Water" sold by "dozen" but stocked by "bottle"
-> Convert to standard unit before calculation:
   Standard qty = qty x (UnitStand / UnitDivide)
   Example: 2 dozen x (12 / 1) = 24 bottles
```

---

## Related Database Tables

### `productbarcode` (MongoDB + PostgreSQL) — Current Product Data

| Field | Description |
|-------|-------------|
| `balanceqty` | Current balance quantity |
| `balanceamount` | Current balance value |
| `averagecost` | Current average cost |

### `docdetail` (PostgreSQL) — All Document Details

| Field | Description |
|-------|-------------|
| `docdatetime` | Document datetime (sort key) |
| `itemcode` | Item code |
| `transflag` | Document type (see table above) |
| `calcflag` | 1=stock increase, 2=stock decrease |
| `totalqty` | Quantity (converted to standard unit) |
| `sumamount` | Total value (price before VAT) |

### `processstockcost` (PostgreSQL) — Cost Calculation Results

| Field | Description |
|-------|-------------|
| `itemcode` | Item code |
| `transflag` | Document type |
| `totalqty` | Quantity changed |
| `calcamount` | Calculated value |
| `averagecost` | Average cost at that point |
| `balanceqty` | Balance after calculation |
| `balanceamount` | Balance value after calculation |
| `unitcost` | Cost per unit |

### `stock_transaction_detail` (PostgreSQL) — Synced from MongoDB

| Field | Description |
|-------|-------------|
| `costperunit` | Cost per unit |
| `totalcost` | Total cost |
| `balanceqty` | Balance quantity |
| `balanceamount` | Balance value |
| `balanceaverage` | Average cost balance |

---

## Key Code Files

### Cost Calculation Engine

| File | Purpose |
|------|---------|
| `backend/pkg/stockcalculator/stockcalculator.go` | Main engine: `ApplyStock()` (in) + `ReduceStock()` (out) |
| `backend/internal/goapi/process/process-stock/process-stock-calc-cost.go` | Batch processing: calculate per itemcode |
| `backend/internal/goapi/process/build/create-database.go` | Create `docdetail`, `processstockcost` tables |

### Consumers Affecting Cost

| File | TransFlag | Effect |
|------|-----------|--------|
| `backend/internal/transaction/transactionconsumer/purchasereceive/` | 310 | Goods receipt -> stock increase + cost recalculation |
| `backend/internal/transaction/transactionconsumer/sale/` | 44 | Sale -> stock decrease + use average cost |
| `backend/internal/transaction/transactionconsumer/stockadjustment/` | 66/68 | Stock adjustment +/- |
| `backend/internal/transaction/transactionconsumer/stocktransfer/` | 72 | Warehouse transfer |

### Frontend

| File | Purpose |
|------|---------|
| `frontend/bcaiaccount/lib/model/transaction_model.dart` | Has `averagecost` field in Detail |

---

## Summary

1. **Goods in** -> use actual purchase price -> recalculate average cost
2. **Goods out** -> use current average cost x qty
3. **Zero stock** -> cost resets to 0
4. **New purchase after zero stock** -> cost = new purchase price
5. **Transfer** -> source warehouse cost passes to destination
6. **Cost adjustment** -> change value without changing quantity
