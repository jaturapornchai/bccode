package cheque

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/shopspring/decimal"
)

// สถานะการลงบัญชีของเอกสารสถานะ (cheque_docs.gl_status)
const (
	GLStatusNone        = "none"
	GLStatusPosted      = "posted"
	GLStatusNoMapping   = "no_mapping"
	GLStatusNotRequired = "not_required"
)

// Ledger คือเครื่องบัญชีที่เอกสารเช็คลงบัญชีผ่าน (ทำงานใน tx เดียวกับเอกสาร) — แยกเป็น interface เพราะ generalledger
// import แพ็กเกจนี้อยู่แล้ว (รหัสธนาคาร/สมุดเงินฝาก); ตัวจริงอยู่ที่ cheque/glposting
type Ledger interface {
	// Lock ถือล็อกบริษัทเดียวกับทุกคำสั่งบัญชี ก่อนแตะแถวของโมดูล เพื่อให้เอกสารกับใบสำคัญเข้าคิวลำดับเดียวกัน (ไม่ deadlock)
	Lock(ctx context.Context, tx *sql.Tx, company string) error
	// CheckDate คือด่านงวดปิด/ปีบัญชีปิดของวันที่เอกสาร (Champ IsPeriodClose)
	CheckDate(ctx context.Context, tx *sql.Tx, scope Scope, date string) error
	// Post ลงบัญชีตามรูปแบบลงบัญชีของชนิดเอกสาร: ไม่มีรูปแบบ = GLStatusNoMapping (เอกสารยังบันทึกได้)
	Post(ctx context.Context, tx *sql.Tx, scope Scope, p Posting) (Posted, error)
	// Reverse กลับรายการใบสำคัญของรอบที่ระบุ (ใบเดิมคงอยู่เป็นหลักฐาน)
	Reverse(ctx context.Context, tx *sql.Tx, scope Scope, r Reversal) error
}

// Posting คือยอดของเอกสารสถานะหนึ่งรอบที่จะลงบัญชี (ยอดปัด 2 ตำแหน่งแล้ว: Total = Net + Fee)
type Posting struct {
	Flag     int
	DocNo    string
	Date     string
	Revision int
	// BookCode/BookAccount คือสมุดเงินฝากของเอกสารและบัญชี GL ของสมุด (ว่าง = ไม่มี) สำหรับรายการ accountfrom bank_book
	BookCode    string
	BookAccount string
	// BookFallback: ใช้บัญชีสำรองในรายการได้เมื่อไม่มีสมุด (เฉพาะยกเลิกเช็ครับในมือ Champ gltlsprocess.cpp:7824)
	BookFallback bool
	Total        decimal.Decimal
	Fee          decimal.Decimal
	Net          decimal.Decimal
}

// Posted คือผลลงบัญชี
type Posted struct {
	Status       string
	JournalID    string
	JournalDocNo string
}

// Reversal คือคำขอกลับรายการใบสำคัญของเอกสาร; Edit = กลับเพราะแก้เอกสาร (ไม่ใช่ยกเลิก)
type Reversal struct {
	Flag      int
	DocNo     string
	Revision  int
	JournalID string
	Edit      bool
	Today     string
}

// NewBookNoGLAccountError: รายการบัญชีต้องใช้บัญชีจากสมุดเงินฝาก แต่เอกสารไม่มีสมุดหรือสมุดยังไม่ระบุรหัสผังบัญชี
// — ไม่ลงบัญชีแทนการข้ามบรรทัดแบบ Champ (gltlsprocess.cpp:7806-7809) เพื่อให้เดบิต = เครดิตเสมอ
func NewBookNoGLAccountError(book string) error {
	if book == "" {
		book = "-" // เอกสารไม่มีสมุด
	}
	return newError(http.StatusBadRequest, "cheque_err_book_no_gl_account", "book_code", book)
}
