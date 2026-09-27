package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ชุดงบการเงินในคำขอเดียว (GET /gl/v2/reports/statement-set — จอพิมพ์ชุดงบการเงิน frontend/src/app/gl/gl-statement-set.tsx).
// ทุกงบคำนวณด้วยเส้นทางเดียวกับ reports/statement (reportContext.statement + finish) ใน transaction อ่านอย่างเดียวแบบ
// REPEATABLE READ ธุรกรรมเดียว — ทุกงบและหมายเหตุเห็นสมุดบัญชีชุดเดียวกัน (snapshot เดียว ตัวเลขข้ามงบจึงกระทบกันได้).
// งบหนึ่งคำนวณไม่สำเร็จ = งบนั้นได้ Err ส่วนงบอื่นยังคืนผล: แต่ละงบอยู่ใน SAVEPOINT ของตัวเอง คำสั่ง SQL ที่ล้มจึงไม่ทำให้
// transaction ทั้งชุดใช้ต่อไม่ได้ (SAVEPOINT ไม่เปลี่ยน snapshot ของ REPEATABLE READ).

// statementSetOrder ลำดับพิมพ์ตามแบบ 2 (บริษัทจำกัด) ประกาศกรมพัฒนาธุรกิจการค้า เรื่อง กำหนดรายการย่อที่ต้องมีในงบการเงิน พ.ศ. 2566
// (https://www.dbd.go.th/storage/law/4941272b-c21b-4d78-b1ca-3e63f83252e7.pdf): งบฐานะการเงิน (หน้า 2-1) → งบกำไรขาดทุน (2-4)
// → งบการเปลี่ยนแปลงส่วนของผู้ถือหุ้น (2-20) → งบกระแสเงินสด (2-23); ชนิดอื่น (งบต้นทุนผลิต / กำหนดเอง) ต่อท้ายเรียงตามรหัส
// และหมายเหตุประกอบงบการเงิน (2-30) พิมพ์ท้ายชุดเสมอ. ต้องตรงกับลำดับของ statementSetRank ใน frontend/src/lib/general-ledger.ts
// (frontend/src/app/gl/gl-statement-set.test.ts อ่านบรรทัดนี้ไปเทียบ — เปลี่ยนรูปแบบบรรทัดต้องแก้ test ด้วย)
var statementSetOrder = []string{"balance_sheet", "pnl", "equity", "cash_flow"}

// statementSetMaxTemplates จำนวนรูปแบบงบสูงสุดต่อคำขอ — กันคำขอเดียวคำนวณงบจำนวนมากจนเกิน timeout 30 วินาทีของ API
const statementSetMaxTemplates = 20

// StatementSetQuery คำขอชุดงบ: Report = ตัวกรองเดียวกับ reports/statement (ปีบัญชี ช่วงวันที่ สาขา แผนก โครงการ สมุด)
type StatementSetQuery struct {
	Report ReportQuery
	// Templates nil = ชุดเริ่มต้น: รูปแบบแรก (ตามรหัส) ที่เปิดใช้ของงบแต่ละชนิดใน statementSetOrder;
	// ว่างแต่ไม่ nil = ไม่พิมพ์งบ (หมายเหตุอย่างเดียว)
	Templates []string
	Notes     bool
}

// StatementSetSection งบหนึ่งงบในชุด: ชื่อ/ชนิด/คอลัมน์หมายเหตุ/ทศนิยมมาจากรูปแบบงบใน snapshot เดียวกับตัวเลข
type StatementSetSection struct {
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	StatementType  string  `json:"statementtype"`
	ShowNoteColumn bool    `json:"shownotecolumn"`
	Scale          int     `json:"scale"`
	Report         *Report `json:"report,omitempty"`
	Error          string  `json:"error,omitempty"`
	// Err เหตุที่คำนวณงบนี้ไม่สำเร็จ — ชั้น HTTP แปลงเป็นข้อความตามภาษาที่ผู้ใช้เลือกลง Error (httpapi/statement_set.go)
	Err error `json:"-"`
}

