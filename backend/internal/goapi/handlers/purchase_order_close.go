package handlers

import (
	"context"
	"fmt"
	"net/http"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myclickhouse"
	"smlcloudplatform/internal/goapi/mypg"
	"time"

	"github.com/labstack/echo/v4"
)

// ManualClosePORequest — request สำหรับปิด/เปิดเอกสารใบสั่งซื้อด้วยมือ
type ManualClosePORequest struct {
	HoldingCode  string `json:"holding_code"`
	DocNo        string `json:"docno"`
	Action       string `json:"action"`         // "close" หรือ "open"
	ActionByCode string `json:"action_by_code"` // รหัสผู้กระทำ
	ActionByName string `json:"action_by_name"` // ชื่อผู้กระทำ
	Reason       string `json:"reason"`         // เหตุผล (ต้องกรอก)
}

// ManualClosePOHandler — ปิด/เปิดเอกสารใบสั่งซื้อด้วยมือ
// POST /api/purchase-order/manual-close
func ManualClosePOHandler(c echo.Context) error {
	var req ManualClosePORequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"status":  "error",
			"code":    400,
			"message": "รูปแบบข้อมูลไม่ถูกต้อง",
		})
	}

	// Validate
	if req.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"status":  "error",
			"code":    400,
			"message": "กรุณาระบุ holding_code",
		})
	}
	if req.DocNo == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"status":  "error",
			"code":    400,
			"message": "กรุณาระบุเลขที่เอกสาร (docno)",
		})
	}
	if req.Action != "close" && req.Action != "open" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"status":  "error",
			"code":    400,
			"message": "action ต้องเป็น 'close' หรือ 'open'",
		})
	}
	if req.Reason == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"status":  "error",
			"code":    400,
			"message": "กรุณาระบุเหตุผล",
		})
	}
	if req.ActionByCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"status":  "error",
			"code":    400,
			"message": "กรุณาระบุรหัสผู้กระทำ (action_by_code)",
		})
	}

	ctx := context.Background()

	// เชื่อมต่อ PostgreSQL
	db, err := mypg.PgSqlFastConnect(req.HoldingCode)
	if err != nil {
		logger.Error("[ManualClosePO] เชื่อมต่อ PostgreSQL ไม่สำเร็จ: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"status":  "error",
			"code":    500,
			"message": "เชื่อมต่อฐานข้อมูลไม่สำเร็จ",
		})
	}

	// ตรวจสอบว่าเอกสารมีอยู่จริง
	var existingDocNo string
	err = db.QueryRowContext(ctx, "SELECT docno FROM doc WHERE docno = $1 AND transflag = 6", req.DocNo).Scan(&existingDocNo)
	if err != nil {
		logger.Error("[ManualClosePO] ไม่พบเอกสาร %s: %v", req.DocNo, err)
		return c.JSON(http.StatusNotFound, map[string]any{
			"status":  "error",
			"code":    404,
			"message": fmt.Sprintf("ไม่พบเอกสาร %s", req.DocNo),
		})
	}

	now := time.Now().UTC()

	if req.Action == "close" {
		// ปิดเอกสาร
		_, err = db.ExecContext(ctx,
			`UPDATE doc SET
				isclosedmanual = true,
				closedmanual_by_code = $1,
				closedmanual_by_name = $2,
				closedmanual_at = $3,
				closedmanual_reason = $4
			WHERE docno = $5 AND transflag = 6`,
			req.ActionByCode, req.ActionByName, now, req.Reason, req.DocNo,
		)
	} else {
		// เปิดเอกสาร (reset)
		_, err = db.ExecContext(ctx,
			`UPDATE doc SET
				isclosedmanual = false,
				closedmanual_by_code = $1,
				closedmanual_by_name = $2,
				closedmanual_at = $3,
				closedmanual_reason = $4
			WHERE docno = $5 AND transflag = 6`,
			req.ActionByCode, req.ActionByName, now, req.Reason, req.DocNo,
		)
	}

	if err != nil {
		logger.Error("[ManualClosePO] อัปเดต PostgreSQL ไม่สำเร็จ: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"status":  "error",
			"code":    500,
			"message": "อัปเดตฐานข้อมูลไม่สำเร็จ",
		})
	}

	// อัปเดต ClickHouse (async)
	go func() {
		isClosedManual := req.Action == "close"
		updateClickHouseManualClose(ctx, req.HoldingCode, req.DocNo, isClosedManual, req.ActionByCode, req.ActionByName, now, req.Reason)
	}()

	actionText := "ปิดเอกสาร"
	if req.Action == "open" {
		actionText = "เปิดเอกสาร"
	}

	logger.Success("[ManualClosePO] %s %s สำเร็จ โดย %s (%s) เหตุผล: %s",
		actionText, req.DocNo, req.ActionByName, req.ActionByCode, req.Reason)

	return c.JSON(http.StatusOK, map[string]any{
		"status":  "success",
		"code":    200,
		"message": fmt.Sprintf("%sสำเร็จ", actionText),
		"data": map[string]any{
			"docno":                req.DocNo,
			"isclosedmanual":       req.Action == "close",
			"closedmanual_by_code": req.ActionByCode,
			"closedmanual_by_name": req.ActionByName,
			"closedmanual_at":      now.Format("2006-01-02T15:04:05.000Z"),
			"closedmanual_reason":  req.Reason,
		},
	})
}

// updateClickHouseManualClose — อัปเดตสถานะปิด/เปิดเอกสารใน ClickHouse
func updateClickHouseManualClose(ctx context.Context, holdingCode, docNo string, isClosedManual bool, byCode, byName string, at time.Time, reason string) {
	conn, err := myclickhouse.ClickHouseFastConnect()
	if err != nil {
		logger.Error("[ManualClosePO] เชื่อมต่อ ClickHouse ไม่สำเร็จ: %v", err)
		return
	}

	tableName := myclickhouse.TableName("doc")
	query := fmt.Sprintf(
		"ALTER TABLE %s UPDATE isclosedmanual = %t, closedmanual_by_code = '%s', closedmanual_by_name = '%s', closedmanual_at = '%s', closedmanual_reason = '%s' WHERE holding_code = '%s' AND docno = '%s'",
		tableName, isClosedManual, byCode, byName, at.Format("2006-01-02 15:04:05"), reason, holdingCode, docNo,
	)

	if err := conn.Exec(ctx, query); err != nil {
		logger.Error("[ManualClosePO] อัปเดต ClickHouse ไม่สำเร็จ: %v", err)
	} else {
		logger.Success("[ManualClosePO] อัปเดต ClickHouse สำเร็จ สำหรับ %s", docNo)
	}
}
