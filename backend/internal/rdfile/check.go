package rdfile

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/whtcert"
)

// Issue - จุดที่ทำให้สร้างไฟล์ไม่ได้: Key = key ภาษา (languages.tsv), Field = ช่องของเอกสาร/แผงสร้างไฟล์ที่ต้องแก้,
// Row = แถวในใบแนบ (0 = หัวแบบ), Args = ค่าที่แทนในข้อความ ({char} อักขระที่พบ, {max} ความยาวสูงสุด)
type Issue struct {
	Key   string            `json:"key"`
	Field string            `json:"field,omitempty"`
	Row   int               `json:"row,omitempty"`
	Args  map[string]string `json:"args,omitempty"`
}

var (
	// N(15,2): จุดทศนิยม 2 ตำแหน่ง ไม่มีคอมมา ยาวรวมไม่เกิน 18 ตัวอักษร (ข้อกำหนดข้อ 8) — ติดลบบล็อกไว้ก่อนในรอบนี้
	moneyPattern = regexp.MustCompile(`^[0-9]{1,15}\.[0-9]{2}$`)
	// N(4,2) อัตราภาษี เช่น 10.00
	ratePattern   = regexp.MustCompile(`^[0-9]{1,4}\.[0-9]{2}$`)
	digitsPattern = regexp.MustCompile(`^[0-9]+$`)
	hundred       = decimal.NewFromInt(100)
)

// forbiddenChars - อักขระพิเศษต้องห้าม (ข้อกำหนดข้อ 10: * + / \ ! $ % # & @ คอมมา single quote double quote)
// รวมเครื่องหมายคำพูดแบบโค้ง และ "|" ที่เป็นตัวคั่นช่อง — พบแล้วบล็อกให้ผู้ใช้แก้ ไม่แทนค่าเอง
const forbiddenChars = `*+/\!$%#&@,'"‘’“”|`

// Check - ตรวจค่าทั้งไฟล์ตามข้อกำหนดของ Format กลาง รวบรวมทุกจุดผิด (หัวแบบก่อน แล้วเรียงตามบรรทัด D)
func Check(kind Kind, h Header, ds []Detail) []Issue {
	c := &checker{}
	if kind != PND53 && kind != PND3 && kind != PND2 {
		c.add("tax_rdfile_unsupported_form", "", 0, nil)
		return c.issues
	}
	if len(ds) == 0 && !h.NilReturn {
		c.add("tax_rdfile_no_rows", "", 0, nil)
	}
	if !whtcert.ValidThaiTaxID(h.NID) {
		c.add("tax_rdfile_company_taxid", "tax_id", 0, nil)
	}
	if !exactDigits(h.BranchNo, 6) {
		c.add("tax_rdfile_branch_required", "branch_no", 0, nil)
	}
	c.text(h.DeptName, 80, "tax_rdfile_dept_name_required", "dept_name", 0)
	if kind != PND2 {
		selected := false
		for _, s := range h.Sections {
			selected = selected || s == "1"
			if s != "0" && s != "1" {
				selected = false
				break
			}
		}
		if !selected {
			c.add("tax_rdfile_section_required", "tax_section", 0, nil)
		}
	}
	if !oneOf(h.LTO, "", "0", "1") {
		c.add("tax_rdfile_value_invalid", "lto", 0, nil)
	}
	if !oneOf(h.BranchType, "", "V", "S") {
		c.add("tax_rdfile_value_invalid", "branch_type", 0, nil)
	}
	if !exactDigits(h.FormType, 2) {
		c.add("tax_rdfile_value_invalid", "additional_no", 0, nil)
	}
	c.text(h.UserID, 20, "tax_rdfile_user_id_required", "media_ref_no", 0)
	c.money(h.SurAmt, "surcharge", 0)
	c.money(h.TotAmt, "total_income", 0)
	c.money(h.TotTax, "total_tax", 0)
	c.money(h.GTotTax, "total_payable", 0)
	if !exactDigits(h.SubmissionNo, 2) {
		c.add("tax_rdfile_submission_invalid", "submission_no", 0, nil)
	}
	for _, d := range ds {
		c.detail(kind, h, d)
	}
	return c.issues
}

