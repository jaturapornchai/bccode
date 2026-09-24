// Package rdfile เขียนไฟล์ยื่นภาษีหัก ณ ที่จ่ายด้วยสื่อ ตาม "รูปแบบข้อมูล (Format กลาง) Version 2.0 ปรับปรุง 16/06/2568"
// ของกรมสรรพากร (ภ.ง.ด.53 / ภ.ง.ด.3 / ภ.ง.ด.2 รายเดือน) สำหรับโปรแกรม SWC-UI ช่องทาง "ยื่นแบบด้วยสื่อฝากไฟล์ออนไลน์".
//
// แหล่งอ้างอิง: FormatPND53V2_0.pdf, FormatPND3V2_0.pdf, FormatPND2V2_0.pdf และไฟล์ตัวอย่างทางการ swc_Data_Test.zip
// (https://www.rd.go.th/fileadmin/user_upload/WHT/Download/) — ลำดับช่องทุกช่องในไฟล์นี้ตรงกับตาราง "ลำดับที่" ของเอกสาร
// และทดสอบย้อนกับไฟล์ตัวอย่างทีละ byte (BC_RDFILE_SAMPLE_ZIP)
//
// แพ็กเกจนี้ไม่มีตรรกะภาษีและไม่แตะฐานข้อมูล: ผู้เรียก (handlers) แปลงเอกสารที่บันทึกเป็น Header/Detail ที่เป็นสตริงแล้ว
// Check ตรวจค่าตามข้อกำหนดของ Format ทุกข้อ (รวบรวมทุกจุดผิด ไม่หยุดที่ข้อแรก) ก่อน Encode
// ไฟล์ .txt นี้ใช้กับ SWC-UI เท่านั้น — อัปโหลดเข้า e-Filing ตรง ๆ ไม่ได้ (SWC แปลงเป็น .rdx ให้)
package rdfile

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// Kind - แบบที่ทำไฟล์ได้ = ค่า TAX_TYPE ในไฟล์และคำนำหน้าชื่อไฟล์
type Kind string

const (
	PND53 Kind = "PND53"
	PND3  Kind = "PND3"
	PND2  Kind = "PND2"
)

// จำนวนช่องต่อบรรทัดตามเอกสาร Format กลาง V2.0 (ตรงกับไฟล์ตัวอย่างทางการ)
const (
	headerFieldsWHT  = 25 // ภ.ง.ด.53 / ภ.ง.ด.3
	detailFieldsWHT  = 38
	headerFieldsPND2 = 22
	detailFieldsPND2 = 27
)

