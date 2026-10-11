package cheque

import (
	"fmt"
	"sort"
	"time"

	"smlcloudplatform/internal/doctype"
)

// ชนิดเช็คในทะเบียน
const (
	TypeReceived = 0 // เช็ครับ (BCChqIn)
	TypeIssued   = 1 // เช็คจ่าย (BCChqOut)
)

// สถานะเช็ครับ (cheque_status_in_0..6) และเช็คจ่าย (cheque_status_out_0..3) ตาม Champ
const (
	InOnHand      = 0
	InDeposited   = 1
	InCleared     = 2
	InReturned    = 3
	InCancelled   = 4
	InDiscounted  = 5
	InRedeposited = 6

	OutOnHand    = 0
	OutCleared   = 1
	OutCancelled = 2
	OutPrepared  = 3
)

// transflag ของทะเบียนเช็ค (doctype registry; ใช้เป็นชุดเลขที่และคีย์ mapping ลงบัญชี)
const (
	FlagRegisterReceived = 61  // บันทึกเช็ครับ CHQ-IN
	FlagRegisterIssued   = 418 // บันทึกเช็คจ่าย CHQ-OUT
)

// transflag ของเอกสารเปลี่ยนสถานะ 8 ตัว (doctype registry)
const (
	FlagDeposit        = 419 // นำฝาก
	FlagReceivedClear  = 420 // เช็ครับผ่าน
	FlagReceivedReturn = 421 // เช็ครับคืน
	FlagRedeposit      = 422 // นำเช็คเข้าใหม่
	FlagReceivedCancel = 423 // ยกเลิกเช็ครับ
	FlagDiscount       = 424 // ขายลดเช็ครับ
	FlagIssuedClear    = 425 // เช็คจ่ายผ่าน
	FlagIssuedCancel   = 426 // ยกเลิกเช็คจ่าย
)

// ชนิดการยกเลิกเช็ครับ (423 CancelStatus): เช็คในมือ / ที่ผ่านแล้ว / ฝาก / คืนฝากธนาคาร
const (
	CancelOnHand      = 0
	CancelCleared     = 1
	CancelDeposited   = 2
	CancelRedeposited = 3
)

type dateRule int

const (
	dateNone     dateRule = iota
	dateDue               // วันที่ถึงกำหนด ≤ วันที่เอกสาร
	dateReceived          // วันที่รับเช็ค ≤ วันที่เอกสาร
)

// docRule คือกติกาของเอกสารหนึ่งชนิด (cheque-bank.md §4.3)
type docRule struct {
	chequeType int
	toStatus   int
	date       dateRule
	posts      bool // ลงบัญชีตามรูปแบบลงบัญชี (นำฝาก 419 ไม่ลงบัญชี)
	fee        bool // มีค่าใช้จ่ายต่อเช็คได้ (จอนำฝาก 5398 ไม่มีคอลัมน์ค่าใช้จ่าย)
}

var docRules = map[int]docRule{
	FlagDeposit:        {chequeType: TypeReceived, toStatus: InDeposited, date: dateReceived},
	FlagReceivedClear:  {chequeType: TypeReceived, toStatus: InCleared, date: dateDue, posts: true, fee: true},
	FlagReceivedReturn: {chequeType: TypeReceived, toStatus: InReturned, date: dateReceived, posts: true, fee: true},
	FlagRedeposit:      {chequeType: TypeReceived, toStatus: InRedeposited, date: dateReceived, posts: true, fee: true},
	FlagReceivedCancel: {chequeType: TypeReceived, toStatus: InCancelled, date: dateReceived, posts: true, fee: true},
	FlagDiscount:       {chequeType: TypeReceived, toStatus: InDiscounted, date: dateReceived, posts: true, fee: true},
	FlagIssuedClear:    {chequeType: TypeIssued, toStatus: OutCleared, date: dateDue, posts: true, fee: true},
	FlagIssuedCancel:   {chequeType: TypeIssued, toStatus: OutCancelled, date: dateNone, posts: true, fee: true},
}

func ruleFor(flag int) (docRule, bool) {
	r, ok := docRules[flag]
	return r, ok
}

