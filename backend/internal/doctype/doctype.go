// Package doctype คือทะเบียนประเภทเอกสารธุรกรรมที่เดียวของระบบ
//
// เอกสารทุกประเภทที่เก็บลงตาราง doc/docdetail/docref/stock_ledger มีแถวเดียวในตารางนี้:
// key ของ route /transaction/:module, เลข transflag, prefix ของเลขที่เอกสาร และผลต่อสต็อก
// โค้ดอื่นห้ามเขียนเลข transflag เอง ให้ใช้ค่าคงที่หรือฟังก์ชันค้นหาของ package นี้
//
// กติกาเลข transflag
//   - เลขเดิมของระบบคงไว้ทุกตัว เลขที่ชนกันให้ประเภทที่มีก่อนคงเลขเดิม
//   - ประเภทที่เพิ่มใหม่ใช้ช่วง 400-499 เรียงตามลำดับที่เพิ่ม (กลุ่มซื้อ เจ้าหนี้ ลูกหนี้ ธนาคาร เช็ค บัตรเครดิต สินค้า)
//     ต่อท้ายเท่านั้น ห้ามเปลี่ยนหรือนำเลขกลับมาใช้ซ้ำ
//   - prefix ใหม่เป็นอักษรอังกฤษตัวใหญ่ 2-4 ตัว ไม่ซ้ำกัน เลขที่เอกสารคือ PREFIX + ปี พ.ศ. 2 หลัก + เดือน + "-NNNN"
//
// ประเภทที่เพิ่มใหม่บันทึกเอกสารอย่างเดียว ไม่ลงบัญชีและไม่ตัดสต็อกตอนบันทึก
// การลงบัญชีอัตโนมัติผูกกับค่าคงที่ที่ตั้งชื่อไว้ด้านล่างเท่านั้น
// ชื่อเมนูภาษาไทยอยู่ในคอมเมนต์ท้ายแต่ละแถว ข้อความที่ผู้ใช้เห็นมาจาก languages.tsv
package doctype

import "sort"

// StockEffect บอกว่าเอกสารประเภทนี้ทำให้สต็อกเคลื่อนไหวอย่างไรในเครื่องคิดต้นทุน
type StockEffect int

const (
	StockNone   StockEffect = iota // ไม่กระทบสต็อก
	StockIn                        // รับเข้าทั้งใบ
	StockOut                       // จ่ายออกทั้งใบ
	StockByLine                    // ทิศทางอยู่ที่ calcflag ของแต่ละบรรทัด (โอนคลัง ปรับปรุงสต็อก)
)

// DocType คือประเภทเอกสารหนึ่งแถวในทะเบียน
type DocType struct {
	Key    string
	Flag   int
	Prefix string
	Stock  StockEffect
}

// ค่าคงที่ของประเภทเอกสารที่โค้ดใช้ตัดสินใจ (ลงบัญชีอัตโนมัติ ตัดสต็อก รายงาน)
const (
	PurchaseOrder     = 6
	Purchase          = 12
	PurchaseReturn    = 16
	StockPickup       = 20
	Expense           = 24
	Paid              = 28
	Pay               = 29
	SaleOrder         = 36
	SaleInvoice       = 44
	SaleInvoiceReturn = 48
	StockTransfer     = 62
	GoodsReceive      = 401
)

