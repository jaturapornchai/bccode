package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"smlcloudplatform/internal/doctype"
	gl "smlcloudplatform/internal/generalledger"
)

// dueDateDocs คือเอกสาร ERP ฝั่งหนี้เจ้าหนี้ที่มีวันเครดิต/วันครบกำหนดบนหัวเอกสาร (Champ CreditDay/DueDate ของ
// BCApInvoice, BCApInvBalance, BCApOtherDebt, BCApDebitNote — design ap-payment-prepare §4.6)
var dueDateDocs = map[string]bool{"purchase": true, "creditorbeginningbalance": true, "apotherdebt": true, "apdebitnote": true}

// creditor คือคู่ค้าในทะเบียนบัญชีย่อยของบริษัท (gl_subledger_partners) เท่าที่เอกสารเจ้าหนี้ใช้
type creditor struct {
	Code       string
	Name       string
	IsSupplier bool
	IsActive   bool
	CreditDays int
}

// lookupPartner อ่านคู่ค้าหนึ่งรายจากทะเบียน; ไม่พบ = found false
func lookupPartner(ctx context.Context, q rowQuerier, company, code string) (creditor, bool, error) {
	var c creditor
	var title, name string
	err := q.QueryRowContext(ctx, `SELECT code, COALESCE(payload->>'title_name', ''), COALESCE(payload->>'name_th', ''),
		COALESCE((payload->>'is_supplier')::boolean, false), COALESCE((payload->>'is_active')::boolean, false), COALESCE((payload->>'credit_days')::int, 0)
		FROM gl_subledger_partners WHERE company = $1 AND code = $2`, company, code).Scan(&c.Code, &title, &name, &c.IsSupplier, &c.IsActive, &c.CreditDays)
	if errors.Is(err, sql.ErrNoRows) {
		return creditor{}, false, nil
	}
	if err != nil {
		return creditor{}, false, fmt.Errorf("lookup partner %s: %w", code, err)
	}
	c.Name = gl.PartnerFullName(title, name)
	return c, true, nil
}

// resolveDueDate กำหนดวันเครดิตและวันครบกำหนดของเอกสารหนี้ตามลำดับ Champ (ChmFrmGoodsPurchasePage1.cpp:813-825):
// วันเครดิตไม่ส่งมา = ของเจ้าหนี้ในทะเบียน; วันครบกำหนดไม่ส่งมา = วันที่เอกสาร + วันเครดิต; ก่อนวันที่เอกสารไม่ได้
// เอกสารประเภทอื่นล้างทั้งสองช่อง (คืน NULL = ไม่แตะ perioddatetime เดิม เช่นวันครบกำหนดของใบขายจากจอปรับวันส่งของ)
func resolveDueDate(ctx context.Context, q rowQuerier, company string, dt doctype.DocType, doc *ErpTransactionDoc, docDate time.Time) (sql.NullTime, error) {
	if !dueDateDocs[dt.Key] {
		doc.CreditDay, doc.DueDate = nil, ""
		return sql.NullTime{}, nil
	}
	creditDay := 0
	if doc.CreditDay != nil {
		creditDay = *doc.CreditDay
		if creditDay < 0 || creditDay > gl.CreditDaysMax {
			return sql.NullTime{}, gl.CreditDaysInvalidError("creditday")
		}
	} else if code := strings.TrimSpace(doc.CustCode); code != "" {
		partner, found, err := lookupPartner(ctx, q, company, code)
		if err != nil {
			return sql.NullTime{}, err
		}
		if found {
			creditDay = partner.CreditDays
		}
	}
	docDay, err := dayStart(businessDay(docDate))
	if err != nil {
		return sql.NullTime{}, fmt.Errorf("document day: %w", err)
	}
	due := docDay.AddDate(0, 0, creditDay)
	if raw := strings.TrimSpace(doc.DueDate); raw != "" {
		given, err := dayStart(raw)
		if err != nil {
			return sql.NullTime{}, dueDateError("erp_due_date_invalid", raw)
		}
		if given.Before(docDay) {
			return sql.NullTime{}, dueDateError("erp_due_before_doc")
		}
		due = given
	}
	doc.CreditDay, doc.DueDate = &creditDay, businessDay(due)
	return sql.NullTime{Time: due, Valid: true}, nil
}

func dueDateError(code string, args ...string) *gl.UserError {
	return erpBadRequest(code, "duedate", args...)
}
