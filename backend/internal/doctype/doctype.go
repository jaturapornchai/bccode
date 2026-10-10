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
	Screen string // menu id ในแค็ตตาล็อกสิทธิ์ (<screen> = เข้าจอ/อ่าน, <screen>:create|update|delete = เขียน)
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
	{"stock-balance", 56, "BAL", StockIn, "stock-balance"},                       // บันทึกสินค้ายกมา
	{"stock-receive-product", 72, "RCV", StockIn, "stock-receive"},               // บันทึกรับสินค้าสำเร็จรูป
	{"stock-pickup-product", StockPickup, "OUT", StockOut, "stock-pickup"},       // บันทึกเบิกใช้สินค้า,วัตถุดิบ
	{"stock-return-product", 21, "RET", StockIn, "stock-return"},                 // บันทึกรับคืนสินค้า,วัตถุดิบ
	{"stock-transfer", StockTransfer, "XFR", StockByLine, "stock-transfer"},      // บันทึกโอนสินค้าระหว่างคลัง
	{"stock-adjustment", 310, "ADJ", StockByLine, "stock-adjust"},                // บันทึกปรับปรุงสินค้าหลังตรวจนับ
	{"stock-issue-request", 430, "OUTR", StockNone, "stock-issue-request"},       // บันทึกใบขอเบิกใช้สินค้า,วัตถุดิบ
	{"stock-transfer-request", 431, "XFRR", StockNone, "stock-transfer-request"}, // บันทึกใบขอโอนสินค้า
	{"goods-inspection", 432, "INSP", StockNone, "goods-inspection"},             // ใบตรวจรับสินค้า
	{"stock-count", 433, "CNT", StockNone, "stock-count"},                        // บันทึกเอกสารเพื่อตรวจนับสินค้า และบันทึกผลการตรวจนับสินค้า

	// ขาย ลูกหนี้
	{"quotation", 34, "QT", StockNone, "quotation"},                                 // บันทึกใบเสนอราคาสินค้า
	{"sale-order", SaleOrder, "SO", StockNone, "sale-order"},                        // บันทึกใบสั่งขาย/สั่งจองสินค้า
	{"sale-invoice", SaleInvoice, "INV", StockOut, "sale"},                          // บันทึกขายสินค้า,บริการ
	{"sale-invoice-return", SaleInvoiceReturn, "SR", StockIn, "sale-return"},        // บันทึกใบรับคืนสินค้า,ลดหนี้
	{"ardebitnote", 50, "DN", StockNone, "debit-note"},                              // บันทึกใบเพิ่มหนี้/เพิ่มสินค้า(ลูกหนี้)
	{"debtorbeginningbalance", 52, "AR-BAL", StockNone, "debtor-beginning-balance"}, // บันทึกเอกสารลูกหนี้ยกมา
	{"receivableother", 54, "AR-OTH", StockNone, "ar-other-debt"},                   // บันทึกเอกสารตั้งลูกหนี้อื่นๆ
	{"billingnote", 26, "BL", StockNone, "billing-note"},                            // บันทึกใบวางบิล
	{"paid", Paid, "RE", StockNone, "paid"},                                         // บันทึกใบเสร็จรับเงิน/รับชำระหนี้
	{"receivedeposit", 32, "RDEP", StockNone, "receive-deposit"},                    // บันทึกใบรับเงินมัดจำ
	{"receiveadvancerefund", 35, "RADR", StockNone, "return-advance"},               // บันทึกคืนเงินรับล่วงหน้า
	{"chequereceived", 61, "CHQ-IN", StockNone, "cheque-received"},                  // บันทึกเช็ครับ

	// ซื้อ เจ้าหนี้
	{"purchase-requisition", 2, "PR", StockNone, "purchase-requisition"},                // บันทึกใบเสนอซื้อสินค้า
	{"purchase-order", PurchaseOrder, "PO", StockNone, "purchase-order"},                // บันทึกใบสั่งซื้อสินค้า
	{"purchase", Purchase, "PUR", StockIn, "purchase"},                                  // บันทึกซื้อสินค้า,บริการ
	{"purchase-return", PurchaseReturn, "PRET", StockOut, "purchase-return"},            // บันทึกส่งคืนสินค้า
	{"creditorbeginningbalance", 53, "AP-BAL", StockNone, "creditor-beginning-balance"}, // บันทึกเอกสารเจ้าหนี้ยกมา
	{"expense", Expense, "EXP", StockNone, "expense-record"},                            // บันทึกจ่ายเงินอื่นๆ
	{"apbillingreceipt", 27, "BAP", StockNone, "ap-billing-receipt"},                    // บันทึกใบรับวางบิล
	{"pay", Pay, "PV", StockNone, "pay"},                                                // บันทึกจ่ายชำระหนี้
	{"deposit", 22, "DEP", StockNone, "deposit"},                                        // บันทึกใบจ่ายเงินมัดจำ
	{"advancepayment", 30, "ADV", StockNone, "advance-payment"},                         // บันทึกใบจ่ายเงินล่วงหน้า
	{"advancepaymentrefund", 31, "ADR", StockNone, "advance-payment-refund"},            // บันทึกรับคืนเงินจ่ายล่วงหน้า

	// ประเภทที่เพิ่มใหม่ ช่วง 400-499 เรียงตามลำดับที่เพิ่ม
	{"rfq", 400, "RFQ", StockNone, "rfq"},                                      // บันทึกใบสืบราคาสินค้ารวม
	{"goodsreceive", GoodsReceive, "GR", StockIn, "purchase-partial"},          // บันทึกรับสินค้าจากการซื้อ
	{"invoicereceive", 402, "IR", StockNone, "accrual-receive"},                // บันทึกตั้งหนี้จากการซื้อ
	{"landedcost", 403, "LC", StockNone, "purchase-landed-cost"},               // บันทึกต้นทุนแฝง
	{"apdebitnote", 404, "APDN", StockNone, "purchase-credit-note"},            // บันทึกเพิ่มหนี้/เพิ่มสินค้า(เจ้าหนี้)
	{"apcreditnote", 405, "APCN", StockNone, "purchase-debit-note"},            // บันทึกส่งคืนสินค้า/ลดหนี้
	{"apotherdebt", 406, "APOD", StockNone, "ap-other-debt"},                   // บันทึกเอกสารตั้งเจ้าหนี้อื่นๆ
	{"apbaddebt", 407, "APBD", StockNone, "ap-bad-debt"},                       // บันทึกตัดหนี้สูญ(เจ้าหนี้)
	{"receiveadvance", 408, "RADV", StockNone, "receive-advance"},              // บันทึกใบรับเงินล่วงหน้า
	{"temporaryreceipt", 409, "TRC", StockNone, "temporary-receipt"},           // บันทึกใบเสร็จชั่วคราว
	{"arbaddebt", 410, "ARBD", StockNone, "ar-bad-debt"},                       // บันทึกตัดหนี้สูญ(ลูกหนี้)
	{"otherincome", 411, "OIN", StockNone, "other-income-receipt"},             // บันทึกรับเงินอื่นๆ
	{"employeeadvance", 412, "EADV", StockNone, "employee-advance"},            // บันทึกขอเบิกเงินทดรองจ่าย
	{"cashdeposit", 413, "CDP", StockNone, "cash-deposit"},                     // บันทึกนำฝากเงินสด
	{"cashwithdrawal", 414, "CWD", StockNone, "cash-withdraw"},                 // บันทึกถอนเงินสด
	{"bankexpense", 415, "BEX", StockNone, "bank-charges"},                     // บันทึกค่าใช้จ่ายธนาคาร
	{"bankincome", 416, "BIN", StockNone, "bank-income"},                       // บันทึกรายได้จากธนาคาร
	{"banktransfer", 417, "BTR", StockNone, "account-transfer"},                // บันทึกโอนเงินระหว่างธนาคาร
	{"chequeissued", 418, "CHQ-OUT", StockNone, "cheque-issued"},               // บันทึกเช็คจ่าย
	{"chequedeposit", 419, "CQDP", StockNone, "cheque-deposit"},                // บันทึกนำฝากเช็ครับ
	{"chequereceivedclear", 420, "CQIP", StockNone, "cheque-received-clear"},   // บันทึกเช็ครับผ่าน
	{"chequereceivedreturn", 421, "CQIR", StockNone, "cheque-received-return"}, // บันทึกเช็ครับคืน
	{"chequeredeposit", 422, "CQRD", StockNone, "cheque-redeposit"},            // บันทึกนำเช็คเข้าใหม่
	{"chequereceivedcancel", 423, "CQIC", StockNone, "cheque-received-cancel"}, // บันทึกยกเลิกเช็ครับ
	{"chequediscount", 424, "CQDS", StockNone, "cheque-discount"},              // บันทึกขายลดเช็ครับ
	{"chequeissuedclear", 425, "CQOP", StockNone, "cheque-issued-clear"},       // บันทึกเช็คจ่ายผ่าน
	{"chequeissuedcancel", 426, "CQOC", StockNone, "cheque-issued-cancel"},     // บันทึกยกเลิกเช็คจ่าย
	{"creditcard", 427, "CC", StockNone, "credit-card-receipts"},               // บันทึกบัตรเครดิต
	{"creditcardsettle", 428, "CCS", StockNone, "credit-card-settle"},          // บันทึกขึ้นเงินบัตรเครดิต
	{"creditcardcancel", 429, "CCC", StockNone, "credit-card-cancel"},          // บันทึกยกเลิกบัตรเครดิต
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

