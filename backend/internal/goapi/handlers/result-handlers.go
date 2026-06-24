package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mypg"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/labstack/echo/v4"
)

const (
	aliasMetadataKey = "__result_alias"
	pdfFontScale     = 0.5
)

var (
	colorWhite = [3]int{255, 255, 255}
	colorBlack = [3]int{0, 0, 0}
)

type pdfStyleBlock struct {
	Background [3]int
	Text       [3]int
	Border     [3]int
	FontStyle  string
}

type pdfTableStyle struct {
	RowSpacing    float64
	ColumnSpacing float64
	GridColor     [3]int
}

type pdfStyleGuide struct {
	Header  pdfStyleBlock
	Detail  pdfStyleBlock
	Summary pdfStyleBlock
	Table   pdfTableStyle
	UseFill bool
}

func buildStyleGuide(cfg PDFStyles) pdfStyleGuide {
	defaultHeader := pdfStyleBlock{Background: colorWhite, Text: [3]int{17, 17, 17}, Border: [3]int{68, 68, 68}, FontStyle: "B"}
	defaultDetail := pdfStyleBlock{Background: colorWhite, Text: [3]int{34, 34, 34}, Border: [3]int{136, 136, 136}}
	defaultSummary := pdfStyleBlock{Background: colorWhite, Text: colorBlack, Border: colorBlack, FontStyle: "B"}
	guide := pdfStyleGuide{
		Header:  defaultHeader,
		Detail:  defaultDetail,
		Summary: defaultSummary,
		UseFill: cfg.UseFill,
		Table: pdfTableStyle{
			GridColor:     [3]int{204, 204, 204},
			RowSpacing:    cfg.Table.RowSpacing,
			ColumnSpacing: cfg.Table.ColumnSpacing,
		},
	}
	if cfg.Header.Background != "" {
		guide.Header.Background = parseHexColor(cfg.Header.Background, guide.Header.Background)
	}
	if cfg.Header.Text != "" {
		guide.Header.Text = parseHexColor(cfg.Header.Text, guide.Header.Text)
	}
	if cfg.Header.Border != "" {
		guide.Header.Border = parseHexColor(cfg.Header.Border, guide.Header.Border)
	}
	if cfg.Header.FontWeight != "" {
		guide.Header.FontStyle = fontWeightToStyle(cfg.Header.FontWeight)
	}
	if cfg.Detail.Background != "" {
		guide.Detail.Background = parseHexColor(cfg.Detail.Background, guide.Detail.Background)
	}
	if cfg.Detail.Text != "" {
		guide.Detail.Text = parseHexColor(cfg.Detail.Text, guide.Detail.Text)
	}
	if cfg.Detail.Border != "" {
		guide.Detail.Border = parseHexColor(cfg.Detail.Border, guide.Detail.Border)
	}
	if cfg.Detail.FontWeight != "" {
		guide.Detail.FontStyle = fontWeightToStyle(cfg.Detail.FontWeight)
	}
	if cfg.Summary.Background != "" {
		guide.Summary.Background = parseHexColor(cfg.Summary.Background, guide.Summary.Background)
	}
	if cfg.Summary.Text != "" {
		guide.Summary.Text = parseHexColor(cfg.Summary.Text, guide.Summary.Text)
	}
	if cfg.Summary.Border != "" {
		guide.Summary.Border = parseHexColor(cfg.Summary.Border, guide.Summary.Border)
	}
	if cfg.Summary.FontWeight != "" {
		guide.Summary.FontStyle = fontWeightToStyle(cfg.Summary.FontWeight)
	}
	if cfg.Table.GridColor != "" {
		guide.Table.GridColor = parseHexColor(cfg.Table.GridColor, guide.Table.GridColor)
	}
	if guide.Table.RowSpacing < 0 {
		guide.Table.RowSpacing = 0
	}
	if guide.Table.ColumnSpacing < 0 {
		guide.Table.ColumnSpacing = 0
	}
	return guide
}

func parseHexColor(value string, fallback [3]int) [3]int {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	trimmed = strings.TrimPrefix(trimmed, "#")
	if len(trimmed) != 6 {
		return fallback
	}
	if parsed, err := strconv.ParseUint(trimmed, 16, 32); err == nil {
		return [3]int{int(parsed >> 16 & 0xFF), int(parsed >> 8 & 0xFF), int(parsed & 0xFF)}
	}
	return fallback
}

func fontWeightToStyle(weight string) string {
	switch strings.ToLower(strings.TrimSpace(weight)) {
	case "bold", "b", "700", "semibold", "600":
		return "B"
	default:
		return ""
	}
}

// SummaryLevel โครงสร้างสำหรับแต่ละระดับของยอดรวม
type SummaryLevel struct {
	GroupByFields []string `json:"groupbyfields"` // เช่น ["docno"], ["docdate","docno"]
	SumFields     []string `json:"sumfields"`     // fields ที่จะ sum
	TypeJSON      int      `json:"typejson"`      // ค่า typejson สำหรับระดับนี้ (1=level1, 2=level2, ...)
}

// SummaryConfig โครงสร้างสำหรับการกำหนดค่าสรุปยอดรวมของแต่ละ query
type SummaryConfig struct {
	Levels         []SummaryLevel `json:"levels"`         // ระดับยอดรวมหลายระดับ
	GrandTotal     bool           `json:"grandtotal"`     // มียอดรวมทั้งหมดหรือไม่
	GrandTotalType int            `json:"grandtotaltype"` // typejson สำหรับยอดรวมทั้งหมด (default=99)
}

// LinkConfig ใข้กำหนดความสัมพันธ์ระหว่าง parent และ child query
type LinkConfig struct {
	ParentAlias string   `json:"parentalias"`
	ParentKeys  []string `json:"parentkeys"`
	ChildKeys   []string `json:"childkeys"`
}

// QueryItem โครงสร้างสำหรับ query พร้อม config ยอดรวมและ alias
type QueryItem struct {
	Alias         string         `json:"alias"`
	Query         string         `json:"query"`         // SQL query
	SummaryConfig *SummaryConfig `json:"summaryconfig"` // optional - config ยอดรวมสำหรับ query นี้
	LinkConfig    *LinkConfig    `json:"linkconfig"`
}

// resultRowPayload เก็บข้อมูลชั่วคราวสำหรับการจัดเรียงผลลัพธ์ตามลำดับ parent-child
type resultRowPayload struct {
	Alias       string
	AliasKey    string
	QueryNumber int
	TypeJSON    int
	DataJSON    string
	Data        map[string]any
	Level       int
	groupKey    string
}

const rootGroupKey = "__root__"

type resultFromQueryPayload struct {
	HoldingCode string      `json:"holdingcode"`
	QueryItems  []QueryItem `json:"queryitems"`
	Queries     []QueryItem `json:"queries"`
	Guid        string      `json:"guid"`
}