var registry = []DocType{
	// สินค้าคงคลัง
	{"stock-balance", 56, "BAL", StockIn},                  // บันทึกสินค้ายกมา
	{"stock-receive-product", 72, "RCV", StockIn},          // บันทึกรับสินค้าสำเร็จรูป
	{"stock-pickup-product", StockPickup, "OUT", StockOut}, // บันทึกเบิกใช้สินค้า,วัตถุดิบ
	{"stock-return-product", 21, "RET", StockIn},           // บันทึกรับคืนสินค้า,วัตถุดิบ
	{"stock-transfer", StockTransfer, "XFR", StockByLine},  // บันทึกโอนสินค้าระหว่างคลัง
	{"stock-adjustment", 310, "ADJ", StockByLine},          // บันทึกปรับปรุงสินค้าหลังตรวจนับ
	{"stock-issue-request", 430, "OUTR", StockNone},        // บันทึกใบขอเบิกใช้สินค้า,วัตถุดิบ
	{"stock-transfer-request", 431, "XFRR", StockNone},     // บันทึกใบขอโอนสินค้า
	{"goods-inspection", 432, "INSP", StockNone},           // ใบตรวจรับสินค้า
	{"stock-count", 433, "CNT", StockNone},                 // บันทึกเอกสารเพื่อตรวจนับสินค้า และบันทึกผลการตรวจนับสินค้า

	// ขาย ลูกหนี้
	{"quotation", 34, "QT", StockNone},                        // บันทึกใบเสนอราคาสินค้า
	{"sale-order", SaleOrder, "SO", StockNone},                // บันทึกใบสั่งขาย/สั่งจองสินค้า
	{"sale-invoice", SaleInvoice, "INV", StockOut},            // บันทึกขายสินค้า,บริการ
	{"sale-invoice-return", SaleInvoiceReturn, "SR", StockIn}, // บันทึกใบรับคืนสินค้า,ลดหนี้
	{"ardebitnote", 50, "DN", StockNone},                      // บันทึกใบเพิ่มหนี้/เพิ่มสินค้า(ลูกหนี้)
	{"debtorbeginningbalance", 52, "AR-BAL", StockNone},       // บันทึกเอกสารลูกหนี้ยกมา
	{"receivableother", 54, "AR-OTH", StockNone},              // บันทึกเอกสารตั้งลูกหนี้อื่นๆ
	{"billingnote", 26, "BL", StockNone},                      // บันทึกใบวางบิล
	{"paid", Paid, "RE", StockNone},                           // บันทึกใบเสร็จรับเงิน/รับชำระหนี้
	{"receivedeposit", 32, "RDEP", StockNone},                 // บันทึกใบรับเงินมัดจำ
	{"receiveadvancerefund", 35, "RADR", StockNone},           // บันทึกคืนเงินรับล่วงหน้า
	{"chequereceived", 61, "CHQ-IN", StockNone},               // บันทึกเช็ครับ

	// ซื้อ เจ้าหนี้
	{"purchase-requisition", 2, "PR", StockNone},          // บันทึกใบเสนอซื้อสินค้า
	{"purchase-order", PurchaseOrder, "PO", StockNone},    // บันทึกใบสั่งซื้อสินค้า
	{"purchase", Purchase, "PUR", StockIn},                // บันทึกซื้อสินค้า,บริการ
	{"purchase-return", PurchaseReturn, "PRET", StockOut}, // บันทึกส่งคืนสินค้า
	{"creditorbeginningbalance", 53, "AP-BAL", StockNone}, // บันทึกเอกสารเจ้าหนี้ยกมา
	{"expense", Expense, "EXP", StockNone},                // บันทึกจ่ายเงินอื่นๆ
	{"apbillingreceipt", 27, "BAP", StockNone},            // บันทึกใบรับวางบิล
	{"pay", Pay, "PV", StockNone},                         // บันทึกจ่ายชำระหนี้
	{"deposit", 22, "DEP", StockNone},                     // บันทึกใบจ่ายเงินมัดจำ
	{"advancepayment", 30, "ADV", StockNone},              // บันทึกใบจ่ายเงินล่วงหน้า
	{"advancepaymentrefund", 31, "ADR", StockNone},        // บันทึกรับคืนเงินจ่ายล่วงหน้า

	// ประเภทที่เพิ่มใหม่ ช่วง 400-499 เรียงตามลำดับที่เพิ่ม
	{"rfq", 400, "RFQ", StockNone},                   // บันทึกใบสืบราคาสินค้ารวม
	{"goodsreceive", GoodsReceive, "GR", StockIn},    // บันทึกรับสินค้าจากการซื้อ
	{"invoicereceive", 402, "IR", StockNone},         // บันทึกตั้งหนี้จากการซื้อ
	{"landedcost", 403, "LC", StockNone},             // บันทึกต้นทุนแฝง
	{"apdebitnote", 404, "APDN", StockNone},          // บันทึกเพิ่มหนี้/เพิ่มสินค้า(เจ้าหนี้)
	{"apcreditnote", 405, "APCN", StockNone},         // บันทึกส่งคืนสินค้า/ลดหนี้
	{"apotherdebt", 406, "APOD", StockNone},          // บันทึกเอกสารตั้งเจ้าหนี้อื่นๆ
	{"apbaddebt", 407, "APBD", StockNone},            // บันทึกตัดหนี้สูญ(เจ้าหนี้)
	{"receiveadvance", 408, "RADV", StockNone},       // บันทึกใบรับเงินล่วงหน้า
	{"temporaryreceipt", 409, "TRC", StockNone},      // บันทึกใบเสร็จชั่วคราว
	{"arbaddebt", 410, "ARBD", StockNone},            // บันทึกตัดหนี้สูญ(ลูกหนี้)
	{"otherincome", 411, "OIN", StockNone},           // บันทึกรับเงินอื่นๆ
	{"employeeadvance", 412, "EADV", StockNone},      // บันทึกขอเบิกเงินทดรองจ่าย
	{"cashdeposit", 413, "CDP", StockNone},           // บันทึกนำฝากเงินสด
	{"cashwithdrawal", 414, "CWD", StockNone},        // บันทึกถอนเงินสด
	{"bankexpense", 415, "BEX", StockNone},           // บันทึกค่าใช้จ่ายธนาคาร
	{"bankincome", 416, "BIN", StockNone},            // บันทึกรายได้จากธนาคาร
	{"banktransfer", 417, "BTR", StockNone},          // บันทึกโอนเงินระหว่างธนาคาร
	{"chequeissued", 418, "CHQ-OUT", StockNone},      // บันทึกเช็คจ่าย
	{"chequedeposit", 419, "CQDP", StockNone},        // บันทึกนำฝากเช็ครับ
	{"chequereceivedclear", 420, "CQIP", StockNone},  // บันทึกเช็ครับผ่าน
	{"chequereceivedreturn", 421, "CQIR", StockNone}, // บันทึกเช็ครับคืน
	{"chequeredeposit", 422, "CQRD", StockNone},      // บันทึกนำเช็คเข้าใหม่
	{"chequereceivedcancel", 423, "CQIC", StockNone}, // บันทึกยกเลิกเช็ครับ
	{"chequediscount", 424, "CQDS", StockNone},       // บันทึกขายลดเช็ครับ
	{"chequeissuedclear", 425, "CQOP", StockNone},    // บันทึกเช็คจ่ายผ่าน
	{"chequeissuedcancel", 426, "CQOC", StockNone},   // บันทึกยกเลิกเช็คจ่าย
	{"creditcard", 427, "CC", StockNone},             // บันทึกบัตรเครดิต
	{"creditcardsettle", 428, "CCS", StockNone},      // บันทึกขึ้นเงินบัตรเครดิต
	{"creditcardcancel", 429, "CCC", StockNone},      // บันทึกยกเลิกบัตรเครดิต
}

