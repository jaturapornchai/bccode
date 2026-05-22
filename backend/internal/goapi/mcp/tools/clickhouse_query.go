package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	myClickHouse "smlcloudplatform/internal/goapi/myclickhouse"
)

// ==================== ClickHouse Query (Readonly) ====================

// ClickHouseQueryRequest คำขอ query ClickHouse
type ClickHouseQueryRequest struct {
	ShopID string `json:"shop_id"`
	Database string `json:"database"` // ถ้าไม่ระบุจะใช้ค่าจาก env CH_DATABASE_NAME
	Query string `json:"query"`    // SQL SELECT query
	Limit int    `json:"limit"`    // จำนวนแถวสูงสุด (default=100, max=1000)
}

// ClickHouseQueryResponse ผลลัพธ์จาก query ClickHouse
type ClickHouseQueryResponse struct {
	Database string                   `json:"database"`
	Query string                   `json:"query"`
	Rows []map[string]interface{} `json:"rows"`
	RowCount int                      `json:"row_count"`
	Truncated bool                     `json:"truncated"`
	ExecutionMs int64                    `json:"execution_ms"`
	GeneratedAt time.Time                `json:"generated_at"`
}

// QueryClickHouse รัน SELECT query บน ClickHouse (readonly)
func QueryClickHouse(ctx context.Context, shopID, database, query string, limit int) (*ClickHouseQueryResponse, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	// ตรวจสอบ readonly — อนุญาตเฉพาะ SELECT และ SHOW
	normalizedQuery := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(normalizedQuery, "SELECT") && !strings.HasPrefix(normalizedQuery, "SHOW") {
		return nil, fmt.Errorf("อนุญาตเฉพาะ SELECT และ SHOW queries เท่านั้น (readonly mode)")
	}

	// ตรวจสอบ keyword ที่ไม่อนุญาต (word boundary — ไม่จับ column names เช่น isdelete, iscreated)
	forbiddenKeywords := []string{
		"DROP ", "DELETE ", "UPDATE ", "INSERT ", "ALTER ", "CREATE ", "TRUNCATE ",
		"GRANT ", "REVOKE ", "ATTACH ", "DETACH ", "RENAME ", "OPTIMIZE ",
		"INTO OUTFILE", "INTO FILE",
	}
	for _, keyword := range forbiddenKeywords {
		if strings.Contains(normalizedQuery, keyword) {
			return nil, fmt.Errorf("query มี keyword ที่ไม่อนุญาต: %s(readonly mode)", strings.TrimSpace(keyword))
		}
	}
	// ตรวจ keyword ที่อยู่ท้ายสุดด้วย (กรณีไม่มี space ตามหลัง)
	for _, keyword := range []string{"DROP", "DELETE", "UPDATE", "INSERT", "ALTER", "CREATE", "TRUNCATE"} {
		if strings.HasSuffix(normalizedQuery, keyword) {
			return nil, fmt.Errorf("query มี keyword ที่ไม่อนุญาต: %s (readonly mode)", keyword)
		}
	}

	// กำหนด limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	// เพิ่ม LIMIT ถ้าไม่มี (เฉพาะ SELECT)
	if strings.HasPrefix(normalizedQuery, "SELECT") && !strings.Contains(normalizedQuery, "LIMIT") {
		query = fmt.Sprintf("%s LIMIT %d", strings.TrimSuffix(strings.TrimSpace(query), ";"), limit)
	}

	// เพิ่ม shopid filter ถ้ามี (ป้องกันการดูข้อมูลข้าม shop)
	if shopID != "" && strings.HasPrefix(normalizedQuery, "SELECT") {
		if strings.Contains(normalizedQuery, "WHERE") {
			// เพิ่ม AND shopid = 'xxx' หลัง WHERE
			query = addShopIDToClickHouseQuery(query, shopID)
		} else {
			// เพิ่ม WHERE shopid = 'xxx'
			query = addShopIDWhereClause(query, shopID)
		}
	}

	// กำหนด database (default = ค่าจาก env CH_DATABASE_NAME)
	if database == "" {
		database = myClickHouse.GetDatabaseName()
	}

	logger.Info("[ClickHouse Query] db=%s, query=%s, limit=%d", database, query, limit)

	// เชื่อมต่อ ClickHouse
	conn, err := myClickHouse.ClickHouseFastConnect()
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ ClickHouse ได้: %w", err)
	}

	startTime := time.Now()

	// ใช้ QuerySelectAll ที่มีอยู่
	rows, err := myClickHouse.QuerySelectAll(conn, query)
	if err != nil {
		return nil, fmt.Errorf("query ล้มเหลว: %w", err)
	}

	if rows == nil {
		rows = []map[string]interface{}{}
	}

	return &ClickHouseQueryResponse{
		Database:    database,
		Query:       query,
		Rows:        rows,
		RowCount:    len(rows),
		Truncated:   len(rows) >= limit,
		ExecutionMs: time.Since(startTime).Milliseconds(),
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== List ClickHouse Tables ====================

// ClickHouseTableInfo ข้อมูล table
type ClickHouseTableInfo struct {
	Name string `json:"name"`
	Engine string `json:"engine"`
	TotalRows int    `json:"total_rows"`
	TotalBytes int    `json:"total_bytes"`
}

// ClickHouseListTablesResponse ผลลัพธ์รายการ tables
type ClickHouseListTablesResponse struct {
	Database string                `json:"database"`
	Tables []ClickHouseTableInfo `json:"tables"`
	Count int                   `json:"count"`
	GeneratedAt time.Time             `json:"generated_at"`
}

// ListClickHouseTables แสดงรายการ tables ใน ClickHouse database
func ListClickHouseTables(ctx context.Context, database string) (*ClickHouseListTablesResponse, error) {
	if database == "" {
		database = myClickHouse.GetDatabaseName()
	}

	logger.Info("[ClickHouse ListTables] db=%s", database)

	conn, err := myClickHouse.ClickHouseFastConnect()
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ ClickHouse ได้: %w", err)
	}

	// ดึงรายการ tables พร้อม engine และ row count
	// ใช้ COALESCE เพื่อแปลง Nullable เป็น non-nullable (QuerySelectAll ไม่รองรับ ptr type)
	query := fmt.Sprintf(`
		SELECT
			name,
			engine,
			COALESCE(total_rows, 0) AS total_rows,
			COALESCE(total_bytes, 0) AS total_bytes
		FROM system.tables
		WHERE database = '%s'
		ORDER BY name
	`, database)

	rows, err := myClickHouse.QuerySelectAll(conn, query)
	if err != nil {
		return nil, fmt.Errorf("ดึงรายการ tables ล้มเหลว: %w", err)
	}

	var tables []ClickHouseTableInfo
	for _, row := range rows {
		info := ClickHouseTableInfo{}
		if name, ok := row["name"].(string); ok {
			info.Name = name
		}
		if engine, ok := row["engine"].(string); ok {
			info.Engine = engine
		}
		if totalRows, ok := row["total_rows"].(int); ok {
			info.TotalRows = totalRows
		}
		if totalBytes, ok := row["total_bytes"].(int); ok {
			info.TotalBytes = totalBytes
		}
		tables = append(tables, info)
	}

	if tables == nil {
		tables = []ClickHouseTableInfo{}
	}

	return &ClickHouseListTablesResponse{
		Database:    database,
		Tables:      tables,
		Count:       len(tables),
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Helper Functions ====================

// addShopIDToClickHouseQuery เพิ่ม shopid condition เข้าไปใน WHERE clause ที่มีอยู่
func addShopIDToClickHouseQuery(query, shopID string) string {
	// หาตำแหน่ง WHERE (case-insensitive)
	upperQuery := strings.ToUpper(query)
	whereIdx := strings.Index(upperQuery, "WHERE")
	if whereIdx == -1 {
		return query
	}

	// แทรก shopid = 'xxx' AND หลัง WHERE
	insertPos := whereIdx + len("WHERE")
	return query[:insertPos] + fmt.Sprintf(" shopid = '%s' AND", shopID) + query[insertPos:]
}

// addShopIDWhereClause เพิ่ม WHERE shopid = 'xxx' เข้าไปก่อน ORDER BY/GROUP BY/HAVING/LIMIT
func addShopIDWhereClause(query, shopID string) string {
	upperQuery := strings.ToUpper(strings.TrimSpace(query))
	whereClause := fmt.Sprintf(" WHERE shopid = '%s'", shopID)

	// หาตำแหน่งแรกสุดจาก keywords ทั้งหมด (ต้องแทรก WHERE ก่อน keyword ที่อยู่ใกล้ FROM ที่สุด)
	insertKeywords := []string{"ORDER BY", "GROUP BY", "HAVING", "LIMIT"}
	minIdx := -1
	for _, keyword := range insertKeywords {
		idx := strings.Index(upperQuery, keyword)
		if idx > 0 && (minIdx == -1 || idx < minIdx) {
			minIdx = idx
		}
	}

	if minIdx > 0 {
		return query[:minIdx] + whereClause + " " + query[minIdx:]
	}

	// ถ้าไม่มี keyword ใดเลย เพิ่มท้าย query (ก่อน ;)
	trimmed := strings.TrimSuffix(strings.TrimSpace(query), ";")
	return trimmed + whereClause
}
