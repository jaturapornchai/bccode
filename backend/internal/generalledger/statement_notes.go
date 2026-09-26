package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

// หมายเหตุประกอบงบการเงิน (ประกาศกรมพัฒนาธุรกิจการค้า เรื่อง กำหนดรายการย่อที่ต้องมีในงบการเงิน พ.ศ. 2566 แบบ 2 ข้อ 5
// หน้า 2-30..2-31; TFRS for NPAEs ย่อหน้า 4.1 — docs/kms/21 DBD-FS-2566-N5). เก็บเป็นข้อมูลหลักชนิด statement-notes
// หนึ่งรายการต่อบริษัทต่อปีบัญชี: code = รหัสปีบัญชี, notes = ข้อความของผู้ใช้ทีละเลขที่หมายเหตุ (ระบบไม่แต่งเนื้อหาเอง).
// งบการเงินตรวจว่าเลขหมายเหตุที่บรรทัดของรูปแบบงบอ้างถึงมีอยู่จริงในหมายเหตุของปีที่ออกงบ (statementNoteWarnings)

const (
	statementNotesMax          = 100
	statementNoteIDMaxRunes    = 64
	statementNoteNoMaxRunes    = 10
	statementNoteTitleMaxRunes = 200
	statementNoteBodyMaxRunes  = 20000
)

// StatementNote หมายเหตุหนึ่งข้อ: เลขที่ (พิมพ์หน้าหัวข้อและอ้างจากคอลัมน์หมายเหตุของงบ), หัวข้อ, เนื้อหา (ขึ้นบรรทัดใหม่ได้)
type StatementNote struct {
	ID     string `json:"id" bson:"id"`
	NoteNo string `json:"noteno" bson:"noteno"`
	Title  string `json:"title" bson:"title"`
	Body   string `json:"body" bson:"body"`
}