var (
	byKey  = make(map[string]DocType, len(registry))
	byFlag = make(map[int]DocType, len(registry))
)

func init() {
	for _, dt := range registry {
		byKey[dt.Key] = dt
		byFlag[dt.Flag] = dt
	}
}

// All คืนสำเนาของทะเบียนทั้งหมด เรียงตามที่ประกาศไว้
func All() []DocType {
	return append([]DocType(nil), registry...)
}

// ByKey ค้นประเภทเอกสารจาก key ของ route (ตรงตัวอักษร)
func ByKey(key string) (DocType, bool) {
	dt, ok := byKey[key]
	return dt, ok
}

// ByFlag ค้นประเภทเอกสารจากเลข transflag
func ByFlag(flag int) (DocType, bool) {
	dt, ok := byFlag[flag]
	return dt, ok
}

// ByDocNo เดาประเภทเอกสารจากเลขที่เอกสาร ด้วย prefix ที่ยาวที่สุดที่ตามด้วยอักขระที่ไม่ใช่ตัวอักษร
// เช่น PRET6910-0001 เป็นส่งคืนสินค้า ไม่ใช่ใบเสนอซื้อ (PR)
func ByDocNo(docNo string) (DocType, bool) {
	var best DocType
	found := false
	for _, dt := range registry {
		p := dt.Prefix
		if len(p) <= len(best.Prefix) || len(docNo) < len(p) || docNo[:len(p)] != p {
			continue
		}
		if len(docNo) > len(p) && isUpperLetter(docNo[len(p)]) {
			continue
		}
		best, found = dt, true
	}
	return best, found
}

// StockFlags คืนเลข transflag ที่กระทบสต็อก (รับเข้า จ่ายออก หรือดูรายบรรทัด) เรียงจากน้อยไปมาก
func StockFlags() []int {
	flags := make([]int, 0, len(registry))
	for _, dt := range registry {
		if dt.Stock != StockNone {
			flags = append(flags, dt.Flag)
		}
	}
	sort.Ints(flags)
	return flags
}

func isUpperLetter(c byte) bool {
	return c >= 'A' && c <= 'Z'
}
