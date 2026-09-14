package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
)

func loadLineAccounts(ctx context.Context, tx *sql.Tx, company string, lines []Line) (map[string]Account, error) {
	codes := make([]string, 0, len(lines))
	for _, line := range lines {
		codes = append(codes, line.AccountCode)
	}
	rows, err := tx.QueryContext(ctx, `SELECT payload FROM gl_records WHERE company=$1 AND kind='accounts' AND code=ANY($2) AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, company, pq.Array(codes))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	accounts := map[string]Account{}
	for rows.Next() {
		var data []byte
		if err = rows.Scan(&data); err != nil {
			return nil, err
		}
		var account Account
		if err = json.Unmarshal(data, &account); err != nil {
			return nil, err
		}
		if _, exists := accounts[account.AccountCode]; exists {
			return nil, fmt.Errorf("รหัสบัญชีซ้ำในบริษัท")
		}
		accounts[account.AccountCode] = account
	}
	return accounts, rows.Err()
}

// projectionCurrency is the column value kept for gl_lines.currency (NOT NULL in schema.sql).
// GL no longer carries a currency concept; the column is satisfied with the base currency code.
const projectionCurrency = "THB"

type projectedLine struct {
	Company        string `json:"company"`
	JournalID      string `json:"journal_id"`
	LineNo         int    `json:"line_no"`
	DocNo          string `json:"doc_no"`
	Date           string `json:"entry_date"`
	FiscalYear     string `json:"fiscal_year"`
	BookCode       string `json:"book_code"`
	BranchCode     string `json:"branch_code"`
	DepartmentCode string `json:"department_code"`
	ProjectCode    string `json:"project_code"`
	Kind           string `json:"kind"`
	Currency       string `json:"currency"`
	Scale          int    `json:"scale"`
	AccountCode    string `json:"account_code"`
	AccountName    string `json:"account_name"`
	AccountType    string `json:"account_type"`
	NormalBalance  string `json:"normal_balance"`
	IsCash         bool   `json:"is_cash"`
	Description    string `json:"description"`
	CashFlow       string `json:"cash_flow"`
	Debit          string `json:"debit"`
	Credit         string `json:"credit"`
}

func insertProjectedLines(ctx context.Context, tx *sql.Tx, lines []projectedLine) error {
	payload, err := json.Marshal(lines)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO gl_lines(company,journal_id,line_no,doc_no,entry_date,fiscal_year,book_code,branch_code,department_code,project_code,kind,currency,scale,account_code,account_name,account_type,normal_balance,is_cash,description,cash_flow,debit,credit)
      SELECT * FROM jsonb_to_recordset($1::jsonb) AS line(company text,journal_id text,line_no integer,doc_no text,entry_date date,fiscal_year text,book_code text,branch_code text,department_code text,project_code text,kind text,currency text,scale integer,account_code text,account_name text,account_type text,normal_balance text,is_cash boolean,description text,cash_flow text,debit numeric(38,8),credit numeric(38,8))`, string(payload))
	return err
}

func verifyProjectedJournal(ctx context.Context, tx *sql.Tx, company string, j Journal, count int) error {
	if count != len(j.Lines) {
		return fmt.Errorf("ห้ามแก้บรรทัดบัญชีที่ผ่านรายการแล้ว")
	}
	payload, err := json.Marshal(j.Lines)
	if err != nil {
		return err
	}
	var matches bool
	err = tx.QueryRowContext(ctx, `WITH incoming AS (SELECT value,ordinality AS line_no FROM jsonb_array_elements($3::jsonb) WITH ORDINALITY) SELECT COALESCE(bool_and(
      p.doc_no=$4 AND p.entry_date=$5::date AND p.fiscal_year=$6 AND p.book_code=$7 AND p.branch_code=$8 AND p.kind=$9 AND p.currency=$10
      AND p.account_code=i.value->>'accountcode' AND p.department_code=i.value->>'departmentcode' AND p.project_code=i.value->>'projectcode'
      AND p.description=i.value->>'description' AND p.cash_flow=i.value->>'cashflow' AND p.debit=(i.value->>'debit')::numeric AND p.credit=(i.value->>'credit')::numeric),false)
      FROM gl_lines p JOIN incoming i USING(line_no) WHERE p.company=$1 AND p.journal_id=$2`, company, j.ID, string(payload), j.DocNo, j.Date, j.FiscalYear, j.BookCode, j.BranchCode, j.Kind, projectionCurrency).Scan(&matches)
	if err != nil {
		return err
	}
	if !matches {
		return fmt.Errorf("ห้ามแก้ยอดบัญชีที่ผ่านรายการแล้ว")
	}
	return nil
}
