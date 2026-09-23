package whtcert

import "smlcloudplatform/internal/pdftext"

const (
	alignLeft   = pdftext.AlignLeft
	alignCenter = pdftext.AlignCenter
	alignRight  = pdftext.AlignRight
)

// พิกัดทุกช่องบนแบบ "หนังสือรับรองการหักภาษี ณ ที่จ่าย ตามมาตรา 50 ทวิ" ของกรมสรรพากร
// (assets/50tawi-rd.pdf = approve_wh3_081156.pdf จาก rd.go.th, หน้า A4 595×842 pt, จุดกำเนิดมุมซ้ายล่าง)
// ค่าทั้งหมดวัดจากฟอร์มจริง: เส้นกรอบ/ช่องจาก raster 8 เท่า, baseline จาก text layer ของฟอร์ม,
// และตำแหน่งช่อง AcroForm เดิมของกรมสรรพากร — แก้เลขใดต้องเทียบกับภาพ render จริงเสมอ

const (
	pageWidth  = 595.0
	pageHeight = 842.0
	fontSize   = 11.0
	minFont    = 7.0
	// digitRise - ระยะจากขอบล่างช่องสี่เหลี่ยมถึง baseline ให้ตัวเลขสูง ~7.6pt อยู่กลางช่องสูง 14pt
	digitRise = 3.2
)

// cellGroups - ช่องเลขประจำตัว: กลุ่มช่องคั่นด้วยขีด (x ซ้าย, x ขวา, จำนวนหลัก)
type cellGroup struct {
	left, right float64
	count       int
}

type idBoxes struct {
	groups   []cellGroup
	baseline float64
}

var (
	payerID13 = idBoxes{groups: []cellGroup{{374.3, 386.3, 1}, {392.3, 440.3, 4}, {446.8, 507.3, 5}, {512.8, 537.3, 2}, {544.3, 556.3, 1}}, baseline: 744.5 + digitRise}
	payeeID13 = idBoxes{groups: []cellGroup{{374.8, 386.8, 1}, {392.8, 441.0, 4}, {447.5, 508.0, 5}, {513.5, 538.0, 2}, {544.8, 557.0, 1}}, baseline: 676.0 + digitRise}
	payerID10 = idBoxes{groups: []cellGroup{{418.5, 430.3, 1}, {436.5, 484.3, 4}, {490.5, 538.3, 4}, {544.5, 556.5, 1}}, baseline: 728.5 + digitRise}
	payeeID10 = idBoxes{groups: []cellGroup{{419.0, 431.0, 1}, {437.0, 485.0, 4}, {491.0, 539.0, 4}, {545.0, 557.0, 1}}, baseline: 657.0 + digitRise}
)

// lineField - ข้อความบนเส้นประ: เริ่มที่ x, baseline, กว้างได้ไม่เกิน maxWidth (ย่อฟอนต์ถ้ายาว)
type lineField struct {
	x, baseline, maxWidth float64
	align                 pdftext.Align
}

var (
	fieldBookNo       = lineField{x: 521, baseline: 781.9, maxWidth: 38}
	fieldRunNo        = lineField{x: 518, baseline: 767.5, maxWidth: 42}
	fieldPayerName    = lineField{x: 56, baseline: 730.3, maxWidth: 258}
	fieldPayerAddress = lineField{x: 58, baseline: 706.9, maxWidth: 492}
	fieldPayeeName    = lineField{x: 56, baseline: 657.5, maxWidth: 258}
	fieldPayeeAddress = lineField{x: 58, baseline: 631.4, maxWidth: 492}
	fieldSequence     = lineField{x: 107.3, baseline: 601.2 + digitRise, maxWidth: 58, align: alignCenter}
	fieldDividendRate = lineField{x: 181, baseline: 389.3, maxWidth: 38, align: alignCenter}    // (1.4) อัตราอื่น ๆ (ระบุ)
	fieldDividendNote = lineField{x: 147, baseline: 273.3, maxWidth: 167}                       // (2.5) อื่น ๆ (ระบุ)
	fieldOtherNote    = lineField{x: 94, baseline: 201.4, maxWidth: 230}                        // 6. อื่น ๆ (ระบุ)
	fieldTaxWords     = lineField{x: 368.5, baseline: 161.6, maxWidth: 370, align: alignCenter} // รวมเงินภาษีที่หักนำส่ง (ตัวอักษร)
	fieldFundGPF      = lineField{x: 258.0, baseline: 144.7, maxWidth: 40, align: alignCenter}  // กบข./กสจ./กองทุนสงเคราะห์ครูโรงเรียนเอกชน
	fieldFundSSF      = lineField{x: 385.3, baseline: 144.7, maxWidth: 41, align: alignCenter}  // กองทุนประกันสังคม
	fieldFundPVD      = lineField{x: 519.5, baseline: 144.7, maxWidth: 43, align: alignCenter}  // กองทุนสำรองเลี้ยงชีพ
	fieldConditionTxt = lineField{x: 470, baseline: 122.1, maxWidth: 87}                        // ผู้จ่ายเงิน (4) อื่น ๆ (ระบุ)
	fieldIssueDay     = lineField{x: 349.6, baseline: 75.9, maxWidth: 34, align: alignCenter}
	fieldIssueMonth   = lineField{x: 395.8, baseline: 75.9, maxWidth: 52, align: alignCenter}
	fieldIssueYear    = lineField{x: 450.5, baseline: 75.9, maxWidth: 50, align: alignCenter}
	fieldCopyLabel    = lineField{x: 560, baseline: 826, maxWidth: 200, align: alignRight} // ที่ว่างมุมขวาบน (ปุ่ม Clear Data เดิม)
)

