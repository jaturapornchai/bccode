# Handlers — Templates และ Patterns

## Handler Function Template (ครบทุกขั้นตอน)

```go
package handlers

import (
    "net/http"

    "goapi/logger"
    mypg "goapi/mypg"

    "github.com/labstack/echo/v4"
)

// Request struct — json tag ตรงกับที่ Flutter ส่งมา
type FeatureListRequest struct {
    ShopID  string `json:"shopid"`
    Keyword string `json:"keyword"`
    Limit   int    `json:"limit"`
    Offset  int    `json:"offset"`
}

// Response item struct
type FeatureItem struct {
    GUID    string `json:"guid"`
    Code    string `json:"code"`
    Name    string `json:"name"`
    ShopID  string `json:"shopid"`
}

func FeatureListHandler(c echo.Context) error {
    // 1. Parse request
    var req FeatureListRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "status": "error", "message": "Invalid request body",
        })
    }

    // 2. Validate required fields
    if req.ShopID == "" {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "status": "error", "message": "Missing required parameter: shopid",
        })
    }

    // 3. Default values
    if req.Limit <= 0 { req.Limit = 20 }
    if req.Limit > 100 { req.Limit = 100 }

    logger.Info("[FeatureList] shopid=%s keyword=%s", req.ShopID, req.Keyword)

    // 4. Connect DB
    db, err := mypg.PgSqlFastConnect(req.ShopID)
    if err != nil {
        logger.Error("[FeatureList] DB connect failed: %v", err)
        return c.JSON(http.StatusInternalServerError, map[string]interface{}{
            "status": "error", "message": "Database connection failed",
        })
    }
    defer db.Close()

    // 5. Query
    rows, err := db.Query(`
        SELECT guid, code, name, shopid
        FROM feature
        WHERE shopid = $1
          AND ($2 = '' OR name ILIKE $3)
        ORDER BY name
        LIMIT $4 OFFSET $5
    `, req.ShopID, req.Keyword, "%"+req.Keyword+"%", req.Limit, req.Offset)
    if err != nil {
        logger.Error("[FeatureList] Query failed: %v", err)
        return c.JSON(http.StatusInternalServerError, map[string]interface{}{
            "status": "error", "message": "Query failed",
        })
    }
    defer rows.Close()

    // 6. Scan results
    var items []FeatureItem
    for rows.Next() {
        var item FeatureItem
        if err := rows.Scan(&item.GUID, &item.Code, &item.Name, &item.ShopID); err != nil {
            continue
        }
        items = append(items, item)
    }

    // 7. Count total (สำหรับ pagination)
    var total int
    db.QueryRow(`
        SELECT COUNT(*) FROM feature
        WHERE shopid = $1 AND ($2 = '' OR name ILIKE $3)
    `, req.ShopID, req.Keyword, "%"+req.Keyword+"%").Scan(&total)

    // 8. Return
    return c.JSON(http.StatusOK, map[string]interface{}{
        "success": true,
        "data":    items,
        "pagination": map[string]interface{}{
            "page":      (req.Offset / req.Limit) + 1,
            "perPage":   req.Limit,
            "total":     total,
            "totalPage": (total + req.Limit - 1) / req.Limit,
        },
    })
}
```

## Save Handler Template (Insert + Transaction)

```go
type FeatureSaveRequest struct {
    ShopID string `json:"shopid"`
    Code   string `json:"code"`
    Name   string `json:"name"`
}

func FeatureSaveHandler(c echo.Context) error {
    var req FeatureSaveRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "status": "error", "message": "Invalid request body",
        })
    }
    if req.ShopID == "" || req.Name == "" {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "status": "error", "message": "shopid and name are required",
        })
    }

    db, err := mypg.PgSqlFastConnect(req.ShopID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]interface{}{
            "status": "error", "message": "Database connection failed",
        })
    }
    defer db.Close()

    tx, err := db.Begin()
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]interface{}{
            "status": "error", "message": "Transaction failed",
        })
    }
    defer func() {
        if r := recover(); r != nil { tx.Rollback() }
    }()

    var guid string
    err = tx.QueryRow(`
        INSERT INTO feature (shopid, code, name)
        VALUES ($1, $2, $3)
        RETURNING guid
    `, req.ShopID, req.Code, req.Name).Scan(&guid)
    if err != nil {
        tx.Rollback()
        logger.Error("[FeatureSave] Insert failed: %v", err)
        return c.JSON(http.StatusInternalServerError, map[string]interface{}{
            "status": "error", "message": err.Error(),
        })
    }

    tx.Commit()
    logger.Success("[FeatureSave] saved guid=%s", guid)

    return c.JSON(http.StatusOK, map[string]interface{}{
        "success": true,
        "id":      guid,
        "message": "บันทึกสำเร็จ",
    })
}
```

## State Transition Handler Template

```go
type ApproveRequest struct {
    ShopID  string `json:"shopid"`
    DocNo   string `json:"docno"`
    Remarks string `json:"remarks"`
}

func ApproveDocHandler(c echo.Context) error {
    var req ApproveRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "status": "error", "message": "Invalid request body",
        })
    }

    db, err := mypg.PgSqlFastConnect(req.ShopID)
    if err != nil { /* ... */ }
    defer db.Close()

    // ตรวจสอบ state ก่อนเสมอ
    var currentStatus string
    db.QueryRow(`SELECT docstatus FROM doc WHERE shopid=$1 AND docno=$2`,
        req.ShopID, req.DocNo).Scan(&currentStatus)

    if currentStatus != "pending" {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "status": "error",
            "message": "เอกสารต้องอยู่ในสถานะ pending จึงจะอนุมัติได้",
        })
    }

    tx, _ := db.Begin()
    defer func() { if r := recover(); r != nil { tx.Rollback() } }()

    _, err = tx.Exec(`
        UPDATE doc SET docstatus='approved', approvedat=NOW(), remarks=$3
        WHERE shopid=$1 AND docno=$2
    `, req.ShopID, req.DocNo, req.Remarks)
    if err != nil {
        tx.Rollback()
        return c.JSON(http.StatusInternalServerError, map[string]interface{}{
            "status": "error", "message": err.Error(),
        })
    }

    tx.Commit()
    return c.JSON(http.StatusOK, map[string]interface{}{
        "success": true, "message": "อนุมัติสำเร็จ",
    })
}
```

## Route Registration ใน main.go

```go
// เพิ่มต่อจาก routes ที่มีอยู่ (เรียงตาม resource)
e.POST("/api/feature/list",   handlers.FeatureListHandler)
e.POST("/api/feature/insert", handlers.FeatureSaveHandler)
e.PUT("/api/feature/:guid",   handlers.FeatureUpdateHandler)
e.DELETE("/api/feature/:guid", handlers.FeatureDeleteHandler)
e.GET("/api/feature/:guid",   handlers.FeatureGetHandler)
```

## Response Format Reference

```go
// Success — single item
map[string]interface{}{"success": true, "data": item}

// Success — list with pagination
map[string]interface{}{
    "success": true,
    "data": items,
    "pagination": map[string]interface{}{
        "page": page, "perPage": limit,
        "total": total, "totalPage": totalPage,
    },
}

// Success — insert (return id)
map[string]interface{}{"success": true, "id": guid, "message": "บันทึกสำเร็จ"}

// Error 400
map[string]interface{}{"status": "error", "message": "เหตุผล"}

// Error 500
map[string]interface{}{"status": "error", "message": "เหตุผล", "error": err.Error()}
```
