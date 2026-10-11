package cheque

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/doctype"
	"smlcloudplatform/internal/goapi/language"
)

// เอกสารเปลี่ยนสถานะเช็ค 8 ชนิด (419..426): บันทึก/แก้/ยกเลิกใน tx เดียวพร้อม event ของเช็คและใบสำคัญ GL
// ลำดับตรวจตามฟอร์มแม่ของ Champ (cqfrmbasetransstatus.cpp; cheque-bank.md §4.3 ข้อ 5):
// อ้างอิง (เอกสารที่ใหม่กว่า) → งวดปิด → เลขที่/วันที่ → ยอด/แถว → สมุด → สถานะเช็ค → ยอดรวม

// DocLineInput คือเช็คหนึ่งแถวในเอกสาร + ค่าใช้จ่ายของแถว (สตริงทศนิยม; ว่าง = 0)
type DocLineInput struct {
	ChequeID  string
	FeeAmount string
}

// DocTotals คือยอดท้ายเอกสารที่หน้าจอคำนวณมา (ถ้าส่งมา backend คำนวณใหม่แล้วต้องตรงกัน)
type DocTotals struct {
	Total string
	Fee   string
	Net   string
}

// DocInput คือเอกสารสถานะที่ผู้ใช้บันทึก; CancelType ใช้เฉพาะ 423, CancelPassed เฉพาะ 426
type DocInput struct {
	Flag         int
	DocNo        string
	Version      int64
	DocDate      string
	BookCode     string
	CancelType   *int
	CancelPassed bool
	Description  string
	Lines        []DocLineInput
	Totals       *DocTotals
}

// DocLine คือเช็คหนึ่งแถวของเอกสาร (ข้อมูลเช็คอ่านจากทะเบียน)
type DocLine struct {
	LineNo         int    `json:"line_no"`
	ChequeID       string `json:"cheque_id"`
	ChequeNo       string `json:"cheque_no"`
	PartnerCode    string `json:"partner_code"`
	PartnerName    string `json:"partner_name"`
	BankCode       string `json:"bank_code"`
	BankBranchCode string `json:"bank_branch_code"`
	ChequeDate     string `json:"cheque_date"`
	DueDate        string `json:"due_date"`
	Description    string `json:"description"`
	FromStatus     int    `json:"from_status"`
	Amount         string `json:"amount"`
	FeeAmount      string `json:"fee_amount"`
	NetAmount      string `json:"net_amount"`
}

// Doc คือเอกสารสถานะหนึ่งฉบับ; journal_docno = ใบสำคัญที่ใช้งานอยู่ (หรือใบล่าสุดถ้าเอกสารถูกยกเลิก)
type Doc struct {
	Transflag    int        `json:"transflag"`
	DocNo        string     `json:"doc_no"`
	DocDate      string     `json:"doc_date"`
	BranchCode   string     `json:"branch_code"`
	BookCode     string     `json:"book_code"`
	CancelType   *int       `json:"cancel_type"`
	CancelPassed bool       `json:"cancel_passed"`
	Description  string     `json:"description"`
	TotalAmount  string     `json:"total_amount"`
	FeeAmount    string     `json:"fee_amount"`
	NetAmount    string     `json:"net_amount"`
	GLStatus     string     `json:"gl_status"`
	JournalDocNo string     `json:"journal_docno"`
	Status       string     `json:"status"`
	CancelReason string     `json:"cancel_reason"`
	CancelledAt  *time.Time `json:"cancelled_at"`
	CancelledBy  string     `json:"cancelled_by"`
	Version      int64      `json:"version"`
	CreatedAt    time.Time  `json:"created_at"`
	CreatedBy    string     `json:"created_by"`
	UpdatedAt    time.Time  `json:"updated_at"`
	UpdatedBy    string     `json:"updated_by"`
	Lines        []DocLine  `json:"lines"`
}

// docValues คือเอกสารที่ตรวจรูปแบบแล้ว (ไม่แตะฐานข้อมูล)
type docValues struct {
	flag                       int
	rule                       docRule
	policy                     bookPolicy
	docNo, date, book, desc    string
	cancelType                 int
	cancelTypeSet, cancelPassd bool
	lines                      []docLineValue
	totals                     *[3]decimal.Decimal // total, fee, net จากหน้าจอ
}

type docLineValue struct {
	chequeID, field string
	fee             decimal.Decimal
}

func lineField(i int, name string) string { return fmt.Sprintf("lines[%d].%s", i, name) }