// point - จุดกึ่งกลางช่องสี่เหลี่ยมสำหรับเครื่องหมาย ✓
type point struct{ x, y float64 }

// formChecks - ช่อง "ลำดับที่ ... ในแบบ" (1)–(7)
var formChecks = map[string]point{
	FormPND1A:        {215.15, 607.85},
	FormPND1ASpecial: {293.20, 607.85},
	FormPND2:         {400.95, 607.90},
	FormPND3:         {477.95, 607.90},
	FormPND2A:        {215.15, 589.45},
	FormPND3A:        {293.15, 589.45},
	FormPND53:        {400.95, 589.45},
}

// conditionChecks - ผู้จ่ายเงิน (1) หัก ณ ที่จ่าย (2) ออกให้ตลอดไป (3) ออกให้ครั้งเดียว (4) อื่น ๆ
var conditionChecks = map[string]point{
	ConditionWithhold: {88.55, 125.75},
	ConditionAlways:   {181.55, 125.75},
	ConditionOnce:     {289.15, 125.75},
	ConditionOther:    {399.80, 125.75},
}

// copyChecks - เครื่องหมายหน้าข้อความ "ฉบับที่ 1 / ฉบับที่ 2" ที่พิมพ์ไว้บนหัวฟอร์ม
var copyChecks = map[int]point{1: {27.5, 819.4}, 2: {27.5, 805.0}}

// ตารางจำนวนเงิน: คอลัมน์ "วัน เดือน หรือปีภาษี ที่จ่าย" | "จำนวนเงินที่จ่าย" (บาท|สตางค์) | "ภาษีที่หัก" (บาท|สตางค์)
const (
	colDateCenter    = 366.0
	colDateWidth     = 74.0
	colPayBahtRight  = 472.9 // เส้นแบ่งสตางค์ที่ 474.9
	colPaySatang     = 482.3 // กึ่งกลางช่องสตางค์ 474.9–489.75
	colPayBahtWidth  = 64.0
	colTaxBahtRight  = 544.3 // เส้นแบ่งสตางค์ที่ 546.25
	colTaxSatang     = 553.4 // กึ่งกลางช่องสตางค์ 546.25–560.5
	colTaxBahtWidth  = 52.0
	totalRowBaseline = 182.5
)

// incomeRowBaselines - baseline ของแต่ละบรรทัดประเภทเงินได้ (= เส้นประของบรรทัดนั้น + 2pt)
var incomeRowBaselines = map[string]float64{
	Income401:   535.4, // 1. เงินเดือน ค่าจ้าง ฯลฯ 40 (1)
	Income402:   520.9, // 2. ค่าธรรมเนียม ค่านายหน้า 40 (2)
	Income403:   506.4, // 3. ค่าแห่งลิขสิทธิ์ 40 (3)
	Income404A:  492.0, // 4. (ก) ดอกเบี้ย 40 (4) (ก)
	IncomeDiv11: 433.9, // 4. (ข) (1.1) ร้อยละ 30
	IncomeDiv12: 419.4, // (1.2) ร้อยละ 25
	IncomeDiv13: 405.0, // (1.3) ร้อยละ 20
	IncomeDiv14: 390.5, // (1.4) อัตราอื่น ๆ
	IncomeDiv21: 361.4, // (2.1) กิจการได้รับยกเว้นภาษี
	IncomeDiv22: 332.4, // (2.2) เงินปันผลที่ได้รับยกเว้น
	IncomeDiv23: 303.5, // (2.3) หักผลขาดทุนสุทธิยกมา
	IncomeDiv24: 288.4, // (2.4) equity method (ฟอร์มไม่มีเส้นประ ใช้ baseline ป้าย + 2)
	IncomeDiv25: 274.4, // (2.5) อื่น ๆ
	Income3Tres: 216.5, // 5. ตามคำสั่งกรมสรรพากรที่ออกตามมาตรา 3 เตรส
	IncomeOther: 202.0, // 6. อื่น ๆ (ระบุ)
}
