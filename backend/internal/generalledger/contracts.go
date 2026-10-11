package generalledger

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

type Change struct {
	Kind    string `json:"kind" bson:"kind"`
	ID      string `json:"id" bson:"id"`
	Code    string `json:"code" bson:"code"`
	Payload string `json:"payload" bson:"payload"`
}

// One command, its audit and its projection intent are committed together.
// Payloads are immutable decimal-string JSON; only delivery metadata may change.
type Event struct {
	ID           string    `json:"id" bson:"_id"`
	HoldingCode  string    `json:"holdingcode" bson:"holdingcode"`
	BusinessCode string    `json:"businesscode" bson:"businesscode"`
	Sequence     int64     `json:"sequence" bson:"sequence"`
	RequestID    string    `json:"requestid" bson:"requestid"`
	RequestHash  string    `json:"requesthash" bson:"requesthash"`
	Action       string    `json:"action" bson:"action"`
	Actor        string    `json:"actor" bson:"actor"`
	Reason       string    `json:"reason" bson:"reason"`
	OccurredAt   time.Time `json:"occurredat" bson:"occurredat"`
	Changes      []Change  `json:"changes" bson:"changes"`
	Delivered    bool      `json:"delivered" bson:"delivered"`
}

func (Event) CollectionName() string { return "gl_events" }