type checker struct {
	issues []Issue
}

func (c *checker) add(key, field string, row int, args map[string]string) {
	c.issues = append(c.issues, Issue{Key: key, Field: field, Row: row, Args: args})
}

// text - ช่องข้อความ (C): บังคับกรอกเมื่อ requiredKey ไม่ว่าง, ไม่มีอักขระต้องห้าม, ยาวไม่เกิน max ตัวอักษร (นับ rune)
func (c *checker) text(value string, max int, requiredKey, field string, row int) {
	if strings.TrimSpace(value) == "" {
		if requiredKey != "" {
			c.add(requiredKey, field, row, nil)
		}
		return
	}
	if ch, bad := forbiddenRune(value); bad {
		c.add("tax_rdfile_forbidden_char", field, row, map[string]string{"char": ch})
	}
	if utf8.RuneCountInString(value) > max {
		c.add("tax_rdfile_too_long", field, row, map[string]string{"max": strconv.Itoa(max)})
	}
}

// money - ช่องเงินที่ต้องมีค่าในรูป N(15,2): ติดลบ ทศนิยมเกิน 2 ตำแหน่ง มีคอมมา หรือยาวเกิน 18 ตัวอักษร = ผิด
func (c *checker) money(value, field string, row int) {
	if !moneyPattern.MatchString(value) {
		c.add("tax_rdfile_amount_invalid", field, row, nil)
	}
}

// forbiddenRune - อักขระต้องห้ามตัวแรกที่พบ (อักขระควบคุมรายงานเป็น U+XXXX เพราะมองไม่เห็นบนจอ)
func forbiddenRune(s string) (string, bool) {
	for _, r := range s {
		if strings.ContainsRune(forbiddenChars, r) {
			return string(r), true
		}
		if r < 0x20 || r == 0x7f {
			return fmt.Sprintf("U+%04X", r), true
		}
	}
	return "", false
}

func (c *checker) detail(kind Kind, h Header, d Detail) {
	row := d.Row
	zeroPINAllowed := kind == PND2 && d.Items[0].IncType == "2" && d.NID == "0000000000000" // [F2] D4: เฉพาะดอกเบี้ย 40(4)(ก)
	if !whtcert.ValidThaiTaxID(d.NID) && !zeroPINAllowed {
		c.add("tax_rdfile_taxid_invalid", "tax_id", row, nil)
	}
	// ภ.ง.ด.3 "กรณีไม่มีให้ใส่ขีด -" — ผู้ใช้พิมพ์เอง ระบบไม่เติมให้
	c.text(d.Title, 100, "tax_rdfile_title_required", "title", row)
	c.text(d.FirstName, 100, "tax_rdfile_name_required", "name", row)
	c.text(d.LastName, 80, "", "surname", row)
	if kind == PND2 {
		// [F2] D6 ACC_NO (M/O): "กรณี มีเงินได้ตาม มาตรา 40(4) (ก) ดอกเบี้ยเงินฝาก และตาม มาตรา 40(4)(ข) ให้ระบุเลขที่บัญชีเงินฝากธนาคาร"
		// บังคับเฉพาะรหัส 3 = 40(4)(ข) เงินปันผล; รหัส 2 = 40(4)(ก) รวมดอกเบี้ยพันธบัตร/ตั๋วเงิน ([F2] INC_TYPE_PND)
		// แยกไม่ได้ว่าเป็นดอกเบี้ยเงินฝากหรือไม่ จึงไม่บังคับ (ส่งตามที่กรอก) — ห้ามเดาจากชื่อรายการ
		accRequired := ""
		if d.Items[0].IncType == "3" {
			accRequired = "tax_rdfile_account_required"
		}
		c.text(d.AccNo, 15, accRequired, "bank_account_no", row)
	}
	c.address(kind, d.Address, row)

	items := 3
	if kind == PND2 {
		items = 1
	}
	if d.Items[0].Empty() {
		c.add("tax_rdfile_item1_required", "l1_date", row, nil)
	}
	for i := 0; i < items; i++ {
		if !d.Items[i].Empty() {
			c.item(kind, h, d.Items[i], i+1, row)
		}
	}
}