// ค่าคงที่ที่ Format กำหนด (ไม่ใช่ข้อความหน้าจอ — ไฟล์เป็นแบบทางการภาษาไทย ไม่แปลตามภาษาผู้ใช้)
const (
	SenderIDMedia  = "0000"       // H2 SENDER_ID: "กรณียื่นสื่อฯ ระบุ 0000"
	SenderRoleSelf = "1"          // H5 SENDER_ROLE: 1 = ผู้หักภาษี ณ ที่จ่าย
	FormFlagMedia  = "1"          // FORM_FLAG: 1 = ยื่นสื่อฝากไฟล์
	NoTIN          = "0000000000" // D5 TIN (เลข 10 หลักแบบเดิม): ระบบไม่เก็บ → "ถ้าไม่มีให้ระบุ 0000000000"
	NoDate         = "00000000"   // PAID_DATE ของรายการที่ไม่มี
	ZeroAmount     = "0.00"       // ช่องเงินที่ไม่มีข้อมูล (ข้อกำหนดข้อ 8, 13)
	// HeadOfficeDept - DEPT_NAME เมื่อไม่แยกนำส่งเป็นแผนก: "ระบุชื่อแผนก/ส่วน/ฝ่ายที่นำส่ง หรือสำนักงานใหญ่ กรณีไม่แยกนำส่งเป็นแผนก"
	HeadOfficeDept = "สำนักงานใหญ่"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// Header - บรรทัด H (ทุกค่าเป็นสตริงตามที่จะเขียนลงไฟล์; Summarize เติมยอดรวมจาก Detail)
type Header struct {
	SenderID     string
	SenderNID    string
	SenderBranch string
	SenderRole   string
	NID          string // เลขประจำตัวผู้เสียภาษีของผู้มีหน้าที่หักภาษี 13 หลัก
	BranchNo     string // 6 หลัก (000000 = สำนักงานใหญ่)
	DeptName     string
	Sections     [3]string // ภ.ง.ด.53: 3 เตรส/65 จัตวา/69 ทวิ; ภ.ง.ด.3: 3 เตรส/48 ทวิ/50(3)(4)(5); ภ.ง.ด.2 ไม่มี
	LTO          string    // "" | "0" | "1"
	TaxMonth     string    // 01-12
	TaxYear      string    // ปี พ.ศ. 4 หลัก
	BranchType   string    // "" | "V" | "S"
	FormType     string    // 00 = ยื่นปกติ, 01-99 = ยื่นเพิ่มเติมครั้งที่
	TotNum       string
	TotAmt       string
	TotTax       string
	SurAmt       string
	GTotTax      string
	TransAmt     string
	UserID       string // ยื่นสื่อ = เลขอ้างอิงการลงทะเบียน
	FormFlag     string
	// SubmissionNo - "ครั้งที่ส่ง (00-99)" ท้ายชื่อไฟล์ (ข้อกำหนดข้อ 3) ไม่อยู่ในเนื้อไฟล์
	SubmissionNo string
	// NilReturn - ฉบับนี้ไม่มีรายการใบแนบและยอดทุกช่องเป็นศูนย์: ยื่นหัวแบบอย่างเดียวได้ (TOT_NUM = 0 ตาม [F53] H18
	// "รวมจำนวนรายของใบแนบ หากไม่มีระบุ 0") — ผู้เรียกตัดสินจากเอกสาร ไม่อยู่ในเนื้อไฟล์
	NilReturn bool
}

// Item - รายการเงินได้หนึ่งรายการ (ภ.ง.ด.53/3 มีได้ 3 รายการต่อบรรทัด, ภ.ง.ด.2 หนึ่งรายการต่อบรรทัด)
type Item struct {
	PaidDate string // ววดดปปปป ปี พ.ศ.
	TaxRate  string // N(4,2)
	PaidAmt  string // N(15,2)
	TaxAmt   string // N(15,2)
	IncType  string // ภ.ง.ด.53/3 = ข้อความประเภทเงินได้; ภ.ง.ด.2 = รหัส 1-5
	PayCon   string // 1 หัก ณ ที่จ่าย, 2 ออกให้ตลอดไป, 3 ออกให้ครั้งเดียว
}

// Empty - รายการที่ไม่มีข้อมูลเลย (เขียนเป็น 00000000|0.00|0.00|0.00|| ตามข้อกำหนด)
func (it Item) Empty() bool {
	return it == Item{}
}

// Address - ที่อยู่ผู้มีเงินได้ 12 ช่องท้ายบรรทัด D ตามลำดับในเอกสาร
type Address struct {
	Building   string // BUILD_NAME
	Room       string // ROOM_NO
	Floor      string // FLOOR_NO
	Village    string // VILLAGE_NAME
	No         string // ADD_NO
	Moo        string // MOO_NO
	Soi        string // SOI
	Street     string // STREET_NAME
	Tambon     string // TAMBON
	Amphur     string // AMPHUR
	Province   string // PROVINCE
	PostalCode string // POSTAL_CODE
}

func (a Address) fields() []string {
	return []string{a.Building, a.Room, a.Floor, a.Village, a.No, a.Moo, a.Soi, a.Street, a.Tambon, a.Amphur, a.Province, a.PostalCode}
}

// Detail - บรรทัด D หนึ่งบรรทัด
type Detail struct {
	// Row - แถวในใบแนบของเอกสาร (เริ่ม 1) ไว้ชี้จุดผิดให้ผู้ใช้ ไม่อยู่ในไฟล์
	Row       int
	Seq       string
	BranchNo  string // สาขาของผู้จ่ายเงินได้ (ผู้หักภาษี) 6 หลัก
	NID       string // ภ.ง.ด.53 NID / ภ.ง.ด.3, ภ.ง.ด.2 PIN ของผู้มีเงินได้
	TIN       string
	AccNo     string // ภ.ง.ด.2 เท่านั้น
	Title     string
	FirstName string
	LastName  string
	Items     [3]Item // ภ.ง.ด.2 ใช้ Items[0] เท่านั้น
	Address   Address
}

// HeaderFields - ช่องของบรรทัด H ตามลำดับในเอกสาร (25 ช่อง ภ.ง.ด.53/3, 22 ช่อง ภ.ง.ด.2)
func HeaderFields(kind Kind, h Header) []string {
	out := []string{"H", h.SenderID, h.SenderNID, h.SenderBranch, h.SenderRole, string(kind), h.NID, h.BranchNo, h.DeptName}
	if kind != PND2 {
		out = append(out, h.Sections[:]...) // H10-H12 ธงมาตรา
	}
	return append(out, h.LTO, h.TaxMonth, h.TaxYear, h.BranchType, h.FormType, h.TotNum,
		h.TotAmt, h.TotTax, h.SurAmt, h.GTotTax, h.TransAmt, h.UserID, h.FormFlag)
}

func itemFields(it Item) []string {
	if it.Empty() {
		return []string{NoDate, ZeroAmount, ZeroAmount, ZeroAmount, "", ""}
	}
	return []string{it.PaidDate, it.TaxRate, it.PaidAmt, it.TaxAmt, it.IncType, it.PayCon}
}

// DetailFields - ช่องของบรรทัด D ตามลำดับในเอกสาร (38 ช่อง ภ.ง.ด.53/3, 27 ช่อง ภ.ง.ด.2)
func DetailFields(kind Kind, d Detail) []string {
	out := []string{"D", d.Seq, d.BranchNo, d.NID, d.TIN}
	if kind == PND2 {
		out = append(out, d.AccNo, d.Title, d.FirstName, d.LastName)
		out = append(out, itemFields(d.Items[0])...)
		return append(out, d.Address.fields()...)
	}
	out = append(out, d.Title, d.FirstName, d.LastName)
	for _, it := range d.Items {
		out = append(out, itemFields(it)...)
	}
	return append(out, d.Address.fields()...)
}

// Summarize - เติม TOT_NUM (จำนวนบรรทัด D ตามไฟล์ตัวอย่างทางการ), TOT_AMT/TOT_TAX (ผลรวมทุกรายการของทุกบรรทัด D)
// และ GTOT_TAX = TOT_TAX + SUR_AMT; ยอดที่รูปแบบผิดไม่ถูกนับ (Check รายงานยอดนั้นแล้ว) — decimal เท่านั้น ห้าม float
// ภ.ง.ด.2 นับเฉพาะ Items[0] — บรรทัด D ของ ภ.ง.ด.2 มีรายการเดียว (DetailFields/Check ใช้ Items[0] เท่านั้น) ยอดหัวแบบจึงต้องเท่ากับเนื้อไฟล์
func Summarize(kind Kind, h Header, ds []Detail) Header {
	amount, tax := decimal.Zero, decimal.Zero
	for _, d := range ds {
		items := d.Items[:]
		if kind == PND2 {
			items = d.Items[:1]
		}
		for _, it := range items {
			amount = amount.Add(parseMoney(it.PaidAmt))
			tax = tax.Add(parseMoney(it.TaxAmt))
		}
	}
	h.TotNum = strconv.Itoa(len(ds))
	h.TotAmt = amount.StringFixed(2)
	h.TotTax = tax.StringFixed(2)
	if h.SurAmt == "" {
		h.SurAmt = ZeroAmount
	}
	if h.TransAmt == "" {
		h.TransAmt = ZeroAmount
	}
	h.GTotTax = tax.Add(parseMoney(h.SurAmt)).StringFixed(2)
	return h
}

// parseMoney - ยอดในรูปแบบของไฟล์ (ตัวเลข.2 หลัก) ไม่งั้นเป็น 0
func parseMoney(s string) decimal.Decimal {
	if !moneyPattern.MatchString(s) {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	return d
}

// Trim - ตัดช่องว่างหัว/ท้ายของทุกช่องข้อความก่อน Check/Encode: ช่องที่มีแต่ช่องว่างกลายเป็นช่องว่างจริง ("||")
// ตามข้อกำหนด "กรณีไม่มี ไม่ต้องเว้นช่องว่าง หรือไม่ระบุค่าใดๆ" — ช่องว่างกลางข้อความคงไว้ตามที่ผู้ใช้กรอก
func Trim(h Header, ds []Detail) (Header, []Detail) {
	t := strings.TrimSpace
	h.SenderID, h.SenderNID, h.SenderBranch, h.SenderRole = t(h.SenderID), t(h.SenderNID), t(h.SenderBranch), t(h.SenderRole)
	h.NID, h.BranchNo, h.DeptName, h.LTO = t(h.NID), t(h.BranchNo), t(h.DeptName), t(h.LTO)
	for i := range h.Sections {
		h.Sections[i] = t(h.Sections[i])
	}
	h.TaxMonth, h.TaxYear, h.BranchType, h.FormType = t(h.TaxMonth), t(h.TaxYear), t(h.BranchType), t(h.FormType)
	h.TotNum, h.TotAmt, h.TotTax, h.SurAmt, h.GTotTax = t(h.TotNum), t(h.TotAmt), t(h.TotTax), t(h.SurAmt), t(h.GTotTax)
	h.TransAmt, h.UserID, h.FormFlag, h.SubmissionNo = t(h.TransAmt), t(h.UserID), t(h.FormFlag), t(h.SubmissionNo)
	out := make([]Detail, len(ds))
	for i, d := range ds {
		d.Seq, d.BranchNo, d.NID, d.TIN, d.AccNo = t(d.Seq), t(d.BranchNo), t(d.NID), t(d.TIN), t(d.AccNo)
		d.Title, d.FirstName, d.LastName = t(d.Title), t(d.FirstName), t(d.LastName)
		for k, it := range d.Items {
			d.Items[k] = Item{PaidDate: t(it.PaidDate), TaxRate: t(it.TaxRate), PaidAmt: t(it.PaidAmt), TaxAmt: t(it.TaxAmt),
				IncType: t(it.IncType), PayCon: t(it.PayCon)}
		}
		a := &d.Address
		a.Building, a.Room, a.Floor, a.Village, a.No, a.Moo = t(a.Building), t(a.Room), t(a.Floor), t(a.Village), t(a.No), t(a.Moo)
		a.Soi, a.Street, a.Tambon, a.Amphur, a.Province, a.PostalCode = t(a.Soi), t(a.Street), t(a.Tambon), t(a.Amphur), t(a.Province), t(a.PostalCode)
		out[i] = d
	}
	return h, out
}

// Lines - H หนึ่งบรรทัด แล้วตามด้วย D ทุกบรรทัด (ไม่มีบรรทัดท้าย — ข้อกำหนดข้อ 4)
func Lines(kind Kind, h Header, ds []Detail) [][]string {
	out := make([][]string, 0, len(ds)+1)
	out = append(out, HeaderFields(kind, h))
	for _, d := range ds {
		out = append(out, DetailFields(kind, d))
	}
	return out
}

// Encode - UTF-8 มี BOM (ตรงไฟล์ตัวอย่างทางการ) คั่นช่องด้วย "|" ไม่มี pipe หัว/ท้าย ขึ้นบรรทัดด้วย CRLF
// ไม่มี CRLF หลังบรรทัดสุดท้าย (ตรงไฟล์ตัวอย่าง) และไม่ตัดช่องว่างหรือเติมค่าใด ๆ — เขียนตามที่ได้รับ (ตรวจค่าที่ Check)
func Encode(lines [][]string) []byte {
	var b bytes.Buffer
	b.Write(utf8BOM)
	for i, fields := range lines {
		if i > 0 {
			b.WriteString("\r\n")
		}
		b.WriteString(strings.Join(fields, "|"))
	}
	return b.Bytes()
}

// FileName - TAX_TYPE_NID_BRANCH_NO_TAX_YEAR_TAX_MONTH_FORM_TYPE_ครั้งที่ส่ง.txt (ข้อกำหนดข้อ 3)
// เช่น PND53_0105555555555_000000_2569_08_00_00.txt
func FileName(kind Kind, h Header) string {
	return strings.Join([]string{string(kind), h.NID, h.BranchNo, h.TaxYear, h.TaxMonth, h.FormType, h.SubmissionNo}, "_") + ".txt"
}