// ResultFromQueryHandler - รับ querys หลายตัวแล้วเก็บผลลัพธ์ใน result table พร้อม querynumber
func ResultFromQueryHandler(c echo.Context) error {
	logger.Info("-> ResultFromQueryHandler called")
	startTime := time.Now()

	// รับ JSON payload จาก request body
	var payload resultFromQueryPayload

	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
			"code":  "INVALID_JSON",
		})
	}

	queries := payload.QueryItems
	if len(queries) == 0 && len(payload.Queries) > 0 {
		queries = payload.Queries
	}

	// Validate holdingcode parameter
	if payload.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing required parameter: holdingcode",
			"code":  "MISSING_HOLDING_CODE",
		})
	}

	// Validate querys parameter
	if len(queries) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing required parameter: query_items (must be array of query items)",
			"code":  "MISSING_QUERY_ITEMS",
		})
	}

	aliasSeen := make(map[string]struct{})

	// Security: Basic query validation - must be SELECT queries
	for idx, item := range queries {
		aliasKey := strings.TrimSpace(item.Alias)
		if aliasKey == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Query %d missing alias", idx+1),
				"code":  "MISSING_ALIAS",
			})
		}
		lowerAlias := strings.ToLower(aliasKey)
		if _, exists := aliasSeen[lowerAlias]; exists {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Duplicate alias detected: %s", aliasKey),
				"code":  "DUPLICATE_ALIAS",
			})
		}
		aliasSeen[lowerAlias] = struct{}{}

		queryLower := strings.ToLower(strings.TrimSpace(item.Query))
		if !strings.HasPrefix(queryLower, "select") {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Only SELECT queries are allowed (query %d is not SELECT)", idx+1),
				"code":  "INVALID_QUERY_TYPE",
			})
		}

		if item.LinkConfig != nil {
			if item.LinkConfig.ParentAlias == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": fmt.Sprintf("Link config for alias %s missing parent_alias", aliasKey),
					"code":  "INVALID_LINK_CONFIG",
				})
			}
			if len(item.LinkConfig.ParentKeys) == 0 || len(item.LinkConfig.ParentKeys) != len(item.LinkConfig.ChildKeys) {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": fmt.Sprintf("Link config for alias %s must specify equal parent_keys and child_keys", aliasKey),
					"code":  "INVALID_LINK_CONFIG",
				})
			}
		}
	}

	// Generate GUID if not provided
	guid := payload.Guid
	if guid == "" {
		guid = uuid.New().String()
	}

	logStage := func(stage string) {
		logger.Info("ResultFromQuery stage=%s guid=%s elapsed=%v", stage, guid, time.Since(startTime))
	}

	// Log sanitized payload for troubleshooting
	logger.Info("ResultFromQuery payload | holdingcode=%s guid=%s queries=%d", payload.HoldingCode, guid, len(queries))

	// Connect to database (ใช้ชื่อ database ตาม holdingcode)
	db, err := mypg.PgSqlFastConnect(payload.HoldingCode)
	if err != nil {
		logger.Error("Database connection error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}
	logStage("db_connected")

	// Process rows and insert into result table using COPY batches
	const (
		maxRows         = 10000
		insertBatchSize = 500
	)

	columnsResult := []string{"guid", "docdatetime", "querynumber", "linenumber", "level", "typejson", "datajson"}
	docDateTime := time.Now()
	aliasSourceRows := make(map[string][]map[string]any)
	aliasRowPayloads := make(map[string][]*resultRowPayload)
	aliasChildren := make(map[string][]string)
	aliasParents := make(map[string]string)
	aliasLinkConfigs := make(map[string]*LinkConfig)
	rootAliasOrder := make([]string, 0)
	rootAliasSeen := make(map[string]bool)

	// Loop through each query
	for queryNumber, queryItem := range queries {
		queryNum := queryNumber + 1
		querySQL := queryItem.Query
		queryAlias := strings.TrimSpace(queryItem.Alias)
		aliasKey := normalizeAliasKey(queryAlias)
		aliasLinkConfigs[aliasKey] = queryItem.LinkConfig
		if queryItem.LinkConfig == nil || strings.TrimSpace(queryItem.LinkConfig.ParentAlias) == "" {
			if !rootAliasSeen[aliasKey] {
				rootAliasOrder = append(rootAliasOrder, aliasKey)
				rootAliasSeen[aliasKey] = true
			}
		} else {
			parentKey := normalizeAliasKey(queryItem.LinkConfig.ParentAlias)
			if parentKey == "" {
				if !rootAliasSeen[aliasKey] {
					rootAliasOrder = append(rootAliasOrder, aliasKey)
					rootAliasSeen[aliasKey] = true
				}
			} else {
				aliasParents[aliasKey] = parentKey
				aliasChildren[parentKey] = append(aliasChildren[parentKey], aliasKey)
			}
		}

		trimmedQuery := strings.TrimSpace(querySQL)
		if len(trimmedQuery) > 500 {
			trimmedQuery = trimmedQuery[:500] + "...<truncated>"
		}
		logger.Info("ResultFromQuery processing query %d/%d | guid=%s alias=%s query=%s", queryNum, len(queries), guid, queryAlias, trimmedQuery) // Execute query with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

		queryStart := time.Now()
		rows, err := db.QueryContext(ctx, querySQL)
		if err != nil {
			cancel()
			logger.Error("Query execution error for query %d: %v", queryNum, err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Query %d failed: %v", queryNum, err),
				"code":  "QUERY_EXECUTION_ERROR",
			})
		}
		logger.Info("ResultFromQuery query_completed query=%d guid=%s duration=%v", queryNum, guid, time.Since(queryStart))

		// Get column names
		columns, err := rows.Columns()
		if err != nil {
			rows.Close()
			cancel()
			logger.Error("Column retrieval error for query %d: %v", queryNum, err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Column retrieval failed for query %d: %v", queryNum, err),
				"code":  "COLUMN_ERROR",
			})
		}

		columnTypes, err := rows.ColumnTypes()
		if err != nil {
			rows.Close()
			cancel()
			logger.Error("Column type retrieval error for query %d: %v", queryNum, err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Column type retrieval failed for query %d: %v", queryNum, err),
				"code":  "COLUMN_ERROR",
			})
		}

		// Prepare scan destinations
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for j := range columns {
			valuePtrs[j] = &values[j]
		}

		sampleRows := make([]map[string]any, 0, 3)
		processedRows := 0

		// Get summary config from this query item
		summaryConfig := queryItem.SummaryConfig
		if summaryConfig != nil && len(summaryConfig.Levels) > 0 {
			logger.Info("Query %d has summary config: %d levels, grand_total=%v", queryNum, len(summaryConfig.Levels), summaryConfig.GrandTotal)
		}

		// Collect all rows for summary processing
		allRowsData := make([]map[string]any, 0, 1000)

		for rows.Next() && processedRows < maxRows {
			if err := rows.Scan(valuePtrs...); err != nil {
				logger.Error("Row scan error query %d: %v", queryNum, err)
				continue
			}

			rowData := make(map[string]any)
			for j, col := range columns {
				var colType *sql.ColumnType
				if j < len(columnTypes) {
					colType = columnTypes[j]
				}
				rowData[col] = normalizeColumnValue(values[j], colType)
			}

			sanitizeRowForJSON(rowData)

			if len(sampleRows) < 3 {
				clone := make(map[string]any, len(rowData))
				for k, v := range rowData {
					clone[k] = v
				}
				sampleRows = append(sampleRows, clone)
			}

			// Always collect data for potential summaries
			allRowsData = append(allRowsData, rowData)
			processedRows++
		}

		rows.Close()
		cancel()

		// Apply parent-child filtering ถ้ามีการกำหนด link_config
		if queryItem.LinkConfig != nil {
			parentAliasKey := normalizeAliasKey(queryItem.LinkConfig.ParentAlias)
			parentRows, ok := aliasSourceRows[parentAliasKey]
			if !ok {
				logger.Error("Link config error: parent alias %s not processed before child %s", queryItem.LinkConfig.ParentAlias, queryAlias)
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": fmt.Sprintf("Parent alias %s not available for child %s", queryItem.LinkConfig.ParentAlias, queryAlias),
					"code":  "INVALID_LINK_SEQUENCE",
				})
			}

			filteredRows := filterRowsByLink(allRowsData, parentRows, queryItem.LinkConfig)
			allRowsData = filteredRows
		}

		aliasSourceRows[normalizeAliasKey(queryAlias)] = cloneRowSlice(allRowsData)

		// Build interleaved data + summary rows
		interleavedRows := buildInterleavedRows(allRowsData, summaryConfig, queryNum, guid, queryAlias, aliasKey)
		aliasRowPayloads[aliasKey] = interleavedRows
		logger.Info("Query %d: generated %d total rows (data + summaries)", queryNum, len(interleavedRows))

		if processedRows >= maxRows {
			logger.Warn("Result truncated to %d rows for query %d", maxRows, queryNum)
		}

		var sampleJSON string
		if len(sampleRows) > 0 {
			if data, err := json.Marshal(sampleRows); err == nil {
				sampleJSON = string(data)
			} else {
				sampleJSON = fmt.Sprintf("sample marshal error: %v", err)
			}
		} else {
			sampleJSON = "[]"
		}

		logger.Info("ResultFromQuery query_result | query=%d guid=%s rows=%d sample=%s", queryNum, guid, len(interleavedRows), sampleJSON)
	}

	aliasDepth := computeAliasDepths(rootAliasOrder, aliasParents, aliasRowPayloads)
	orderedRows := orderRowsByHierarchy(rootAliasOrder, aliasChildren, aliasRowPayloads, aliasLinkConfigs, aliasDepth)
	if len(orderedRows) == 0 {
		logStage("all_queries_processed")
		logStage("completed")
		return c.JSON(http.StatusOK, map[string]any{
			"status":  "success",
			"guid":    guid,
			"count":   0,
			"queries": len(queries),
			"message": "No rows generated for the provided queries",
		})
	}

	lineNumber := 1
	batchRows := make([][]any, 0, insertBatchSize)
	totalInsertCount := 0
	flushBatch := func() error {
		if len(batchRows) == 0 {
			return nil
		}
		batchLen := len(batchRows)
		batchStart := time.Now()
		insertCtx, insertCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer insertCancel()

		if err := mypg.BulkInsertWithCopy(insertCtx, db, "result", columnsResult, batchRows); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				logger.Error("Bulk insert deadline exceeded | guid=%s batch=%d total=%d elapsed=%v", guid, batchLen, totalInsertCount+batchLen, time.Since(startTime))
			} else {
				logger.Error("Bulk insert error | guid=%s batch=%d total=%d err=%v", guid, batchLen, totalInsertCount+batchLen, err)
			}
			return err
		}

		totalInsertCount += batchLen
		batchRows = batchRows[:0]
		logger.Info("ResultFromQuery batch_insert | guid=%s batch=%d total=%d batch_elapsed=%v overall_elapsed=%v", guid, batchLen, totalInsertCount, time.Since(batchStart), time.Since(startTime))
		return nil
	}

	for _, row := range orderedRows {
		batchRows = append(batchRows, []any{guid, docDateTime, row.QueryNumber, lineNumber, row.Level, row.TypeJSON, row.DataJSON})
		lineNumber++
		if len(batchRows) >= insertBatchSize {
			if err := flushBatch(); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to persist result rows",
					"code":  "RESULT_INSERT_ERROR",
				})
			}
		}
	}

	if err := flushBatch(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to persist result rows",
			"code":  "RESULT_INSERT_ERROR",
		})
	}

	logStage("all_queries_processed")
	logger.Success("Inserted %d rows total into result table with guid: %s (%d queries)", totalInsertCount, guid, len(queries))
	logStage("completed")

	// Return success response
	return c.JSON(http.StatusOK, map[string]any{
		"status":  "success",
		"guid":    guid,
		"count":   totalInsertCount,
		"queries": len(queries),
		"message": fmt.Sprintf("Query results saved successfully (%d rows from %d queries)", totalInsertCount, len(queries)),
	})
}

// ResultGetHandler - ดึงข้อมูลจาก result table พร้อม pagination filter
func ResultGetHandler(c echo.Context) error {
	logger.Info("-> ResultGetHandler called")

	// รับ JSON payload จาก request body
	var payload struct {
		HoldingCode string `json:"holdingcode"`
		Guid        string `json:"guid"`
		Limit       int    `json:"limit"`
		Offset      int    `json:"offset"`
	}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
			"code":  "INVALID_JSON",
		})
	}

	// Validate holdingcode parameter
	if payload.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing required parameter: holdingcode",
			"code":  "MISSING_HOLDING_CODE",
		})
	}

	// Validate guid parameter
	if payload.Guid == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing required parameter: guid",
			"code":  "MISSING_GUID",
		})
	}

	// Set default values for limit and offset
	if payload.Limit <= 0 {
		payload.Limit = 100
	}
	if payload.Offset < 0 {
		payload.Offset = 0
	}

	// Security: Limit max rows to prevent abuse
	if payload.Limit > 1000 {
		payload.Limit = 1000
	}

	// Connect to database (ใช้ชื่อ database ตาม holdingcode)
	db, err := mypg.PgSqlFastConnect(payload.HoldingCode)
	if err != nil {
		logger.Error("Database connection error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	// Build SELECT query (no count query)
	selectSQL := `
		SELECT querynumber, linenumber, level, typejson, datajson
		FROM result
		WHERE guid = $1
		ORDER BY linenumber
		LIMIT $2 OFFSET $3
	`
	selectArgs := []any{payload.Guid, payload.Limit, payload.Offset}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, selectSQL, selectArgs...)
	if err != nil {
		logger.Error("Query execution error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Query failed: %v", err),
			"code":  "QUERY_ERROR",
		})
	}
	defer rows.Close()

	// Collect results - return ทุกอย่าง (รวม summary และ grand total)
	flatRows := make([]map[string]any, 0, payload.Limit)
	aliasBuckets := make(map[string][]map[string]any)
	aliasOrder := make([]string, 0)

	for rows.Next() {
		var (
			queryNumber int
			lineNumber  int
			levelValue  int
			typeJSON    int
			dataJSON    []byte
		)

		if err := rows.Scan(&queryNumber, &lineNumber, &levelValue, &typeJSON, &dataJSON); err != nil {
			logger.Error("Row scan error: %v", err)
			continue
		}

		// Parse JSON data
		var jsonData map[string]any
		if err := json.Unmarshal(dataJSON, &jsonData); err != nil {
			logger.Error("JSON unmarshal error: %v", err)
			continue
		}

		aliasValue := ""
		if aliasRaw, ok := jsonData[aliasMetadataKey]; ok {
			aliasValue = fmt.Sprintf("%v", aliasRaw)
			delete(jsonData, aliasMetadataKey)
		}
		if aliasValue == "" {
			aliasValue = fmt.Sprintf("query_%d", queryNumber)
		}

		rowPayload := map[string]any{
			"alias":       aliasValue,
			"level":       levelValue,
			"typejson":    typeJSON,
			"is_summary":  typeJSON != 0,
			"querynumber": queryNumber,
			"linenumber":  lineNumber,
			"data":        jsonData,
		}

		flatRows = append(flatRows, rowPayload)
		if _, exists := aliasBuckets[aliasValue]; !exists {
			aliasOrder = append(aliasOrder, aliasValue)
		}
		aliasBuckets[aliasValue] = append(aliasBuckets[aliasValue], rowPayload)
	}

	sections := make([]map[string]any, 0, len(aliasOrder))
	for _, alias := range aliasOrder {
		rowsForAlias := aliasBuckets[alias]
		sections = append(sections, map[string]any{
			"alias": alias,
			"count": len(rowsForAlias),
			"rows":  rowsForAlias,
		})
	}

	// สร้าง response (ไม่มี total count)
	response := map[string]any{
		"status":   "success",
		"guid":     payload.Guid,
		"sections": sections,
		"rows":     flatRows,
		"pagination": map[string]any{
			"limit":  payload.Limit,
			"offset": payload.Offset,
			"count":  len(flatRows),
		},
	}

	return c.JSON(http.StatusOK, response)
}

