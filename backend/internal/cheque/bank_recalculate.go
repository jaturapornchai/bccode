package cheque

import (
	"context"
	"database/sql"
	"encoding/json"
)

// BankBalanceResult คือผลการคำนวณยอดคงเหลือของสมุดบัญชีธนาคารหนึ่งเล่ม
type BankBalanceResult struct {
	BankAccountCode string `json:"bank_account_code"`
	AccountNumber   string `json:"account_number"`
	AccountName     string `json:"account_name"`
	GLAccountCode   string `json:"gl_account_code"`
	GLBalance       string `json:"gl_balance"`
	ChequeNet       string `json:"cheque_net"`
	TotalBalance    string `json:"total_balance"`
}

// BankRecalcSummary คือสรุปผลการคำนวณยอดคงเหลือสมุดบัญชีใหม่ (Champ CASHReprocess)
type BankRecalcSummary struct {
	TotalBooks int                 `json:"total_books"`
	Books      []BankBalanceResult `json:"books"`
}

// RecalculateBankBalances คำนวณยอดคงเหลือสะสมของสมุดบัญชีเงินฝากธนาคารทุกเล่มในบริษัทตาม Champ CashDlgBankReprocess.cpp
func RecalculateBankBalances(ctx context.Context, db *sql.DB, company string) (BankRecalcSummary, error) {
	summary := BankRecalcSummary{Books: []BankBalanceResult{}}

	// ดึงรายการสมุดเงินฝากธนาคารทั้งหมด
	rows, err := db.QueryContext(ctx,
		`SELECT code, payload FROM gl_subledger_bank_accounts WHERE company = $1 ORDER BY code`, company)
	if err != nil {
		return summary, err
	}
	defer rows.Close()

	type rawBook struct {
		code    string
		payload []byte
	}
	var rawBooks []rawBook
	for rows.Next() {
		var rb rawBook
		if err = rows.Scan(&rb.code, &rb.payload); err != nil {
			return summary, err
		}
		rawBooks = append(rawBooks, rb)
	}
	if err = rows.Err(); err != nil {
		return summary, err
	}

	for _, rb := range rawBooks {
		var parsed struct {
			AccountNumber string `json:"account_number"`
			AccountName   string `json:"account_name"`
			GLAccountCode string `json:"gl_account_code"`
		}
		_ = json.Unmarshal(rb.payload, &parsed)

		// 1. คำนวณยอดเคลื่อนไหวเช็คสุทธิ (Cheque Net Movement)
		movements, err := BookMovements(ctx, db, company, rb.code, "", "")
		chequeNet := "0.00"
		if err == nil && len(movements) > 0 {
			chequeNet = movements[0].Net
		}

		// 2. คำนวณยอดคงเหลือในบัญชีแยกประเภท (GL Balance: Debit - Credit)
		glBalance := "0.00"
		if parsed.GLAccountCode != "" {
			var bal sql.NullString
			err = db.QueryRowContext(ctx,
				`SELECT COALESCE(SUM(debit - credit), 0)::numeric(18,2)::text
				 FROM gl_lines
				 WHERE company = $1 AND account_code = $2`, company, parsed.GLAccountCode).Scan(&bal)
			if err == nil && bal.Valid {
				glBalance = bal.String
			}
		}

		// 3. ยอดคงเหลือสุทธิ = GL Balance (ซึ่งสะท้อนยอดเงินจริงตามหลักบัญชีคู่)
		totalBalance := glBalance

		// อัปเดต payload ใน gl_subledger_bank_accounts ให้มี gl_balance ล่าสุด
		var currentPayload map[string]any
		if json.Unmarshal(rb.payload, &currentPayload) == nil {
			currentPayload["gl_balance"] = totalBalance
			if updatedJSON, err := json.Marshal(currentPayload); err == nil {
				_, _ = db.ExecContext(ctx,
					`UPDATE gl_subledger_bank_accounts
					 SET payload = $1
					 WHERE company = $2 AND code = $3`, string(updatedJSON), company, rb.code)
			}
		}

		summary.Books = append(summary.Books, BankBalanceResult{
			BankAccountCode: rb.code,
			AccountNumber:   parsed.AccountNumber,
			AccountName:     parsed.AccountName,
			GLAccountCode:   parsed.GLAccountCode,
			GLBalance:       glBalance,
			ChequeNet:       chequeNet,
			TotalBalance:    totalBalance,
		})
	}

	summary.TotalBooks = len(summary.Books)
	return summary, nil
}
