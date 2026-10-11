package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// JournalTemplateKind is the master kind of saved journal entry formats — Champ "รูปแบบรายการรายวัน" (recurring):
// the Recur button names an entry and the "..." picker fills the screen from it (glfrmjournal.cpp OnRecur / OnSelRecure,
// mybasedbformview.cpp BaseRecur). A template is a pattern, never a voucher: it keeps the book, reference, description,
// tax type and the account/detail/dimension of every line, but no document number, date, debit/credit amount or invoice
// figure (Champ clears all of them after loading, OnRecurClearSceen :2034). It is never posted and never reaches gl_lines.
const (
	JournalTemplateKind = "journal-templates"
	// JournalTemplateNameMaxRunes follows Champ BCTrans.RecurName VARCHAR(40) (Script/SQLSERVER_Script.sql:9404).
	JournalTemplateNameMaxRunes = 40
	JournalTemplateMaxLines     = 500
	journalTemplateCodePrefix   = "T"
)

// JournalTemplate is the pattern stored in Master.Template of a journal-templates master.
type JournalTemplate struct {
	BookCode    string `json:"bookcode,omitempty" bson:"bookcode,omitempty"`
	Reference   string `json:"reference,omitempty" bson:"reference,omitempty"`
	RefDate     string `json:"refdate,omitempty" bson:"refdate,omitempty"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
	// TaxType follows the details.vats tax_type of a journal: 0 none, 1 purchase, 2 sale.
	TaxType int                   `json:"taxtype,omitempty" bson:"taxtype,omitempty"`
	Lines   []JournalTemplateLine `json:"lines" bson:"lines"`
}

type JournalTemplateLine struct {
	AccountCode    string `json:"accountcode" bson:"accountcode"`
	Description    string `json:"description,omitempty" bson:"description,omitempty"`
	DepartmentCode string `json:"departmentcode,omitempty" bson:"departmentcode,omitempty"`
	ProjectCode    string `json:"projectcode,omitempty" bson:"projectcode,omitempty"`
	AllocateCode   string `json:"allocatecode,omitempty" bson:"allocatecode,omitempty"`
	JobCode        string `json:"jobcode,omitempty" bson:"jobcode,omitempty"`
	PartCode       string `json:"partcode,omitempty" bson:"partcode,omitempty"`
	SideCode       string `json:"sidecode,omitempty" bson:"sidecode,omitempty"`
	CashFlow       string `json:"cashflow,omitempty" bson:"cashflow,omitempty"`
}

// asLine is the template line as a voucher line, so the dimension rules of a voucher apply to it unchanged.
func (l JournalTemplateLine) asLine() Line {
	return Line{AccountCode: l.AccountCode, DepartmentCode: l.DepartmentCode, ProjectCode: l.ProjectCode,
		AllocateCode: l.AllocateCode, JobCode: l.JobCode, PartCode: l.PartCode, SideCode: l.SideCode}
}

// storedTemplateLines reads the lines of the stored template with this code ("" code = a new template: none), so editing a
// template that still names a code of a dimension switched off since is not refused for the value it already holds.
func storedTemplateLines(ctx context.Context, tx *sql.Tx, company, code string) ([]Line, error) {
	if code == "" {
		return nil, nil
	}
	var data []byte
	err := tx.QueryRowContext(ctx, `SELECT COALESCE(payload->'template'->'lines','[]'::jsonb) FROM gl_records WHERE company=$1 AND kind=$2 AND code=$3 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, company, JournalTemplateKind, code).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var stored []JournalTemplateLine
	if err = json.Unmarshal(data, &stored); err != nil {
		return nil, err
	}
	lines := make([]Line, len(stored))
	for i, line := range stored {
		lines[i] = line.asLine()
	}
	return lines, nil
}

// prepareJournalTemplate validates a journal-templates master and rewrites it to the stored shape: only name, code,
// active flag and the template body survive, so a client cannot park unrelated master fields in the record.
// code is the existing code on update; empty on create, which numbers the template (T0001, T0002 … never reused —
// deleted templates keep their record). The user only ever types the name, like Champ's Recur Name dialog.
func prepareJournalTemplate(ctx context.Context, tx *sql.Tx, company string, m *Master, code string) error {
	name := NormalizeCode(m.Name)
	if name == "" {
		return fieldError("journal_template_name_required", "name", "กรุณาระบุชื่อรูปแบบรายการ")
	}
	if strings.IndexFunc(name, func(r rune) bool { return unicode.Is(unicode.C, r) }) >= 0 {
		return fmt.Errorf("ชื่อรูปแบบรายการมีอักขระที่มองไม่เห็นปนอยู่ (มักติดมาจากการคัดลอก) กรุณาลบแล้วพิมพ์ใหม่")
	}
	if n := utf8.RuneCountInString(name); n > JournalTemplateNameMaxRunes {
		return fieldError("journal_template_name_too_long", "name", fmt.Sprintf("ชื่อรูปแบบรายการยาวได้ไม่เกิน %d ตัวอักษร (ตอนนี้ %d ตัว) กรุณาย่อให้สั้นลง", JournalTemplateNameMaxRunes, n))
	}
	if m.Template == nil {
		return fmt.Errorf("ไม่พบข้อมูลรูปแบบรายการ")
	}
	body, err := cleanJournalTemplate(*m.Template)
	if err != nil {
		return err
	}
	if body.BookCode != "" {
		if _, found, err := loadJournalBook(ctx, tx, company, body.BookCode); err != nil {
			return err
		} else if !found {
			return journalBookNotFound(body.BookCode)
		}
	}
	lines := make([]Line, len(body.Lines))
	for i, line := range body.Lines {
		lines[i] = line.asLine()
	}
	accounts, err := loadLineAccounts(ctx, tx, company, lines)
	if err != nil {
		return err
	}
	for i, line := range lines {
		// The same rule as a voucher line (Journal.Validate): a template must not point at an account the entry could not use.
		if a, ok := accounts[line.AccountCode]; !ok || !a.IsActive || a.IsDeleted || !a.AllowPosting {
			return fmt.Errorf("บรรทัด %d: กรุณาเลือกบัญชีที่เปิดใช้งานและลงรายการได้", i+1)
		}
	}
	if lineNamesDimension(lines) {
		oldLines, err := storedTemplateLines(ctx, tx, company, code)
		if err != nil {
			return err
		}
		if err = checkJournalDimensions(ctx, tx, company, lines, accounts, oldLines); err != nil {
			return err
		}
	}
	if code == "" {
		if code, err = nextJournalTemplateCode(ctx, tx, company); err != nil {
			return err
		}
	}
	*m = Master{Kind: JournalTemplateKind, Code: code, Name: name, IsActive: true, Template: &body}
	return nil
}

