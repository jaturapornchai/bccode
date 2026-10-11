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

// ListPettyCash คืนรายการวงเงินสดย่อยทั้งหมดของบริษัท เรียงตามรหัส
func ListPettyCash(ctx context.Context, q querier, company string) ([]PettyCash, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT code, name, amount::numeric(18,2)::text, begin_amount::numeric(18,2)::text,
		        balance::numeric(18,2)::text, version
		 FROM bc_petty_cash
		 WHERE company = $1
		 ORDER BY code ASC LIMIT $2`,
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
