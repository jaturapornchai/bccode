package handlers

import (
	"strconv"
	"strings"

	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/rdform"
	"smlcloudplatform/internal/whtcert"
)

// ยอดรวม/ยอดสุทธิของแบบคำนวณตามสูตรที่พิมพ์บนแบบฟอร์มของกรมสรรพากรเท่านั้น (เช่น ภ.พ.30 ข้อ 4 = 1 - 2 - 3)
// ยอดรายการและฐานภาษีเป็นค่าที่ผู้ใช้กรอก/แก้ได้เสมอ — ระบบไม่คำนวณย้อนหรือเดาฐานภาษี
// ทุกยอดเป็น decimal ทศนิยม 2 ตำแหน่ง (ห้าม float); แบบที่ไม่มีในตารางนี้ไม่มีบรรทัดรวมอัตโนมัติ

var taxFormComputers = map[string]func(*rdform.Spec, *rdform.Document) error{
	"pnd3":  computeWithholdingCover,
	"pnd53": computeWithholdingCover,
	"pnd2":  computePnd2Cover,
	"pnd2a": computePnd2Cover,
	"pp30":  computePP30,
	"pp36":  computePP36,
	"pbt40": computePBT40,
	"pnd50": computePND50,
	"pnd51": computePND51,
}

func computeTaxForm(code string, doc *rdform.Document) error {
	if doc.Values == nil {
		doc.Values = map[string]string{}
	}
	fn := taxFormComputers[code]
	if fn == nil {
		return nil
	}
	cover, err := rdform.Lookup(code)
	if err != nil {
		return err
	}
	return fn(cover, doc)
}

// amounts - ตัวอ่านยอดเงินจาก map (ว่าง = 0) ที่จำ error แรกไว้ ให้สูตรเขียนอ่านง่ายเป็นบรรทัดเดียว
type amounts struct {
	values map[string]string
	err    error
}