// StatementSet ผลของ reports/statement-set: งบเรียงตามแบบ 2 เสมอ (ไม่ขึ้นกับลำดับที่ส่งมา) และหมายเหตุของปีบัญชีที่ออกงบ
type StatementSet struct {
	Sequence   int64                 `json:"sequence"`
	FiscalYear string                `json:"fiscalyear"`
	From       string                `json:"from"`
	To         string                `json:"to"`
	Sections   []StatementSetSection `json:"sections"`
	Notes      []StatementNote       `json:"notes"`
}

// ParseStatementSetTemplates แยกรหัสรูปแบบงบจากพารามิเตอร์ templates (คั่นด้วยจุลภาค — รหัสใช้จุลภาคไม่ได้ ตาม checkCode):
// ตัดช่องว่างหัวท้ายเหมือน reports/statement, ข้ามช่องว่างและรหัสซ้ำ, ไม่เกิน statementSetMaxTemplates รูปแบบ
func ParseStatementSetTemplates(raw string) ([]string, error) {
	codes, seen := []string{}, map[string]bool{}
	for _, part := range strings.Split(raw, ",") {
		code := strings.TrimSpace(part)
		if code == "" || seen[code] {
			continue
		}
		seen[code] = true
		codes = append(codes, code)
	}
	if len(codes) > statementSetMaxTemplates {
		err := fieldError("statement_set_too_many", "templates", fmt.Sprintf("พิมพ์ชุดงบการเงินได้ครั้งละไม่เกิน %d รูปแบบ (ตอนนี้เลือก %d รูปแบบ) กรุณาแบ่งพิมพ์เป็นหลายชุด", statementSetMaxTemplates, len(codes)))
		err.Args = []string{strconv.Itoa(statementSetMaxTemplates), strconv.Itoa(len(codes))}
		return nil, err
	}
	return codes, nil
}

// ParseStatementSetNotes พารามิเตอร์ notes: ไม่ส่ง = รวมหมายเหตุ; รับเฉพาะ true / false
func ParseStatementSetNotes(raw string) (bool, error) {
	switch raw {
	case "", "true":
		return true, nil
	case "false":
		return false, nil
	}
	return false, fieldError("statement_set_notes_invalid", "notes", "ค่า notes ต้องเป็น true หรือ false")
}

func statementSetRank(statementType string) int {
	for i, item := range statementSetOrder {
		if item == statementType {
			return i
		}
	}
	return len(statementSetOrder)
}

// sortStatementSet เรียงตามแบบ 2 แล้วตามรหัส (เทียบแบบ byte เหมือน a.code < b.code ของ frontend)
func sortStatementSet(templates []Master) {
	sort.SliceStable(templates, func(i, j int) bool {
		ri, rj := statementSetRank(templates[i].StatementType), statementSetRank(templates[j].StatementType)
		if ri != rj {
			return ri < rj
		}
		return templates[i].Code < templates[j].Code
	})
}

// selectStatementSetTemplates เลือกรูปแบบงบของชุดจากรูปแบบที่ยังไม่ถูกลบทั้งหมด (all): codes nil = ชุดเริ่มต้น;
// รหัสที่ส่งมาต้องมีอยู่ (ไม่ถูกลบ) และเปิดใช้งาน ไม่งั้นทั้งคำขอไม่ผ่าน พร้อมบอกรหัสที่ผิด
func selectStatementSetTemplates(all []Master, codes []string) ([]Master, error) {
	chosen := []Master{}
	if codes == nil {
		active := []Master{}
		for _, template := range all {
			if template.IsActive {
				active = append(active, template)
			}
		}
		sortStatementSet(active)
		taken := map[string]bool{}
		for _, template := range active {
			if statementSetRank(template.StatementType) < len(statementSetOrder) && !taken[template.StatementType] {
				taken[template.StatementType] = true
				chosen = append(chosen, template)
			}
		}
		return chosen, nil
	}
	byCode := map[string]Master{}
	for _, template := range all {
		byCode[template.Code] = template
	}
	for _, code := range codes {
		template, ok := byCode[code]
		if !ok {
			err := fieldError("statement_set_template_not_found", "templates", fmt.Sprintf("ไม่พบรูปแบบงบการเงินรหัส %s หรือถูกลบไปแล้ว กรุณาปิดแล้วเปิดหน้าพิมพ์ชุดงบใหม่", code))
			err.Args = []string{code}
			return nil, err
		}
		if !template.IsActive {
			err := fieldError("statement_set_template_inactive", "templates", fmt.Sprintf("รูปแบบงบการเงินรหัส %s ปิดใช้งานอยู่ กรุณาเปิดใช้งานก่อน หรือยกเลิกการเลือก", code))
			err.Args = []string{code}
			return nil, err
		}
		chosen = append(chosen, template)
	}
	sortStatementSet(chosen)
	return chosen, nil
}