// ResultToPDFHandler - สร้าง PDF จาก result table
func ResultToPDFHandler(c echo.Context) error {
	logger.Info("-> ResultToPDFHandler called")

	// รับ JSON payload จาก request body
	var payload resultToPDFPayload

	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
			"code":  "INVALID_JSON",
		})
	}

	// Validate parameters
	if payload.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing required parameter: holdingcode",
			"code":  "MISSING_HOLDING_CODE",
		})
	}

	if payload.Guid == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing required parameter: guid",
			"code":  "MISSING_GUID",
		})
	}

	// Set default PDF config values
	if payload.PDFConfig.Title == "" {
		payload.PDFConfig.Title = "Report"
	}
	if payload.PDFConfig.Orientation == "" {
		payload.PDFConfig.Orientation = "P" // Portrait
	}
	if payload.PDFConfig.PageSize == "" {
		payload.PDFConfig.PageSize = "A4"
	}

	// Connect to database
	db, err := mypg.PgSqlFastConnect(payload.HoldingCode)
	if err != nil {
		logger.Error("Database connection error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	// Query data from result table
	selectSQL := `
		SELECT querynumber, linenumber, level, typejson, datajson
		FROM result
		WHERE guid = $1
		ORDER BY querynumber, linenumber
	`

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, selectSQL, payload.Guid)
	if err != nil {
		logger.Error("Query execution error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Query failed: %v", err),
			"code":  "QUERY_ERROR",
		})
	}
	defer rows.Close()

	// Collect all data with metadata
	resultRows := make([]pdfResultRow, 0, 512)
	for rows.Next() {
		var (
			queryNumber int
			lineNumber  int
			levelValue  int
			typeJSON    int
			dataJSON    []byte
		)
		if err := rows.Scan(&queryNumber, &lineNumber, &levelValue, &typeJSON, &dataJSON); err != nil {
			logger.Error("Row scan error: %v", err)
			continue
		}

		var jsonData map[string]any
		if err := json.Unmarshal(dataJSON, &jsonData); err != nil {
			logger.Error("JSON unmarshal error: %v", err)
			continue
		}

		aliasValue := ""
		if aliasRaw, ok := jsonData[aliasMetadataKey]; ok {
			aliasValue = fmt.Sprintf("%v", aliasRaw)
			delete(jsonData, aliasMetadataKey)
		}
		if aliasValue == "" {
			aliasValue = fmt.Sprintf("query_%d", queryNumber)
		}

		resultRows = append(resultRows, pdfResultRow{
			Alias:       aliasValue,
			QueryNumber: queryNumber,
			LineNumber:  lineNumber,
			Level:       levelValue,
			TypeJSON:    typeJSON,
			Data:        jsonData,
		})
	}

	if len(resultRows) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "No data found for the given guid",
			"code":  "NO_DATA",
		})
	}

	var (
		pdfDoc    *gofpdf.Fpdf
		renderErr error
	)
	if len(payload.LayoutConfig.Sections) > 0 {
		pdfDoc, renderErr = renderPDFWithLayout(resultRows, payload)
	} else {
		pdfDoc, renderErr = renderSimplePDF(resultRows, payload)
	}
	if renderErr != nil {
		logger.Error("Failed to render PDF: %v", renderErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Unable to render PDF",
			"code":  "PDF_RENDER_ERROR",
		})
	}

	filePath, err := savePDFDocument(pdfDoc, payload.Guid)
	if err != nil {
		logger.Error("Failed to persist PDF: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create PDF",
			"code":  "PDF_CREATION_ERROR",
		})
	}

	logger.Success("PDF created successfully: %s", filePath)

	// Return file as download
	return c.File(filePath)
}

func renderPDFWithLayout(rows []pdfResultRow, payload resultToPDFPayload) (*gofpdf.Fpdf, error) {
	pdf, baseFont := createBasePDF(payload.PDFConfig)
	if baseFont == "" {
		baseFont = "Arial"
	}
	setupPageHeader(pdf, baseFont, payload.PDFConfig.Title)
	pdf.AddPage()
	pdf.SetTextColor(colorBlack[0], colorBlack[1], colorBlack[2])

	left, _, right, _ := pdf.GetMargins()
	pageWidth, _ := pdf.GetPageSize()
	tableWidth := pageWidth - left - right
	styleGuide := buildStyleGuide(payload.LayoutConfig.Styles)
	columnConfigMap := columnConfigLookup(payload.LayoutConfig)

	aliasRows := groupRowsByAlias(rows)
	aliasLevels := make(map[string][]int, len(aliasRows))
	for alias, rowSet := range aliasRows {
		aliasLevels[alias] = collectSectionLevels(rowSet)
	}
	logger.Debug("PDF layout config | title=%s sections=%d font=%s scale=%.2f", payload.PDFConfig.Title, len(payload.LayoutConfig.Sections), baseFont, pdfFontScale)
	visibleSections := make([]PDFSectionConfig, 0, len(payload.LayoutConfig.Sections))
	for _, section := range payload.LayoutConfig.Sections {
		if sectionVisible(section) {
			enriched := applySchemaToSection(section, payload.LayoutConfig)
			visibleSections = append(visibleSections, enriched)
		}
	}

	detailGroups := make(map[string]map[string][]pdfResultRow)
	detailRenderers := make(map[string]*sectionRenderer)
	detailDocRendered := make(map[string]map[string]bool)
	detailOrder := make([]PDFSectionConfig, 0)
	for _, section := range visibleSections {
		if strings.EqualFold(section.RowType, "detail") {
			rowsForAlias := aliasRows[section.Alias]
			if len(rowsForAlias) == 0 {
				continue
			}
			detailGroups[section.Alias] = groupRowsByField(rowsForAlias, "docno")
			detailRenderers[section.Alias] = newSectionRenderer(pdf, section, payload.ColumnNames, payload.ColumnOrder, tableWidth, 0, baseFont, styleGuide, columnConfigMap, aliasLevels[section.Alias])
			detailDocRendered[section.Alias] = make(map[string]bool)
			detailOrder = append(detailOrder, section)
		}
	}

	renderedAliases := make(map[string]bool)
	for _, section := range visibleSections {
		if renderedAliases[section.Alias] {
			continue
		}

		rowsForAlias := aliasRows[section.Alias]
		if len(rowsForAlias) == 0 {
			continue
		}

		rowType := strings.ToLower(strings.TrimSpace(section.RowType))
		renderer := newSectionRenderer(pdf, section, payload.ColumnNames, payload.ColumnOrder, tableWidth, 0, baseFont, styleGuide, columnConfigMap, aliasLevels[section.Alias])
		if renderer == nil {
			continue
		}
		logger.Debug("Rendering section | alias=%s rows=%d levels=%v type=%s", section.Alias, len(rowsForAlias), aliasLevels[section.Alias], rowType)
		renderer.ensureSectionHeader()
		for _, row := range rowsForAlias {
			if rowType == "header" && row.TypeJSON == 0 {
				detailHeight := estimateDetailLeadHeight(row, detailOrder, detailGroups, detailRenderers)
				renderer.ensurePairSpacing(row, detailHeight)
			}
			renderer.renderDataRow(row)
			if rowType == "header" && row.TypeJSON == 0 {
				docRef := stringifyValue(row.Data["docno"])
				if docRef == "" {
					continue
				}
				for _, detailSection := range detailOrder {
					detailRenderer := detailRenderers[detailSection.Alias]
					if detailRenderer == nil {
						continue
					}
					if detailDocRendered[detailSection.Alias][docRef] {
						continue
					}
					rowsForDoc := detailGroups[detailSection.Alias][docRef]
					if len(rowsForDoc) == 0 {
						continue
					}
					detailRenderer.mergeWithPreviousHeader()
					detailRenderer.ensureSectionHeader()
					for _, detailRow := range rowsForDoc {
						detailRenderer.renderDataRow(detailRow)
					}
					pdf.Ln(2 * pdfFontScale)
					detailDocRendered[detailSection.Alias][docRef] = true
					renderedAliases[detailSection.Alias] = true
				}
			}
		}
		pdf.Ln(2 * pdfFontScale)
		renderedAliases[section.Alias] = true
	}

	// Render any remaining visible sections that haven't been emitted yet (e.g. detail-only sections)
	for _, section := range visibleSections {
		if renderedAliases[section.Alias] {
			continue
		}
		rowsForAlias := aliasRows[section.Alias]
		if len(rowsForAlias) == 0 {
			continue
		}
		renderer := newSectionRenderer(pdf, section, payload.ColumnNames, payload.ColumnOrder, tableWidth, 0, baseFont, styleGuide, columnConfigMap, aliasLevels[section.Alias])
		if renderer == nil {
			continue
		}
		renderer.ensureSectionHeader()
		for _, row := range rowsForAlias {
			renderer.renderDataRow(row)
		}
		pdf.Ln(2 * pdfFontScale)
		renderedAliases[section.Alias] = true
	}

	return pdf, nil
}

