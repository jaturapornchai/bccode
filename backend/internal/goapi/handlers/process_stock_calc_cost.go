package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myglobal"
	mypg "smlcloudplatform/internal/goapi/mypg"
	processstock "smlcloudplatform/internal/goapi/process/process-stock"

	"github.com/labstack/echo/v4"
)

const processStockCalcCostCommandID = "processstockcalccost"

// processStockCalcCostRequest defines the payload structure accepted by ProcessStockCalcCostHandler.
type processStockCalcCostRequest struct {
	HoldingCode  string          `json:"holdingcode"`
	CommandID    string          `json:"commandid"`
	ItemCodeList json.RawMessage `json:"itemcodelist"`
	DeleteFirst  *bool           `json:"deletefirst"`
	PointQty     *int            `json:"pointqty"`
	PointAmount  *int            `json:"pointamount"`
	PointCost    *int            `json:"pointcost"`
	Incremental  *bool           `json:"incremental"` // ใช้ incremental calculation (ตรวจ checksum ก่อน)
	MinimalLog   *bool           `json:"minimallog"`  // ใช้ UPSERT แทน DELETE+INSERT เพื่อลด WAL
}

// processStockCalcCostItemResult holds per-item execution metadata for API responses.
type processStockCalcCostItemResult struct {
	ItemCode   string `json:"itemcode"`
	Status     string `json:"status"`
	DurationMs int64  `json:"durationms"`
}

// ProcessStockCalcCostHandler executes ProductCalcCost for a list of item codes.
func ProcessStockCalcCostHandler(c echo.Context) error {
	start := time.Now()
	var payload processStockCalcCostRequest

	if err := c.Bind(&payload); err != nil {
		logger.Warn("ProcessStockCalcCostHandler invalid JSON: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid JSON payload",
			"code":  "INVALID_JSON",
		})
	}

	payload.HoldingCode = strings.TrimSpace(payload.HoldingCode)
	if payload.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "holdingcode is required",
			"code":  "MISSING_HOLDING_CODE",
		})
	}

	if payload.CommandID == "" {
		payload.CommandID = processStockCalcCostCommandID
	} else if strings.ToLower(payload.CommandID) != processStockCalcCostCommandID {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Unsupported command_id",
			"code":  "INVALID_COMMAND",
		})
	}

	itemCodes, err := payload.extractItemCodes()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": err.Error(),
			"code":  "INVALID_ITEM_CODE_LIST",
		})
	}
	if len(itemCodes) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "item_code_list must contain at least one item",
			"code":  "EMPTY_ITEM_LIST",
		})
	}

	pointQty := resolvePoint(payload.PointQty, myglobal.ConfigSystem.StockQtyPoint)
	pointAmount := resolvePoint(payload.PointAmount, myglobal.ConfigSystem.StockAmountPoint)
	pointCost := resolvePoint(payload.PointCost, myglobal.ConfigSystem.StockCostPoint)

	deleteFirst := true
	if payload.DeleteFirst != nil {
		deleteFirst = *payload.DeleteFirst
	}

	// Incremental mode (default: false for API to maintain backward compatibility)
	incremental := false
	if payload.Incremental != nil {
		incremental = *payload.Incremental
	}

	// MinimalLog mode (default: false for API to maintain backward compatibility)
	minimalLog := false
	if payload.MinimalLog != nil {
		minimalLog = *payload.MinimalLog
	}

	results, err := runProcessStockCalcCost(payload.HoldingCode, itemCodes, pointQty, pointAmount, pointCost, deleteFirst, incremental, minimalLog)
	if err != nil {
		logger.Error("ProcessStockCalcCostHandler failed: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to process stock cost",
			"code":  "PROCESS_STOCK_COST_ERROR",
		})
	}

	logger.Success("ProcessStockCalcCostHandler completed | shop=%s items=%d duration=%v", payload.HoldingCode, len(results), time.Since(start))

	return c.JSON(http.StatusOK, map[string]any{
		"status":          "success",
		"code":            200,
		"holdingcode":     payload.HoldingCode,
		"command_id":      payload.CommandID,
		"processed_items": len(results),
		"duration_ms":     time.Since(start).Milliseconds(),
		"delete_first":    deleteFirst,
		"incremental":     incremental,
		"minimal_log":     minimalLog,
		"point_qty":       pointQty,
		"point_amount":    pointAmount,
		"point_cost":      pointCost,
		"items":           results,
	})
}

func (p processStockCalcCostRequest) extractItemCodes() ([]string, error) {
	return parseItemCodesFromJSON(p.ItemCodeList)
}

func sanitizeItemCodeList(items []string) []string {
	seen := make(map[string]struct{})
	sanitized := make([]string, 0, len(items))
	for _, raw := range items {
		cleaned := strings.ToUpper(strings.TrimSpace(raw))
		if cleaned == "" {
			continue
		}
		if _, exists := seen[cleaned]; exists {
			continue
		}
		seen[cleaned] = struct{}{}
		sanitized = append(sanitized, cleaned)
	}
	return sanitized
}

func resolvePoint(custom *int, fallback int) int {
	if custom == nil {
		return fallback
	}
	value := *custom
	if value < 0 {
		return 0
	}
	if value > 6 {
		return 6
	}
	return value
}

func parseItemCodesFromJSON(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 {
		return nil, errors.New("item_code_list is required")
	}

	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		return sanitizeItemCodeList(arr), nil
	}

	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		if strings.TrimSpace(single) == "" {
			return nil, errors.New("item_code_list is empty")
		}
		return sanitizeItemCodeList(strings.Split(single, ",")), nil
	}

	return nil, fmt.Errorf("item_code_list must be array or string, got: %s", string(raw))
}

func runProcessStockCalcCost(holdingCode string, itemCodes []string, pointQty, pointAmount, pointCost int, deleteFirst, incremental, minimalLog bool) ([]processStockCalcCostItemResult, error) {
	if len(itemCodes) == 0 {
		return nil, errors.New("item_code_list must contain at least one item")
	}

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return nil, fmt.Errorf("database connection error: %w", err)
	}

	results := make([]processStockCalcCostItemResult, 0, len(itemCodes))
	for idx, itemCode := range itemCodes {
		logger.Info("ProcessStockCalcCost processing %s (%d/%d) for shop %s (incremental=%v, minimalLog=%v)",
			itemCode, idx+1, len(itemCodes), holdingCode, incremental, minimalLog)
		itemStart := time.Now()

		if incremental {
			// Use incremental mode with checksum checking
			processstock.ProductCalcCostIncremental(db, holdingCode, itemCode, pointQty, pointAmount, pointCost, incremental, minimalLog)
		} else {
			// Legacy mode
			processstock.ProductCalcCost(db, holdingCode, itemCode, pointQty, pointAmount, pointCost, deleteFirst)
		}

		results = append(results, processStockCalcCostItemResult{
			ItemCode:   itemCode,
			Status:     "completed",
			DurationMs: time.Since(itemStart).Milliseconds(),
		})
	}

	return results, nil
}