// normalizeDoc ตรวจรูปแบบเอกสารตามลำดับช่องบนจอ
func normalizeDoc(in DocInput) (docValues, error) {
	var (
		v   = docValues{flag: in.Flag}
		ok  bool
		err error
	)
	if v.rule, ok = ruleFor(in.Flag); !ok {
		return v, badRequest("cheque_err_transflag_invalid", "transflag", strconv.Itoa(in.Flag))
	}
	if !validCancelType(in.Flag, in.CancelType) {
		return v, badRequest("cheque_err_cancel_type_invalid", "cancel_type")
	}
	if in.CancelType != nil {
		v.cancelType, v.cancelTypeSet = *in.CancelType, true
	}
	v.cancelPassd = in.Flag == FlagIssuedCancel && in.CancelPassed
	v.policy = bookPolicyFor(in.Flag, v.cancelType)
	if v.docNo, err = cleanText(in.DocNo, "doc_no", maxDocNoLen, false); err != nil {
		return v, err
	}
	v.date = strings.TrimSpace(in.DocDate)
	if v.date == "" {
		return v, badRequest("cheque_err_required_field", "doc_date")
	}
	if !validISODate(v.date) {
		return v, badRequest("cheque_err_date_invalid", "doc_date")
	}
	if v.desc, err = cleanText(in.Description, "description", maxDescriptionLen, false); err != nil {
		return v, err
	}
	if v.book, err = cleanText(in.BookCode, "book_code", maxRefCodeLen, false); err != nil {
		return v, err
	}
	if v.policy.none {
		v.book = "" // ยกเลิกเช็คในมือไม่ใช้สมุด (cqfrmcheqincancel.cpp)
	}
	if v.policy.required && v.book == "" {
		return v, badRequest("cheque_err_book_required", "book_code")
	}
	if len(in.Lines) == 0 {
		return v, badRequest("cheque_err_no_lines", "lines")
	}
	seen := map[string]bool{}
	for i, l := range in.Lines {
		id := strings.ToLower(strings.TrimSpace(l.ChequeID))
		if !uuidPattern.MatchString(id) {
			return v, notFound("cheque_err_not_found", lineField(i, "cheque_id"))
		}
		if seen[id] {
			return v, badRequest("cheque_err_duplicate_line", lineField(i, "cheque_id"), strconv.Itoa(i+1))
		}
		seen[id] = true
		fee := decimal.Zero
		if s := strings.TrimSpace(l.FeeAmount); s != "" {
			if fee, ok = parseMoney(s); !ok {
				return v, badRequest("cheque_err_fee_invalid", lineField(i, "fee_amount"), strconv.Itoa(i+1))
			}
		}
		if !v.rule.fee && !fee.IsZero() {
			return v, badRequest("cheque_err_fee_invalid", lineField(i, "fee_amount"), strconv.Itoa(i+1))
		}
		v.lines = append(v.lines, docLineValue{chequeID: id, field: lineField(i, "cheque_id"), fee: fee})
	}
	if t := in.Totals; t != nil {
		var totals [3]decimal.Decimal
		for i, s := range []string{t.Total, t.Fee, t.Net} {
			if totals[i], ok = parseMoney(s); !ok {
				return v, badRequest("cheque_err_totals_mismatch", "totals")
			}
		}
		v.totals = &totals
	}
	return v, nil
}

// docLine คือแถวที่ตรวจกับทะเบียนแล้ว (หลังล็อกแถวเช็ค)
type docLine struct {
	cheque         lineCheque
	amount, fee    decimal.Decimal
	net            decimal.Decimal
	bookAfterwards string
}