func estimateDetailLeadHeight(headerRow pdfResultRow, detailOrder []PDFSectionConfig, detailGroups map[string]map[string][]pdfResultRow, renderers map[string]*sectionRenderer) float64 {
	docRef := stringifyValue(headerRow.Data["docno"])
	if docRef == "" {
		return 0
	}
	for _, section := range detailOrder {
		rowsForDoc := detailGroups[section.Alias][docRef]
		if len(rowsForDoc) == 0 {
			continue
		}
		renderer := renderers[section.Alias]
		if renderer == nil {
			continue
		}
		return renderer.estimateRowHeight(rowsForDoc[0])
	}
	return 0
}

func renderSimplePDF(rows []pdfResultRow, payload resultToPDFPayload) (*gofpdf.Fpdf, error) {
	pdf, baseFont := createBasePDF(payload.PDFConfig)
	if baseFont == "" {
		baseFont = "Arial"
	}
	setupPageHeader(pdf, baseFont, payload.PDFConfig.Title)
	pdf.AddPage()
	pdf.SetTextColor(colorBlack[0], colorBlack[1], colorBlack[2])
	styleGuide := buildStyleGuide(payload.LayoutConfig.Styles)

	columns := make([]string, 0, len(payload.ColumnOrder))
	columns = append(columns, payload.ColumnOrder...)
	if len(columns) == 0 && len(rows) > 0 {
		for key := range rows[0].Data {
			columns = append(columns, key)
		}
	}
	if len(columns) == 0 {
		return nil, fmt.Errorf("no columns to render")
	}

	left, _, right, _ := pdf.GetMargins()
	pageWidth, _ := pdf.GetPageSize()
	colWidth := (pageWidth - left - right) / float64(len(columns))
	rowTypeMap := sectionRowTypeLookup(payload.LayoutConfig)
	columnConfigMap := columnConfigLookup(payload.LayoutConfig)
	tableWidth := pageWidth - left - right

	drawHeader := func() {
		headerFill := styleGuide.Header.Background
		if !styleGuide.UseFill {
			headerFill = colorWhite
		}
		pdf.SetFont(baseFont, styleGuide.Header.FontStyle, 12*pdfFontScale)
		pdf.SetTextColor(styleGuide.Header.Text[0], styleGuide.Header.Text[1], styleGuide.Header.Text[2])
		pdf.SetFillColor(headerFill[0], headerFill[1], headerFill[2])
		pdf.SetDrawColor(styleGuide.Header.Border[0], styleGuide.Header.Border[1], styleGuide.Header.Border[2])
		startY := pdf.GetY()
		rowHeight := 7.0 * pdfFontScale
		pdf.Line(left, startY, left+tableWidth, startY)
		x := left
		for _, col := range columns {
			label := col
			if payload.ColumnNames != nil {
				if custom, ok := payload.ColumnNames[col]; ok && custom != "" {
					label = custom
				}
			}
			pdf.SetXY(x, startY)
			pdf.CellFormat(colWidth, rowHeight, label, "", 0, "C", styleGuide.UseFill, 0, "")
			x += colWidth
		}
		pdf.Line(left, startY+rowHeight, left+tableWidth, startY+rowHeight)
		pdf.SetY(startY + rowHeight + pdfFontScale)
		pdf.SetFont(baseFont, "", 11*pdfFontScale)
	}

	drawHeader()
	lineHeight := 4.8 * pdfFontScale
	_, pageHeight := pdf.GetPageSize()
	_, bottomMargin := pdf.GetAutoPageBreak()
	for _, row := range rows {
		texts := make([]string, len(columns))
		maxLines := 1
		for idx, col := range columns {
			cfg, hasCfg := columnConfigMap[col]
			if hasCfg {
				texts[idx] = formatValueForColumn(row.Data[col], cfg)
			} else {
				texts[idx] = formatDisplayValue(row.Data[col])
			}
			wrapped := pdf.SplitLines([]byte(texts[idx]), colWidth)
			lineCount := len(wrapped)
			if lineCount == 0 {
				lineCount = 1
			}
			if lineCount > maxLines {
				maxLines = lineCount
			}
		}
		rowHeight := lineHeight * float64(maxLines)
		if pdf.GetY()+rowHeight+bottomMargin > pageHeight {
			pdf.AddPage()
			drawHeader()
		}
		startX := left
		startY := pdf.GetY()
		rowType := ""
		if rowTypeMap != nil {
			rowType = rowTypeMap[normalizeAliasKey(row.Alias)]
		}
		rowStyle := computeRowStyle(row, rowType, styleGuide)
		pdf.SetFillColor(rowStyle.Background[0], rowStyle.Background[1], rowStyle.Background[2])
		pdf.SetDrawColor(rowStyle.Border[0], rowStyle.Border[1], rowStyle.Border[2])
		pdf.SetTextColor(rowStyle.Text[0], rowStyle.Text[1], rowStyle.Text[2])
		pdf.SetFont(baseFont, rowStyle.FontStyle, 11*pdfFontScale)
		for idx, text := range texts {
			align := "L"
			if cfg, ok := columnConfigMap[columns[idx]]; ok {
				align = resolveAlignValue(cfg)
			}
			pdf.SetXY(startX, startY)
			pdf.MultiCell(colWidth, lineHeight, text, "", align, styleGuide.UseFill)
			startX += colWidth
		}
		pdf.SetY(startY + rowHeight)
	}

	return pdf, nil
}

func savePDFDocument(pdf *gofpdf.Fpdf, guid string) (string, error) {
	tempDir := "temp"
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return "", err
	}
	filename := fmt.Sprintf("result_%s_%s.pdf", guid, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(tempDir, filename)
	if err := pdf.OutputFileAndClose(filePath); err != nil {
		return "", err
	}

	// Upload ไป Cloudflare R2 แล้วลบ temp file
	uploadPDFToS3AndCleanup(filePath, "results/"+filename)

	return filePath, nil
}

// uploadPDFToS3AndCleanup - upload ไฟล์ไป S3 แล้วลบ temp (async)
func uploadPDFToS3AndCleanup(filePath string, objectKey string) {
	go func() {
		defer os.Remove(filePath)

		client, err := GetR2Client()
		if err != nil {
			logger.Warn("S3 not available for PDF cleanup: %v", err)
			return
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			logger.Warn("Failed to read temp PDF: %v", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err = client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(r2BucketName),
			Key:         aws.String(objectKey),
			Body:        bytes.NewReader(data),
			ContentType: aws.String("application/pdf"),
		})
		if err != nil {
			logger.Warn("Failed to upload PDF to S3: %v", err)
			return
		}
		logger.Info("PDF uploaded to S3: %s", objectKey)
	}()
}

func createBasePDF(cfg PDFConfig) (*gofpdf.Fpdf, string) {
	orientation := cfg.Orientation
	if orientation == "" {
		orientation = "P"
	}
	pageSize := cfg.PageSize
	if pageSize == "" {
		pageSize = "A4"
	}
	pdf := gofpdf.New(orientation, "mm", pageSize, "")
	pdf.SetMargins(12, 20, 12)
	pdf.SetAutoPageBreak(true, 15)
	font := registerUTF8Fonts(pdf)
	if font != "" {
		pdf.SetFont(font, "", 12*pdfFontScale)
	} else {
		pdf.SetFont("Arial", "", 12*pdfFontScale)
	}
	return pdf, font
}

func registerUTF8Fonts(pdf *gofpdf.Fpdf) string {
	type fontCandidate struct {
		family  string
		regular string
		bold    string
	}
	candidates := []fontCandidate{
		{family: "THSarabunNew", regular: filepath.Join("fonts", "THSarabunNew.ttf"), bold: filepath.Join("fonts", "THSarabunNew-Bold.ttf")},
		{family: "NotoSansThai", regular: filepath.Join("fonts", "NotoSansThai-Regular.ttf"), bold: filepath.Join("fonts", "NotoSansThai-Bold.ttf")},
		{family: "NotoSansThai_SemiCondensed", regular: filepath.Join("fonts", "NotoSansThai_SemiCondensed-Regular.ttf"), bold: filepath.Join("fonts", "NotoSansThai_SemiCondensed-Bold.ttf")},
		{family: "NotoSansThai_Condensed", regular: filepath.Join("fonts", "NotoSansThai_Condensed-Regular.ttf"), bold: filepath.Join("fonts", "NotoSansThai_Condensed-Bold.ttf")},
		{family: "NotoSansThai_ExtraCondensed", regular: filepath.Join("fonts", "NotoSansThai_ExtraCondensed-Regular.ttf"), bold: filepath.Join("fonts", "NotoSansThai_ExtraCondensed-Bold.ttf")},
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate.regular); err != nil {
			continue
		}
		pdf.AddUTF8Font(candidate.family, "", candidate.regular)
		if _, err := os.Stat(candidate.bold); err == nil {
			pdf.AddUTF8Font(candidate.family, "B", candidate.bold)
		} else {
			pdf.AddUTF8Font(candidate.family, "B", candidate.regular)
		}
		return candidate.family
	}

	return ""
}

func setupPageHeader(pdf *gofpdf.Fpdf, fontFamily, title string) {
	font := fontFamily
	if font == "" {
		font = "Arial"
	}
	pdf.AliasNbPages("{nb}")
	pdf.SetHeaderFuncMode(func() {
		left, top, right, _ := pdf.GetMargins()
		pdf.SetY(10)
		pdf.SetFont(font, "B", 12*pdfFontScale)
		pdf.SetTextColor(colorBlack[0], colorBlack[1], colorBlack[2])
		if title != "" {
			pdf.SetX(left)
			pdf.CellFormat(0, 6*pdfFontScale, title, "", 0, "L", false, 0, "")
		}
		pageLabel := fmt.Sprintf("หน้า %d/{nb}", pdf.PageNo())
		pageWidth, _ := pdf.GetPageSize()
		labelWidth := pdf.GetStringWidth(pageLabel)
		pdf.SetXY(pageWidth-right-labelWidth, pdf.GetY())
		pdf.CellFormat(labelWidth, 6*pdfFontScale, pageLabel, "", 0, "L", false, 0, "")
		pdf.Ln(4 * pdfFontScale)
		pdf.SetY(top)
	}, true)
}

