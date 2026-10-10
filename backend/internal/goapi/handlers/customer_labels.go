package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"smlcloudplatform/internal/doctype"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
)

type CustomerLabelsRequest struct {
	HoldingCode string `json:"holdingcode"`
	FromCode    string `json:"from_code,omitempty"`
	ToCode      string `json:"to_code,omitempty"`
	FromDate    string `json:"from_date,omitempty"`
	ToDate      string `json:"to_date,omitempty"`
	FromAmount  string `json:"from_amount,omitempty"`
	ToAmount    string `json:"to_amount,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
}

type CustomerLabelItem struct {
	CustomerCode    string `json:"customer_code"`
	NameTh          string `json:"name_th"`
	NameEn          string `json:"name_en"`
	TaxID           string `json:"tax_id"`
	BranchNo        string `json:"branch_no"`
	BillingAddress  string `json:"billing_address"`
	ShippingAddress string `json:"shipping_address"`
	District        string `json:"district"`
	Province        string `json:"province"`
	Postcode        string `json:"postcode"`
	Phone           string `json:"phone"`
	TotalSales      string `json:"total_sales"`
	DocCount        int    `json:"doc_count"`
}

type sqlQueryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// CustomerLabelsHandler ค้นหารายการลูกค้าและที่อยู่สำหรับพิมพ์ป้ายผนึก (Customer Address Labels)
// ตามสเปก Champ พิมพ์ป้ายผนึกชื่อลูกค้า สมาชิก
func CustomerLabelsHandler(c echo.Context) error {
	var req CustomerLabelsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	req.HoldingCode = strings.TrimSpace(req.HoldingCode)
	if req.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode is required",
			"code":  "MISSING_HOLDING_CODE",
		})
	}

	businessCode, scopeErr := reportCompanyScope(c, req.HoldingCode)
	if scopeErr != nil {
		return scopeErr.respond(c)
	}

	if req.Limit <= 0 || req.Limit > 2000 {
		req.Limit = 500
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	db, err := mypg.PgSqlFastConnect(req.HoldingCode)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get DB pool for %s: %v", req.HoldingCode, err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to connect database",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 30*time.Second)
	defer cancel()

	items, err := queryCustomerLabels(ctx, db, businessCode, req)
	if err != nil {
		logger.Error(fmt.Sprintf("Customer labels query failed: %v", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to query customer labels: " + err.Error(),
			"code":  "QUERY_ERROR",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"data":    items,
		"count":   len(items),
	})
}

func queryCustomerLabels(ctx context.Context, db sqlQueryer, businessCode string, req CustomerLabelsRequest) ([]CustomerLabelItem, error) {
	var whereClauses []string
	var args []any
	argIdx := 1

	whereClauses = append(whereClauses, fmt.Sprintf("(p.company = $%d OR p.company = '01' OR $%d = '')", argIdx, argIdx))
	args = append(args, businessCode)
	argIdx++

	// เฉพาะบทบาทลูกค้า
	whereClauses = append(whereClauses, "COALESCE((p.payload->>'is_customer')::boolean, true) = true")
	whereClauses = append(whereClauses, "COALESCE((p.payload->>'is_active')::boolean, true) = true")

	fromCode := strings.TrimSpace(req.FromCode)
	toCode := strings.TrimSpace(req.ToCode)
	if fromCode != "" {
		if toCode != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("p.code BETWEEN $%d AND $%d", argIdx, argIdx+1))
			args = append(args, fromCode, toCode)
			argIdx += 2
		} else {
			whereClauses = append(whereClauses, fmt.Sprintf("p.code >= $%d", argIdx))
			args = append(args, fromCode)
			argIdx++
		}
	} else if toCode != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("p.code <= $%d", argIdx))
		args = append(args, toCode)
		argIdx++
	}

	// เงื่อนไขยอดขาย/วันที่ขาย
	saleWhere := fmt.Sprintf("s.businesscode = p.company AND s.custcode = p.code AND s.transflag = %d AND NOT s.isdelete AND NOT s.iscancel", doctype.SaleInvoice)
	fromDate := strings.TrimSpace(req.FromDate)
	toDate := strings.TrimSpace(req.ToDate)
	if fromDate != "" {
		if toDate != "" {
			saleWhere += fmt.Sprintf(" AND s.docdatetime::date BETWEEN $%d AND $%d", argIdx, argIdx+1)
			args = append(args, fromDate, toDate)
			argIdx += 2
		} else {
			saleWhere += fmt.Sprintf(" AND s.docdatetime::date >= $%d", argIdx)
			args = append(args, fromDate)
			argIdx++
		}
	} else if toDate != "" {
		saleWhere += fmt.Sprintf(" AND s.docdatetime::date <= $%d", argIdx)
		args = append(args, toDate)
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT p.code,
		       COALESCE(p.payload->>'name_th', p.payload->>'name', p.code),
		       COALESCE(p.payload->>'name_en', ''),
		       COALESCE(p.payload->>'tax_id', ''),
		       COALESCE(p.payload->>'tax_branch_no', '00000'),
		       COALESCE(p.payload->>'address', ''),
		       COALESCE(p.payload->>'shipping_address', p.payload->>'address', ''),
		       COALESCE(p.payload->>'addr_district', ''),
		       COALESCE(p.payload->>'addr_province', ''),
		       COALESCE(p.payload->>'addr_postcode', ''),
		       COALESCE(p.payload->>'phone', p.payload->>'telephone', ''),
		       COALESCE((
		           SELECT TRIM(TRAILING '.' FROM TRIM(TRAILING '0' FROM ROUND(SUM(s.totalamount), 2)::text))
		           FROM doc s
		           WHERE %s
		       ), '0'),
		       COALESCE((
		           SELECT COUNT(*)
		           FROM doc s
		           WHERE %s
		       ), 0)
		FROM gl_subledger_partners p
		WHERE %s
		ORDER BY p.code ASC
		LIMIT $%d OFFSET $%d
	`, saleWhere, saleWhere, strings.Join(whereClauses, " AND "), argIdx, argIdx+1)

	args = append(args, req.Limit, req.Offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query customer labels: %w", err)
	}
	defer rows.Close()

	fromAmountStr := strings.TrimSpace(req.FromAmount)
	toAmountStr := strings.TrimSpace(req.ToAmount)
	var minAmt, maxAmt float64
	hasMinAmt := false
	hasMaxAmt := false
	if fromAmountStr != "" {
		if v, err := strconv.ParseFloat(fromAmountStr, 64); err == nil {
			minAmt = v
			hasMinAmt = true
		}
	}
	if toAmountStr != "" {
		if v, err := strconv.ParseFloat(toAmountStr, 64); err == nil {
			maxAmt = v
			hasMaxAmt = true
		}
	}

	var items []CustomerLabelItem
	for rows.Next() {
		var item CustomerLabelItem
		if err := rows.Scan(
			&item.CustomerCode,
			&item.NameTh,
			&item.NameEn,
			&item.TaxID,
			&item.BranchNo,
			&item.BillingAddress,
			&item.ShippingAddress,
			&item.District,
			&item.Province,
			&item.Postcode,
			&item.Phone,
			&item.TotalSales,
			&item.DocCount,
		); err != nil {
			return nil, err
		}

		if item.TotalSales == "" {
			item.TotalSales = "0"
		}

		if hasMinAmt || hasMaxAmt {
			amt, _ := strconv.ParseFloat(item.TotalSales, 64)
			if hasMinAmt && amt < minAmt {
				continue
			}
			if hasMaxAmt && amt > maxAmt {
				continue
			}
		}

		items = append(items, item)
	}

	if items == nil {
		items = []CustomerLabelItem{}
	}
	return items, nil
}