// address - ภ.ง.ด.3 บังคับอำเภอ/จังหวัด/รหัสไปรษณีย์ 5 หลัก ([F3] ข้อ 36-38 เป็น M); ช่องอื่นไม่บังคับทุกแบบ
func (c *checker) address(kind Kind, a Address, row int) {
	required := ""
	if kind == PND3 {
		required = "tax_rdfile_address_required"
	}
	c.text(a.Building, 40, "", "addr_building", row)
	c.text(a.Room, 20, "", "addr_room", row)
	c.text(a.Floor, 20, "", "addr_floor", row)
	c.text(a.Village, 100, "", "addr_village", row)
	c.text(a.No, 20, "", "addr_no", row)
	c.text(a.Moo, 20, "", "addr_moo", row)
	c.text(a.Soi, 100, "", "addr_soi", row)
	c.text(a.Street, 100, "", "addr_road", row)
	c.text(a.Tambon, 50, "", "addr_subdistrict", row)
	c.text(a.Amphur, 50, required, "addr_district", row)
	c.text(a.Province, 50, required, "addr_province", row)
	switch {
	case a.PostalCode == "":
		if required != "" {
			c.add(required, "addr_postcode", row, nil)
		}
	case !exactDigits(a.PostalCode, 5):
		c.add("tax_rdfile_postcode_invalid", "addr_postcode", row, nil)
	}
}

// item - รายการเงินได้ที่มีข้อมูลต้องครบ: วันที่จ่ายในเดือนภาษี, อัตรา, ยอดเงิน ≥ 0, ภาษี, ประเภทเงินได้, เงื่อนไข
func (c *checker) item(kind Kind, h Header, it Item, n, row int) {
	p := "l" + strconv.Itoa(n) + "_"
	switch {
	case !realDate(it.PaidDate):
		c.add("tax_rdfile_date_invalid", p+"date", row, nil)
	case it.PaidDate[2:4] != h.TaxMonth || it.PaidDate[4:] != h.TaxYear:
		// งวดของภาษีหัก = เดือนที่จ่ายเงินได้ (docs/kms/21 §9: RD-MOF-WHT-EXT, RD-WHT-GUIDE-3-53)
		c.add("tax_rdfile_date_outside_month", p+"date", row, nil)
	}
	if !validRate(it.TaxRate) {
		c.add("tax_rdfile_rate_invalid", p+"rate", row, nil)
	}
	c.money(it.PaidAmt, p+"amount", row)
	c.money(it.TaxAmt, p+"tax", row)
	if kind == PND2 {
		if !oneOf(it.IncType, "1", "2", "3", "4", "5") {
			c.add("tax_rdfile_income_type_required", "income_type", row, nil)
		}
	} else {
		c.text(it.IncType, 100, "tax_rdfile_income_type_required", p+"income_type", row)
	}
	if !oneOf(it.PayCon, "1", "2", "3") {
		c.add("tax_rdfile_condition_invalid", p+"condition", row, nil)
	}
}

// realDate - ววดดปปปป ปี พ.ศ. ที่เป็นวันที่มีอยู่จริง (31/02 ไม่ผ่าน)
func realDate(s string) bool {
	if !exactDigits(s, 8) {
		return false
	}
	day, _ := strconv.Atoi(s[:2])
	month, _ := strconv.Atoi(s[2:4])
	year, _ := strconv.Atoi(s[4:])
	if month < 1 || month > 12 || day < 1 || year <= 543 {
		return false
	}
	d := time.Date(year-543, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return d.Day() == day && int(d.Month()) == month
}

func validRate(s string) bool {
	if !ratePattern.MatchString(s) {
		return false
	}
	d, err := decimal.NewFromString(s)
	return err == nil && d.LessThanOrEqual(hundred)
}

func exactDigits(s string, n int) bool {
	return len(s) == n && digitsPattern.MatchString(s)
}

func oneOf(s string, options ...string) bool {
	for _, o := range options {
		if s == o {
			return true
		}
	}
	return false
}