func sectionRowTypeLookup(cfg PDFLayoutConfig) map[string]string {
	if len(cfg.Sections) == 0 {
		return nil
	}
	lookup := make(map[string]string, len(cfg.Sections))
	for _, section := range cfg.Sections {
		lookup[normalizeAliasKey(section.Alias)] = strings.ToLower(strings.TrimSpace(section.RowType))
	}
	return lookup
}

func columnConfigLookup(cfg PDFLayoutConfig) map[string]PDFColumnConfig {
	lookup := make(map[string]PDFColumnConfig)
	for field, schema := range cfg.ColumnSchema {
		fieldKey := strings.TrimSpace(field)
		if fieldKey == "" {
			continue
		}
		lookup[fieldKey] = columnConfigFromSchema(fieldKey, schema, cfg.NumberFormats[fieldKey])
	}
	for _, section := range cfg.Sections {
		for _, col := range section.Columns {
			field := strings.TrimSpace(col.Field)
			if field == "" {
				continue
			}
			enriched := enrichColumnConfig(col, cfg.ColumnSchema, cfg.NumberFormats)
			if existing, ok := lookup[field]; ok {
				lookup[field] = mergeColumnConfigs(existing, enriched)
			} else {
				lookup[field] = enriched
			}
		}
	}
	if len(lookup) == 0 {
		return nil
	}
	return lookup
}

func mergeColumnConfigs(base PDFColumnConfig, incoming PDFColumnConfig) PDFColumnConfig {
	result := base
	if result.Label == "" && incoming.Label != "" {
		result.Label = incoming.Label
	}
	if (result.Flex == 0 || incoming.Flex > 0) && incoming.Flex > 0 {
		result.Flex = incoming.Flex
	}
	if result.Align == "" && incoming.Align != "" {
		result.Align = incoming.Align
	}
	if !result.Numeric && incoming.Numeric {
		result.Numeric = true
	}
	if !result.HideWhenSummary && incoming.HideWhenSummary {
		result.HideWhenSummary = true
	}
	if result.DecimalPlaces == nil && incoming.DecimalPlaces != nil {
		result.DecimalPlaces = incoming.DecimalPlaces
	}
	if result.FormatNumber == "" && incoming.FormatNumber != "" {
		result.FormatNumber = incoming.FormatNumber
	}
	if result.DataType == "" && incoming.DataType != "" {
		result.DataType = incoming.DataType
	}
	if result.Format == "" && incoming.Format != "" {
		result.Format = incoming.Format
	}
	if !result.UseBuddhistYear && incoming.UseBuddhistYear {
		result.UseBuddhistYear = true
	}
	return result
}

func applySchemaToSection(section PDFSectionConfig, layout PDFLayoutConfig) PDFSectionConfig {
	if len(section.Columns) == 0 {
		return section
	}
	enriched := section
	enriched.Columns = make([]PDFColumnConfig, len(section.Columns))
	for idx, col := range section.Columns {
		enriched.Columns[idx] = enrichColumnConfig(col, layout.ColumnSchema, layout.NumberFormats)
	}
	return enriched
}

func enrichColumnConfig(col PDFColumnConfig, schemaMap map[string]PDFColumnSchema, numberFormats map[string]string) PDFColumnConfig {
	result := col
	schema, hasSchema := schemaMap[col.Field]
	if hasSchema {
		if result.Label == "" {
			result.Label = schema.Label
		}
		if result.Flex == 0 && schema.Flex > 0 {
			result.Flex = schema.Flex
		}
		if result.Align == "" && schema.Align != "" {
			result.Align = schema.Align
		}
		if result.DataType == "" {
			result.DataType = schema.DataType
		}
		if result.Format == "" {
			result.Format = schema.Format
		}
		if schema.HideWhenSummary {
			result.HideWhenSummary = true
		}
		if schema.UseBuddhistYear {
			result.UseBuddhistYear = true
		}
	}
	dataType := strings.ToLower(strings.TrimSpace(result.DataType))
	if !result.Numeric && dataType == "number" {
		result.Numeric = true
	}
	if result.FormatNumber == "" {
		if format, ok := numberFormats[strings.TrimSpace(col.Field)]; ok && format != "" {
			result.FormatNumber = format
		} else if dataType == "number" && result.Format != "" {
			result.FormatNumber = result.Format
		}
	}
	if result.Align == "" {
		if result.Numeric {
			result.Align = "R"
		} else {
			result.Align = "L"
		}
	}
	if result.Flex == 0 {
		result.Flex = 1
	}
	return result
}

func columnConfigFromSchema(field string, schema PDFColumnSchema, numberFormat string) PDFColumnConfig {
	config := PDFColumnConfig{
		Field:           field,
		Label:           schema.Label,
		Flex:            schema.Flex,
		Align:           schema.Align,
		HideWhenSummary: schema.HideWhenSummary,
		DataType:        schema.DataType,
		Format:          schema.Format,
		UseBuddhistYear: schema.UseBuddhistYear,
	}
	if strings.ToLower(strings.TrimSpace(schema.DataType)) == "number" {
		config.Numeric = true
	}
	if config.Align == "" {
		if config.Numeric {
			config.Align = "R"
		} else {
			config.Align = "L"
		}
	}
	if config.Flex == 0 {
		config.Flex = 1
	}
	if numberFormat != "" {
		config.FormatNumber = numberFormat
	} else if config.Numeric && config.Format != "" {
		config.FormatNumber = config.Format
	}
	return config
}

func computeRowStyle(row pdfResultRow, sectionRowType string, styles pdfStyleGuide) pdfStyleBlock {
	if row.TypeJSON != 0 {
		return styles.Summary
	}
	if strings.EqualFold(sectionRowType, "header") {
		return styles.Header
	}
	return styles.Detail
}

type sectionRenderer struct {
	pdf            *gofpdf.Fpdf
	section        PDFSectionConfig
	columnNames    map[string]string
	columns        []PDFColumnConfig
	widths         []float64
	indent         float64
	fontFamily     string
	headerDrawn    bool
	lastHeaderPage int
	leftMargin     float64
	styleGuide     pdfStyleGuide
	globalConfig   map[string]PDFColumnConfig
	lineHeight     float64
	rowSpacing     float64
	columnSpacing  float64
	levelBadges    []int
	skipTopBorder  bool
}

func newSectionRenderer(pdf *gofpdf.Fpdf, section PDFSectionConfig, columnNames map[string]string, fallbackOrder []string, tableWidth, indent float64, fontFamily string, styleGuide pdfStyleGuide, global map[string]PDFColumnConfig, levels []int) *sectionRenderer {
	columns := section.Columns
	if len(columns) == 0 && len(fallbackOrder) > 0 {
		columns = make([]PDFColumnConfig, 0, len(fallbackOrder))
		for _, field := range fallbackOrder {
			if cfg, ok := global[field]; ok {
				columns = append(columns, cfg)
				continue
			}
			columns = append(columns, PDFColumnConfig{Field: field, Label: field, Flex: 1, Align: "left"})
		}
	}
	if len(columns) == 0 {
		return nil
	}
	colCopy := make([]PDFColumnConfig, len(columns))
	copy(colCopy, columns)
	usableWidth := tableWidth - indent
	columnSpacing := styleGuide.Table.ColumnSpacing * pdfFontScale
	if columnSpacing < 0 {
		columnSpacing = 0
	}
	if usableWidth <= 0 {
		usableWidth = tableWidth
	}
	if columnSpacing > 0 && len(colCopy) > 1 {
		spacer := columnSpacing * float64(len(colCopy)-1)
		if usableWidth-spacer > 0 {
			usableWidth -= spacer
		}
	}
	if usableWidth <= 0 {
		usableWidth = tableWidth
	}
	widths := computeColumnWidths(colCopy, usableWidth)
	left, _, _, _ := pdf.GetMargins()
	rowSpacing := 0.0
	return &sectionRenderer{
		pdf:           pdf,
		section:       section,
		columnNames:   columnNames,
		columns:       colCopy,
		widths:        widths,
		indent:        indent,
		fontFamily:    fontFamily,
		leftMargin:    left,
		styleGuide:    styleGuide,
		globalConfig:  global,
		lineHeight:    4.8 * pdfFontScale,
		rowSpacing:    rowSpacing,
		columnSpacing: columnSpacing,
		levelBadges:   levels,
	}
}

func (sr *sectionRenderer) baseFont() string {
	if sr.fontFamily == "" {
		return "Arial"
	}
	return sr.fontFamily
}

func (sr *sectionRenderer) totalWidth() float64 {
	sum := 0.0
	for _, w := range sr.widths {
		sum += w
	}
	return sum
}

func (sr *sectionRenderer) renderSectionHeader() {
	sr.refreshHeaderState()
	skipTopBorder := sr.consumeHeaderSkip()
	if sr.shouldDisplayTitle() {
		sr.pdf.SetFont(sr.baseFont(), "B", 13*pdfFontScale)
		sr.pdf.SetTextColor(sr.styleGuide.Header.Text[0], sr.styleGuide.Header.Text[1], sr.styleGuide.Header.Text[2])
		sr.pdf.CellFormat(0, 8*pdfFontScale, sr.section.Title, "", 1, "L", false, 0, "")
		sr.pdf.Ln(pdfFontScale)
	}
	fill, text := sr.headerColors()
	fontStyle := sr.styleGuide.Header.FontStyle
	sr.pdf.SetFont(sr.baseFont(), fontStyle, 11*pdfFontScale)
	sr.pdf.SetFillColor(fill[0], fill[1], fill[2])
	sr.pdf.SetTextColor(text[0], text[1], text[2])
	sr.pdf.SetDrawColor(sr.styleGuide.Header.Border[0], sr.styleGuide.Header.Border[1], sr.styleGuide.Header.Border[2])
	startX := sr.leftMargin + sr.indent
	startY := sr.pdf.GetY()
	rowHeight := 7.0 * pdfFontScale
	widthTotal := sr.totalRowWidth()
	sr.pdf.SetX(startX)
	if !skipTopBorder {
		sr.pdf.Line(startX, startY, startX+widthTotal, startY)
	}
	for idx, col := range sr.columns {
		label := col.Label
		if label == "" && sr.columnNames != nil {
			if custom, ok := sr.columnNames[col.Field]; ok && custom != "" {
				label = custom
			}
		}
		if label == "" {
			label = col.Field
		}
		align := sr.resolveAlign(col)
		sr.pdf.SetXY(startX, startY)
		sr.pdf.CellFormat(sr.widths[idx], rowHeight, label, "", 0, align, sr.styleGuide.UseFill, 0, "")
		startX += sr.widths[idx] + sr.columnSpacing
	}
	sr.pdf.Line(sr.leftMargin+sr.indent, startY+rowHeight, sr.leftMargin+sr.indent+widthTotal, startY+rowHeight)
	spacing := pdfFontScale
	if skipTopBorder {
		spacing = 0
	}
	sr.pdf.SetY(startY + rowHeight + spacing)
	sr.headerDrawn = true
	sr.lastHeaderPage = sr.pdf.PageNo()
}