// normalizeStatementNotes ตัดช่องว่างหัวท้ายของเลขที่และหัวข้อ ใส่ id ให้หมายเหตุที่ยังไม่มี แล้วตรวจขอบเขต (นับตัวอักษร ไม่ใช่ byte)
func normalizeStatementNotes(notes []StatementNote) ([]StatementNote, error) {
	if len(notes) > statementNotesMax {
		return nil, fieldError("statement_notes_too_many", "notes", fmt.Sprintf("หมายเหตุประกอบงบการเงินมีได้ไม่เกิน %d หัวข้อต่อปี (ตอนนี้ %d หัวข้อ) กรุณารวมหัวข้อย่อยไว้ในเนื้อหาของหัวข้อหลัก", statementNotesMax, len(notes)))
	}
	out := make([]StatementNote, 0, len(notes))
	seen, seenID := map[string]bool{}, map[string]bool{}
	for i, note := range notes {
		note.ID = strings.TrimSpace(note.ID)
		// id มาจาก client (รวม API token) ห้ามเชื่อ: ว่าง ซ้ำ หรือยาวผิดปกติ = ออก id ใหม่ ไม่งั้นจอแก้ไขที่ผูกหมายเหตุด้วย id
		// จะแก้/ลบหลายข้อพร้อมกัน
		if note.ID == "" || seenID[note.ID] || utf8.RuneCountInString(note.ID) > statementNoteIDMaxRunes {
			note.ID = uuid.NewString()
		}
		seenID[note.ID] = true
		// จุดท้ายเลขที่ตัดออก (พิมพ์หมายเหตุเติม ". " ต่อท้ายเลขที่เอง); เลขไทยคงตามที่พิมพ์ แต่เทียบซ้ำด้วยเลขอารบิก ("๕" ซ้ำกับ "5")
		note.NoteNo = strings.TrimSpace(strings.TrimRight(strings.TrimSpace(note.NoteNo), "."))
		note.Title = strings.TrimSpace(note.Title)
		field := fmt.Sprintf("notes[%d].", i)
		if note.NoteNo == "" {
			return nil, fieldError("statement_note_no_required", field+"noteno", fmt.Sprintf("กรุณาระบุเลขที่หมายเหตุของหัวข้อที่ %d", i+1))
		}
		if n := utf8.RuneCountInString(note.NoteNo); n > statementNoteNoMaxRunes {
			return nil, fieldError("statement_note_no_too_long", field+"noteno", fmt.Sprintf("เลขที่หมายเหตุของหัวข้อที่ %d ยาวได้ไม่เกิน %d ตัวอักษร (ตอนนี้ %d ตัว)", i+1, statementNoteNoMaxRunes, n))
		}
		key := canonicalStatementNoteNo(note.NoteNo)
		if seen[key] {
			return nil, fieldError("statement_note_no_duplicate", field+"noteno", fmt.Sprintf("เลขที่หมายเหตุ %s ซ้ำกัน กรุณาใช้เลขที่ไม่ซ้ำกันในปีเดียวกัน", note.NoteNo))
		}
		seen[key] = true
		if note.Title == "" {
			return nil, fieldError("statement_note_title_required", field+"title", fmt.Sprintf("กรุณาระบุหัวข้อของหมายเหตุ %s", note.NoteNo))
		}
		if n := utf8.RuneCountInString(note.Title); n > statementNoteTitleMaxRunes {
			return nil, fieldError("statement_note_title_too_long", field+"title", fmt.Sprintf("หัวข้อของหมายเหตุ %s ยาวได้ไม่เกิน %d ตัวอักษร (ตอนนี้ %d ตัว)", note.NoteNo, statementNoteTitleMaxRunes, n))
		}
		if n := utf8.RuneCountInString(note.Body); n > statementNoteBodyMaxRunes {
			return nil, fieldError("statement_note_body_too_long", field+"body", fmt.Sprintf("เนื้อหาของหมายเหตุ %s ยาวได้ไม่เกิน %d ตัวอักษร (ตอนนี้ %d ตัว) กรุณาแยกเป็นหมายเหตุหลายข้อ", note.NoteNo, statementNoteBodyMaxRunes, n))
		}
		out = append(out, note)
	}
	return out, nil
}