// cleanJournalTemplate checks the pattern's own fields (no database lookups) and returns the trimmed copy.
func cleanJournalTemplate(in JournalTemplate) (JournalTemplate, error) {
	out := JournalTemplate{
		BookCode: NormalizeCode(in.BookCode), Reference: strings.TrimSpace(in.Reference), RefDate: strings.TrimSpace(in.RefDate),
		Description: strings.TrimSpace(in.Description), TaxType: in.TaxType,
	}
	if out.BookCode != "" {
		if err := checkCode(out.BookCode, "bookcode", "สมุดรายวัน", BookCodeMaxRunes); err != nil {
			return out, err
		}
	}
	if utf8.RuneCountInString(out.Reference) > JournalReferenceMaxRunes {
		return out, fieldError("journal_reference_too_long", "reference", fmt.Sprintf("เลขที่อ้างอิงต้องไม่เกิน %d ตัวอักษร — กรุณาใส่เฉพาะเลขที่เอกสารอ้างอิง", JournalReferenceMaxRunes))
	}
	if out.RefDate != "" && !validDate(out.RefDate) {
		return out, fieldError("journal_refdate_invalid", "refdate", "วันที่เอกสารอ้างอิงไม่ถูกต้อง กรุณาเลือกวันที่ให้ครบ วัน เดือน ปี")
	}
	if utf8.RuneCountInString(out.Description) > JournalDescriptionMaxRunes {
		return out, fieldError("journal_description_too_long", "description", fmt.Sprintf("คำอธิบายรายการต้องไม่เกิน %d ตัวอักษร — กรุณาย่อให้สั้นลง แล้วใส่รายละเอียดที่เหลือในคำอธิบายบรรทัด", JournalDescriptionMaxRunes))
	}
	if out.TaxType < 0 || out.TaxType > 2 {
		return out, fmt.Errorf("ประเภทภาษีของรูปแบบรายการไม่ถูกต้อง")
	}
	if len(in.Lines) < 1 || len(in.Lines) > JournalTemplateMaxLines {
		return out, fmt.Errorf("รูปแบบรายการต้องมี 1–%d บรรทัด", JournalTemplateMaxLines)
	}
	out.Lines = make([]JournalTemplateLine, len(in.Lines))
	for i, line := range in.Lines {
		field := fmt.Sprintf("lines[%d].accountcode", i)
		clean := JournalTemplateLine{
			AccountCode: NormalizeCode(line.AccountCode), Description: strings.TrimSpace(line.Description),
			DepartmentCode: NormalizeCode(line.DepartmentCode), ProjectCode: NormalizeCode(line.ProjectCode),
			AllocateCode: NormalizeCode(line.AllocateCode), JobCode: NormalizeCode(line.JobCode), PartCode: NormalizeCode(line.PartCode), SideCode: NormalizeCode(line.SideCode),
			CashFlow: line.CashFlow,
		}
		if err := checkCode(clean.AccountCode, field, "รหัสบัญชี", AccountCodeMaxRunes); err != nil {
			return out, err
		}
		if utf8.RuneCountInString(clean.Description) > JournalLineDescriptionMaxRunes {
			return out, fieldError("journal_line_description_too_long", fmt.Sprintf("lines[%d].description", i), fmt.Sprintf("บรรทัดที่ %d: คำอธิบายบรรทัดต้องไม่เกิน %d ตัวอักษร — กรุณาย่อให้สั้นลง", i+1, JournalLineDescriptionMaxRunes))
		}
		asLine := clean.asLine()
		for _, d := range dimensionRegistry {
			if utf8.RuneCountInString(asLine.dimensionCode(d.kind)) > codeMaxRunes {
				return out, fmt.Errorf("บรรทัด %d: รหัส%sยาวเกินกำหนด", i+1, d.label)
			}
		}
		if clean.CashFlow != "" && !contains([]string{"operating", "investing", "financing"}, clean.CashFlow) {
			return out, fmt.Errorf("ประเภทกระแสเงินสดไม่ถูกต้อง")
		}
		out.Lines[i] = clean
	}
	return out, nil
}

// nextJournalTemplateCode numbers a new template. Execute holds the company lock for the whole command, so two
// creates never read the same maximum; deleted templates still count, so a code is never reused.
func nextJournalTemplateCode(ctx context.Context, tx *sql.Tx, company string) (string, error) {
	var next int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(substring(code from '[0-9]+$')::bigint), 0) + 1 FROM gl_records WHERE company=$1 AND kind=$2`, company, JournalTemplateKind).Scan(&next); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%04d", journalTemplateCodePrefix, next), nil
}