func (sr *sectionRenderer) consumeHeaderSkip() bool {
	if sr.skipTopBorder {
		sr.skipTopBorder = false
		return true
	}
	return false
}

func (sr *sectionRenderer) shouldDisplayTitle() bool {
	if sr.section.Title == "" {
		return false
	}
	rowType := strings.ToLower(strings.TrimSpace(sr.section.RowType))
	if rowType == "header" || rowType == "detail" {
		return false
	}
	return true
}

func (sr *sectionRenderer) headerColors() ([3]int, [3]int) {
	fill := sr.styleGuide.Header.Background
	if !sr.styleGuide.UseFill {
		fill = colorWhite
	}
	return fill, sr.styleGuide.Header.Text
}

func (sr *sectionRenderer) refreshHeaderState() {
	if sr.lastHeaderPage != sr.pdf.PageNo() {
		sr.headerDrawn = false
	}
}

func (sr *sectionRenderer) ensureSectionHeader() {
	sr.refreshHeaderState()
	if !sr.headerDrawn {
		sr.renderSectionHeader()
	}
}

func (sr *sectionRenderer) ensureRowSpace(rowHeight float64) {
	sr.refreshHeaderState()
	totalHeight := rowHeight
	if sr.rowSpacing > 0 {
		totalHeight += sr.rowSpacing
	}
	if sr.willOverflow(totalHeight) {
		sr.forcePageBreak()
	} else if !sr.headerDrawn {
		sr.renderSectionHeader()
	}
}

func (sr *sectionRenderer) renderDataRow(row pdfResultRow) {
	sr.ensureSectionHeader()
	texts, rowHeight := sr.measureRow(row)
	sr.ensureRowSpace(rowHeight)
	style := sr.resolveRowStyle(row)
	sr.pdf.SetFont(sr.baseFont(), style.FontStyle, 11*pdfFontScale)
	sr.pdf.SetFillColor(style.Background[0], style.Background[1], style.Background[2])
	sr.pdf.SetDrawColor(style.Border[0], style.Border[1], style.Border[2])
	sr.pdf.SetTextColor(style.Text[0], style.Text[1], style.Text[2])
	startX := sr.leftMargin + sr.indent
	startY := sr.pdf.GetY()
	fill := sr.styleGuide.UseFill
	for idx, col := range sr.columns {
		align := sr.resolveAlign(col)
		sr.pdf.SetXY(startX, startY)
		sr.pdf.MultiCell(sr.widths[idx], sr.lineHeight, texts[idx], "", align, fill)
		startX += sr.widths[idx] + sr.columnSpacing
	}
	sr.pdf.SetY(startY + rowHeight)
	if sr.rowSpacing > 0 {
		sr.pdf.SetY(sr.pdf.GetY() + sr.rowSpacing)
	}
}

func (sr *sectionRenderer) resolveColumnText(row pdfResultRow, col PDFColumnConfig) string {
	if row.TypeJSON != 0 && col.HideWhenSummary {
		return ""
	}
	value := row.Data[col.Field]
	if value == nil {
		return ""
	}
	if row.TypeJSON != 0 && strings.EqualFold(col.Field, "docno") {
		dateCfg, ok := sr.globalConfig["docdate"]
		if !ok {
			dateCfg = PDFColumnConfig{Field: "docdate", DataType: "date", Format: "dd/MM/yyyy"}
		}
		dateText := formatValueForColumn(row.Data["docdate"], dateCfg)
		if dateText != "" {
			return fmt.Sprintf("รวมวันที่ %s", dateText)
		}
		return "รวมวันที่"
	}
	return formatValueForColumn(value, col)
}

func (sr *sectionRenderer) resolveRowStyle(row pdfResultRow) pdfStyleBlock {
	return computeRowStyle(row, sr.section.RowType, sr.styleGuide)
}

func (sr *sectionRenderer) totalRowWidth() float64 {
	width := sr.totalWidth()
	if len(sr.columns) > 1 {
		width += sr.columnSpacing * float64(len(sr.columns)-1)
	}
	return width
}

func (sr *sectionRenderer) measureRow(row pdfResultRow) ([]string, float64) {
	texts := make([]string, len(sr.columns))
	maxLines := 1
	sr.pdf.SetFont(sr.baseFont(), "", 11*pdfFontScale)
	for idx, col := range sr.columns {
		text := sr.resolveColumnText(row, col)
		texts[idx] = text
		availableWidth := sr.widths[idx]
		if availableWidth <= 0 {
			availableWidth = 1
		}
		wrapped := sr.pdf.SplitLines([]byte(text), availableWidth)
		lineCount := len(wrapped)
		if lineCount == 0 {
			lineCount = 1
		}
		if lineCount > maxLines {
			maxLines = lineCount
		}
	}
	rowHeight := sr.lineHeight * float64(maxLines)
	if rowHeight == 0 {
		rowHeight = sr.lineHeight
	}
	return texts, rowHeight
}

func (sr *sectionRenderer) estimateRowHeight(row pdfResultRow) float64 {
	_, height := sr.measureRow(row)
	return height
}

func (sr *sectionRenderer) ensurePairSpacing(row pdfResultRow, pairedHeight float64) {
	if pairedHeight <= 0 {
		return
	}
	sr.ensureSectionHeader()
	rowHeight := sr.estimateRowHeight(row)
	total := rowHeight + pairedHeight
	if sr.rowSpacing > 0 {
		total += sr.rowSpacing
	}
	if sr.willOverflow(total) {
		sr.forcePageBreak()
	}
}

func (sr *sectionRenderer) willOverflow(height float64) bool {
	_, pageHeight := sr.pdf.GetPageSize()
	_, margin := sr.pdf.GetAutoPageBreak()
	return sr.pdf.GetY()+height+margin > pageHeight
}

func (sr *sectionRenderer) forcePageBreak() {
	sr.pdf.AddPage()
	sr.headerDrawn = false
	sr.renderSectionHeader()
}

func (sr *sectionRenderer) mergeWithPreviousHeader() {
	sr.skipTopBorder = true
}

func (sr *sectionRenderer) resolveAlign(col PDFColumnConfig) string {
	return resolveAlignValue(col)
}

func resolveAlignValue(col PDFColumnConfig) string {
	trimmedAlign := strings.ToUpper(strings.TrimSpace(col.Align))
	if trimmedAlign == "" {
		if col.Numeric || strings.ToLower(strings.TrimSpace(col.DataType)) == "number" {
			return "R"
		}
		return "L"
	}
	alignRune := []rune(trimmedAlign)[0]
	if alignRune != 'L' && alignRune != 'C' && alignRune != 'R' {
		return "L"
	}
	return string(alignRune)
}

func computeColumnWidths(columns []PDFColumnConfig, totalWidth float64) []float64 {
	if len(columns) == 0 {
		return nil
	}
	totalFlex := 0
	for _, col := range columns {
		flex := col.Flex
		if flex <= 0 {
			flex = 1
		}
		totalFlex += flex
	}
	if totalFlex == 0 {
		totalFlex = len(columns)
	}
	widths := make([]float64, len(columns))
	for idx, col := range columns {
		flex := col.Flex
		if flex <= 0 {
			flex = 1
		}
		widths[idx] = totalWidth * (float64(flex) / float64(totalFlex))
	}
	return widths
}

func groupRowsByAlias(rows []pdfResultRow) map[string][]pdfResultRow {
	result := make(map[string][]pdfResultRow)
	for _, row := range rows {
		result[row.Alias] = append(result[row.Alias], row)
	}
	return result
}

func collectSectionLevels(rows []pdfResultRow) []int {
	levels := make(map[int]struct{})
	for _, row := range rows {
		levels[row.TypeJSON] = struct{}{}
	}
	if len(levels) == 0 {
		return nil
	}
	result := make([]int, 0, len(levels))
	for lvl := range levels {
		result = append(result, lvl)
	}
	slices.Sort(result)
	return result
}

func groupRowsByField(rows []pdfResultRow, field string) map[string][]pdfResultRow {
	result := make(map[string][]pdfResultRow)
	for _, row := range rows {
		key := stringifyValue(row.Data[field])
		if key == "" {
			continue
		}
		result[key] = append(result[key], row)
	}
	return result
}

func stringifyValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case json.Number:
		return v.String()
	case time.Time:
		return v.Format("2006-01-02")
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatDisplayValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	case json.Number:
		return v.String()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatValueForColumn(value any, col PDFColumnConfig) string {
	if value == nil {
		return ""
	}
	dataType := strings.ToLower(strings.TrimSpace(col.DataType))
	switch dataType {
	case "number":
		return formatNumericByColumn(value, col)
	case "date":
		return formatDateValue(value, col.Format, col.UseBuddhistYear)
	case "time":
		return formatTimeValue(value, col.Format)
	default:
		if col.Numeric {
			return formatNumericByColumn(value, col)
		}
		return formatDisplayValue(value)
	}
}

func formatDateValue(value any, pattern string, useBuddhist bool) string {
	date, ok := parseDateTimeValue(value)
	if !ok {
		return formatDisplayValue(value)
	}
	layout := convertDateFormatPattern(pattern)
	if layout == "" {
		layout = "02/01/2006"
	}
	formatted := date.Format(layout)
	if useBuddhist {
		formatted = applyBuddhistYear(formatted, layout, date)
	}
	return formatted
}

func formatTimeValue(value any, pattern string) string {
	timeValue, ok := parseTimeOfDay(value)
	if !ok {
		return formatDisplayValue(value)
	}
	layout := convertDateFormatPattern(pattern)
	if layout == "" {
		layout = "15:04"
	}
	return timeValue.Format(layout)
}

func parseDateTimeValue(value any) (time.Time, bool) {
	switch v := value.(type) {
	case time.Time:
		return v, true
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return time.Time{}, false
		}
		layouts := []string{
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02",
			"02/01/2006",
			"20060102",
		}
		for _, layout := range layouts {
			if parsed, err := time.Parse(layout, trimmed); err == nil {
				return parsed, true
			}
		}
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return time.Unix(int64(f), 0), true
		}
	}
	return time.Time{}, false
}