// DocChequeType คือชนิดเช็คที่เอกสารสถานะใช้ (ok = false เมื่อไม่ใช่ transflag ของเอกสารสถานะ)
func DocChequeType(flag int) (int, bool) {
	r, ok := docRules[flag]
	return r.chequeType, ok
}

// OwnsDocFlag คือ true เมื่อ transflag นั้นเป็นของโมดูลเช็ค (ทะเบียนเช็ค 2 ชนิด + เอกสารเปลี่ยนสถานะ 8 ชนิด);
// เส้นทางเอกสารทั่วไปใช้ปฏิเสธชนิดเหล่านี้ เพราะตารางและการลงบัญชีของเช็คอยู่ที่ /cheque/v1 เท่านั้น
func OwnsDocFlag(flag int) bool {
	if flag == FlagRegisterReceived || flag == FlagRegisterIssued {
		return true
	}
	_, ok := docRules[flag]
	return ok
}

// DocFlags คือ transflag ของเอกสารเปลี่ยนสถานะทั้ง 8 ตัว เรียงจากน้อยไปมาก
func DocFlags() []int {
	flags := make([]int, 0, len(docRules))
	for flag := range docRules {
		flags = append(flags, flag)
	}
	sort.Ints(flags)
	return flags
}

// MappingSourceDocs คือ doctype key ของเอกสารเช็คที่ลงบัญชีด้วยรูปแบบลงบัญชี (420–426) — generalledger ใช้กำหนดว่า
// รูปแบบของเอกสารเหล่านี้ใช้แหล่งยอดตระกูล cheque และบัญชีจากสมุดเงินฝากได้
func MappingSourceDocs() []string {
	keys := []string{}
	for _, flag := range DocFlags() {
		if !docRules[flag].posts {
			continue
		}
		if dt, ok := doctype.ByFlag(flag); ok {
			keys = append(keys, dt.Key)
		}
	}
	return keys
}

// validCancelType: 423 ต้องเลือกชนิดการยกเลิก 0..3; เอกสารอื่นต้องไม่มี
func validCancelType(flag int, cancelType *int) bool {
	if flag != FlagReceivedCancel {
		return cancelType == nil
	}
	return cancelType != nil && *cancelType >= CancelOnHand && *cancelType <= CancelRedeposited
}

// selectableStatuses คือสถานะของเช็คที่เอกสารชนิดนี้เลือกได้ (Champ ตัวกรองของจอเอกสารสถานะแต่ละจอ)
func selectableStatuses(flag, cancelType int, cancelPassed bool) []int {
	switch flag {
	case FlagDeposit:
		return []int{InOnHand}
	case FlagReceivedClear, FlagDiscount:
		return []int{InOnHand, InDeposited}
	case FlagReceivedReturn:
		return []int{InCleared, InRedeposited}
	case FlagRedeposit:
		return []int{InReturned}
	case FlagReceivedCancel:
		switch cancelType {
		case CancelOnHand:
			return []int{InOnHand}
		case CancelCleared:
			return []int{InCleared}
		case CancelDeposited:
			return []int{InDeposited}
		case CancelRedeposited:
			return []int{InRedeposited}
		}
	case FlagIssuedClear:
		return []int{OutOnHand}
	case FlagIssuedCancel:
		if cancelPassed {
			return []int{OutCleared}
		}
		return []int{OutOnHand}
	}
	return nil
}

// bookPolicy คือกติกาสมุดเงินฝากของเอกสาร: required = ต้องระบุ, match = เช็คต้องอยู่สมุดเดียวกับเอกสาร,
// none = เอกสารไม่ใช้สมุด (ยกเลิกเช็คในมือ)
type bookPolicy struct{ required, match, none bool }

