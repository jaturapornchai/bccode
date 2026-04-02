# Database Queries — SQL Patterns & Common Queries

## PostgreSQL Connection

```go
// Per-shop connection (most common)
db, err := mypg.PgSqlFastConnect(shopID) // DB name = shopID
defer db.Close()

// Always use parameterized queries — never string concatenation
// db.Query(`SELECT * FROM product WHERE shopid=$1`, shopID)
// NEVER: db.Query(`SELECT * FROM product WHERE shopid='` + shopID + `'`)
```

## Common Query Patterns

### List with Search + Pagination

```go
rows, err := db.Query(`
    SELECT guid, code, name, shopid
    FROM {table}
    WHERE shopid = $1
      AND ($2 = '' OR name ILIKE $3)
    ORDER BY name
    LIMIT $4 OFFSET $5
`, shopID, keyword, "%"+keyword+"%", limit, offset)

// Count total
var total int
db.QueryRow(`
    SELECT COUNT(*) FROM {table}
    WHERE shopid = $1 AND ($2 = '' OR name ILIKE $3)
`, shopID, keyword, "%"+keyword+"%").Scan(&total)
```

### Insert + Return ID

```go
var guid string
err = db.QueryRow(`
    INSERT INTO {table} (shopid, code, name, createdat)
    VALUES ($1, $2, $3, NOW())
    RETURNING guid
`, shopID, code, name).Scan(&guid)
```

### Upsert (Insert or Update)

```go
_, err = db.Exec(`
    INSERT INTO {table} (shopid, code, name)
    VALUES ($1, $2, $3)
    ON CONFLICT (shopid, code)
    DO UPDATE SET name = EXCLUDED.name, updatedat = NOW()
`, shopID, code, name)
```

### Update + Check Affected Rows

```go
result, err := db.Exec(`
    UPDATE {table} SET name=$3, updatedat=NOW()
    WHERE shopid=$1 AND guid=$2
`, shopID, guid, name)

rowsAffected, _ := result.RowsAffected()
if rowsAffected == 0 {
    return c.JSON(404, map[string]interface{}{
        "status": "error", "message": "Record not found",
    })
}
```

### Soft Delete

```go
_, err = db.Exec(`
    UPDATE {table} SET deletedat=NOW(), isdeleted=true
    WHERE shopid=$1 AND guid=$2
`, shopID, guid)
```

## Document Queries (table: doc + docdetail)

### Fetch Document + Line Items

```go
// Header
var doc DocModel
db.QueryRow(`
    SELECT docno, doctype, docstatus, docdatetime, totalamount, vatamount, grandtotal
    FROM doc
    WHERE shopid=$1 AND docno=$2
`, shopID, docNo).Scan(&doc.DocNo, &doc.DocType, &doc.DocStatus,
    &doc.DocDateTime, &doc.TotalAmount, &doc.VatAmount, &doc.GrandTotal)

// Detail
rows, _ := db.Query(`
    SELECT linenumber, itemcode, itemname, qty, price, amount
    FROM docdetail
    WHERE shopid=$1 AND docno=$2
    ORDER BY linenumber
`, shopID, docNo)
```

### Update Document Status

```go
_, err = tx.Exec(`
    UPDATE doc
    SET docstatus=$3, updatedat=NOW()
    WHERE shopid=$1 AND docno=$2 AND docstatus=$4
`, shopID, docNo, newStatus, expectedCurrentStatus)
```

## Stock Queries

### Check Stock Balance

```go
rows, err := db.Query(`
    SELECT sb.itemcode, p.name0, sb.whcode, sb.balanceqty
    FROM stockbalance sb
    JOIN product p ON p.shopid = sb.shopid AND p.itemcode = sb.itemcode
    WHERE sb.shopid = $1
      AND sb.balanceqty > 0
    ORDER BY p.name0
`, shopID)
```

## MongoDB Queries

```go
import (
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "goapi/myglobal"
)

db := myglobal.GetMongoDatabase()
collection := db.Collection("collection_name")

// Find one
var result bson.M
err := collection.FindOne(context.TODO(),
    bson.D{{Key: "shopid", Value: shopID}, {Key: "code", Value: code}},
).Decode(&result)

// Find many
cursor, err := collection.Find(context.TODO(),
    bson.D{{Key: "shopid", Value: shopID}},
)
var results []bson.M
cursor.All(context.TODO(), &results)

// Upsert
collection.UpdateOne(context.TODO(),
    bson.D{{Key: "shopid", Value: shopID}, {Key: "code", Value: code}},
    bson.D{{Key: "$set", Value: updateData}},
    options.Update().SetUpsert(true),
)
```

## ClickHouse Queries (Analytics only)

```go
import "goapi/myclickhouse"

ch, err := myclickhouse.Connect()
if err != nil { /* ... */ }
defer ch.Close()

rows, err := ch.Query(`
    SELECT toDate(event_time) AS date, count() AS cnt
    FROM analytics_events
    WHERE shopid = ? AND event_time >= ?
    GROUP BY date
    ORDER BY date
`, shopID, startDate)
```

## Scan Helper — Nullable Fields

```go
import "database/sql"

// For fields that may be NULL
var nullableField sql.NullString
rows.Scan(&nullableField)
value := nullableField.String // Returns "" if null

var nullableFloat sql.NullFloat64
rows.Scan(&nullableFloat)
amount := nullableFloat.Float64 // Returns 0 if null
```