// lockDocCheques ล็อกเช็คทุกใบของเอกสาร (เรียงตาม id กัน deadlock) แล้วตรวจกับกติกาของเอกสาร:
// สถานะที่เลือกได้ + วันที่ + สมุด (checkSelectable), วันที่เอกสารไม่ก่อน event สถานะล่าสุดของเช็ค, ค่าใช้จ่าย ≤ ยอดเช็ค
func lockDocCheques(ctx context.Context, tx *sql.Tx, company string, v docValues) ([]docLine, [3]decimal.Decimal, error) {
	var totals [3]decimal.Decimal
	ids := make([]string, len(v.lines))
	for i, l := range v.lines {
		ids[i] = l.chequeID
	}
	sorted := append([]string(nil), ids...)
	sort.Strings(sorted)
	rows, err := tx.QueryContext(ctx, `SELECT id::text, cheque_no, to_char(doc_date, 'YYYY-MM-DD'), to_char(due_date, 'YYYY-MM-DD'),
			COALESCE(book_code, ''), cheque_type, status, home_amount::text
		FROM cheque_registers WHERE company = $1 AND id = ANY($2::uuid[]) AND deleted_at IS NULL ORDER BY id FOR UPDATE`, company, pq.Array(sorted))
	if err != nil {
		return nil, totals, err
	}
	found := map[string]docLine{}
	for rows.Next() {
		var (
			l   docLine
			amt string
		)
		if err = rows.Scan(&l.cheque.id, &l.cheque.chequeNo, &l.cheque.docDate, &l.cheque.dueDate, &l.cheque.book,
			&l.cheque.chequeType, &l.cheque.status, &amt); err != nil {
			rows.Close()
			return nil, totals, err
		}
		l.amount = decimal.RequireFromString(amt)
		found[l.cheque.id] = l
	}
	if err = rows.Close(); err != nil {
		return nil, totals, err
	}
	lastDates, err := lastStatusDates(ctx, tx, company, sorted)
	if err != nil {
		return nil, totals, err
	}
	lines := make([]docLine, 0, len(v.lines))
	for i, in := range v.lines {
		l, ok := found[in.chequeID]
		if !ok {
			return nil, totals, notFound("cheque_err_not_found", in.field)
		}
		if err = checkSelectable(v.flag, v.cancelType, v.cancelPassd, v.date, v.book, l.cheque, in.field); err != nil {
			return nil, totals, err
		}
		// สถานะของเช็ค derive จาก event ล่าสุดตามวันที่ จึงห้ามเอกสารลงวันที่ก่อน event สถานะล่าสุดของเช็ค
		if last := lastDates[in.chequeID]; last > v.date {
			return nil, totals, badRequest("cheque_err_doc_before_last_event", "doc_date", l.cheque.chequeNo, thaiDate(last))
		}
		if in.fee.GreaterThan(l.amount) {
			return nil, totals, badRequest("cheque_err_fee_invalid", lineField(i, "fee_amount"), strconv.Itoa(i+1))
		}
		l.fee = in.fee
		l.net = l.amount.Sub(in.fee)
		l.bookAfterwards = bookAfter(v.flag, v.book, l.cheque.book)
		totals[0], totals[1], totals[2] = totals[0].Add(l.amount), totals[1].Add(l.fee), totals[2].Add(l.net)
		lines = append(lines, l)
	}
	if v.totals != nil && (!v.totals[0].Equal(totals[0]) || !v.totals[1].Equal(totals[1]) || !v.totals[2].Equal(totals[2])) {
		return nil, totals, badRequest("cheque_err_totals_mismatch", "totals")
	}
	return lines, totals, nil
}

