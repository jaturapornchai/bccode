// Package glposting ลงบัญชีเอกสารเปลี่ยนสถานะเช็คผ่านเครื่องบัญชีกลาง (generalledger) ใน tx เดียวกับเอกสาร —
// แยกจาก package cheque เพราะ generalledger import cheque อยู่แล้ว (สมุดเงินฝากใช้รหัสธนาคารของ cheque)
package glposting

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/cheque"
	"smlcloudplatform/internal/doctype"
	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/goapi/language"
)

const (
	actionCreate  = "create"
	actionPost    = "post"
	actionReverse = "reverse"
)

// Ledger คือ cheque.Ledger ที่ลงบัญชีด้วยรูปแบบลงบัญชี (mappings) ของ doctype เช็ค
type Ledger struct{}

var _ cheque.Ledger = Ledger{}

func glScope(s cheque.Scope) gl.Scope {
	return gl.Scope{Holding: s.Holding, Company: s.Company, Branch: s.Branch, Actor: s.Actor}
}

// Lock ใช้ล็อกบริษัทเดียวกับคำสั่งบัญชีทุกคำสั่ง ก่อนเอกสารเช็คแตะแถวของตัวเอง (กัน deadlock กับเครื่องบัญชี)
func (Ledger) Lock(ctx context.Context, tx *sql.Tx, company string) error {
	return gl.LockCompanyTx(ctx, tx, company)
}

// CheckDate: วันที่เอกสารต้องอยู่ในปีบัญชีที่เปิดและงวดที่ไม่ปิด (ตรวจทุกเอกสาร แม้นำฝากที่ไม่ลงบัญชี)
func (Ledger) CheckDate(ctx context.Context, tx *sql.Tx, scope cheque.Scope, date string) error {
	_, err := gl.CheckSourceDateOpen(ctx, tx, glScope(scope), date)
	return err
}

// requestID เป็นรหัสคำขอคงที่ของขั้นหนึ่งของเอกสารหนึ่งรอบ — ส่งซ้ำแล้วเครื่องบัญชีคืนผลเดิม ไม่ลงซ้ำ
func requestID(company string, flag int, docNo string, revision int, action string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("cheque\x00%s\x00%d\x00%s\x00%d\x00%s", company, flag, docNo, revision, action)))
	return "chq-" + hex.EncodeToString(sum[:])[:40]
}

// thText คือข้อความไทยจาก languages.tsv แทนค่า {0}, {1}, … (ไทยเป็นต้นฉบับของใบสำคัญ)
func thText(key string, args ...string) string {
	text := language.Text(key, "th")
	for i, arg := range args {
		text = strings.ReplaceAll(text, "{"+strconv.Itoa(i)+"}", arg)
	}
	return text
}

// docLabel คือชื่อเมนูของเอกสาร เช่น "บันทึกเช็ครับผ่าน" (key = menu id ของ doctype)
func docLabel(dt doctype.DocType) string {
	return thText(strings.ReplaceAll(dt.Screen, "-", "_"))
}

// mappingLines สร้างรายการบัญชีจากรูปแบบลงบัญชี; รายการที่ยอดเป็นศูนย์ (ค่าธรรมเนียม 0) ข้ามไป
// บัญชีของรายการ bank_book = บัญชีของสมุดในหัวเอกสาร; ไม่มีสมุด → บัญชีสำรองเฉพาะยกเลิกเช็ครับในมือ (gltlsprocess.cpp:7824)
// กรณีอื่น Champ ข้ามบรรทัดจนใบสำคัญไม่สมดุล (:7806-7809) — ที่นี่ปฏิเสธทั้งเอกสารแทน เพื่อให้เดบิต = เครดิตเสมอ
func mappingLines(rules []gl.MappingRule, p cheque.Posting, description string) ([]gl.Line, error) {
	lines := make([]gl.Line, 0, len(rules))
	for _, rule := range rules {
		var amount decimal.Decimal
		switch rule.Source {
		case gl.MappingSourceChequeTotal:
			amount = p.Total
		case gl.MappingSourceChequeNet:
			amount = p.Net
		case gl.MappingSourceChequeFee:
			amount = p.Fee
		}
		if !amount.IsPositive() {
			continue
		}
		account := rule.AccountCode
		if rule.AccountFrom == gl.MappingAccountFromBankBook {
			switch {
			case p.BookAccount != "":
				account = p.BookAccount
			case p.BookFallback && rule.AccountCode != "":
			default:
				return nil, cheque.NewBookNoGLAccountError(p.BookCode)
			}
		}
		line := gl.Line{AccountCode: account, Description: description, Debit: "0", Credit: "0"}
		text := gl.Amount(amount.StringFixed(2))
		if rule.Side == gl.MappingSideDebit {
			line.Debit = text
		} else {
			line.Credit = text
		}
		lines = append(lines, line)
	}
	return lines, nil
}