// StatementSet คำนวณชุดงบการเงินและอ่านหมายเหตุของปีบัญชีใน transaction เดียว (อธิบายต้นไฟล์)
func (p *Postgres) StatementSet(ctx context.Context, scope Scope, q StatementSetQuery) (StatementSet, error) {
	db, err := p.database(ctx, scope.Holding)
	if err != nil {
		return StatementSet{}, err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return StatementSet{}, err
	}
	defer tx.Rollback()
	rc, err := newReportContext(ctx, tx, scope, q.Report)
	if err != nil {
		return StatementSet{}, err
	}
	all, err := rc.statementSetTemplates(ctx)
	if err != nil {
		return StatementSet{}, err
	}
	chosen, err := selectStatementSetTemplates(all, q.Templates)
	if err != nil {
		return StatementSet{}, err
	}
	if len(chosen) == 0 && !q.Notes {
		return StatementSet{}, fieldError("statement_set_empty", "templates", "เลือกงบหรือหมายเหตุอย่างน้อย 1 รายการ")
	}
	set := StatementSet{FiscalYear: rc.fiscal.Code, From: rc.query.From, To: rc.query.To, Sections: make([]StatementSetSection, 0, len(chosen)), Notes: []StatementNote{}}
	for _, template := range chosen {
		section := StatementSetSection{Code: template.Code, Name: template.Name, StatementType: template.StatementType, ShowNoteColumn: statementShowsNoteColumn(template), Scale: int(statementScale(template))}
		if err = rc.statementSection(ctx, &section); err != nil {
			return StatementSet{}, err
		}
		set.Sections = append(set.Sections, section)
	}
	if q.Notes {
		if set.Notes, _, err = rc.fiscalYearNotes(ctx); err != nil {
			return StatementSet{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return StatementSet{}, err
	}
	return set, nil
}

// statementSetTemplates รูปแบบงบที่ยังไม่ถูกลบของบริษัท — ไม่อ่านแถว/คอลัมน์ของงบ (statement() อ่านเต็มเองตอนคำนวณ ใน snapshot เดียวกัน)
func (r reportContext) statementSetTemplates(ctx context.Context) ([]Master, error) {
	rows, err := r.tx.QueryContext(ctx, `SELECT code, payload - 'rows' - 'columns' FROM gl_records WHERE company=$1 AND kind='statement-templates' AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, r.scope.Company)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	templates := []Master{}
	for rows.Next() {
		var code string
		var payload []byte
		if err = rows.Scan(&code, &payload); err != nil {
			return nil, err
		}
		var template Master
		if err = json.Unmarshal(payload, &template); err != nil {
			return nil, err
		}
		// รหัสที่ statement() ค้นคือคอลัมน์ code — ใช้ค่าเดียวกันเสมอ
		template.Code = code
		templates = append(templates, template)
	}
	return templates, rows.Err()
}

// statementSection คำนวณงบหนึ่งงบของชุดผ่านเส้นทางเดียวกับ reports/statement&template=code ลง section.Report;
// คำนวณไม่สำเร็จ = section.Err (ย้อนกลับถึง SAVEPOINT แล้วงบถัดไปคำนวณต่อได้); คืน error = transaction ใช้ต่อไม่ได้ ทั้งชุดไม่สำเร็จ
func (r reportContext) statementSection(ctx context.Context, section *StatementSetSection) error {
	if _, err := r.tx.ExecContext(ctx, `SAVEPOINT gl_statement_set_section`); err != nil {
		return err
	}
	r.query.Template = section.Code
	report, err := r.statement(ctx)
	if err != nil {
		section.Err = err
		if _, err = r.tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT gl_statement_set_section`); err != nil {
			return err
		}
	} else {
		report = r.finish(report)
		section.Report = &report
	}
	_, err = r.tx.ExecContext(ctx, `RELEASE SAVEPOINT gl_statement_set_section`)
	return err
}