// lastStatusDates คือวันที่ของ event สถานะล่าสุดที่ยังไม่ถูกยกเลิกของเช็คแต่ละใบ
func lastStatusDates(ctx context.Context, tx *sql.Tx, company string, ids []string) (map[string]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT cheque_id::text, to_char(MAX(event_date), 'YYYY-MM-DD') FROM cheque_events
		WHERE company = $1 AND cheque_id = ANY($2::uuid[]) AND voided_at IS NULL AND to_status IS NOT NULL GROUP BY cheque_id`, company, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dates := map[string]string{}
	for rows.Next() {
		var id, date string
		if err = rows.Scan(&id, &date); err != nil {
			return nil, err
		}
		dates[id] = date
	}
	return dates, rows.Err()
}

// writeDocLines เขียนแถวเอกสาร + event status_doc + สถานะ/สมุดใหม่ในทะเบียน (ครั้งละหนึ่งคำสั่งต่อตาราง)
func writeDocLines(ctx context.Context, tx *sql.Tx, scope Scope, v docValues, docNo string, lines []docLine) error {
	n := len(lines)
	ids, books := make([]string, n), make([]string, n)
	from := make([]int64, n)
	amounts, fees, nets := make([]string, n), make([]string, n), make([]string, n)
	for i, l := range lines {
		ids[i], books[i], from[i] = l.cheque.id, l.bookAfterwards, int64(l.cheque.status)
		amounts[i], fees[i], nets[i] = money2(l.amount), money2(l.fee), money2(l.net)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO cheque_doc_lines (company, transflag, doc_no, line_no, cheque_id, from_status, amount, fee_amount, net_amount)
		SELECT $1, $2::integer, $3, t.ord, t.id::uuid, t.from_status, t.amount::numeric, t.fee::numeric, t.net::numeric
		FROM unnest($4::text[], $5::smallint[], $6::text[], $7::text[], $8::text[]) WITH ORDINALITY AS t(id, from_status, amount, fee, net, ord)`,
		scope.Company, v.flag, docNo, pq.Array(ids), pq.Array(from), pq.Array(amounts), pq.Array(fees), pq.Array(nets)); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO cheque_events (company, cheque_id, event_kind, doc_flag, doc_no, event_date, from_status, to_status, book_code, actor)
		SELECT $1, t.id::uuid, 'status_doc', $2::integer, $3, $4::date, t.from_status, $5::smallint, NULLIF(t.book, ''), $6
		FROM unnest($7::text[], $8::smallint[], $9::text[]) WITH ORDINALITY AS t(id, from_status, book, ord) ORDER BY t.ord`,
		scope.Company, v.flag, docNo, v.date, v.rule.toStatus, scope.Actor, pq.Array(ids), pq.Array(from), pq.Array(books)); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `UPDATE cheque_registers r SET status = $2, book_code = NULLIF(t.book, ''), version = r.version + 1, updated_at = now(), updated_by = $3
		FROM unnest($4::text[], $5::text[]) AS t(id, book) WHERE r.company = $1 AND r.id = t.id::uuid`,
		scope.Company, v.rule.toStatus, scope.Actor, pq.Array(ids), pq.Array(books))
	return err
}

// postDoc ลงบัญชีเอกสารรอบใหม่ (นำฝาก 419 ไม่ลงบัญชี) แล้วผูกใบสำคัญกับเอกสาร
func postDoc(ctx context.Context, tx *sql.Tx, ledger Ledger, scope Scope, v docValues, docNo string, book bankBook, totals [3]decimal.Decimal) (string, error) {
	if !v.rule.posts {
		return GLStatusNotRequired, nil
	}
	var revision int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(revision), 0) + 1 FROM cheque_doc_journals WHERE company = $1 AND transflag = $2 AND doc_no = $3`,
		scope.Company, v.flag, docNo).Scan(&revision); err != nil {
		return "", err
	}
	posted, err := ledger.Post(ctx, tx, scope, Posting{
		Flag: v.flag, DocNo: docNo, Date: v.date, Revision: revision,
		BookCode: book.code, BookAccount: book.glAccount,
		BookFallback: v.flag == FlagReceivedCancel && v.cancelType == CancelOnHand,
		Total:        totals[0], Fee: totals[1], Net: totals[2],
	})
	if err != nil {
		return "", err
	}
	if posted.Status != GLStatusPosted {
		return posted.Status, nil
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO cheque_doc_journals (company, transflag, doc_no, revision, journal_id, journal_docno) VALUES ($1, $2, $3, $4, $5, $6)`,
		scope.Company, v.flag, docNo, revision, posted.JournalID, posted.JournalDocNo)
	return GLStatusPosted, err
}

// reverseDocJournal กลับรายการใบสำคัญที่ใช้งานอยู่ของเอกสาร (ถ้ามี) แล้วทำเครื่องหมายว่าเลิกใช้
func reverseDocJournal(ctx context.Context, tx *sql.Tx, ledger Ledger, scope Scope, flag int, docNo string, edit bool) error {
	var (
		revision  int
		journalID string
	)
	err := tx.QueryRowContext(ctx, `SELECT revision, journal_id FROM cheque_doc_journals WHERE company = $1 AND transflag = $2 AND doc_no = $3 AND voided_at IS NULL FOR UPDATE`,
		scope.Company, flag, docNo).Scan(&revision, &journalID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if err = ledger.Reverse(ctx, tx, scope, Reversal{Flag: flag, DocNo: docNo, Revision: revision, JournalID: journalID, Edit: edit, Today: businessDay(time.Now())}); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE cheque_doc_journals SET voided_at = now() WHERE company = $1 AND transflag = $2 AND doc_no = $3 AND revision = $4`,
		scope.Company, flag, docNo, revision)
	return err
}

// docBook อ่านสมุดเงินฝากของเอกสาร (ว่าง = เอกสารไม่ระบุสมุด)
func docBook(ctx context.Context, tx *sql.Tx, company, code string) (bankBook, error) {
	if code == "" {
		return bankBook{}, nil
	}
	return requireBook(ctx, tx, company, code)
}