// ตาม ConfirmSaveData (CANNOT_SAVE_NO_BOOKBANK) และตัวกรองตอนเลือกเช็ค (BookNo = สมุดของเอกสาร) ของแต่ละจอ Champ
func bookPolicyFor(flag, cancelType int) bookPolicy {
	switch flag {
	case FlagDeposit:
		return bookPolicy{required: true}
	case FlagReceivedClear:
		// cqfrmcheqinpass.cpp:379-383 บังคับสมุด; :739-748 เลือกได้เฉพาะเช็คของสมุดนั้น
		return bookPolicy{required: true, match: true}
	case FlagReceivedReturn:
		// cqfrmcheqinreturn.cpp:301-305 บังคับสมุด (ซ้ำที่ :356-364 ถูก comment ออก); :698-711 เลือกเฉพาะเช็คของสมุดนั้น
		return bookPolicy{required: true, match: true}
	case FlagRedeposit:
		// cqfrmcheqindeposit2.cpp:711-714 บังคับสมุด และเลือกได้เฉพาะเช็คคืนของสมุดนั้น
		return bookPolicy{required: true, match: true}
	case FlagDiscount:
		// cqfrmcheqinsale.cpp:374-382 การบังคับสมุดถูก comment ออก; :676-681 ระบุสมุดแล้วกรองเฉพาะเช็คของสมุดนั้น
		return bookPolicy{match: true}
	case FlagReceivedCancel:
		if cancelType == CancelOnHand {
			return bookPolicy{none: true}
		}
		return bookPolicy{required: true, match: true}
	case FlagIssuedClear:
		// cqfrmcheqoutpass.cpp:362-366 บังคับสมุด; :664-679 เลือกเฉพาะเช็คของสมุดนั้น
		return bookPolicy{required: true, match: true}
	case FlagIssuedCancel:
		// cqfrmcheqoutcancel.cpp:359-363 บังคับสมุดทั้งยกเลิกเช็คในมือและเช็คผ่าน; :666-687 เลือกเฉพาะเช็คของสมุดนั้น
		return bookPolicy{required: true, match: true}
	}
	return bookPolicy{}
}

// lineCheque คือข้อมูลของเช็คที่ใช้ตรวจตอนบันทึกเอกสาร (อ่านหลังล็อกแถวแล้ว)
type lineCheque struct {
	id, chequeNo, docDate, dueDate, book string
	chequeType, status                   int
}

// checkSelectable ตรวจเช็คหนึ่งใบกับกติกาของเอกสาร: ชนิด/สถานะ → วันที่ → สมุด
func checkSelectable(flag, cancelType int, cancelPassed bool, docDate, docBook string, c lineCheque, field string) error {
	rule := docRules[flag]
	if c.chequeType != rule.chequeType || !containsInt(selectableStatuses(flag, cancelType, cancelPassed), c.status) {
		return newError(409, "cheque_err_status_not_selectable", field, c.chequeNo)
	}
	switch rule.date {
	case dateDue:
		if c.dueDate > docDate {
			return badRequest("cheque_err_not_due", field, c.chequeNo, thaiDate(c.dueDate))
		}
	case dateReceived:
		if c.docDate > docDate {
			return badRequest("cheque_err_received_after_doc", field, c.chequeNo, thaiDate(c.docDate))
		}
	}
	if bookPolicyFor(flag, cancelType).match && docBook != "" && c.book != docBook {
		return badRequest("cheque_err_book_mismatch", field, c.chequeNo, docBook)
	}
	return nil
}

// bookAfter คือสมุดของเช็คหลังเอกสาร: เช็ครับรับสมุดของเอกสาร (Champ BookNo) ถ้าเอกสารระบุ (423 ชนิด 0 ไม่มีสมุด จึงคงค่าเดิม);
// เช็คจ่ายคงสมุดเดิมเสมอ
func bookAfter(flag int, docBook, current string) string {
	if docRules[flag].chequeType == TypeReceived && docBook != "" {
		return docBook
	}
	return current
}

func containsInt(values []int, v int) bool {
	for _, x := range values {
		if x == v {
			return true
		}
	}
	return false
}

// thaiDate แสดงวันที่ ISO เป็น วว/ดด/ปปปป พ.ศ. ในข้อความผิดพลาด (ผู้ใช้หลักอ่านปีพุทธศักราช)
func thaiDate(iso string) string {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return iso
	}
	return fmt.Sprintf("%02d/%02d/%d", t.Day(), int(t.Month()), t.Year()+543)
}
