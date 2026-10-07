package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Adjust purchase VAT (Champ 5498, CLibDlgAdjInputTax in LibDlgUserRemind.cpp:197-441).
//
// GET /gl/v2/vat-adjustments lists the purchase VAT rows (details.vats, tax_type 1) of POSTED journals filtered by
// document date, tax invoice date and creditor code (Champ OnProcess :317-400).
//
// The command (resource "vat-adjustments", action "update") changes the claim period (tax_period_year/month = Champ
// TaxDate2) and the claim status of many rows at once. Every journal goes through mutateReconciliation, the very path
// of the single-journal reconcile screen, so validation, the 6-month claim window, the version check and the audit
// (gl_subledger_audit vat_replace + reconcile, with the reason) are the same. The whole command is ONE transaction,
// so one rejected row leaves every other row untouched.
//
// Mode "invoice_date" is Champ's button "Update filing date = tax invoice date": the claim period of the rows matching
// the filter becomes the month of their tax invoice date. Champ's second button updates the whole BCInputTax table
// (:431-440); BC never does that, only the rows of the filter, which the server re-selects inside the transaction.
const (
	VatAdjustKind            = "vat-adjustments"
	VatAdjustModeRows        = "rows"
	VatAdjustModeInvoiceDate = "invoice_date"
	// One command rewrites at most this many journals / VAT rows: every journal payload is stored in the audit event.
	VatAdjustMaxJournals = 200
	VatAdjustMaxRows     = 1000
	vatAdjustDefaultPage = 100
	vatAdjustMaxPage     = 500
)

// VatAdjustFilter mirrors the Champ dialog. Champ semantics (OnProcess): with only the "from" value the filter is an
// exact match, with both a BETWEEN; "to" alone is an upper bound here (Champ ignores it). At least one date range of
// the document date / tax invoice date is required (Champ refuses to load otherwise).
type VatAdjustFilter struct {
	DocDateFrom string `json:"docdatefrom,omitempty"`
	DocDateTo   string `json:"docdateto,omitempty"`
	InvDateFrom string `json:"invdatefrom,omitempty"`
	InvDateTo   string `json:"invdateto,omitempty"`
	PartnerFrom string `json:"partnerfrom,omitempty"`
	PartnerTo   string `json:"partnerto,omitempty"`
}

// VatAdjustItem is one row the user changed, with the journal version the screen loaded.
type VatAdjustItem struct {
	JournalID      string `json:"journalid"`
	Version        int64  `json:"version"`
	VatID          string `json:"vatid"`
	TaxPeriodYear  int    `json:"taxperiodyear"`
	TaxPeriodMonth int    `json:"taxperiodmonth"`
	ClaimStatus    int    `json:"claimstatus"`
}

// VatAdjustInput: mode "rows" carries Items; mode "invoice_date" carries Filter and the row count the screen showed.
type VatAdjustInput struct {
	Mode          string          `json:"mode"`
	Filter        VatAdjustFilter `json:"filter"`
	ExpectedCount int             `json:"expectedcount,omitempty"`
	Items         []VatAdjustItem `json:"items,omitempty"`
}

// VatAdjustRow is a listed VAT row with its journal; Version is the journal version to send back on save.
type VatAdjustRow struct {
	JournalID  string `json:"journalid"`
	Version    int64  `json:"version"`
	DocNo      string `json:"docno"`
	DocDate    string `json:"docdate"`
	BranchCode string `json:"branchcode,omitempty"`
	SubledgerVat
}

func normalizeVatAdjustFilter(f VatAdjustFilter) VatAdjustFilter {
	for _, value := range []*string{&f.DocDateFrom, &f.DocDateTo, &f.InvDateFrom, &f.InvDateTo} {
		*value = strings.TrimSpace(*value)
	}
	f.PartnerFrom, f.PartnerTo = NormalizeCode(f.PartnerFrom), NormalizeCode(f.PartnerTo)
	return f
}