// CreateDoc บันทึกเอกสารสถานะใหม่ (เลขที่ว่าง = ออกเลขอัตโนมัติตาม prefix ของ doctype)
func CreateDoc(ctx context.Context, db *sql.DB, ledger Ledger, scope Scope, requestID string, in DocInput) (Doc, error) {
	v, err := normalizeDoc(in)
	if err != nil {
		return Doc{}, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Doc{}, err
	}
	defer tx.Rollback()
	var out Doc
	if replayed, err := claimRequest(ctx, tx, scope, requestID, "docs:create", &out); err != nil || replayed {
		return out, err
	}
	if err = ledger.Lock(ctx, tx, scope.Company); err != nil {
		return Doc{}, err
	}
	if err = ledger.CheckDate(ctx, tx, scope, v.date); err != nil {
		return Doc{}, err
	}
	docNo := v.docNo
	if docNo == "" {
		dt, _ := doctype.ByFlag(v.flag)
		if err = lockNumbering(ctx, tx, scope.Company, "doc:"+strconv.Itoa(v.flag)); err != nil {
			return Doc{}, err
		}
		if docNo, err = nextDocNo(ctx, tx, `SELECT doc_no FROM cheque_docs WHERE company = $1 AND transflag = $2 AND doc_no LIKE $3 ORDER BY doc_no DESC LIMIT 1`,
			dt.Prefix, v.date, scope.Company, v.flag); err != nil {
			return Doc{}, err
		}
	}
	book, err := docBook(ctx, tx, scope.Company, v.book)
	if err != nil {
		return Doc{}, err
	}
	lines, totals, err := lockDocCheques(ctx, tx, scope.Company, v)
	if err != nil {
		return Doc{}, err
	}
	glStatus := GLStatusNone
	if !v.rule.posts {
		glStatus = GLStatusNotRequired
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO cheque_docs (company, transflag, doc_no, doc_date, branch_code, book_code, cancel_type, cancel_passed,
			description, total_amount, fee_amount, net_amount, gl_status, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $14)`,
		scope.Company, v.flag, docNo, v.date, scope.Branch, nullable(v.book), cancelTypeValue(v), v.cancelPassd, v.desc,
		money2(totals[0]), money2(totals[1]), money2(totals[2]), glStatus, scope.Actor)
	if err != nil {
		if isUniqueViolation(err) {
			return Doc{}, newError(http.StatusConflict, "cheque_err_duplicate_doc_no", "doc_no", docNo)
		}
		return Doc{}, err
	}
	if err = writeDocLines(ctx, tx, scope, v, docNo, lines); err != nil {
		return Doc{}, err
	}
	if glStatus, err = postDoc(ctx, tx, ledger, scope, v, docNo, book, totals); err != nil {
		return Doc{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE cheque_docs SET gl_status = $4 WHERE company = $1 AND transflag = $2 AND doc_no = $3`,
		scope.Company, v.flag, docNo, glStatus); err != nil {
		return Doc{}, err
	}
	return finishDoc(ctx, tx, scope, requestID, v.flag, docNo)
}

func cancelTypeValue(v docValues) any {
	if !v.cancelTypeSet {
		return nil
	}
	return v.cancelType
}

// finishDoc อ่านเอกสารที่บันทึกแล้ว เก็บผลไว้ตอบคำขอซ้ำ แล้ว commit
func finishDoc(ctx context.Context, tx *sql.Tx, scope Scope, requestID string, flag int, docNo string) (Doc, error) {
	out, err := GetDoc(ctx, tx, scope.Company, flag, docNo)
	if err != nil {
		return Doc{}, err
	}
	if err = completeRequest(ctx, tx, scope, requestID, out); err != nil {
		return Doc{}, err
	}
	return out, tx.Commit()
}

// lockedDoc คือหัวเอกสารที่ล็อกแถวแล้ว
type lockedDoc struct {
	date, status string
	version      int64
}

// lockDoc ล็อกหัวเอกสาร แล้วตรวจว่ายังใช้งาน + version ตรง + ไม่มีเช็คในเอกสารที่มี event ใหม่กว่า (Champ IsChqReference)
func lockDoc(ctx context.Context, tx *sql.Tx, company string, flag int, docNo string, version int64) (lockedDoc, error) {
	var d lockedDoc
	err := tx.QueryRowContext(ctx, `SELECT to_char(doc_date, 'YYYY-MM-DD'), status, version FROM cheque_docs
		WHERE company = $1 AND transflag = $2 AND doc_no = $3 FOR UPDATE`, company, flag, docNo).Scan(&d.date, &d.status, &d.version)
	if errors.Is(err, sql.ErrNoRows) {
		return d, notFound("cheque_err_doc_not_found", "doc_no", docNo)
	}
	if err != nil {
		return d, err
	}
	if d.status != "active" {
		return d, newError(http.StatusConflict, "cheque_err_already_cancelled", "doc_no", docNo)
	}
	if d.version != version {
		return d, versionConflict()
	}
	chequeNo, newer, err := newerEvent(ctx, tx, company, flag, docNo)
	if err != nil {
		return d, err
	}
	if newer != "" {
		return d, newError(http.StatusConflict, "cheque_err_newer_event", "doc_no", chequeNo, newer)
	}
	return d, nil
}