// screenAliases คือ menu id เพิ่มเติมที่เปิดจอของประเภทเดียวกัน (จอใบตรวจนับและจอผลตรวจนับใช้ stock-count ร่วมกัน)
var screenAliases = map[string][]string{
	"stock-count":            {"stock-count-sheet"},
	"purchase-requisition":   {"purchase-requisition-approve", "purchase-requisition-cancel", "purchase-requisition-approve-cancel"},
	"purchase-order":         {"purchase-order-approve", "purchase-order-cancel"},
	"quotation":              {"quotation-approve", "quotation-cancel"},
	"sale-order":             {"sale-order-approve", "sale-order-cancel"},
	"stock-issue-request":    {"stock-issue-request-cancel"},
	"stock-transfer-request": {"stock-transfer-request-cancel"},
	"sale-invoice":           {"sale-delivery-date"},
	"purchase":               {"purchase-receipt-date-adjust"},
	"apcreditnote":           {"purchase-reduce-debt"},
	"rfq":                    {"rfq-price-table"},
}

// Screens คืน menu id ที่เปิดประเภทเอกสารนี้ได้ (ตัวแรกคือเมนูหลัก) ถ้าไม่รู้จัก key คืน nil
func Screens(key string) []string {
	dt, ok := byKey[key]
	if !ok {
		return nil
	}
	return append([]string{dt.Screen}, screenAliases[key]...)
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

// postingFlags คือประเภทเอกสารที่ลงบัญชีตอนบันทึก (ประเภทอื่นบันทึกเอกสารอย่างเดียว)
// บัญชีและสมุดรายวันของแต่ละประเภทมาจากรูปแบบลงบัญชี (GL mappings) ที่ผู้ใช้ตั้งไว้ ไม่ได้ฝังในโค้ด
var postingFlags = map[int]bool{
	Purchase: true, PurchaseReturn: true, Expense: true, Paid: true, Pay: true, SaleInvoice: true, SaleInvoiceReturn: true,
}

// PostsToGL บอกว่าเอกสารประเภทนี้ลงบัญชีตอนบันทึกหรือไม่
func PostsToGL(flag int) bool { return postingFlags[flag] }

// PostingKeys คืน key ของประเภทเอกสารที่ลงบัญชีได้ เรียงตามเลข transflag
func PostingKeys() []string {
	flags := make([]int, 0, len(postingFlags))
	for flag := range postingFlags {
		flags = append(flags, flag)
	}
	sort.Ints(flags)
	keys := make([]string, 0, len(flags))
	for _, flag := range flags {
		keys = append(keys, byFlag[flag].Key)
	}
	return keys
}

func isUpperLetter(c byte) bool {
	return c >= 'A' && c <= 'Z'
}