// Post ลงบัญชีเอกสารหนึ่งรอบ: ไม่มีรูปแบบลงบัญชี → no_mapping (เอกสารยังบันทึกได้ แนวเดียวกับเอกสาร ERP)
func (Ledger) Post(ctx context.Context, tx *sql.Tx, scope cheque.Scope, p cheque.Posting) (cheque.Posted, error) {
	dt, ok := doctype.ByFlag(p.Flag)
	if !ok {
		return cheque.Posted{}, fmt.Errorf("cheque doc flag %d has no doctype", p.Flag)
	}
	s := glScope(scope)
	year, err := gl.CheckSourceDateOpen(ctx, tx, s, p.Date)
	if err != nil {
		return cheque.Posted{}, err
	}
	mapping, found, err := gl.MappingForSourceDoc(ctx, tx, s, dt.Key)
	if err != nil {
		return cheque.Posted{}, err
	}
	if !found {
		return cheque.Posted{Status: cheque.GLStatusNoMapping}, nil
	}
	description := thText("erp_gl_journal_desc", docLabel(dt), p.DocNo)
	lines, err := mappingLines(mapping.Rules, p, description)
	if err != nil {
		return cheque.Posted{}, err
	}
	if len(lines) == 0 {
		return cheque.Posted{Status: cheque.GLStatusNotRequired}, nil
	}
	journal := gl.Journal{
		Date:           p.Date,
		BookCode:       mapping.BookCode,
		FiscalYear:     year,
		Description:    description,
		Reference:      p.DocNo,
		BranchCode:     s.Branch,
		Kind:           "manual",
		SourceSystem:   gl.SourceSystemCheque,
		SourceRecordID: fmt.Sprintf("%s:%s:r%d", dt.Key, p.DocNo, p.Revision),
		Lines:          lines,
	}
	created, err := gl.ExecuteSourceTx(ctx, tx, s, gl.Command{
		Resource:  "journals",
		Action:    actionCreate,
		RequestID: requestID(s.Company, p.Flag, p.DocNo, p.Revision, actionCreate),
		Journal:   &journal,
	})
	if err != nil {
		return cheque.Posted{}, err
	}
	if _, err = gl.ExecuteSourceTx(ctx, tx, s, gl.Command{
		Resource:  "journals",
		Action:    actionPost,
		ID:        created.ID,
		Version:   created.Version,
		RequestID: requestID(s.Company, p.Flag, p.DocNo, p.Revision, actionPost),
	}); err != nil {
		return cheque.Posted{}, err
	}
	info, ok, err := gl.SourceJournal(ctx, tx, s, created.ID)
	if err != nil {
		return cheque.Posted{}, err
	}
	if !ok {
		return cheque.Posted{}, fmt.Errorf("journal %s of cheque document %s not found after posting", created.ID, p.DocNo)
	}
	return cheque.Posted{Status: cheque.GLStatusPosted, JournalID: created.ID, JournalDocNo: info.DocNo}, nil
}

// Reverse กลับรายการใบสำคัญของเอกสาร (ใบเดิมคงอยู่เป็นหลักฐาน); ผู้ใช้กลับรายการที่จอบัญชีไปก่อนแล้ว → ไม่กลับซ้ำ
func (Ledger) Reverse(ctx context.Context, tx *sql.Tx, scope cheque.Scope, r cheque.Reversal) error {
	s := glScope(scope)
	info, found, err := gl.SourceJournal(ctx, tx, s, r.JournalID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("journal %s of cheque document %s not found", r.JournalID, r.DocNo)
	}
	switch info.Status {
	case "posted":
	case "reversed":
		return nil
	default:
		return fmt.Errorf("journal %s of cheque document %s is %q, expected posted", r.JournalID, r.DocNo, info.Status)
	}
	date, err := gl.SourceReversalDate(ctx, tx, s, info.Date, r.Today)
	if err != nil {
		return err
	}
	reasonKey := "erp_gl_reverse_cancel"
	if r.Edit {
		reasonKey = "erp_gl_reverse_edit"
	}
	_, err = gl.ExecuteSourceTx(ctx, tx, s, gl.Command{
		Resource:  "journals",
		Action:    actionReverse,
		ID:        r.JournalID,
		Version:   info.Version,
		Date:      date,
		DocNo:     info.DocNo + "-R",
		Reason:    thText(reasonKey, r.DocNo),
		RequestID: requestID(s.Company, r.Flag, r.DocNo, r.Revision, actionReverse),
	})
	return err
}