// releaseDoc ถอนผลของเอกสารเดิม: void event ของเช็คทุกใบ → คำนวณสถานะเช็คใหม่จากประวัติ → กลับรายการใบสำคัญ
func releaseDoc(ctx context.Context, tx *sql.Tx, ledger Ledger, scope Scope, flag int, docNo, reason string, edit bool) error {
	ids, err := voidDocEvents(ctx, tx, scope, flag, docNo, reason)
	if err != nil {
		return err
	}
	if err = restoreStates(ctx, tx, scope, ids); err != nil {
		return err
	}
	return reverseDocJournal(ctx, tx, ledger, scope, flag, docNo, edit)
}

// UpdateDoc แก้เอกสารสถานะ: ถอนผลเดิมทั้งหมด (event void + ใบสำคัญกลับรายการ) แล้วบันทึกเนื้อหาใหม่เป็นรอบใหม่
// — เอกสารที่ลงบัญชีแล้วไม่ถูกแก้เงียบ ใบเดิมคงอยู่เป็นหลักฐาน (cheque-bank.md §4.4)
func UpdateDoc(ctx context.Context, db *sql.DB, ledger Ledger, scope Scope, requestID string, in DocInput) (Doc, error) {
	v, err := normalizeDoc(in)
	if err != nil {
		return Doc{}, err
	}
	if v.docNo == "" {
		return Doc{}, badRequest("cheque_err_required_field", "doc_no")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Doc{}, err
	}
	defer tx.Rollback()
	var out Doc
	if replayed, err := claimRequest(ctx, tx, scope, requestID, "docs:update", &out); err != nil || replayed {
		return out, err
	}
	if err = ledger.Lock(ctx, tx, scope.Company); err != nil {
		return Doc{}, err
	}
	cur, err := lockDoc(ctx, tx, scope.Company, v.flag, v.docNo, in.Version)
	if err != nil {
		return Doc{}, err
	}
	if err = ledger.CheckDate(ctx, tx, scope, cur.date); err != nil {
		return Doc{}, err
	}
	if v.date != cur.date {
		if err = ledger.CheckDate(ctx, tx, scope, v.date); err != nil {
			return Doc{}, err
		}
	}
	if err = releaseDoc(ctx, tx, ledger, scope, v.flag, v.docNo, docReasonText("erp_gl_reverse_edit", v.docNo), true); err != nil {
		return Doc{}, err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM cheque_doc_lines WHERE company = $1 AND transflag = $2 AND doc_no = $3`, scope.Company, v.flag, v.docNo); err != nil {
		return Doc{}, err
	}
	book, err := docBook(ctx, tx, scope.Company, v.book)
	if err != nil {
		return Doc{}, err
	}
	lines, totals, err := lockDocCheques(ctx, tx, scope.Company, v)
	if err != nil {
		return Doc{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE cheque_docs SET doc_date = $4, book_code = $5, cancel_type = $6, cancel_passed = $7, description = $8,
			total_amount = $9, fee_amount = $10, net_amount = $11, version = version + 1, updated_at = now(), updated_by = $12
		WHERE company = $1 AND transflag = $2 AND doc_no = $3`,
		scope.Company, v.flag, v.docNo, v.date, nullable(v.book), cancelTypeValue(v), v.cancelPassd, v.desc,
		money2(totals[0]), money2(totals[1]), money2(totals[2]), scope.Actor); err != nil {
		return Doc{}, err
	}
	if err = writeDocLines(ctx, tx, scope, v, v.docNo, lines); err != nil {
		return Doc{}, err
	}
	glStatus, err := postDoc(ctx, tx, ledger, scope, v, v.docNo, book, totals)
	if err != nil {
		return Doc{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE cheque_docs SET gl_status = $4 WHERE company = $1 AND transflag = $2 AND doc_no = $3`,
		scope.Company, v.flag, v.docNo, glStatus); err != nil {
		return Doc{}, err
	}
	return finishDoc(ctx, tx, scope, requestID, v.flag, v.docNo)
}

// docReasonText คือเหตุผลของการถอนผลเอกสาร (ภาษาไทย เก็บเป็นหลักฐานใน event/ใบกลับรายการ)
func docReasonText(key, docNo string) string {
	return strings.ReplaceAll(language.Text(key, "th"), "{0}", docNo)
}

// CancelDoc ยกเลิกเอกสารสถานะพร้อมเหตุผล: สถานะเช็คกลับตามประวัติ ใบสำคัญถูกกลับรายการ (ไม่ลบ)
func CancelDoc(ctx context.Context, db *sql.DB, ledger Ledger, scope Scope, requestID string, flag int, docNo string, version int64, reason string) (Doc, error) {
	if _, ok := ruleFor(flag); !ok {
		return Doc{}, badRequest("cheque_err_transflag_invalid", "transflag", strconv.Itoa(flag))
	}
	docNo = strings.TrimSpace(docNo)
	reason, err := cleanText(reason, "reason", maxDescriptionLen, true)
	if err != nil {
		if e, ok := AsError(err); ok && e.Key == "cheque_err_required_field" {
			return Doc{}, badRequest("cheque_err_reason_required", "reason")
		}
		return Doc{}, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Doc{}, err
	}
	defer tx.Rollback()
	var out Doc
	if replayed, err := claimRequest(ctx, tx, scope, requestID, "docs:cancel", &out); err != nil || replayed {
		return out, err
	}
	if err = ledger.Lock(ctx, tx, scope.Company); err != nil {
		return Doc{}, err
	}
	cur, err := lockDoc(ctx, tx, scope.Company, flag, docNo, version)
	if err != nil {
		return Doc{}, err
	}
	if err = ledger.CheckDate(ctx, tx, scope, cur.date); err != nil {
		return Doc{}, err
	}
	if err = releaseDoc(ctx, tx, ledger, scope, flag, docNo, reason, false); err != nil {
		return Doc{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE cheque_docs SET status = 'cancelled', cancel_reason = $4, cancelled_at = now(), cancelled_by = $5,
			version = version + 1, updated_at = now(), updated_by = $5
		WHERE company = $1 AND transflag = $2 AND doc_no = $3`, scope.Company, flag, docNo, reason, scope.Actor); err != nil {
		return Doc{}, err
	}
	return finishDoc(ctx, tx, scope, requestID, flag, docNo)
}

const docSelect = `SELECT d.transflag, d.doc_no, to_char(d.doc_date, 'YYYY-MM-DD'), d.branch_code, COALESCE(d.book_code, ''), d.cancel_type,
	d.cancel_passed, d.description, d.total_amount::text, d.fee_amount::text, d.net_amount::text, d.gl_status,
	COALESCE((SELECT j.journal_docno FROM cheque_doc_journals j WHERE j.company = d.company AND j.transflag = d.transflag AND j.doc_no = d.doc_no
		ORDER BY (j.voided_at IS NULL) DESC, j.revision DESC LIMIT 1), ''),
	d.status, COALESCE(d.cancel_reason, ''), d.cancelled_at, COALESCE(d.cancelled_by, ''), d.version, d.created_at, d.created_by, d.updated_at, d.updated_by
FROM cheque_docs d`

func scanDoc(row rowScanner) (Doc, error) {
	var (
		d          Doc
		cancelType sql.NullInt64
		cancelled  sql.NullTime
	)
	err := row.Scan(&d.Transflag, &d.DocNo, &d.DocDate, &d.BranchCode, &d.BookCode, &cancelType, &d.CancelPassed, &d.Description,
		&d.TotalAmount, &d.FeeAmount, &d.NetAmount, &d.GLStatus, &d.JournalDocNo, &d.Status, &d.CancelReason, &cancelled, &d.CancelledBy,
		&d.Version, &d.CreatedAt, &d.CreatedBy, &d.UpdatedAt, &d.UpdatedBy)
	d.CancelType = nullInt(cancelType)
	if cancelled.Valid {
		d.CancelledAt = &cancelled.Time
	}
	return d, err
}

// GetDoc คือเอกสารหนึ่งฉบับพร้อมแถวเช็ค
func GetDoc(ctx context.Context, q querier, company string, flag int, docNo string) (Doc, error) {
	docNo = strings.TrimSpace(docNo)
	d, err := scanDoc(q.QueryRowContext(ctx, docSelect+` WHERE d.company = $1 AND d.transflag = $2 AND d.doc_no = $3`, company, flag, docNo))
	if errors.Is(err, sql.ErrNoRows) {
		return Doc{}, notFound("cheque_err_doc_not_found", "doc_no", docNo)
	}
	if err != nil {
		return Doc{}, err
	}
	rows, err := q.QueryContext(ctx, `SELECT l.line_no, l.cheque_id::text, r.cheque_no, r.partner_code, COALESCE(p.payload->>'name_th', ''),
			r.bank_code, r.bank_branch_code, to_char(r.doc_date, 'YYYY-MM-DD'), to_char(r.due_date, 'YYYY-MM-DD'), r.description,
			l.from_status, l.amount::text, l.fee_amount::text, l.net_amount::text
		FROM cheque_doc_lines l
		JOIN cheque_registers r ON r.company = l.company AND r.id = l.cheque_id
		LEFT JOIN gl_subledger_partners p ON p.company = r.company AND p.code = r.partner_code
		WHERE l.company = $1 AND l.transflag = $2 AND l.doc_no = $3 ORDER BY l.line_no`, company, flag, docNo)
	if err != nil {
		return Doc{}, err
	}
	defer rows.Close()
	d.Lines = []DocLine{}
	for rows.Next() {
		var l DocLine
		if err = rows.Scan(&l.LineNo, &l.ChequeID, &l.ChequeNo, &l.PartnerCode, &l.PartnerName, &l.BankCode, &l.BankBranchCode,
			&l.ChequeDate, &l.DueDate, &l.Description, &l.FromStatus, &l.Amount, &l.FeeAmount, &l.NetAmount); err != nil {
			return Doc{}, err
		}
		d.Lines = append(d.Lines, l)
	}
	return d, rows.Err()
}

// DocFilter คือเงื่อนไขค้นเอกสารสถานะ (Status ว่าง = ทุกสถานะ, active, cancelled)
type DocFilter struct {
	Query    string
	DateFrom string
	DateTo   string
	Status   string
	Page     int
	Limit    int
}

// DocPage คือผลค้นหนึ่งหน้า (หัวเอกสารเท่านั้น)
type DocPage struct {
	Items []Doc
	Total int
	Page  int
	Limit int
}

// ListDocs ค้นเอกสารสถานะชนิดหนึ่ง เรียงวันที่ล่าสุดก่อน
func ListDocs(ctx context.Context, q querier, company string, flag int, f DocFilter) (DocPage, error) {
	if _, ok := ruleFor(flag); !ok {
		return DocPage{}, badRequest("cheque_err_transflag_invalid", "transflag", strconv.Itoa(flag))
	}
	where := []string{"d.company = $1", "d.transflag = $2"}
	args := []any{company, flag}
	add := func(cond string, value any) {
		args = append(args, value)
		where = append(where, strings.ReplaceAll(cond, "?", "$"+strconv.Itoa(len(args))))
	}
	switch s := strings.TrimSpace(f.Status); s {
	case "":
	case "active", "cancelled":
		add("d.status = ?", s)
	default:
		return DocPage{}, badRequest("cheque_err_status_invalid", "status")
	}
	for _, d := range []struct{ value, field, cond string }{{f.DateFrom, "date_from", "d.doc_date >= ?"}, {f.DateTo, "date_to", "d.doc_date <= ?"}} {
		if s := strings.TrimSpace(d.value); s != "" {
			if !validISODate(s) {
				return DocPage{}, badRequest("cheque_err_date_invalid", d.field)
			}
			add(d.cond, s)
		}
	}
	if s := strings.TrimSpace(f.Query); s != "" {
		add("(d.doc_no ILIKE ? OR d.description ILIKE ?)", likePattern(s))
	}
	page, limit, offset := pageWindow(f.Page, f.Limit, defaultRegisterLimit, maxRegisterLimit)
	result := DocPage{Items: []Doc{}, Page: page, Limit: limit}
	cond := strings.Join(where, " AND ")
	if err := q.QueryRowContext(ctx, `SELECT count(*) FROM cheque_docs d WHERE `+cond, args...).Scan(&result.Total); err != nil {
		return DocPage{}, err
	}
	n := len(args)
	rows, err := q.QueryContext(ctx, docSelect+` WHERE `+cond+fmt.Sprintf(` ORDER BY d.doc_date DESC, d.doc_no DESC LIMIT $%d OFFSET $%d`, n+1, n+2),
		append(args, limit, offset)...)
	if err != nil {
		return DocPage{}, err
	}
	defer rows.Close()
	for rows.Next() {
		d, err := scanDoc(rows)
		if err != nil {
			return DocPage{}, err
		}
		result.Items = append(result.Items, d)
	}
	return result, rows.Err()
}
