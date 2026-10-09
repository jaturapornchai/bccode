package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"smlcloudplatform/internal/doctype"
	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/pkg/microservice"
)

type DeliveryDateInvoiceItem struct {
	DocNo          string `json:"docno"`
	DocDate        string `json:"docdate"`
	CustCode       string `json:"custcode"`
	CustName       string `json:"custname"`
	CreditOfBill   int    `json:"credit_of_bill"`
	OldDueDate     string `json:"old_duedate"`
	DeliveryDate   string `json:"delivery_date"`
	CreditDelivery int    `json:"credit_delivery"`
	NewDueDate     string `json:"new_duedate"`
}

type SaveDeliveryDateItem struct {
	DocNo          string `json:"docno"`
	DeliveryDate   string `json:"delivery_date"`
	CreditDelivery int    `json:"credit_delivery"`
	NewDueDate     string `json:"new_duedate"`
}

type SaveDeliveryDatesPayload struct {
	Items []SaveDeliveryDateItem `json:"items"`
}

// ListDeliveryDateInvoices ดึงรายการบิลที่ยังไม่ถูกยกเลิก/ลบ เพื่อนำมาปรับปรุงวันที่ส่งของ/วันที่รับสินค้า
func (s *Store) ListDeliveryDateInvoices(ctx context.Context, db *sql.DB, company string, transflag int, partner string) ([]DeliveryDateInvoiceItem, error) {
	partner = strings.TrimSpace(partner)
	query := `
		SELECT docno, 
		       COALESCE(docdatetime, now())::timestamptz,
		       COALESCE(perioddatetime, docdatetime, now())::timestamptz,
		       custcode, custname
		FROM doc
		WHERE businesscode = $1
		  AND transflag = $2
		  AND ($3 = '' OR custcode = $3)
		  AND NOT isdelete
		  AND NOT iscancel
		ORDER BY docdatetime DESC, docno DESC
		LIMIT 200
	`

	rows, err := db.QueryContext(ctx, query, company, transflag, partner)
	if err != nil {
		return nil, fmt.Errorf("list delivery date invoices: %w", err)
	}
	defer rows.Close()

	var items []DeliveryDateInvoiceItem
	for rows.Next() {
		var docno, custcode, custname string
		var docDate, dueDate time.Time
		if err := rows.Scan(&docno, &docDate, &dueDate, &custcode, &custname); err != nil {
			return nil, err
		}

		creditDays := int(dueDate.Sub(docDate).Hours() / 24)
		if creditDays < 0 {
			creditDays = 0
		}

		docDateStr := docDate.Format("2006-01-02")
		dueDateStr := dueDate.Format("2006-01-02")

		items = append(items, DeliveryDateInvoiceItem{
			DocNo:          docno,
			DocDate:        docDateStr,
			CustCode:       custcode,
			CustName:       custname,
			CreditOfBill:   creditDays,
			OldDueDate:     dueDateStr,
			DeliveryDate:   docDateStr,
			CreditDelivery: creditDays,
			NewDueDate:     dueDateStr,
		})
	}
	if items == nil {
		items = []DeliveryDateInvoiceItem{}
	}
	return items, nil
}

// SaveDeliveryDates บันทึกปรับปรุงวันที่ส่งของและวันครบกำหนดใหม่ทั้งในตาราง doc และ gl_subledger_documents
func (s *Store) SaveDeliveryDates(ctx context.Context, db *sql.DB, scope gl.Scope, dt doctype.DocType, items []SaveDeliveryDateItem) (int, error) {
	if len(items) == 0 {
		return 0, nil
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	gl.LockCompanyTx(ctx, tx, scope.Company)

	updatedCount := 0
	for _, item := range items {
		docno := strings.TrimSpace(item.DocNo)
		newDueDateStr := strings.TrimSpace(item.NewDueDate)
		if docno == "" || newDueDateStr == "" {
			continue
		}

		parsedDueDate, err := time.Parse("2006-01-02", newDueDateStr)
		if err != nil {
			continue
		}

		// 1. ปรับปรุงตาราง doc ฟิลด์ perioddatetime (วันครบกำหนด)
		res, err := tx.ExecContext(ctx, `
			UPDATE doc 
			SET perioddatetime = $1, modified_at = now(), modifier_code = $2
			WHERE businesscode = $3 AND transflag = $4 AND docno = $5 AND NOT isdelete AND NOT iscancel
		`, parsedDueDate, scope.Actor, scope.Company, dt.Flag, docno)
		if err != nil {
			return 0, fmt.Errorf("update doc date: %w", err)
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			continue
		}

		// 2. ปรับปรุงวันครบกำหนดใน gl_subledger_documents เพื่อให้รายงานอายุหนี้ (Aging) แสดงผลถูกต้อง
		var hasSubledger bool
		_ = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'gl_subledger_documents')`).Scan(&hasSubledger)
		if hasSubledger {
			_, _ = tx.ExecContext(ctx, `
				UPDATE gl_subledger_documents
				SET payload = jsonb_set(COALESCE(payload::jsonb, '{}'::jsonb), '{due_date}', to_jsonb($1::text))
				WHERE company = $2 AND (id = $3 OR id = 'AR-' || $3 OR id = 'AP-' || $3)
			`, newDueDateStr, scope.Company, docno)
		}

		// 3. บันทึก audit trail
		_, _ = tx.ExecContext(ctx, `
			INSERT INTO erp_doc_events (businesscode, transflag, docno, action, reason, actor, created_at)
			VALUES ($1, $2, $3, 'update', $4, $5, now())
		`, scope.Company, dt.Flag, docno, fmt.Sprintf("ปรับปรุงวันครบกำหนดเป็น %s", newDueDateStr), scope.Actor)

		updatedCount++
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return updatedCount, nil
}

func (h *Http) listDeliveryDates(request microservice.IContext) error {
	access := h.authorize(request, actionRead)
	if access == nil {
		return nil
	}
	partner := request.QueryParam("partner")
	if partner == "" {
		partner = request.QueryParam("custcode")
	}

	items, err := h.store.ListDeliveryDateInvoices(access.ctx, access.db, access.scope.Company, access.dt.Flag, partner)
	if err != nil {
		return docFailure(request, err, "", "โหลดรายการเอกสารไม่สำเร็จ")
	}

	return success(request, map[string]any{
		"success": true,
		"data":    items,
	})
}

func (h *Http) saveDeliveryDates(request microservice.IContext) error {
	access := h.authorize(request, actionUpdate)
	if access == nil {
		return nil
	}

	var payload SaveDeliveryDatesPayload
	if err := json.NewDecoder(request.Request().Body).Decode(&payload); err != nil {
		return fail(request, http.StatusBadRequest, "รูปแบบข้อมูลไม่ถูกต้อง")
	}

	count, err := h.store.SaveDeliveryDates(access.ctx, access.db, access.scope, access.dt, payload.Items)
	if err != nil {
		return docFailure(request, err, "", "บันทึกการปรับปรุงวันที่ไม่สำเร็จ")
	}

	return success(request, map[string]any{
		"success": true,
		"count":   count,
		"message": "ops_save_delivery_dates_success",
	})
}