type Page struct {
	Sequence int64             `json:"sequence"`
	Items    []json.RawMessage `json:"items"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
}

type ReportQuery struct {
	From           string
	To             string
	FiscalYear     string
	AccountCode    string
	AccountFrom    string // Champ "จากรหัสบัญชี" (รวมค่านี้) — ว่าง = ไม่จำกัดต้นช่วง
	AccountTo      string // Champ "ถึงรหัสบัญชี" (รวมค่านี้) — ว่าง = ไม่จำกัดท้ายช่วง
	BranchCode     string
	DepartmentCode string
	ProjectCode    string
	// AllocateCode / JobCode / PartCode / SideCode - ตัวกรองทะเบียนมิติที่เหลือ (Champ GLRepTrialBalanceDlg.cpp:160-190); คั่นจุลภาค = หลายรหัส
	// ใช้ได้เมื่อบริษัทเปิดสวิตช์ของมิตินั้น (หน้าจอแสดงตัวกรองแต่ disable เมื่อปิด) — backend ไม่บังคับ เพราะข้อมูลเก่ายังอ่านได้
	AllocateCode string
	JobCode      string
	PartCode     string
	SideCode     string
	BookCode     string
	BudgetCode     string
	BudgetCodeTo   string // budgetdetail: ถึงรหัสงบประมาณ — ว่าง = เฉพาะรหัสใน BudgetCode (Champ 5525 GLRepBudgetDlg.cpp:244-251)
	Template       string // statement: รหัสรูปแบบงบการเงิน
	// ตัวเลือกรายงานแบบ Champ (docs/kms/champ/09-gl.md §5.3-5.6) — รายงานที่ไม่รู้จักตัวเลือกนั้นไม่สนใจค่า
	HideOpening   bool // ledger: ไม่แสดงยอดยกมา — ยอดคงเหลือสะสมเริ่มจากรายการในช่วง (Champ 5534)
	HideZero      bool // trialbalance/workingpaper: ไม่แสดงรายการผังบัญชีที่ยอดทุกช่องเป็นศูนย์ (Champ 5506/5499)
	ControlTotals bool // trialbalance/workingpaper/ledger summary: เพิ่มแถวผลรวมบัญชีคุมตามต้นไม้บัญชีแม่
	Summary       bool // gljournal: แถวละเอกสาร (Champ 5533); ledger: ยอดสะสมสุทธิ ณ วันที่ต่อบัญชี (Champ 5535)
	Gross         bool // trialbalance: ยอดรวมเดบิต/เครดิตของช่วงแบบไม่หักกลบ 2 คอลัมน์ (Champ 5507 งบทดลอง 2)
	DayTotals     bool // ledger: รวมรายวัน — แทรกแถวผลรวมเดบิต/เครดิตหลังรายการสุดท้ายของแต่ละวัน (Champ 5534 m_iSumOfDay); gljournal 5532: เรียงตามวันที่เท่านั้น
	// gljournal (Champ 5532/5533 GLRepTransDlg.cpp:167-246): เลขที่เอกสาร จาก–ถึง (จากอย่างเดียว = เอกสารนั้นใบเดียว) และเรียงตาม เลขที่เอกสาร (ค่าเริ่มต้น) / วันที่เอกสาร
	DocNoFrom string
	DocNoTo   string
	OrderBy   string // "docno" | "date"
	// trialbalance/workingpaper/ledger summary (Champ 5506/5499/5535 "แยกแผนก"/"แยกโครงการ" + "มียอดรวม"): หนึ่งแถวต่อบัญชี × รหัสมิติ;
	// DimTotals = แถวรวมของแต่ละบัญชีที่ตั้งแยกแผนก (มีผลเมื่อ ByDepartment เท่านั้น เหมือน Champ m_iSumDepart)
	ByDepartment bool
	ByProject    bool
	ByAllocate   bool
	ByJob        bool
	ByPart       bool
	BySide       bool
	DimTotals    bool
	// PostStatus = Champ "สถานะ" combo of 5532/5533/5534/5535/5507 (GLRepTransDlg.cpp:82-88,206-226; GLRepGL2Dlg.cpp:216-222,425-436):
	// "" | "posted" = posted+reversed journals (gl_lines), "unposted" = draft journals, "all" = both
	PostStatus string
	Page       int
	Limit      int
	// RowCap and Timeout are set only by in-process callers (report registry print/export), never from HTTP input:
	// RowCap > 0 returns page 1 with up to RowCap rows in this one snapshot instead of pageBounds; Timeout > 0 sets
	// statement_timeout for the report's read-only transaction
	RowCap  int
	Timeout time.Duration
}

type ReportColumn struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Amount bool   `json:"amount,omitempty"`
	// Group = หัวคอลัมน์ชั้นบนแบบ Champ (เช่น "ยอดยกมา" คร่อม เดบิต | เครดิต) — คอลัมน์ติดกันที่ Group เดียวกันรวมเป็นแถบเดียว
	Group string `json:"group,omitempty"`
}
type Report struct {
	Sequence  int64               `json:"sequence"`
	Columns   []ReportColumn      `json:"columns"`
	Rows      []map[string]string `json:"rows"`
	Totals    map[string]string   `json:"totals"`
	TotalRows int64               `json:"totalrows"`
	Warnings  []string            `json:"warnings"`
	AsOf      string              `json:"asof"`
	Periods   []ReportPeriod      `json:"periods,omitempty"`
	Checks    []ReportCheck       `json:"checks,omitempty"`
	// Unassigned: บัญชีที่มียอดแต่ไม่อยู่ในบรรทัดใดของงบฐานะการเงิน/งบกำไรขาดทุน (statements_unassigned.go)
	Unassigned []StatementUnassigned `json:"unassigned,omitempty"`
	// Footer: แถวท้ายรายงานที่พิมพ์ต่อจากแถวรวม (เช่น กำไรขาดทุนสุทธิ + รวมทั้งสิ้น ของกระดาษทำการ) — labelkey = key ภาษาของป้ายแถว
	Footer []map[string]string `json:"footer,omitempty"`
}

type Projection interface {
	Project(context.Context, Event) error
	Version(context.Context, Scope) (int64, error)
	List(context.Context, Scope, string, string, int, int, ListFilter) (Page, error)
	Get(context.Context, Scope, string, string) (json.RawMessage, error)
	Report(context.Context, Scope, string, ReportQuery) (Report, error)
}

var ErrProjectionPending = errors.New("บันทึกแล้ว กำลังเตรียมข้อมูลรายงาน กรุณารอสักครู่แล้วโหลดใหม่")

var ErrNotFound = errors.New("ไม่พบข้อมูลบัญชีที่ต้องการ")

type ListFilter struct {
	BookCode string
	Kind     string
	Status   string
	// DateFrom and DateTo (YYYY-MM-DD, both inclusive) bound the journal document date; empty = open end.
	DateFrom string
	DateTo   string
	// CodeFrom and CodeTo bound the record code of any master list (print ranges like Champ 5538/5521): CodeFrom alone matches
	// that code exactly, both are an inclusive range, CodeTo alone is ignored (Champ GLRepDeaprtmentDlg.cpp:87-98).
	CodeFrom string
	CodeTo   string
}