func (a *amounts) get(key string) decimal.Decimal {
	raw := strings.ReplaceAll(strings.TrimSpace(a.values[key]), ",", "")
	if raw == "" {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(raw)
	if err != nil && a.err == nil {
		a.err = &rdform.FieldError{Key: key}
	}
	return d
}

func (a *amounts) set(key string, d decimal.Decimal) { a.values[key] = d.StringFixed(2) }

// setPositive - บรรทัด "ถ้า ... มากกว่า ..." : มีค่าเมื่อเป็นบวก ไม่งั้นว่าง
func (a *amounts) setPositive(key string, d decimal.Decimal) {
	if d.IsPositive() {
		a.set(key, d)
	} else {
		delete(a.values, key)
	}
}

// sumRows - รวมคอลัมน์ที่ชื่อ = name หรือ l<n>_name ของทุกแถว
func sumRows(rows []map[string]string, name string) (decimal.Decimal, error) {
	total := decimal.Zero
	for i, r := range rows {
		a := amounts{values: r}
		for k := range r {
			if k == name || (strings.HasPrefix(k, "l") && strings.HasSuffix(k, "_"+name)) {
				total = total.Add(a.get(k))
			}
		}
		if a.err != nil {
			fe := a.err.(*rdform.FieldError)
			fe.Row = i + 1
			return total, fe
		}
	}
	return total, nil
}

// computeWithholdingCover - ภ.ง.ด.3/53: 1. รวมยอดเงินได้ 2. รวมภาษี (จากใบแนบ) 4. = 2. + 3.
// ยื่นทางสื่อบันทึก (ไม่มีรายการใบแนบในระบบ) ใช้ยอดที่ผู้ใช้กรอกเอง
func computeWithholdingCover(_ *rdform.Spec, doc *rdform.Document) error {
	a := &amounts{values: doc.Values}
	if len(doc.Rows) > 0 {
		income, err := sumRows(doc.Rows, "amount")
		if err != nil {
			return err
		}
		tax, err := sumRows(doc.Rows, "tax")
		if err != nil {
			return err
		}
		a.set("total_income", income)
		a.set("total_tax", tax)
	}
	a.set("total_payable", a.get("total_tax").Add(a.get("surcharge")))
	return a.err
}

var pnd2Kinds = []string{"royalty", "interest", "dividend", "share_transfer", "other_404"}

// computePnd2Cover - ภ.ง.ด.2/2ก: จำนวนราย/เงินได้/ภาษี แยกตามประเภทเงินได้ของใบแนบ แล้ว 6. รวม, 8. = 6. + 7.
// (ภ.ง.ด.2ก ไม่มีบรรทัด 40(3) และไม่มีเงินเพิ่ม — ข้ามช่องที่แบบไม่มี)
func computePnd2Cover(cover *rdform.Spec, doc *rdform.Document) error {
	a := &amounts{values: doc.Values}
	if len(doc.Rows) > 0 {
		for _, kind := range pnd2Kinds {
			var rows []map[string]string
			for _, r := range doc.Rows {
				if r["income_type"] == kind {
					rows = append(rows, r)
				}
			}
			income, err := sumRows(rows, "amount")
			if err != nil {
				return err
			}
			tax, err := sumRows(rows, "tax")
			if err != nil {
				return err
			}
			if cover.Field(kind+"_count") != nil {
				doc.Values[kind+"_count"] = strconv.Itoa(len(rows))
				a.set(kind+"_income", income)
				a.set(kind+"_tax", tax)
			}
		}
		// 6. รวม = ทุกแถวของใบแนบ รวมแถวที่ยังไม่เลือกประเภทเงินได้ (ระบบไม่เดาประเภทให้) — ยอดภาษีที่นำส่งต้องไม่ตกหล่นเงียบ ๆ
		totalIncome, err := sumRows(doc.Rows, "amount")
		if err != nil {
			return err
		}
		totalTax, err := sumRows(doc.Rows, "tax")
		if err != nil {
			return err
		}
		doc.Values["total_count"] = strconv.Itoa(len(doc.Rows))
		a.set("total_income", totalIncome)
		a.set("total_tax", totalTax)
	}
	if cover.Field("total_payable") != nil {
		a.set("total_payable", a.get("total_tax").Add(a.get("surcharge")))
	}
	return a.err
}

// pp30RowColumns - ยอดของใบแนบ ภ.พ.30 (ยื่นรวมหลายสาขา) ที่รวมขึ้นหน้าแบบ
var pp30RowColumns = []string{"sales_amount", "sales_zero_rate", "sales_exempt", "sales_taxable", "output_tax", "purchase_amount", "input_tax"}

// pp30AdditionalColumns - ช่องของการยื่นเพิ่มเติม (ยอดขายต่ำไป/ซื้อสูงไป ฯลฯ) รวมขึ้นหน้าแบบเฉพาะเมื่อมีสาขากรอกไว้
// การยื่นปกติต้องเว้นว่าง ไม่ใช่ "0.00"
var pp30AdditionalColumns = []string{"add_sales_under", "add_purchase_over", "add_purchase_under", "add_sales_over"}

func anyRowFilled(rows []map[string]string, key string) bool {
	for _, r := range rows {
		if strings.TrimSpace(r[key]) != "" {
			return true
		}
	}
	return false
}

// computePP30 - ตามสูตรบนแบบ ภ.พ.30: 4 = 1-2-3, 8 = 5-7 (ถ้า 5>7), 9 = 7-5 (ถ้า 5<7), 11 = 8-10 (ถ้า 8>10),
// 12 = (10-8 ถ้า 10>8) หรือ (9+10), 15 = (11+13+14) หรือ (13+14-12), 16 = 12-13-14
func computePP30(_ *rdform.Spec, doc *rdform.Document) error {
	a := &amounts{values: doc.Values}
	for i, r := range doc.Rows {
		ra := &amounts{values: r}
		ra.set("sales_taxable", ra.get("sales_amount").Sub(ra.get("sales_zero_rate")).Sub(ra.get("sales_exempt")))
		ra.set("tax_net", ra.get("output_tax").Sub(ra.get("input_tax")))
		if ra.err != nil {
			fe := ra.err.(*rdform.FieldError)
			fe.Row = i + 1
			return fe
		}
	}
	if len(doc.Rows) > 0 {
		columns := pp30RowColumns
		for _, k := range pp30AdditionalColumns {
			if anyRowFilled(doc.Rows, k) {
				columns = append(columns[:len(columns):len(columns)], k)
			} else {
				// หน้าแบบเป็นผลรวมของใบแนบเมื่อมีสาขา — สาขาล้างช่องยื่นเพิ่มเติมหมดแล้ว ยอดรวมเดิมบนหน้าแบบต้องหายด้วย
				// (เดิมค้างยอดเก่าไว้ ทำให้แบบยื่นปกติมียอด "ยื่นเพิ่มเติม" ที่ไม่มีสาขาใดกรอก)
				delete(doc.Values, k)
			}
		}
		for _, k := range columns {
			total, err := sumRows(doc.Rows, k)
			if err != nil {
				return err
			}
			a.set(k, total)
		}
	}
	a.set("sales_taxable", a.get("sales_amount").Sub(a.get("sales_zero_rate")).Sub(a.get("sales_exempt")))
	diff := a.get("output_tax").Sub(a.get("input_tax"))
	a.setPositive("tax_payable", diff)
	a.setPositive("tax_excess", diff.Neg())
	if diff.IsZero() {
		a.set("tax_payable", decimal.Zero)
	}
	forward := a.get("excess_brought_forward")
	payable, excess := decimal.Zero, decimal.Zero
	if diff.IsPositive() {
		payable = diff.Sub(forward)
		if payable.IsNegative() {
			payable, excess = decimal.Zero, payable.Neg()
		}
	} else {
		excess = diff.Neg().Add(forward)
	}
	a.setPositive("net_payable", payable)
	a.setPositive("net_excess", excess)
	penalties := a.get("surcharge").Add(a.get("penalty"))
	delete(doc.Values, "total_payable")
	delete(doc.Values, "total_excess")
	delete(doc.Values, "net_result")
	switch {
	case payable.IsPositive():
		doc.Values["net_result"] = "payable"
		a.set("total_payable", payable.Add(penalties))
	case excess.IsPositive():
		doc.Values["net_result"] = "excess"
		a.setPositive("total_payable", penalties.Sub(excess))
		a.setPositive("total_excess", excess.Sub(penalties))
	default:
		a.setPositive("total_payable", penalties)
	}
	return a.err
}

// computePP36 - 5. = 2. + 3. + 4. และจำนวนเงินเป็นตัวอักษรของข้อ 2 และ 5
func computePP36(_ *rdform.Spec, doc *rdform.Document) error {
	a := &amounts{values: doc.Values}
	vat := a.get("vat_amount")
	total := vat.Add(a.get("surcharge")).Add(a.get("penalty"))
	if strings.TrimSpace(doc.Values["vat_amount"]) != "" {
		doc.Values["vat_amount_text"] = whtcert.BahtText(vat)
	}
	if total.IsPositive() {
		a.set("total_payable", total)
		doc.Values["total_payable_text"] = whtcert.BahtText(total)
	}
	return a.err
}

// ภ.ง.ด.50 รายการที่ 1 ข้อ 3–6 / ภ.ง.ด.51 รายการที่ 2 ข้อ 5–8 ตามคู่มือวิธีกรอกแบบของกรมสรรพากร (docs/kms/21-thai-tax-form-references.md §4):
// รวมเครดิต = ผลรวมช่อง "หัก" ทุกช่อง, คงเหลือ = ภาษีที่คำนวณได้ − รวมเครดิต (บวก = ชำระเพิ่มเติม, ลบ = ชำระไว้เกิน), รวมภาษี = คงเหลือ + เงินเพิ่ม
// ภาษีที่คำนวณได้ (อัตราตามกรณีของกิจการ) และเงินเพิ่ม (ลดได้ตามระเบียบอธิบดี) ผู้ใช้กรอกเอง — ระบบไม่เดาอัตรา
var pnd50Credits = []string{"less_exempt_tax_rd18_463", "less_exempt_tax_rd300", "less_wht", "less_pnd51_paid", "less_reduced_rate_tax", "less_pnd50_paid"}
var pnd51Credits = []string{"r2_5_1_wht", "r2_5_2_reduced_rate_relief", "r2_5_3_prior_pnd51_paid"}

func computePND50(_ *rdform.Spec, doc *rdform.Document) error {
	a := &amounts{values: doc.Values}
	credits := sumCredits(a, pnd50Credits, "less_total")
	settleTax(a, "tax_computed", credits, "tax_balance", "tax_balance_type", "surcharge", "tax_net", "tax_net_type")
	return a.err
}

func computePND51(_ *rdform.Spec, doc *rdform.Document) error {
	a := &amounts{values: doc.Values}
	credits := sumCredits(a, pnd51Credits, "r2_5_total_credits")
	settleTax(a, "r2_4_tax_computed", credits, "r2_6_balance", "r2_6_sign", "r2_7_surcharge", "r2_8_total", "r2_8_sign")
	return a.err
}

// sumCredits - ช่องรวมเครดิตมีค่าเมื่อผู้ใช้/ระบบกรอกช่องเครดิตอย่างน้อยหนึ่งช่อง (ไม่มีเลย = ว่าง)
func sumCredits(a *amounts, keys []string, totalKey string) decimal.Decimal {
	total, filled := decimal.Zero, false
	for _, k := range keys {
		if strings.TrimSpace(a.values[k]) != "" {
			total, filled = total.Add(a.get(k)), true
		}
	}
	if filled {
		a.set(totalKey, total)
	} else {
		delete(a.values, totalKey)
	}
	return total
}

// settleTax - คำนวณเมื่อกรอกภาษีที่คำนวณได้แล้วเท่านั้น (ช่องว่าง ≠ 0: ไม่ให้แบบขึ้น "ชำระไว้เกิน" ทั้งที่ยังไม่กรอก)
func settleTax(a *amounts, taxKey string, credits decimal.Decimal, balanceKey, balanceSign, surchargeKey, netKey, netSign string) {
	if strings.TrimSpace(a.values[taxKey]) == "" {
		return
	}
	balance := a.get(taxKey).Sub(credits)
	setSigned(a, balanceKey, balanceSign, balance)
	setSigned(a, netKey, netSign, balance.Add(a.get(surchargeKey)))
}

func setSigned(a *amounts, key, signKey string, d decimal.Decimal) {
	a.set(key, d.Abs())
	if d.IsNegative() {
		a.values[signKey] = "overpaid"
	} else {
		a.values[signKey] = "payable"
	}
}

// computePBT40 - 12. รวมภาษีธุรกิจเฉพาะ = ผลรวมช่องภาษีทุกประเภทกิจการ, 15. = 12+13+14, 17. = 15+16
// ใบแนบรายสถานประกอบการ: ยอดรวมภาษีของแผ่น และหน้าแบบ = ผลรวมทุกแผ่น
// ข้อ 16 รายได้ส่วนท้องถิ่น ผู้ใช้กรอกตามอัตราที่กฎหมายกำหนด (ระบบไม่ใส่อัตราเอง)
func computePBT40(_ *rdform.Spec, doc *rdform.Document) error {
	a := &amounts{values: doc.Values}
	for i, s := range doc.Sheets {
		sa := &amounts{values: s}
		sa.set("est_total_tax", sumSuffix(sa, "_tax"))
		if sa.err != nil {
			return &rdform.FieldError{Key: sa.err.(*rdform.FieldError).Key, Row: i + 1}
		}
	}
	if len(doc.Sheets) > 0 {
		keys := map[string]bool{}
		for _, s := range doc.Sheets {
			for k := range s {
				if strings.HasPrefix(k, "biz") && (strings.HasSuffix(k, "_receipts") || strings.HasSuffix(k, "_tax")) {
					keys[k] = true
				}
			}
		}
		for k := range keys {
			total, err := sumRows(doc.Sheets, k)
			if err != nil {
				return err
			}
			a.set(k, total)
		}
	}
	sbt := sumSuffix(a, "_tax")
	a.set("total_sbt_tax", sbt)
	withPenalty := sbt.Add(a.get("surcharge")).Add(a.get("penalty"))
	a.set("total_with_surcharge_penalty", withPenalty)
	a.set("total_payable", withPenalty.Add(a.get("local_tax")))
	return a.err
}

// sumSuffix - รวมช่องประเภทกิจการ biz*_<suffix>
func sumSuffix(a *amounts, suffix string) decimal.Decimal {
	total := decimal.Zero
	for k := range a.values {
		if strings.HasPrefix(k, "biz") && strings.HasSuffix(k, suffix) {
			total = total.Add(a.get(k))
		}
	}
	return total
}