func parseTimeOfDay(value any) (time.Time, bool) {
	switch v := value.(type) {
	case time.Time:
		return v, true
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return time.Time{}, false
		}
		layouts := []string{"15:04:05", "15:04"}
		for _, layout := range layouts {
			if parsed, err := time.Parse(layout, trimmed); err == nil {
				return parsed, true
			}
		}
		if parsed, ok := parseDateTimeValue(trimmed); ok {
			return parsed, true
		}
	default:
		return parseDateTimeValue(v)
	}
	return time.Time{}, false
}

func convertDateFormatPattern(pattern string) string {
	trimmed := strings.TrimSpace(pattern)
	if trimmed == "" {
		return ""
	}
	replacements := []struct {
		token string
		value string
	}{
		{"yyyy", "2006"},
		{"yy", "06"},
		{"MM", "01"},
		{"dd", "02"},
		{"HH", "15"},
		{"mm", "04"},
		{"ss", "05"},
		{"M", "1"},
		{"d", "2"},
		{"H", "15"},
		{"m", "4"},
		{"s", "5"},
	}
	result := trimmed
	for _, repl := range replacements {
		result = strings.ReplaceAll(result, repl.token, repl.value)
	}
	return result
}

func applyBuddhistYear(formatted string, layout string, date time.Time) string {
	yearIdx := strings.Index(layout, "2006")
	if yearIdx < 0 {
		return formatted
	}
	buddhistYear := fmt.Sprintf("%04d", date.Year()+543)
	if len(formatted) < yearIdx+4 {
		return formatted
	}
	return formatted[:yearIdx] + buddhistYear + formatted[yearIdx+4:]
}

func formatNumberWithPrecision(value float64, decimals int) string {
	if decimals < 0 {
		decimals = 0
	}
	formatted := fmt.Sprintf("%.*f", decimals, value)
	return insertThousands(formatted)
}

func insertThousands(formatted string) string {
	parts := strings.SplitN(formatted, ".", 2)
	intPart := parts[0]
	fracPart := ""
	if len(parts) == 2 {
		fracPart = parts[1]
	}
	sign := ""
	if strings.HasPrefix(intPart, "-") {
		sign = "-"
		intPart = intPart[1:]
	}
	var builder strings.Builder
	for i, r := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			builder.WriteByte(',')
		}
		builder.WriteRune(r)
	}
	result := builder.String()
	if sign != "" {
		result = sign + result
	}
	if fracPart != "" {
		return result + "." + fracPart
	}
	return result
}

func formatNumberByPattern(value float64, pattern string) string {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return formatNumberWithPrecision(value, 2)
	}
	if strings.Contains(pattern, ";") {
		pattern = strings.Split(pattern, ";")[0]
	}
	decimals := 0
	if idx := strings.Index(pattern, "."); idx >= 0 {
		for _, r := range pattern[idx+1:] {
			if r == '0' || r == '#' {
				decimals++
			}
		}
	}
	return formatNumberWithPrecision(value, decimals)
}

func formatNumericByColumn(value any, col PDFColumnConfig) string {
	if value == nil {
		return ""
	}
	var (
		f      float64
		parsed bool
	)
	switch v := value.(type) {
	case string:
		if v == "" {
			return ""
		}
		if fv, ok := parseNumericString(v); ok {
			f = fv
			parsed = true
		} else {
			return v
		}
	case fmt.Stringer:
		if fv, ok := parseNumericString(v.String()); ok {
			f = fv
			parsed = true
		} else {
			return v.String()
		}
	default:
		f = toFloat64Value(value)
		parsed = true
	}
	if !parsed {
		return formatDisplayValue(value)
	}
	if col.FormatNumber != "" {
		return formatNumberByPattern(f, col.FormatNumber)
	}
	if col.DecimalPlaces != nil {
		return formatNumberWithPrecision(f, *col.DecimalPlaces)
	}
	return formatNumberWithPrecision(f, 2)
}

