package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myglobal"
	mypg "smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/goapi/process/stockengine"

	"github.com/labstack/echo/v4"
)

const processStockCalcCostCommandID = "processstockcalccost"

// processStockCalcCostRequest defines the payload structure accepted by ProcessStockCalcCostHandler.
type processStockCalcCostRequest struct {
	HoldingCode  string          `json:"holdingcode"`
	BusinessCode string          `json:"businesscode"`
	CommandID    string          `json:"commandid"`
	ItemCodeList json.RawMessage `json:"itemcodelist"`
	PointQty     *int            `json:"pointqty"`
	PointAmount  *int            `json:"pointamount"`
	PointCost    *int            `json:"pointcost"`
}

// processStockCalcCostItemResult holds per-item execution metadata for API responses.
type processStockCalcCostItemResult struct {
	ItemCode   string `json:"itemcode"`
	Status     string `json:"status"`
	Rows       int    `json:"rows"`
	DurationMs int64  `json:"durationms"`
	Error      string `json:"error,omitempty"`
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

	holdingCode, businessCode, scopeErr := authenticatedCompanyContext(c, payload.HoldingCode, payload.BusinessCode)
	if scopeErr != nil {
		return c.JSON(scopeErr.Status, map[string]any{
			"success": false,
			"code":    scopeErr.Code,
			"message": scopeErr.Message,
		})
	}
	payload.HoldingCode = holdingCode
	payload.BusinessCode = businessCode

	if payload.CommandID == "" {
		payload.CommandID = processStockCalcCostCommandID
	} else if strings.ToLower(payload.CommandID) != processStockCalcCostCommandID {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Unsupported commandid",
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
			"error": "itemcodelist must contain at least one item",
			"code":  "EMPTY_ITEM_LIST",
		})
	}

	pointQty := resolvePoint(payload.PointQty, myglobal.ConfigSystem.StockQtyPoint)
	pointAmount := resolvePoint(payload.PointAmount, myglobal.ConfigSystem.StockAmountPoint)
	pointCost := resolvePoint(payload.PointCost, myglobal.ConfigSystem.StockCostPoint)

	results, err := runProcessStockCalcCost(payload.HoldingCode, payload.BusinessCode, itemCodes, pointQty, pointAmount, pointCost)
	if err != nil {
		logger.Error("ProcessStockCalcCostHandler failed: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to process stock cost",
			"code":  "PROCESS_STOCK_COST_ERROR",
		})
	}

	logger.Success("ProcessStockCalcCostHandler completed | shop=%s items=%d duration=%v", payload.HoldingCode, len(results), time.Since(start))

	return c.JSON(http.StatusOK, map[string]any{
		"status":         "success",
		"code":           200,
		"holdingcode":    payload.HoldingCode,
		"businesscode":   payload.BusinessCode,
		"commandid":      payload.CommandID,
		"processeditems": len(results),
		"durationms":     time.Since(start).Milliseconds(),
		"pointqty":       pointQty,
		"pointamount":    pointAmount,
		"pointcost":      pointCost,
		"items":          results,
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

// runProcessStockCalcCost คำนวณต้นทุนใหม่ทั้งชีวิตของสินค้าที่ระบุ แล้วรอจนเสร็จ
//
// ใช้ตอนผู้ใช้สั่งประมวลผลเองจากจอเครื่องมือ จึงคำนวณทันทีไม่ผ่านคิว
// เอกสารที่บันทึกตามปกติจะเข้าคิวให้ตัวประมวลผลเบื้องหลังแทน (stockengine.MarkDirty)
func runProcessStockCalcCost(holdingCode, businessCode string, itemCodes []string, pointQty, pointAmount, pointCost int) ([]processStockCalcCostItemResult, error) {
	if len(itemCodes) == 0 {
		return nil, errors.New("itemcodelist must contain at least one item")
	}

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return nil, fmt.Errorf("database connection error: %w", err)
	}

	options := stockengine.DefaultOptions()
	options.PointQty = pointQty
	options.PointAmount = pointAmount
	options.PointCost = pointCost

	ctx := context.Background()
	results := make([]processStockCalcCostItemResult, 0, len(itemCodes))

	for idx, itemCode := range itemCodes {
		logger.Info("ProcessStockCalcCost processing %s (%d/%d) for holding %s company %s",
			itemCode, idx+1, len(itemCodes), holdingCode, businessCode)
		itemStart := time.Now()

		rows, err := stockengine.Recalculate(ctx, db,
			stockengine.Scope{BusinessCode: businessCode, ItemCode: itemCode},
			myglobal.TransFlagsToProcess, options)

		item := processStockCalcCostItemResult{
			ItemCode:   itemCode,
			Status:     "completed",
			Rows:       rows,
			DurationMs: time.Since(itemStart).Milliseconds(),
		}
		if err != nil {
			// สินค้าตัวที่พังต้องไม่ฉุดตัวอื่น ผู้ใช้เห็นรายตัวว่าอันไหนไม่ผ่าน
			logger.Error("ProcessStockCalcCost %s: %v", itemCode, err)
			item.Status = "failed"
			item.Error = err.Error()
		}
		results = append(results, item)
	}

	return results, nil
}