// validateStatementNotesMaster ตรวจรายการหมายเหตุของปีบัญชีก่อนสร้าง/แก้ไข: ปีบัญชี (code) ต้องมีอยู่ในบริษัท และหมายเหตุทุกข้อถูกต้อง
func validateStatementNotesMaster(ctx context.Context, tx *sql.Tx, company string, m *Master) error {
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_records WHERE company=$1 AND kind='fiscal-years' AND code=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false))`, company, m.Code).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return fieldError("statement_notes_fiscal_year_not_found", "code", fmt.Sprintf("ไม่พบปีบัญชี %s กรุณาเลือกปีบัญชีจากรายการ", m.Code))
	}
	notes, err := normalizeStatementNotes(m.Notes)
	if err != nil {
		return err
	}
	m.Notes = notes
	return nil
}

// เลขที่หมายเหตุในช่องหมายเหตุของบรรทัดงบอ้างได้หลายข้อ เช่น "4, 5" / "4 และ 5" / "5.1"
// \s ของ Go เป็นช่องว่าง ASCII เท่านั้น จึงเพิ่ม \p{Z} ให้ช่องว่างไม่ตัดบรรทัด (NBSP) ที่ติดมาจาก Word/Excel แยกเลขที่ได้ด้วย
var statementNoteRefSeparator = regexp.MustCompile(`[,;\s\p{Z}]+|และ`)

// canonicalStatementNoteNo เลขที่หมายเหตุสำหรับเทียบกัน: ตัดช่องว่าง/จุดท้าย และเลขไทย ๐-๙ เป็น 0-9 ("๕." = "5")
func canonicalStatementNoteNo(no string) string {
	no = strings.Map(func(r rune) rune {
		if r >= '๐' && r <= '๙' {
			return '0' + (r - '๐')
		}
		return r
	}, strings.TrimSpace(no))
	return strings.TrimSpace(strings.TrimRight(no, "."))
}

// statementReportNoteNos เลขที่หมายเหตุที่บรรทัดงบที่พิมพ์จริงอ้างถึง (หลังซ่อนแถวศูนย์) แยกทีละเลขที่ ไม่ซ้ำ เรียงตามที่พบ
func statementReportNoteNos(rows []map[string]string) []string {
	referenced, seen := []string{}, map[string]bool{}
	for _, row := range rows {
		for _, part := range statementNoteRefSeparator.Split(row["noteno"], -1) {
			part = canonicalStatementNoteNo(part)
			if part != "" && !seen[part] {
				seen[part] = true
				referenced = append(referenced, part)
			}
		}
	}
	return referenced
}

// statementPrintedNoteNos เลขที่หมายเหตุที่งบพิมพ์จริง: ไม่ติ๊ก "แสดงคอลัมน์หมายเหตุประกอบงบ" = ไม่พิมพ์เลขที่เลย จึงไม่ต้องตรวจ
// (ไม่ระบุ = แสดงตามค่าเริ่มของจอ)
func statementPrintedNoteNos(template Master, rows []map[string]string) []string {
	if style := template.GlobalStyle; style != nil && style.ShowNoteColumn != nil && !*style.ShowNoteColumn {
		return nil
	}
	return statementReportNoteNos(rows)
}

// statementNoteWarnings เตือนเมื่อเลขที่หมายเหตุที่บรรทัดงบอ้างถึงยังไม่มีในหมายเหตุประกอบงบการเงินของปีที่ออกงบ (ปีปัจจุบันเท่านั้น;
// รายการที่ลบแล้วนับว่าไม่มี) — ไม่มีบรรทัดใดอ้างหมายเหตุ = ไม่ต้องอ่านหมายเหตุเลย. เลขที่ย่อย "5.1" นับว่ามีเมื่อมีหมายเหตุ "5.1"
// หรือหมายเหตุหลัก "5" (หัวข้อย่อยเขียนรวมในเนื้อหาของหัวข้อหลักได้)
func (r reportContext) statementNoteWarnings(ctx context.Context, referenced []string) ([]string, error) {
	if len(referenced) == 0 {
		return nil, nil
	}
	year := r.fiscal.Code
	var payload []byte
	err := r.tx.QueryRowContext(ctx, `SELECT payload FROM gl_records WHERE company=$1 AND kind='statement-notes' AND code=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, r.scope.Company, year).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return []string{"ยังไม่มีหมายเหตุประกอบงบการเงินของปี " + year + " แต่มีบรรทัดอ้างหมายเหตุ"}, nil
	}
	if err != nil {
		return nil, err
	}
	var record Master
	if err = json.Unmarshal(payload, &record); err != nil {
		return nil, err
	}
	present := map[string]bool{}
	for _, note := range record.Notes {
		present[canonicalStatementNoteNo(note.NoteNo)] = true
	}
	return missingStatementNotes(referenced, present, year), nil
}

func missingStatementNotes(referenced []string, present map[string]bool, year string) []string {
	warnings := []string{}
	for _, no := range referenced {
		if statementNotePresent(no, present) {
			continue
		}
		warnings = append(warnings, "หมายเหตุ "+no+" ที่อ้างในงบยังไม่มีในหมายเหตุประกอบงบการเงินปี "+year)
	}
	return warnings
}

// เลขที่ย่อยนับว่ามีเมื่อมีเลขที่นั้นเอง หรือหมายเหตุแม่ชั้นใดก็ได้ ("5.1.2" → "5.1" → "5")
func statementNotePresent(no string, present map[string]bool) bool {
	for {
		if present[no] {
			return true
		}
		i := strings.LastIndex(no, ".")
		if i < 0 {
			return false
		}
		no = no[:i]
	}
}