func validateVatAdjustFilter(f VatAdjustFilter) error {
	for _, d := range []struct{ field, value string }{{"docdatefrom", f.DocDateFrom}, {"docdateto", f.DocDateTo}, {"invdatefrom", f.InvDateFrom}, {"invdateto", f.InvDateTo}} {
		if d.value != "" && !validDate(d.value) {
			return fieldError("vat_adjust_date_invalid", d.field, "วันที่ในเงื่อนไขไม่ถูกต้อง กรุณาเลือกวันที่ให้ครบ วัน เดือน ปี")
		}
	}
	for _, r := range []struct{ field, from, to string }{{"docdateto", f.DocDateFrom, f.DocDateTo}, {"invdateto", f.InvDateFrom, f.InvDateTo}, {"partnerto", f.PartnerFrom, f.PartnerTo}} {
		if r.from != "" && r.to != "" && r.from > r.to {
			return fieldError("vat_adjust_range_invalid", r.field, "ค่าสิ้นสุดต้องไม่น้อยกว่าค่าเริ่มต้นของช่วงเงื่อนไข")
		}
	}
	if f.DocDateFrom == "" && f.DocDateTo == "" && f.InvDateFrom == "" && f.InvDateTo == "" {
		return fieldError("vat_adjust_filter_required", "docdatefrom", "กรุณาระบุช่วงวันที่เอกสารหรือวันที่ใบกำกับภาษีอย่างน้อย 1 ช่วง ก่อนประมวลผล")
	}
	return nil
}

// vatAdjustSelect: the purchase VAT rows of posted, not-reversed journals matching the filter. $1 company, $2 branch
// scope (” = company wide); the filter conditions follow from $3. ORDER BY is the caller's (list pages, command locks).
func vatAdjustSelect(f VatAdjustFilter, scope Scope) (string, []any) {
	args := []any{scope.Company, scope.Branch}
	conds := []string{}
	add := func(expr, from, to string) {
		switch {
		case from != "" && to != "":
			args = append(args, from, to)
			conds = append(conds, fmt.Sprintf("%s BETWEEN $%d AND $%d", expr, len(args)-1, len(args)))
		case from != "":
			args = append(args, from)
			conds = append(conds, fmt.Sprintf("%s = $%d", expr, len(args)))
		case to != "":
			args = append(args, to)
			conds = append(conds, fmt.Sprintf("%s <= $%d", expr, len(args)))
		}
	}
	add("COALESCE(r.payload->>'date', '')", f.DocDateFrom, f.DocDateTo)
	add("LEFT(COALESCE(v.item->>'tax_invoice_date', ''), 10)", f.InvDateFrom, f.InvDateTo)
	add("COALESCE(v.item->>'partner_code', '')", f.PartnerFrom, f.PartnerTo)
	where := ""
	for _, cond := range conds {
		where += " AND " + cond
	}
	return `SELECT r.id AS journal_id, r.version AS version, r.code AS doc_no, COALESCE(r.payload->>'date', '') AS doc_date,
  COALESCE(r.payload->>'branchcode', '') AS branch_code, v.item AS item, v.ord AS ord
FROM gl_records r
CROSS JOIN LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(r.payload->'details'->'vats') = 'array'
  THEN r.payload->'details'->'vats' ELSE '[]'::jsonb END) WITH ORDINALITY AS v(item, ord)
WHERE r.company = $1 AND r.kind = 'journals'
  AND r.payload->>'status' = 'posted' AND NOT COALESCE((r.payload->>'isdeleted')::boolean, false)
  AND COALESCE(r.payload->>'kind', '') <> 'reversal'
  AND ($2 = '' OR COALESCE(r.payload->>'branchcode', '') = $2)
  AND v.item @> '{"tax_type":1}'::jsonb` + where, args
}

const vatAdjustOrder = ` ORDER BY v.item->>'tax_invoice_date', r.code, v.ord`