func parseNumericString(input string) (float64, bool) {
	cleaned := strings.ReplaceAll(strings.TrimSpace(input), ",", "")
	if cleaned == "" {
		return 0, false
	}
	if strings.EqualFold(cleaned, "nan") {
		return 0, false
	}
	f, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func sectionVisible(section PDFSectionConfig) bool {
	if section.Visible == nil {
		return true
	}
	return *section.Visible
}

// buildInterleavedRows - สร้าง rows ที่มีทั้งข้อมูลและยอดรวมแบบ interleave
func buildInterleavedRows(dataRows []map[string]any, config *SummaryConfig, queryNum int, guid string, alias string, aliasKey string) []*resultRowPayload {
	result := make([]*resultRowPayload, 0, len(dataRows)*2)
	lineNumber := 1
	appendRow := func(row map[string]any, typeJSON int) {
		rowClone := cloneMap(row)
		jsonPayload := marshalRowToJSON(rowClone, alias, guid, queryNum, lineNumber)
		payload := &resultRowPayload{
			Alias:       alias,
			AliasKey:    aliasKey,
			QueryNumber: queryNum,
			TypeJSON:    typeJSON,
			DataJSON:    jsonPayload,
			Data:        rowClone,
		}
		result = append(result, payload)
		lineNumber++
	}

	// Helper to convert value to float64
	toFloat := func(v any) float64 {
		switch val := v.(type) {
		case float64:
			return val
		case float32:
			return float64(val)
		case int:
			return float64(val)
		case int64:
			return float64(val)
		case string:
			f, _ := strconv.ParseFloat(val, 64)
			return f
		default:
			return 0
		}
	}

	// Helper to make group key
	makeGroupKey := func(row map[string]any, fields []string) string {
		parts := make([]string, len(fields))
		for i, field := range fields {
			parts[i] = fmt.Sprintf("%v", row[field])
		}
		return strings.Join(parts, "|")
	}

	// If no summary config, just return data rows with typejson=0
	if config == nil || len(config.Levels) == 0 {
		for _, row := range dataRows {
			appendRow(row, 0)
		}
		return result
	}

	// Track current group keys for each level
	currentKeys := make([]string, len(config.Levels))
	for i := range currentKeys {
		currentKeys[i] = ""
	}

	// Accumulators for each level
	levelAccums := make([]map[string]float64, len(config.Levels))
	for i := range levelAccums {
		levelAccums[i] = make(map[string]float64)
	}

	// Grand total accumulator
	grandTotalAccum := make(map[string]float64)

	// Collect all sum fields from all levels
	allSumFields := make(map[string]bool)
	for _, level := range config.Levels {
		for _, field := range level.SumFields {
			allSumFields[field] = true
		}
	}

	// Process each data row
	for rowIdx, row := range dataRows {
		// Check each level for group changes (from coarsest to finest)
		// ต้อง check จากหยาบ→ละเอียด เพราะถ้า level หยาบเปลี่ยน ต้อง emit summaries ของ level ละเอียดก่อน
		for levelIdx := 0; levelIdx < len(config.Levels); levelIdx++ {
			level := config.Levels[levelIdx]
			newKey := makeGroupKey(row, level.GroupByFields)

			if currentKeys[levelIdx] != "" && currentKeys[levelIdx] != newKey {
				// Group changed - emit summaries from finest to this level
				// เมื่อ group level นี้เปลี่ยน ต้อง emit ยอดรวมของ level ละเอียดกว่าก่อน แล้วค่อย emit level นี้
				for emitIdx := len(config.Levels) - 1; emitIdx >= levelIdx; emitIdx-- {
					if currentKeys[emitIdx] != "" {
						emitLevel := config.Levels[emitIdx]
						summaryRow := make(map[string]any)
						// Copy group fields from previous row
						if rowIdx > 0 {
							prevRow := dataRows[rowIdx-1]
							for _, field := range emitLevel.GroupByFields {
								summaryRow[field] = prevRow[field]
							}
						}
						// Add accumulated sums
						for _, field := range emitLevel.SumFields {
							summaryRow[field] = levelAccums[emitIdx][field]
						}
						typeJSON := emitLevel.TypeJSON
						if typeJSON == 0 {
							typeJSON = emitIdx + 1 // default: 1, 2, 3...
						}
						appendRow(summaryRow, typeJSON)
					}
				}

				// Reset this level and all finer levels
				for i := levelIdx; i < len(config.Levels); i++ {
					levelAccums[i] = make(map[string]float64)
					currentKeys[i] = ""
				}

				// ออกจาก loop เพราะ level หยาบเปลี่ยนแล้ว ไม่ต้อง check level ละเอียดต่อ
				break
			}
		}

		// Insert data row with typejson=0
		appendRow(row, 0)

		// Update current keys for all levels
		for levelIdx, level := range config.Levels {
			currentKeys[levelIdx] = makeGroupKey(row, level.GroupByFields)
		}

		// Accumulate sums for all levels and grand total
		for levelIdx, level := range config.Levels {
			for _, field := range level.SumFields {
				if val, ok := row[field]; ok {
					levelAccums[levelIdx][field] += toFloat(val)
				}
			}
		}
		for field := range allSumFields {
			if val, ok := row[field]; ok {
				grandTotalAccum[field] += toFloat(val)
			}
		}
	}

	// Emit remaining summaries for all levels (from finest to coarsest)
	for levelIdx := len(config.Levels) - 1; levelIdx >= 0; levelIdx-- {
		if currentKeys[levelIdx] != "" {
			level := config.Levels[levelIdx]
			summaryRow := make(map[string]any)
			if len(dataRows) > 0 {
				lastRow := dataRows[len(dataRows)-1]
				for _, field := range level.GroupByFields {
					summaryRow[field] = lastRow[field]
				}
			}
			for _, field := range level.SumFields {
				summaryRow[field] = levelAccums[levelIdx][field]
			}
			typeJSON := level.TypeJSON
			if typeJSON == 0 {
				typeJSON = levelIdx + 1
			}
			appendRow(summaryRow, typeJSON)
		}
	}

	// Emit grand total if requested
	if config.GrandTotal {
		grandTotalRow := make(map[string]any)
		for field := range allSumFields {
			grandTotalRow[field] = grandTotalAccum[field]
		}
		grandTypeJSON := config.GrandTotalType
		if grandTypeJSON == 0 {
			grandTypeJSON = 99 // default
		}
		appendRow(grandTotalRow, grandTypeJSON)
	}

	return result
}

func computeAliasDepths(rootAliasKeys []string, aliasParents map[string]string, aliasRows map[string][]*resultRowPayload) map[string]int {
	depthCache := make(map[string]int)
	var resolve func(string) int
	resolve = func(aliasKey string) int {
		if aliasKey == "" {
			return 0
		}
		if depth, ok := depthCache[aliasKey]; ok {
			return depth
		}
		parentKey, ok := aliasParents[aliasKey]
		if !ok || parentKey == "" {
			depthCache[aliasKey] = 0
			return 0
		}
		depth := resolve(parentKey) + 1
		depthCache[aliasKey] = depth
		return depth
	}

	for _, aliasKey := range rootAliasKeys {
		resolve(aliasKey)
	}
	for aliasKey := range aliasParents {
		resolve(aliasKey)
	}
	for aliasKey := range aliasRows {
		resolve(aliasKey)
	}

	return depthCache
}

func orderRowsByHierarchy(rootAliasKeys []string, aliasChildren map[string][]string, aliasRows map[string][]*resultRowPayload, aliasLinkConfigs map[string]*LinkConfig, aliasDepth map[string]int) []*resultRowPayload {
	grouped := make(map[string]map[string][]*resultRowPayload)
	totalRows := 0
	for aliasKey, rows := range aliasRows {
		totalRows += len(rows)
		cfg := aliasLinkConfigs[aliasKey]
		group := make(map[string][]*resultRowPayload)
		if cfg == nil || len(cfg.ChildKeys) == 0 {
			group[rootGroupKey] = rows
			for _, row := range rows {
				row.groupKey = rootGroupKey
			}
		} else {
			for _, row := range rows {
				key := buildCompositeKey(row.Data, cfg.ChildKeys)
				row.groupKey = key
				group[key] = append(group[key], row)
			}
		}
		grouped[aliasKey] = group
	}

	if len(rootAliasKeys) == 0 {
		for aliasKey := range aliasRows {
			rootAliasKeys = append(rootAliasKeys, aliasKey)
		}
	}

	ordered := make([]*resultRowPayload, 0, totalRows)
	visited := make(map[*resultRowPayload]bool)
	var emit func(aliasKey string, groupKey string, depth int)
	emit = func(aliasKey string, groupKey string, depth int) {
		group := grouped[aliasKey]
		if group == nil {
			return
		}
		rows := group[groupKey]
		if len(rows) == 0 {
			return
		}
		for _, row := range rows {
			if visited[row] {
				continue
			}
			visited[row] = true
			row.Level = depth
			ordered = append(ordered, row)

			if row.TypeJSON != 0 {
				continue
			}
			for _, childAlias := range aliasChildren[aliasKey] {
				childCfg := aliasLinkConfigs[childAlias]
				if childCfg == nil || len(childCfg.ParentKeys) == 0 {
					emit(childAlias, rootGroupKey, depth+1)
					continue
				}
				childGroupKey := buildCompositeKey(row.Data, childCfg.ParentKeys)
				if childGroupKey == "" {
					continue
				}
				emit(childAlias, childGroupKey, depth+1)
			}
		}
	}

	seenRoots := make(map[string]bool)
	for _, aliasKey := range rootAliasKeys {
		if seenRoots[aliasKey] {
			continue
		}
		seenRoots[aliasKey] = true
		emit(aliasKey, rootGroupKey, 0)
	}

	for aliasKey, rows := range aliasRows {
		depth := aliasDepth[aliasKey]
		for _, row := range rows {
			if visited[row] {
				continue
			}
			row.Level = depth
			ordered = append(ordered, row)
		}
	}

	return ordered
}

// sanitizeRowForJSON replaces NaN/Inf float values with nil so PostgreSQL JSONB accepts the payload
func sanitizeRowForJSON(row map[string]any) {
	for key, value := range row {
		switch v := value.(type) {
		case float64:
			if math.IsNaN(v) || math.IsInf(v, 0) {
				row[key] = nil
			}
		case float32:
			f := float64(v)
			if math.IsNaN(f) || math.IsInf(f, 0) {
				row[key] = nil
			}
		}
	}
}

// marshalRowToJSON marshals data rows while logging any unexpected serialization failures
func marshalRowToJSON(row map[string]any, alias string, guid string, queryNum int, lineNumber int) string {
	sanitizeRowForJSON(row)
	if alias != "" {
		row[aliasMetadataKey] = alias
	}
	jsonData, err := json.Marshal(row)
	if alias != "" {
		delete(row, aliasMetadataKey)
	}
	if err != nil {
		logger.Error("JSON marshal error guid=%s query=%d line=%d: %v row=%+v", guid, queryNum, lineNumber, err, row)
		return "{}"
	}
	return string(jsonData)
}

// PDFConfig - โครงสร้างการตั้งค่า PDF
type PDFConfig struct {
	Title       string `json:"title"`       // ชื่อรายงาน
	Orientation string `json:"orientation"` // "P" (Portrait) หรือ "L" (Landscape)
	PageSize    string `json:"pagesize"`    // "A4", "A3", "Letter" เป็นต้น
}

type PDFColumnConfig struct {
	Field           string `json:"field"`
	Label           string `json:"label"`
	Flex            int    `json:"flex"`
	Align           string `json:"align"`
	Numeric         bool   `json:"numeric"`
	HideWhenSummary bool   `json:"hidewhensummary"`
	DecimalPlaces   *int   `json:"decimalplaces"`
	FormatNumber    string `json:"formatnumber"`
	DataType        string `json:"datatype"`
	Format          string `json:"format"`
	UseBuddhistYear bool   `json:"usebuddhistyear"`
}

type PDFSectionConfig struct {
	Alias   string            `json:"alias"`
	Title   string            `json:"title"`
	RowType string            `json:"rowtype"`
	Visible *bool             `json:"visible"`
	Columns []PDFColumnConfig `json:"columns"`
}

type PDFStyleBlockConfig struct {
	Background string `json:"background"`
	Text       string `json:"text"`
	Border     string `json:"border"`
	FontWeight string `json:"fontweight"`
}

type PDFTableStyleConfig struct {
	RowSpacing    float64 `json:"rowspacing"`
	ColumnSpacing float64 `json:"columnspacing"`
	GridColor     string  `json:"gridcolor"`
}

type PDFStyles struct {
	Palette string              `json:"palette"`
	UseFill bool                `json:"usefill"`
	Header  PDFStyleBlockConfig `json:"header"`
	Detail  PDFStyleBlockConfig `json:"detail"`
	Summary PDFStyleBlockConfig `json:"summary"`
	Table   PDFTableStyleConfig `json:"table"`
}

type PDFColumnSchema struct {
	Label           string `json:"label"`
	Flex            int    `json:"flex"`
	Align           string `json:"align"`
	DataType        string `json:"datatype"`
	Format          string `json:"format"`
	HideWhenSummary bool   `json:"hidewhensummary"`
	UseBuddhistYear bool   `json:"usebuddhistyear"`
}

type PDFLayoutConfig struct {
	SchemaVersion int                        `json:"schemaversion"`
	Sections      []PDFSectionConfig         `json:"sections"`
	Styles        PDFStyles                  `json:"styles"`
	ColumnSchema  map[string]PDFColumnSchema `json:"columnschema"`
	NumberFormats map[string]string          `json:"numberformats"`
}

type resultToPDFPayload struct {
	HoldingCode  string            `json:"holdingcode"`
	Guid         string            `json:"guid"`
	PDFConfig    PDFConfig         `json:"pdfconfig"`
	ColumnOrder  []string          `json:"columnorder"`
	ColumnNames  map[string]string `json:"columnnames"`
	LayoutConfig PDFLayoutConfig   `json:"layoutconfig"`
}

type pdfResultRow struct {
	Alias       string
	QueryNumber int
	LineNumber  int
	Level       int
	TypeJSON    int
	Data        map[string]any
}

func normalizeAliasKey(alias string) string {
	return strings.ToLower(strings.TrimSpace(alias))
}

func cloneRowSlice(rows []map[string]any) []map[string]any {
	if len(rows) == 0 {
		return nil
	}
	cloned := make([]map[string]any, len(rows))
	copy(cloned, rows)
	return cloned
}

func cloneMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dup := make(map[string]any, len(src))
	for k, v := range src {
		dup[k] = v
	}
	return dup
}

func filterRowsByLink(childRows []map[string]any, parentRows []map[string]any, cfg *LinkConfig) []map[string]any {
	if cfg == nil {
		return childRows
	}
	if len(parentRows) == 0 {
		return []map[string]any{}
	}

	parentKeySet := make(map[string]struct{}, len(parentRows))
	for _, parentRow := range parentRows {
		key := buildCompositeKey(parentRow, cfg.ParentKeys)
		parentKeySet[key] = struct{}{}
	}

	filtered := make([]map[string]any, 0, len(childRows))
	for _, row := range childRows {
		key := buildCompositeKey(row, cfg.ChildKeys)
		if _, ok := parentKeySet[key]; ok {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func buildCompositeKey(row map[string]any, fields []string) string {
	parts := make([]string, len(fields))
	for i, field := range fields {
		parts[i] = fmt.Sprintf("%v", row[field])
	}
	return strings.Join(parts, "|")
}

func normalizeColumnValue(value any, columnType *sql.ColumnType) any {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case []byte:
		strVal := strings.TrimSpace(string(v))
		if strVal == "" {
			return nil
		}
		dbType := ""
		if columnType != nil {
			dbType = strings.ToUpper(columnType.DatabaseTypeName())
		}
		if isBooleanDatabaseType(dbType) {
			if boolVal, err := strconv.ParseBool(strVal); err == nil {
				return boolVal
			}
		}
		if isNumericDatabaseType(dbType) {
			if strings.Contains(dbType, "INT") && !strings.Contains(dbType, "POINT") {
				if intVal, err := strconv.ParseInt(strVal, 10, 64); err == nil {
					return intVal
				}
			}
			if floatVal, err := strconv.ParseFloat(strVal, 64); err == nil {
				return floatVal
			}
		}
		if floatVal, err := strconv.ParseFloat(strVal, 64); err == nil {
			return floatVal
		}
		return strVal
	default:
		return value
	}
}

func toFloat64Value(value any) float64 {
	if value == nil {
		return 0
	}
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return f
		}
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return f
		}
	}
	return 0
}

func isNumericDatabaseType(dbType string) bool {
	if dbType == "" {
		return false
	}
	if strings.Contains(dbType, "NUMERIC") || strings.Contains(dbType, "DECIMAL") || strings.Contains(dbType, "FLOAT") || strings.Contains(dbType, "DOUBLE") || strings.Contains(dbType, "REAL") || strings.Contains(dbType, "MONEY") {
		return true
	}
	switch dbType {
	case "INT", "INT2", "INT4", "INT8", "INTEGER", "SMALLINT", "BIGINT", "SERIAL", "BIGSERIAL":
		return true
	}
	return false
}

func isBooleanDatabaseType(dbType string) bool {
	return strings.Contains(dbType, "BOOL")
}
