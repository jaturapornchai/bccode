package generalledger

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// AccountReference: how many records of one kind still use the account. Kind is a gl_records kind
// ("journals", "accounts" = child accounts, "fiscal-years", "mappings", …) or "budgets" (gl_budget_lines).
type AccountReference struct {
	Kind  string `json:"kind"`
	Count int    `json:"count"`
}

// AccountReferences answers accounts action "references" (nothing is saved): Champ's "รายการที่ใช้บัญชีนี้"
// panel (GLFrmChartofAccountPage1.cpp:269-483). Champ lists 6 sources of its own; these are the same
// questions for the records the GL module owns — the exact set that blocks deleting the account.
type AccountReferences struct {
	Items []AccountReference `json:"items"`
}

// accountReferencedByRecord: SQL predicate (columns of one gl_records row, account code = $2) — the row
// uses the account. Shared by deletePGAccountGuard and accountReferenceCounts so the delete block and the
// panel always agree.
const accountReferencedByRecord = `((kind='journals' AND EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(payload->'lines','[]'::jsonb)) l WHERE l->>'accountcode'=$2)) OR payload->>'parentaccountcode'=$2 OR payload->>'accountcode'=$2 AND kind<>'accounts' OR payload->>'profitlossaccount'=$2 OR payload->>'retainedearningsaccount'=$2 OR payload->>'itemaccount'=$2 OR payload->>'costaccount'=$2 OR payload->>'revenueaccount'=$2 OR EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(payload->'rules','[]'::jsonb)) r WHERE r->>'accountcode'=$2) OR (kind='allocations' AND EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(payload->'allocaterules','[]'::jsonb)) ar WHERE ar->>'accountcode'=$2)) OR (kind='journal-templates' AND EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(payload->'template'->'lines','[]'::jsonb)) tl WHERE tl->>'accountcode'=$2)))`

// accountGroupListsAccount: the account-groups row t lists the account on its own — a single code in the
// group's entries (the end of a range is not a member: ranges compare as text, so they do not tie to the account).
const accountGroupListsAccount = `EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(t.payload->'accountentries','[]'::jsonb)) e WHERE e->>'from'=$2 AND COALESCE(e->>'to','')='')`

// accountReferenceCounts counts, per kind, the rows that use the account — one query for gl_records
// (GROUP BY kind) plus one for budget lines. statementTemplateReadsAccount counts only the positions the
// statement engine really reads, so a code left in a row of another type does not show up.
func accountReferenceCounts(ctx context.Context, tx *sql.Tx, company, code string) ([]AccountReference, error) {
	items := []AccountReference{}
	var budgets int
	if err := tx.QueryRowContext(ctx, `SELECT count(DISTINCT budget_code) FROM gl_budget_lines WHERE company=$1 AND account_code=$2`, company, code).Scan(&budgets); err != nil {
		return nil, err
	}
	if budgets > 0 {
		items = append(items, AccountReference{Kind: "budgets", Count: budgets})
	}
	rows, err := tx.QueryContext(ctx, `SELECT t.kind, count(*) FROM gl_records t WHERE t.company=$1 AND NOT COALESCE((t.payload->>'isdeleted')::boolean,false) AND (`+
		accountReferencedByRecord+
		` OR (t.kind='statement-templates' AND `+statementTemplateReadsAccount+`)`+
		` OR (t.kind='account-groups' AND `+accountGroupListsAccount+`)) GROUP BY t.kind ORDER BY t.kind`, company, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item AccountReference
		if err = rows.Scan(&item.Kind, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// accountReferences answers accounts action "references": read-only, one account by id.
func (s *PostgresStore) accountReferences(ctx context.Context, scope Scope, cmd Command) (Result, error) {
	if cmd.ID == "" {
		return Result{}, fmt.Errorf("กรุณาเลือกรายการบัญชี")
	}
	db, err := s.pg.database(ctx, scope.Holding)
	if err != nil {
		return Result{}, err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback()
	var account Account
	if err = s.loadRecord(ctx, tx, scope.Company, "accounts", cmd.ID, &account); err != nil {
		return Result{}, err
	}
	items, err := accountReferenceCounts(ctx, tx, scope.Company, account.AccountCode)
	if err != nil {
		return Result{}, err
	}
	return Result{References: &AccountReferences{Items: items}}, nil
}

// checkDuplicateAccountName: Champ saves no second account with the same Name1 (GLFrmChartofAccountFormView.cpp:158-196 —
// `name1 = '…'` on SQL Server: case-insensitive, trailing blanks ignored; the message names the other account's code).
// Name1 here is the Thai name. BC also trims both ends and compares in NFC so a Thai mark typed in another order matches.
// Checked for a new account or when the Thai name changes, so an older duplicate can still be deactivated or fixed.
func checkDuplicateAccountName(ctx context.Context, tx *sql.Tx, company string, next Account, old *Account) error {
	name := NormalizeCode(next.ThaiName())
	if name == "" {
		return nil
	}
	excludeID := ""
	if old != nil {
		if strings.EqualFold(NormalizeCode(old.ThaiName()), name) {
			return nil
		}
		excludeID = old.ID
	}
	var otherCode string
	err := tx.QueryRowContext(ctx, `SELECT r.code FROM gl_records r WHERE r.company=$1 AND r.kind='accounts' AND r.id<>$2 AND NOT COALESCE((r.payload->>'isdeleted')::boolean,false)
		AND EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(r.payload->'names','[]'::jsonb)) n WHERE n->>'code'='th' AND lower(normalize(btrim(n->>'name', E' \t\r\n'), NFC))=lower($3::text))
		ORDER BY r.code LIMIT 1`, company, excludeID, name).Scan(&otherCode)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	e := fieldError("account_name_duplicate", "names", fmt.Sprintf("ชื่อบัญชี “%s” ซ้ำกับบัญชีรหัส %s กรุณาใช้ชื่ออื่น", name, otherCode))
	e.Args = []string{name, otherCode}
	return e
}
