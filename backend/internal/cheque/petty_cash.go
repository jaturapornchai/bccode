package cheque

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// ความยาวสูงสุดตาม Champ BCPettyCash (Code VARCHAR(20), Name VARCHAR(100))
const (
	maxPettyCashCodeLen    = 20
	maxPettyCashNameLen    = 100
	MaxPettyCashPerCompany = 500
)

// PettyCash คือแถวกำหนดวงเงินสดย่อยหนึ่งแถว
type PettyCash struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Amount      string `json:"amount"`
	BeginAmount string `json:"begin_amount"`
	Balance     string `json:"balance"`
	Version     int64  `json:"version"`
}

// PettyCashInput คือแถวที่ส่งมาจากหน้าจอ
type PettyCashInput struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Amount      string `json:"amount"`
	BeginAmount string `json:"begin_amount"`
}

// ListPettyCash คืนรายการวงเงินสดย่อยทั้งหมดของบริษัท เรียงตามรหัส พร้อมยอดคงเหลือคำนวณจริงตาม Champ CASHFrmPettyCash.cpp
func ListPettyCash(ctx context.Context, q querier, company string) ([]PettyCash, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT c.code, c.name, c.amount::numeric(18,2)::text, c.begin_amount::numeric(18,2)::text,
		        (c.begin_amount + COALESCE(SUM(CASE WHEN t.trans_type = 1 THEN t.amount ELSE -t.amount END), 0))::numeric(18,2)::text AS balance,
		        c.version
		 FROM bc_petty_cash c
		 LEFT JOIN bc_petty_cash_trans t ON t.company = c.company AND t.petty_cash_code = c.code
		 WHERE c.company = $1
		 GROUP BY c.company, c.code, c.name, c.amount, c.begin_amount, c.version
		 ORDER BY c.code ASC LIMIT $2`,
		company, MaxPettyCashPerCompany)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []PettyCash{}
	for rows.Next() {
		var item PettyCash
		if err = rows.Scan(&item.Code, &item.Name, &item.Amount, &item.BeginAmount, &item.Balance, &item.Version); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// SavePettyCash บันทึก/อัปเดต/ลบรายการวงเงินสดย่อยทั้งชุดใน ACID transaction เดียวตาม Champ CASHFrmPettyCash.cpp
func SavePettyCash(ctx context.Context, db *sql.DB, scope Scope, inputs []PettyCashInput) ([]PettyCash, error) {
	if len(inputs) > MaxPettyCashPerCompany {
		return nil, badRequest("petty_cash_err_too_many", "items", fmt.Sprint(MaxPettyCashPerCompany))
	}

	type normalizedRow struct {
		code        string
		name        string
		amount      decimal.Decimal
		beginAmount decimal.Decimal
		balance     decimal.Decimal
	}

	normRows := make([]normalizedRow, len(inputs))
	seen := make(map[string]bool, len(inputs))

	for i, in := range inputs {
		line := fmt.Sprint(i + 1)
		code, err := cleanText(in.Code, fmt.Sprintf("items[%d].code", i), maxPettyCashCodeLen, true)
		if err != nil {
			return nil, err
		}
		key := strings.ToUpper(code)
		if seen[key] {
			return nil, badRequest("petty_cash_err_duplicate", fmt.Sprintf("items[%d].code", i), code, line)
		}
		seen[key] = true

		name, err := cleanText(in.Name, fmt.Sprintf("items[%d].name", i), maxPettyCashNameLen, true)
		if err != nil {
			return nil, err
		}

		amountStr := strings.TrimSpace(in.Amount)
		if amountStr == "" {
			amountStr = "0.00"
		}
		amt, ok := parseMoney(amountStr)
		if !ok || amt.IsNegative() {
			return nil, badRequest("petty_cash_err_amount_invalid", fmt.Sprintf("items[%d].amount", i), line)
		}

		beginStr := strings.TrimSpace(in.BeginAmount)
		if beginStr == "" {
			beginStr = "0.00"
		}
		beginAmt, ok := parseMoney(beginStr)
		if !ok {
			return nil, badRequest("petty_cash_err_begin_amount_invalid", fmt.Sprintf("items[%d].begin_amount", i), line)
		}

		// ยอดคงเหลือเริ่มต้น = ยอดยกมา (ตาม Champ CASHFrmPettyCash.cpp)
		balance := beginAmt

		normRows[i] = normalizedRow{
			code:        code,
			name:        name,
			amount:      amt,
			beginAmount: beginAmt,
			balance:     balance,
		}
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if len(normRows) == 0 {
		if _, err = tx.ExecContext(ctx, `DELETE FROM bc_petty_cash WHERE company = $1`, scope.Company); err != nil {
			return nil, err
		}
	} else {
		codes := make([]string, len(normRows))
		for i, r := range normRows {
			codes[i] = r.code
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM bc_petty_cash WHERE company = $1 AND code NOT = ANY($2)`, scope.Company, pq.Array(codes)); err != nil {
			return nil, err
		}

		for _, r := range normRows {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO bc_petty_cash (company, code, name, amount, begin_amount, balance, version, created_by, updated_by, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6, 1, $7, $7, now())
				 ON CONFLICT (company, code) DO UPDATE SET
				   name = EXCLUDED.name,
				   amount = EXCLUDED.amount,
				   begin_amount = EXCLUDED.begin_amount,
				   balance = EXCLUDED.balance,
				   version = bc_petty_cash.version + 1,
				   updated_by = EXCLUDED.updated_by,
				   updated_at = now()`,
				scope.Company, r.code, r.name, r.amount.StringFixed(2), r.beginAmount.StringFixed(2), r.balance.StringFixed(2), scope.Actor)
			if err != nil {
				return nil, err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return ListPettyCash(ctx, db, scope.Company)
}

// PettyCashTrans คือแถวประวัติความเคลื่อนไหวเงินสดย่อยหนึ่งแถว
type PettyCashTrans struct {
	ID            string `json:"id"`
	DocNo         string `json:"doc_no"`
	DocDate       string `json:"doc_date"`
	PettyCashCode string `json:"petty_cash_code"`
	PettyCashName string `json:"petty_cash_name,omitempty"`
	TransType     int    `json:"trans_type"` // 1=รับ/เบิกชดเชย, 2=จ่าย
	Amount        string `json:"amount"`
	Description   string `json:"description"`
	PayeePayer    string `json:"payee_payer"`
	CreatedAt     string `json:"created_at"`
	CreatedBy     string `json:"created_by"`
}

// PettyCashTransInput คือข้อมูลสำหรับบันทึกรายการเคลื่อนไหวเงินสดย่อย
type PettyCashTransInput struct {
	DocNo         string `json:"doc_no"`
	DocDate       string `json:"doc_date"`
	PettyCashCode string `json:"petty_cash_code"`
	TransType     int    `json:"trans_type"` // 1=รับ/เบิกชดเชย, 2=จ่าย
	Amount        string `json:"amount"`
	Description   string `json:"description"`
	PayeePayer    string `json:"payee_payer"`
}

// CreatePettyCashTrans บันทึกรายการเคลื่อนไหวเงินสดย่อย (รับชดเชย/จ่าย) ใน ACID transaction เดียว
func CreatePettyCashTrans(ctx context.Context, db *sql.DB, scope Scope, in PettyCashTransInput) (*PettyCashTrans, error) {
	docNo, err := cleanText(in.DocNo, "doc_no", 30, true)
	if err != nil {
		return nil, err
	}

	docDate := strings.TrimSpace(in.DocDate)
	if !validISODate(docDate) {
		return nil, badRequest("cheque_err_date_invalid", "doc_date")
	}

	code, err := cleanText(in.PettyCashCode, "petty_cash_code", maxPettyCashCodeLen, true)
	if err != nil {
		return nil, err
	}

	if in.TransType != 1 && in.TransType != 2 {
		return nil, badRequest("petty_cash_err_trans_type_invalid", "trans_type")
	}

	amtStr := strings.TrimSpace(in.Amount)
	amt, ok := parseMoney(amtStr)
	if !ok || amt.LessThanOrEqual(decimal.Zero) {
		return nil, badRequest("petty_cash_err_amount_positive", "amount")
	}

	desc, err := cleanText(in.Description, "description", 255, false)
	if err != nil {
		return nil, err
	}

	payee, err := cleanText(in.PayeePayer, "payee_payer", 100, false)
	if err != nil {
		return nil, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// ตรวจสอบว่ามีรหัสวงเงินสดย่อยนี้จริง และล็อกแถว
	var pcName string
	var beginAmt, curBalance decimal.Decimal
	err = tx.QueryRowContext(ctx,
		`SELECT name, begin_amount,
		        (begin_amount + COALESCE((SELECT SUM(CASE WHEN trans_type = 1 THEN amount ELSE -amount END)
		                                  FROM bc_petty_cash_trans
		                                  WHERE company = $1 AND petty_cash_code = $2), 0))
		 FROM bc_petty_cash
		 WHERE company = $1 AND code = $2
		 FOR UPDATE`,
		scope.Company, code).Scan(&pcName, &beginAmt, &curBalance)
	if err == sql.ErrNoRows {
		return nil, badRequest("petty_cash_err_code_not_found", "petty_cash_code", code)
	}
	if err != nil {
		return nil, err
	}

	// ถ้าเป็นการจ่าย (trans_type = 2) ตรวจสอบว่ายอดเงินในวงเงินสดย่อยมีเพียงพอ
	newBalance := curBalance
	if in.TransType == 1 {
		newBalance = curBalance.Add(amt)
	} else {
		newBalance = curBalance.Sub(amt)
		if newBalance.IsNegative() {
			return nil, badRequest("petty_cash_err_insufficient_balance", "amount", curBalance.StringFixed(2))
		}
	}

	var row PettyCashTrans
	row.DocNo = docNo
	row.DocDate = docDate
	row.PettyCashCode = code
	row.PettyCashName = pcName
	row.TransType = in.TransType
	row.Amount = amt.StringFixed(2)
	row.Description = desc
	row.PayeePayer = payee
	row.CreatedBy = scope.Actor

	err = tx.QueryRowContext(ctx,
		`INSERT INTO bc_petty_cash_trans
		   (company, doc_no, doc_date, petty_cash_code, trans_type, amount, description, payee_payer, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		 RETURNING id::text, created_at::timestamptz::text`,
		scope.Company, docNo, docDate, code, in.TransType, amt.StringFixed(2), desc, payee, scope.Actor).Scan(&row.ID, &row.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, badRequest("petty_cash_err_doc_no_duplicate", "doc_no", docNo)
		}
		return nil, err
	}

	// ปรับปรุงยอด balance cache ใน bc_petty_cash
	_, err = tx.ExecContext(ctx,
		`UPDATE bc_petty_cash
		 SET balance = $1, version = version + 1, updated_by = $2, updated_at = now()
		 WHERE company = $3 AND code = $4`,
		newBalance.StringFixed(2), scope.Actor, scope.Company, code)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &row, nil
}

// ListPettyCashTrans คืนรายการเคลื่อนไหวเงินสดย่อยตามช่วงวันที่และรหัสเงินสดย่อย
func ListPettyCashTrans(ctx context.Context, q querier, company, pettyCashCode, fromDate, toDate string, limit int) ([]PettyCashTrans, error) {
	if limit <= 0 || limit > 1000 {
		limit = 500
	}

	query := `SELECT t.id::text, t.doc_no, t.doc_date::text, t.petty_cash_code, c.name,
	                 t.trans_type, t.amount::numeric(18,2)::text, t.description, t.payee_payer,
	                 t.created_at::timestamptz::text, t.created_by
	          FROM bc_petty_cash_trans t
	          JOIN bc_petty_cash c ON c.company = t.company AND c.code = t.petty_cash_code
	          WHERE t.company = $1`
	args := []any{company}

	if strings.TrimSpace(pettyCashCode) != "" {
		args = append(args, strings.TrimSpace(pettyCashCode))
		query += fmt.Sprintf(" AND t.petty_cash_code = $%d", len(args))
	}
	if strings.TrimSpace(fromDate) != "" {
		args = append(args, strings.TrimSpace(fromDate))
		query += fmt.Sprintf(" AND t.doc_date >= $%d", len(args))
	}
	if strings.TrimSpace(toDate) != "" {
		args = append(args, strings.TrimSpace(toDate))
		query += fmt.Sprintf(" AND t.doc_date <= $%d", len(args))
	}

	query += fmt.Sprintf(" ORDER BY t.doc_date DESC, t.created_at DESC LIMIT $%d", len(args)+1)
	args = append(args, limit)

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []PettyCashTrans{}
	for rows.Next() {
		var item PettyCashTrans
		if err = rows.Scan(&item.ID, &item.DocNo, &item.DocDate, &item.PettyCashCode, &item.PettyCashName,
			&item.TransType, &item.Amount, &item.Description, &item.PayeePayer, &item.CreatedAt, &item.CreatedBy); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// DeletePettyCashTrans ลบรายการเคลื่อนไหวเงินสดย่อย
func DeletePettyCashTrans(ctx context.Context, db *sql.DB, scope Scope, id string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var code string
	var transType int
	var amt decimal.Decimal
	err = tx.QueryRowContext(ctx,
		`SELECT petty_cash_code, trans_type, amount
		 FROM bc_petty_cash_trans
		 WHERE company = $1 AND id = $2
		 FOR UPDATE`,
		scope.Company, id).Scan(&code, &transType, &amt)
	if err == sql.ErrNoRows {
		return badRequest("petty_cash_err_trans_not_found", "id", id)
	}
	if err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM bc_petty_cash_trans WHERE company = $1 AND id = $2`, scope.Company, id); err != nil {
		return err
	}

	// อัปเดต balance ของ bc_petty_cash
	var newBal decimal.Decimal
	err = tx.QueryRowContext(ctx,
		`SELECT (begin_amount + COALESCE((SELECT SUM(CASE WHEN trans_type = 1 THEN amount ELSE -amount END)
		                                  FROM bc_petty_cash_trans
		                                  WHERE company = $1 AND petty_cash_code = $2), 0))
		 FROM bc_petty_cash
		 WHERE company = $1 AND code = $2`,
		scope.Company, code).Scan(&newBal)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE bc_petty_cash SET balance = $1, version = version + 1, updated_by = $2, updated_at = now()
		 WHERE company = $3 AND code = $4`,
		newBal.StringFixed(2), scope.Actor, scope.Company, code)
	if err != nil {
		return err
	}

	return tx.Commit()
}