func scanVatAdjustRows(rows *sql.Rows) ([]VatAdjustRow, error) {
	defer rows.Close()
	out := []VatAdjustRow{}
	for rows.Next() {
		var row VatAdjustRow
		var item []byte
		var ord int
		if err := rows.Scan(&row.JournalID, &row.Version, &row.DocNo, &row.DocDate, &row.BranchCode, &item, &ord); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(item, &row.SubledgerVat); err != nil {
			return nil, fmt.Errorf("parse vat row of %s: %w", row.DocNo, err)
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// VatAdjustList pages the filtered purchase VAT rows (LIMIT/OFFSET over a fixed ORDER BY: tax invoice date, doc no, row order).
func (p *Postgres) VatAdjustList(ctx context.Context, scope Scope, f VatAdjustFilter, page, limit int) (Page, error) {
	out := Page{Items: []json.RawMessage{}}
	f = normalizeVatAdjustFilter(f)
	if err := validateVatAdjustFilter(f); err != nil {
		return out, err
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = vatAdjustDefaultPage
	}
	if limit > vatAdjustMaxPage {
		limit = vatAdjustMaxPage
	}
	db, err := p.database(ctx, scope.Holding)
	if err != nil {
		return out, err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	selectSQL, args := vatAdjustSelect(f, scope)
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM (`+selectSQL+`) counted`, args...).Scan(&out.Total); err != nil {
		return out, err
	}
	rows, err := tx.QueryContext(ctx, selectSQL+vatAdjustOrder+fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2), append(args, limit, (page-1)*limit)...)
	if err != nil {
		return out, err
	}
	list, err := scanVatAdjustRows(rows)
	if err != nil {
		return out, err
	}
	for _, row := range list {
		raw, err := json.Marshal(row)
		if err != nil {
			return out, err
		}
		out.Items = append(out.Items, raw)
	}
	out.Page, out.Limit = page, limit
	return out, nil
}

type vatAdjustTarget struct {
	vatID                    string
	year, month, claimStatus int
}

type vatAdjustJournal struct {
	id      string
	version int64
	targets []vatAdjustTarget
}

func vatAdjustRowsInvalid() *UserError {
	return fieldError("vat_adjust_rows_invalid", "items", "รายการที่ส่งมาไม่ถูกต้อง (ว่าง ซ้ำ หรือคนละรุ่นของใบสำคัญเดียวกัน) กรุณาประมวลผลใหม่แล้วบันทึกอีกครั้ง")
}

func vatAdjustTooMany() *UserError {
	return &UserError{Code: "vat_adjust_too_many", Field: "items", Status: http.StatusBadRequest, Args: []string{strconv.Itoa(VatAdjustMaxJournals), strconv.Itoa(VatAdjustMaxRows)},
		Message: fmt.Sprintf("ปรับปรุงได้ครั้งละไม่เกิน %d ใบสำคัญ และไม่เกิน %d รายการภาษี กรุณากำหนดเงื่อนไขให้แคบลงแล้วทำเป็นหลายรอบ", VatAdjustMaxJournals, VatAdjustMaxRows)}
}

// vatAdjustPlan turns the command into per-journal targets (journals sorted by id so every run locks in the same order).
func (s *PostgresStore) vatAdjustPlan(ctx context.Context, tx *sql.Tx, scope Scope, in VatAdjustInput) ([]vatAdjustJournal, error) {
	byJournal := map[string]*vatAdjustJournal{}
	add := func(journalID string, version int64, t vatAdjustTarget) {
		j := byJournal[journalID]
		if j == nil {
			j = &vatAdjustJournal{id: journalID, version: version}
			byJournal[journalID] = j
		}
		j.targets = append(j.targets, t)
	}
	rowCount := 0
	switch in.Mode {
	case VatAdjustModeRows:
		if len(in.Items) == 0 {
			return nil, vatAdjustRowsInvalid()
		}
		seen := map[string]bool{}
		versions := map[string]int64{}
		for _, item := range in.Items {
			item.JournalID, item.VatID = strings.TrimSpace(item.JournalID), strings.TrimSpace(item.VatID)
			key := item.JournalID + "\x00" + item.VatID
			if item.JournalID == "" || item.VatID == "" || item.Version < 1 || seen[key] || (versions[item.JournalID] != 0 && versions[item.JournalID] != item.Version) {
				return nil, vatAdjustRowsInvalid()
			}
			seen[key], versions[item.JournalID] = true, item.Version
			add(item.JournalID, item.Version, vatAdjustTarget{vatID: item.VatID, year: item.TaxPeriodYear, month: item.TaxPeriodMonth, claimStatus: item.ClaimStatus})
			rowCount++
		}
	case VatAdjustModeInvoiceDate:
		filter := normalizeVatAdjustFilter(in.Filter)
		if err := validateVatAdjustFilter(filter); err != nil {
			return nil, err
		}
		if len(in.Items) != 0 || in.ExpectedCount < 1 {
			return nil, vatAdjustRowsInvalid()
		}
		selectSQL, args := vatAdjustSelect(filter, scope)
		var actual int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM (`+selectSQL+`) counted`, args...).Scan(&actual); err != nil {
			return nil, err
		}
		if actual != in.ExpectedCount {
			return nil, conflictError("vat_adjust_filter_changed", "expectedcount", "รายการตามเงื่อนไขเปลี่ยนไปจากที่แสดงบนจอ (มีผู้อื่นแก้ไขหรือผ่านรายการเพิ่ม) กรุณากดประมวลผลใหม่แล้วตรวจรายการก่อนบันทึก")
		}
		if actual > VatAdjustMaxRows {
			return nil, vatAdjustTooMany()
		}
		rows, err := tx.QueryContext(ctx, selectSQL+vatAdjustOrder, args...)
		if err != nil {
			return nil, err
		}
		list, err := scanVatAdjustRows(rows)
		if err != nil {
			return nil, err
		}
		for _, row := range list {
			invoiceDate, err := time.Parse("2006-01-02", row.TaxInvoiceDate)
			if err != nil {
				return nil, vatAdjustJournalError(row.DocNo, fieldError("vat_invoice_date_invalid", "tax_invoice_date", "กรุณาระบุวันที่ใบกำกับภาษีให้ถูกต้อง"))
			}
			add(row.JournalID, row.Version, vatAdjustTarget{vatID: row.ID, year: invoiceDate.Year(), month: int(invoiceDate.Month()), claimStatus: row.ClaimStatus})
			rowCount++
		}
	default:
		return nil, vatAdjustRowsInvalid()
	}
	if len(byJournal) > VatAdjustMaxJournals || rowCount > VatAdjustMaxRows {
		return nil, vatAdjustTooMany()
	}
	out := make([]vatAdjustJournal, 0, len(byJournal))
	for _, j := range byJournal {
		out = append(out, *j)
	}
	sort.Slice(out, func(a, b int) bool { return out[a].id < out[b].id })
	return out, nil
}

// vatAdjustJournalError tells which voucher refused the change; code, field and the other-language text stay those of the
// underlying rule (the voucher number is only added to the Thai text).
func vatAdjustJournalError(docNo string, err error) error {
	if user, ok := AsUserError(err); ok {
		named := *user
		named.Message = "ใบสำคัญ " + docNo + ": " + user.Message
		return &named
	}
	return err
}

// mutateVatAdjust applies the plan journal by journal through mutateReconciliation inside the caller's transaction.
func (s *PostgresStore) mutateVatAdjust(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	if cmd.Action != "update" {
		return nil, fmt.Errorf("ไม่รองรับคำสั่งบัญชีนี้")
	}
	if cmd.VatAdjust == nil {
		return nil, vatAdjustRowsInvalid()
	}
	plan, err := s.vatAdjustPlan(ctx, tx, scope, *cmd.VatAdjust)
	if err != nil {
		return nil, err
	}
	changes := []Change{}
	for _, target := range plan {
		var current Journal
		if err = s.loadRecord(ctx, tx, scope.Company, "journals", target.id, &current); err != nil {
			return nil, err
		}
		if current.Details == nil {
			return nil, vatAdjustJournalError(current.DocNo, fieldError("vat_adjust_row_not_found", "items", "ไม่พบรายการภาษีซื้อของใบสำคัญนี้ (ถูกแก้ไขหรือลบแล้ว) กรุณาประมวลผลใหม่"))
		}
		vats := append([]SubledgerVat{}, current.Details.Vats...)
		changed := false
		for _, t := range target.targets {
			index := -1
			for i := range vats {
				if vats[i].ID == t.vatID && vats[i].TaxType == 1 {
					index = i
				}
			}
			if index < 0 {
				return nil, vatAdjustJournalError(current.DocNo, fieldError("vat_adjust_row_not_found", "items", "ไม่พบรายการภาษีซื้อของใบสำคัญนี้ (ถูกแก้ไขหรือลบแล้ว) กรุณาประมวลผลใหม่"))
			}
			row := &vats[index]
			if row.TaxPeriodYear == t.year && row.TaxPeriodMonth == t.month && row.ClaimStatus == t.claimStatus {
				continue
			}
			row.TaxPeriodYear, row.TaxPeriodMonth, row.ClaimStatus = t.year, t.month, t.claimStatus
			changed = true
		}
		if !changed {
			continue
		}
		one, err := s.mutateReconciliation(ctx, tx, scope, Command{Resource: "journals", Action: "reconcile", ID: target.id, Version: target.version, Reason: cmd.Reason,
			Journal: &Journal{Details: &JournalDetails{Vats: vats}}}, now)
		if err != nil {
			return nil, vatAdjustJournalError(current.DocNo, err)
		}
		if len(one) != 1 {
			return nil, fmt.Errorf("vat adjust of %s touched %d records, want 1", current.DocNo, len(one))
		}
		changes = append(changes, one...)
	}
	if len(changes) == 0 {
		return nil, userError("vat_adjust_nothing_changed", "ไม่มีรายการที่ต้องเปลี่ยน — งวดและสถานะการใช้สิทธิของรายการที่เลือกตรงกับที่ต้องการอยู่แล้ว")
	}
	return changes, nil
}
