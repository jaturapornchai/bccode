package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/internal/rdform"
)

// TaxFormNote - ข้อความเตือนหลังเติมค่า (key ภาษา + จำนวนรายการ/ยอดอ้างอิงถ้ามี) ให้ผู้ใช้ตรวจก่อนยื่น
type TaxFormNote struct {
	Key    string `json:"key"`
	Count  int    `json:"count,omitempty"`
	Amount string `json:"amount,omitempty"`
}

// profileKeys - หัวแบบที่ยกจากฉบับล่าสุด: ทะเบียนบริษัทยังไม่มีที่อยู่/สาขา/ผู้ลงนาม ผู้ใช้จึงกรอกครั้งแรกครั้งเดียว
// media_ref_no = เลขอ้างอิงการลงทะเบียนยื่นด้วยสื่อ (USER_ID ของไฟล์ยื่นกรมสรรพากร) กรอกครั้งเดียวแล้วยกไปฉบับถัดไป
var profileKeys = []string{"branch_no", "establishment_name", "name_line2", "addr_building", "addr_room", "addr_floor",
	"addr_village", "addr_no", "addr_moo", "addr_soi", "addr_junction", "addr_road", "addr_subdistrict", "addr_district",
	"addr_province", "addr_postcode", "phone", "website", "signer_name", "signer_position", "signer2_name", "signer2_position",
	"media_ref_no"}

// ictZone - เวลาประเทศไทย (ไม่มี DST) ไม่พึ่ง tzdata ใน container
var ictZone = time.FixedZone("ICT", 7*3600)

var thaiShortMonths = [12]string{"ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.", "ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค."}

// formFiller - ตัวช่วยเติมค่าเฉพาะช่องที่แบบนี้มีจริง (Validate ปฏิเสธ key ที่แบบไม่มี)
type formFiller struct {
	cover  *rdform.Spec
	keys   map[string]bool
	values map[string]string
}

func newFormFiller(code string) (*formFiller, error) {
	cover, err := rdform.Lookup(code)
	if err != nil {
		return nil, err
	}
	schema, err := rdform.SchemaFor(code)
	if err != nil {
		return nil, err
	}
	f := &formFiller{cover: cover, keys: map[string]bool{}, values: map[string]string{}}
	for _, field := range schema.Fields {
		f.keys[field.Key] = true
	}
	return f, nil
}

func (f *formFiller) set(key, value string) {
	if f.keys[key] && strings.TrimSpace(value) != "" {
		f.values[key] = value
	}
}

// digitWidth - จำนวนช่องตัวเลขของช่อง (0 = ช่องข้อความ)
func (f *formFiller) digitWidth(key string) int {
	field := f.cover.Field(key)
	if field == nil || field.Type != "digits" {
		return 0
	}
	if len(field.Cells) > 0 {
		return len(field.Cells)
	}
	return field.Comb
}

// setYear - ปี พ.ศ.; ช่องที่พิมพ์ "25" ไว้แล้วมีแค่ 2 ช่อง → 2 หลักท้าย
func (f *formFiller) setYear(key string, yearCE int) {
	be := strconv.Itoa(yearCE + 543)
	if n := f.digitWidth(key); n > 0 && n < len(be) {
		be = be[len(be)-n:]
	}
	f.set(key, be)
}

// setDay / setMonth - ช่องตัวเลขใส่ 2 หลัก, ช่องข้อความใส่ชื่อเดือน (ช่องแคบใช้ชื่อย่อ)
func (f *formFiller) setDay(key string, day int) {
	if f.digitWidth(key) >= 2 {
		f.set(key, fmt.Sprintf("%02d", day))
		return
	}
	f.set(key, strconv.Itoa(day))
}

func (f *formFiller) setMonth(key string, month int) {
	if f.digitWidth(key) >= 2 {
		f.set(key, fmt.Sprintf("%02d", month))
		return
	}
	if field := f.cover.Field(key); field != nil && field.Rect[2]-field.Rect[0] < 40 {
		f.set(key, thaiShortMonths[month-1])
		return
	}
	f.set(key, rdform.ThaiMonth(month))
}

func (f *formFiller) setDate(prefix string, d time.Time) {
	f.setDay(prefix+"_day", d.Day())
	f.setMonth(prefix+"_month", int(d.Month()))
	f.setYear(prefix+"_year_be", d.Year())
}

// prefillTaxForm - ค่าเริ่มต้นของแบบ: หัวแบบ (ฉบับก่อน + ทะเบียนบริษัท) งวด วันที่ลงนาม และยอดจากบัญชีแยกประเภท
func prefillTaxForm(ctx context.Context, db *sql.DB, holding, company, code string, year, month int, now time.Time) (rdform.Document, []TaxFormNote, error) {
	f, err := newFormFiller(code)
	if err != nil {
		return rdform.Document{}, nil, err
	}
	meta := taxForms[code]
	if err := f.copyProfile(ctx, db, company, code, meta.Individual); err != nil {
		return rdform.Document{}, nil, err
	}
	if !meta.Individual {
		header, err := whtCompanyHeader(ctx, holding, company)
		if err != nil {
			return rdform.Document{}, nil, err
		}
		if id := digitsOnly(header.TaxID); len(id) == 13 {
			f.set("tax_id", id)
		}
		f.set("name", header.Name)
		if f.values["branch_no"] == "" {
			f.set("branch_no", "00000") // ทะเบียนยังไม่มีสาขา: สำนักงานใหญ่ แก้ได้
		}
	}
	f.set("filing_type", "normal")
	if month > 0 {
		f.set("month", strconv.Itoa(month))
	}
	f.setYear("year_be", year)
	f.setYear("tax_year_be", year)
	today := now.In(ictZone)
	f.setDate("sign", today)
	f.set("sign_date", fmt.Sprintf("%d %s %d", today.Day(), rdform.ThaiMonth(int(today.Month())), today.Year()+543))

	doc := rdform.Document{Values: f.values}
	var notes []TaxFormNote
	switch meta.Source {
	case "wht":
		notes, err = fillWithholdingForm(ctx, db, company, code, year, month, &doc)
	case "vat":
		notes, err = fillVatForm(ctx, db, company, year, month, f)
	case "cit":
		notes, err = fillCitForm(ctx, db, company, code, year, f)
	default:
		notes = []TaxFormNote{{Key: "tax_form_note_manual"}}
	}
	if err != nil {
		return rdform.Document{}, nil, err
	}
	if err := computeTaxForm(code, &doc); err != nil {
		return rdform.Document{}, nil, err
	}
	return doc, notes, nil
}

// copyProfile - ที่อยู่/ผู้ลงนามจากฉบับล่าสุดของบริษัท เฉพาะแบบตระกูลเดียวกัน: แบบของบริษัทยกจากแบบของบริษัทเท่านั้น
// (ผู้ยื่น ภ.ง.ด.93/94 เป็นบุคคลธรรมดา ที่อยู่/ผู้ลงนามคนละรายกับบริษัท ห้ามปนข้ามกัน);
// แบบบุคคลธรรมดายกเฉพาะจากแบบเดียวกัน พร้อมชื่อ/เลขผู้เสียภาษี
func (f *formFiller) copyProfile(ctx context.Context, db *sql.DB, company, code string, individual bool) error {
	family := []string{code}
	keys := profileKeys
	if individual {
		keys = append([]string{"tax_id", "name"}, profileKeys...)
	} else {
		family = family[:0]
		for form, meta := range taxForms {
			if !meta.Individual {
				family = append(family, form)
			}
		}
	}
	var raw []byte
	err := db.QueryRowContext(ctx, `SELECT document FROM tax_filings WHERE company_code=$1 AND form_code = ANY($2)
ORDER BY updated_at DESC LIMIT 1`, company, pq.StringArray(family)).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load last tax filing: %w", err)
	}
	var last rdform.Document
	if err := json.Unmarshal(raw, &last); err != nil {
		return fmt.Errorf("parse last tax filing: %w", err)
	}
	for _, k := range keys {
		f.set(k, last.Values[k])
	}
	return nil
}

func digitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ---- ภาษีหัก ณ ที่จ่าย (ภ.ง.ด.3 / 53 / 2 / 2ก) จากรายการภาษีหักที่บันทึกในใบสำคัญ ----

var whtFormOf = map[string]string{"pnd3": "3", "pnd53": "53", "pnd2": "2", "pnd2a": "2"}

func fillWithholdingForm(ctx context.Context, db *sql.DB, company, code string, year, month int, doc *rdform.Document) ([]TaxFormNote, error) {
	if err := generalledger.EnsureSchema(ctx, db); err != nil {
		return nil, err
	}
	months := []int{month}
	if code == "pnd2a" { // แบบรายปี: รวมทั้งปีภาษี
		months = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	}
	var rows []TaxWithholdingRow
	unknownForm := 0
	for _, m := range months {
		report, err := buildWithholdingReport(ctx, db, company, year, m, "paid", []string{whtFormOf[code]})
		if err != nil {
			return nil, err
		}
		rows = append(rows, report.Rows...)
		unknownForm += report.UnknownForm
	}
	switch code {
	case "pnd2", "pnd2a":
		doc.Rows = pnd2Rows(rows, code == "pnd2a")
	default:
		doc.Rows = payeeRows(rows, code == "pnd3")
	}
	notes := withholdingNotes(rows)
	// ยอดที่ไม่รู้แบบไม่ถูกเติมลงแบบนี้ (ไม่เดาว่าเป็น ภ.ง.ด.3 หรือ 53) — บอกจำนวนให้ผู้ใช้บันทึกรายละเอียดและเลือกแบบ
	if unknownForm > 0 {
		notes = append(notes, TaxFormNote{Key: "tax_form_note_wht_form_unknown", Count: unknownForm})
	}
	notes = append(notes, missingTaxIDNotes(doc.Rows)...)
	return append(notes, missingIncomeTypeNotes(code, doc.Rows)...), nil
}

func withholdingNotes(rows []TaxWithholdingRow) []TaxFormNote {
	if len(rows) == 0 {
		return []TaxFormNote{{Key: "tax_form_note_no_withholding"}}
	}
	inferred := 0
	for _, r := range rows {
		if r.TaxBaseSource != "recorded" {
			inferred++
		}
	}
	var notes []TaxFormNote
	if inferred > 0 {
		notes = append(notes, TaxFormNote{Key: "tax_form_note_inferred_base", Count: inferred})
	}
	// กลับรายการในเดือนหลัง: ยังเป็นรายการของแบบเดือนนี้ (ยื่นแล้ว) — บอกทางแก้ภาษีที่ยื่นไปแล้ว (whtReversalMonthSQL)
	if reversed := reversedLaterCount(rows); reversed > 0 {
		notes = append(notes, TaxFormNote{Key: "tax_form_note_wht_reversed_later", Count: reversed})
	}
	return notes
}

// missingTaxIDNotes - แถวใบแนบที่ยังไม่มีเลขประจำตัวผู้เสียภาษี 13 หลัก นับจากเอกสารปัจจุบัน
// (prefill และ compute ใช้ตัวเดียวกัน — ผู้ใช้กรอกเลขแล้วกดคำนวณ/บันทึก หมายเหตุหายเอง)
func missingTaxIDNotes(rows []map[string]string) []TaxFormNote {
	missing := 0
	for _, r := range rows {
		if len(digitsOnly(r["tax_id"])) != 13 {
			missing++
		}
	}
	if missing == 0 {
		return []TaxFormNote{}
	}
	return []TaxFormNote{{Key: "tax_form_note_missing_taxid", Count: missing}}
}

// missingIncomeTypeNotes - รายการในใบแนบที่ยังไม่มีประเภทเงินได้ (ระบบไม่เดาให้: ภ.ง.ด.2 รหัสที่ไม่ใช่ 40(3)/40(4), ภ.ง.ด.3/53 ที่มีแต่ชื่อมาตรา)
// นับจากเอกสารปัจจุบันเหมือน missingTaxIDNotes — ผู้ใช้เลือก/กรอกแล้วกดคำนวณ หมายเหตุหายเอง
func missingIncomeTypeNotes(code string, rows []map[string]string) []TaxFormNote {
	missing := 0
	for _, r := range rows {
		if code == "pnd2" || code == "pnd2a" {
			if strings.TrimSpace(r["income_type"]) == "" {
				missing++
			}
			continue
		}
		for k := 1; k <= 3; k++ {
			p := "l" + strconv.Itoa(k) + "_"
			if strings.TrimSpace(r[p+"amount"]) != "" && strings.TrimSpace(r[p+"income_type"]) == "" {
				missing++
			}
		}
	}
	if missing == 0 {
		return []TaxFormNote{}
	}
	return []TaxFormNote{{Key: "tax_form_note_missing_income_type", Count: missing}}
}

// payeeRows - ใบแนบ ภ.ง.ด.3/53: หนึ่งแถวต่อผู้มีเงินได้ ไม่เกิน 3 รายการเงินได้ต่อแถว (เกินขึ้นแถวใหม่ของรายเดิม)
func payeeRows(rows []TaxWithholdingRow, individual bool) []map[string]string {
	var out []map[string]string
	current := map[string]int{}
	lines := map[int]int{}
	for _, r := range rows {
		key := payeeKey(r)
		i, ok := current[key]
		if !ok || lines[i] == 3 {
			out = append(out, payeeHeader(r, individual))
			i = len(out) - 1
			current[key] = i
		}
		lines[i]++
		p := fmt.Sprintf("l%d_", lines[i])
		out[i][p+"date"] = thaiDate(firstNonEmpty(r.PaidDate, r.DocDate))
		out[i][p+"income_type"] = attachmentIncomeType(r)
		out[i][p+"rate"] = rateText(r.RatePercent)
		out[i][p+"amount"] = r.BaseAmount
		out[i][p+"tax"] = r.WhtAmount
		out[i][p+"condition"] = conditionText(r.Condition)
	}
	return out
}

// payeeKey - ผู้มีเงินได้รายเดียวกัน = เลขผู้เสียภาษี + ชื่อเดียวกัน; แถวที่ไม่มีทั้งสองอย่างห้ามรวมเป็นรายเดียว
// (คนละรายจะถูกพิมพ์ในบรรทัดเดียวกัน) — ใช้รหัสคู่ค้า ถ้าไม่มีใช้เลขใบสำคัญ
func payeeKey(r TaxWithholdingRow) string {
	taxID := digitsOnly(r.TaxID)
	switch {
	case taxID != "" || strings.TrimSpace(r.PartnerName) != "":
		return "payee:" + taxID + "|" + r.PartnerName
	case r.PartnerCode != "":
		return "partner:" + r.PartnerCode
	}
	return "journal:" + r.JournalID
}

func payeeHeader(r TaxWithholdingRow, individual bool) map[string]string {
	row := map[string]string{"seq": ""}
	if id := digitsOnly(r.TaxID); len(id) == 13 {
		row["tax_id"] = id
	}
	setNonEmpty(row, "title", r.Title) // คำนำหน้าจากทะเบียนคู่ค้า (ไฟล์ยื่นด้วยสื่อบังคับ) ผู้ใช้แก้ได้ในใบแนบ
	address := strings.Join(strings.Fields(r.Address), " ")
	if individual { // ใบแนบ ภ.ง.ด.3 แยกชื่อ/ชื่อสกุล และมีที่อยู่บรรทัดเดียว (+อำเภอ/จังหวัด/รหัสไปรษณีย์ที่ไฟล์ ภ.ง.ด.3 บังคับ)
		row["name"], row["surname"] = splitPersonName(r.PartnerName, r.Title)
		row["address1"] = address
		setNonEmpty(row, "addr_district", r.District)
		setNonEmpty(row, "addr_province", r.Province)
		setNonEmpty(row, "addr_postcode", r.Postcode)
		return row
	}
	row["name"] = r.PartnerName
	row["address1"], row["address2"] = splitAddress(address, 60)
	return row
}

func setNonEmpty(row map[string]string, key, value string) {
	if value = strings.TrimSpace(value); value != "" {
		row[key] = value
	}
}

// splitPersonName - ชื่อ/ชื่อสกุลของใบแนบและไฟล์ยื่น: "นายสมชาย ใจดี" → ("นายสมชาย", "ใจดี"); คำเดียว = ชื่ออย่างเดียว
// รู้คำนำหน้า (ทะเบียนคู่ค้า): ชื่อคือคำแรกหลังคำนำหน้า ที่เหลือเป็นชื่อสกุล — Format กลาง ภ.ง.ด.3 ช่อง 8 / ภ.ง.ด.2 ช่อง 9 SNAME
// "กรณีมีชื่อกลาง ให้ระบุชื่อกลาง+1 ช่องว่าง+นามสกุล" ("Mr. John Michael Smith" → "Mr. John" / "Michael Smith", "นายสมชาย ณ อยุธยา" → "ณ อยุธยา");
// คำนำหน้าที่เขียนแยกคำ ("นาย สมชาย ใจดี") อยู่ในช่องชื่อตามแบบ ("ให้ระบุให้ชัดเจนว่าเป็น นาย นาง นางสาว หรือยศ")
// ไม่รู้คำนำหน้า: แยกคำนำหน้าแยกคำกับชื่อกลางไม่ออก จึงใช้คำสุดท้ายเป็นชื่อสกุลแบบเดิม (ไฟล์ยื่นบังคับคำนำหน้าอยู่แล้ว)
func splitPersonName(full, title string) (string, string) {
	parts := strings.Fields(full)
	if len(parts) < 2 {
		return strings.TrimSpace(full), ""
	}
	titleWords := strings.Fields(title)
	if len(titleWords) == 0 || strings.TrimSpace(title) == "-" {
		return strings.Join(parts[:len(parts)-1], " "), parts[len(parts)-1]
	}
	first := 1
	if len(parts) > len(titleWords)+1 && strings.Join(parts[:len(titleWords)], " ") == strings.Join(titleWords, " ") {
		first = len(titleWords) + 1
	}
	return strings.Join(parts[:first], " "), strings.Join(parts[first:], " ")
}

// splitAddress - ที่อยู่ยาวเกินบรรทัด: ตัดที่ช่องว่างใกล้ครึ่งแรกที่ไม่เกิน max ตัวอักษร
func splitAddress(address string, max int) (string, string) {
	if utf8.RuneCountInString(address) <= max {
		return address, ""
	}
	words := strings.Fields(address)
	first := ""
	for i, w := range words {
		next := strings.TrimSpace(first + " " + w)
		if utf8.RuneCountInString(next) > max && first != "" {
			return first, strings.Join(words[i:], " ")
		}
		first = next
	}
	return first, ""
}

// thaiDate - "2026-09-15" → "15/09/2569"
func thaiDate(iso string) string {
	iso = strings.TrimSpace(iso)
	if len(iso) > 10 {
		iso = iso[:10] // "2026-09-15T00:00:00Z"
	}
	d, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%02d/%02d/%d", d.Day(), int(d.Month()), d.Year()+543)
}

// rateText - "3.00" → "3", "1.50" → "1.5"
func rateText(rate string) string {
	d, err := decimal.NewFromString(strings.TrimSpace(rate))
	if err != nil {
		return ""
	}
	return d.String()
}

func conditionText(c int) string {
	if c >= 1 && c <= 3 {
		return strconv.Itoa(c)
	}
	return "1"
}

// incomeTypeText - ประเภทเงินได้: คำอธิบายที่ผู้ใช้บันทึก ถ้าไม่มีใช้ชื่อมาตราของรหัสประเภทเงินได้ตามภาษา lang (languages.tsv)
// ใบแนบ ภ.ง.ด. ส่ง "th"; รายงานบนจอส่งภาษาของผู้ใช้ — ห้ามฝังข้อความไทยใน API (กฎ i18n ของ AGENTS.md)
func incomeTypeText(r TaxWithholdingRow, lang string) string {
	if d := strings.TrimSpace(r.Description); d != "" {
		return d
	}
	if key := incomeTypeKey(r.IncomeType); key != "" {
		return language.Text(key, lang)
	}
	return ""
}

// attachmentIncomeType - ประเภทเงินได้ในใบแนบ ภ.ง.ด.3/53 (INC_TYPE_PND ของไฟล์ยื่น) ต้องบอกว่าเป็นค่าอะไร —
// Format กลาง ภ.ง.ด.53 ช่อง 13 "ให้ระบุว่าเป็นค่าอะไร เช่น ค่าเช่าอาคาร ค่าบริการ", ภ.ง.ด.3 "ต้องระบุประเภทเงินได้"
// (https://www.rd.go.th/fileadmin/user_upload/WHT/Download/FormatPND53V2_0.pdf) — "มาตรา 3 เตรส" เป็นชื่อมาตรา ไม่ใช่ค่าอะไร
// จึงเว้นว่างให้ผู้ใช้กรอก (missingIncomeTypeNotes แจ้ง, ไฟล์ยื่นขึ้น tax_rdfile_income_type_required) แทนการเขียนชื่อมาตราลงแบบ;
// ใบแนบ ภ.ง.ด. เป็นแบบทางการภาษาไทยเสมอ
func attachmentIncomeType(r TaxWithholdingRow) string {
	if r.IncomeType == "3_tres" && strings.TrimSpace(r.Description) == "" {
		return ""
	}
	return incomeTypeText(r, "th")
}

// incomeTypeKey - key ชื่อมาตราของรหัสประเภทเงินได้แบบ 50 ทวิ; เงินปันผล 40(4)(ข) ทุกกรณีย่อย (40_4b*) ใช้ชื่อเดียวกัน
func incomeTypeKey(code string) string {
	switch {
	case code == "3_tres":
		return "tax_income_type_3_tres"
	case code == "40_2":
		return "tax_income_type_40_2"
	case code == "40_3":
		return "tax_income_type_40_3"
	case code == "40_4a":
		return "tax_income_type_40_4a"
	case strings.HasPrefix(code, "40_4b"):
		return "tax_income_type_40_4b"
	}
	return ""
}

// pnd2IncomeType - รหัสประเภทเงินได้ของ 50 ทวิ → ตัวเลือก "ประเภทเงินได้" ของ ภ.ง.ด.2/2ก (หนึ่งแผ่นต่อหนึ่งประเภท)
// Format กลาง ภ.ง.ด.2 V2.0 ช่อง 14 INC_TYPE_PND มีแค่ 40(3), 40(4)(ก), 40(4)(ข), 40(4)(ช), 40(4) อื่น ๆ
// (https://www.rd.go.th/fileadmin/user_upload/WHT/Download/FormatPND2V2_0.pdf) — รหัสนอกตระกูลนี้ (40(2), มาตรา 3 เตรส, อื่น ๆ)
// และ 40(3) ใน ภ.ง.ด.2ก (แบบไม่มีบรรทัด 40(3)) คืนค่าว่างให้ผู้ใช้เลือกเอง — เดิมตกไปเป็น "40(4) อื่น ๆ" เงียบ ๆ (review 2026-09-24)
func pnd2IncomeType(code string, annual bool) string {
	switch {
	case code == "40_3" && !annual:
		return "royalty"
	case code == "40_4a":
		return "interest"
	case strings.HasPrefix(code, "40_4b"):
		return "dividend"
	}
	return ""
}

// pnd2Rows - ใบแนบ ภ.ง.ด.2 (รายการจ่ายแต่ละครั้ง) / ภ.ง.ด.2ก (รวมทั้งปีต่อผู้รับต่อประเภท)
func pnd2Rows(rows []TaxWithholdingRow, annual bool) []map[string]string {
	var out []map[string]string
	index := map[string]int{}
	for _, r := range rows {
		kind := pnd2IncomeType(r.IncomeType, annual)
		key := kind + "|" + payeeKey(r)
		if i, ok := index[key]; ok && annual {
			out[i]["l1_amount"] = addMoneyText(out[i]["l1_amount"], r.BaseAmount)
			out[i]["l1_tax"] = addMoneyText(out[i]["l1_tax"], r.WhtAmount)
			if out[i]["l1_rate"] != rateText(r.RatePercent) {
				out[i]["l1_rate"] = "" // หลายอัตราในปีเดียวกัน: ให้ผู้ใช้ระบุเอง
			}
			continue
		}
		row := map[string]string{"seq": "", "income_type": kind, "l1_rate": rateText(r.RatePercent),
			"l1_amount": r.BaseAmount, "l1_tax": r.WhtAmount, "l1_condition": conditionText(r.Condition)}
		if id := digitsOnly(r.TaxID); len(id) == 13 {
			row["tax_id"] = id
		}
		row["name"], row["surname"] = splitPersonName(r.PartnerName, r.Title)
		if annual {
			row["address1"] = strings.Join(strings.Fields(r.Address), " ")
		} else { // ใบแนบ ภ.ง.ด.2ก ไม่มีช่องคำนำหน้า (ไฟล์ ภ.ง.ด.2ก เลื่อนไปก่อน)
			row["l1_date"] = thaiDate(firstNonEmpty(r.PaidDate, r.DocDate))
			setNonEmpty(row, "title", r.Title)
		}
		index[key] = len(out)
		out = append(out, row)
	}
	return out
}

func addMoneyText(a, b string) string {
	x, _ := decimal.NewFromString(strings.TrimSpace(a))
	y, _ := decimal.NewFromString(strings.TrimSpace(b))
	return x.Add(y).StringFixed(2)
}

// ---- ภาษีเงินได้นิติบุคคล (ภ.ง.ด.50 / 51) จากบัญชีแยกประเภท ----

type fiscalPeriod struct {
	code       string
	start, end time.Time
}

// fiscalYearEnding - รอบบัญชีที่สิ้นสุดในปี ค.ศ. year (ไม่มี = ปีปฏิทิน)
func fiscalYearEnding(ctx context.Context, db *sql.DB, company string, year int) (fiscalPeriod, bool, error) {
	var code, start, end string
	err := db.QueryRowContext(ctx, `SELECT code, payload->>'startdate', payload->>'enddate' FROM gl_records
WHERE company=$1 AND kind='fiscal-years' AND NOT COALESCE((payload->>'isdeleted')::boolean,false)
AND EXTRACT(YEAR FROM (payload->>'enddate')::date)=$2 ORDER BY code LIMIT 1`, company, year).Scan(&code, &start, &end)
	calendar := fiscalPeriod{start: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), end: time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)}
	if errors.Is(err, sql.ErrNoRows) {
		return calendar, false, nil
	}
	if err != nil {
		return calendar, false, fmt.Errorf("load fiscal year: %w", err)
	}
	s, err1 := time.Parse("2006-01-02", start)
	e, err2 := time.Parse("2006-01-02", end)
	if err1 != nil || err2 != nil {
		return calendar, false, nil
	}
	return fiscalPeriod{code: code, start: s, end: e}, true, nil
}

// fillCitForm - รอบบัญชี + กำไรสุทธิทางบัญชีและยอดรวมงบแสดงฐานะการเงิน (ภ.ง.ด.50) หรือกำไร 6 เดือนแรกเป็นยอดอ้างอิง (ภ.ง.ด.51)
// เติมเฉพาะยอดรวมที่ได้จากประเภทบัญชีโดยตรง — การแยกรายการตามแบบ (รายได้/ต้นทุน/รายจ่ายต้องห้าม) ผู้ทำบัญชีต้องกรอกเอง
func fillCitForm(ctx context.Context, db *sql.DB, company, code string, year int, f *formFiller) ([]TaxFormNote, error) {
	if err := generalledger.EnsureSchema(ctx, db); err != nil {
		return nil, err
	}
	fy, found, err := fiscalYearEnding(ctx, db, company, year)
	if err != nil {
		return nil, err
	}
	notes := []TaxFormNote{}
	if !found {
		notes = append(notes, TaxFormNote{Key: "tax_form_note_no_fiscal_year"})
	}
	periodEnd := fy.end
	if code == "pnd51" { // ครึ่งปีแรกของรอบบัญชี
		periodEnd = fy.start.AddDate(0, 6, -1)
	}
	for _, prefix := range []string{"period", "ds_period"} {
		f.setDate(prefix+"_start", fy.start)
		f.setDate(prefix+"_end", fy.end)
	}
	f.set("ds_company_name", f.values["name"])
	credits, err := fillCitCredits(ctx, db, company, code, year, fy.start, periodEnd, f)
	if err != nil {
		return nil, err
	}
	notes = append(notes, credits...)
	if !found {
		return notes, nil
	}
	revenue, expense, err := profitAndLoss(ctx, db, company, fy.code, fy.start, periodEnd)
	if err != nil {
		return nil, err
	}
	profit := revenue.Sub(expense)
	if code == "pnd51" {
		return append(notes, TaxFormNote{Key: "tax_form_note_cit_half_profit", Amount: profit.StringFixed(2)}), nil
	}
	setProfit(f, "r2_09_net_profit_per_accounts_total", "r2_09_net_result", profit)
	if err := fillBalanceSheetTotals(ctx, db, company, fy, f); err != nil {
		return nil, err
	}
	return append(notes, TaxFormNote{Key: "tax_form_note_cit_totals_only"}), nil
}

func setProfit(f *formFiller, amountKey, resultKey string, profit decimal.Decimal) {
	f.set(amountKey, profit.Abs().StringFixed(2))
	if profit.IsNegative() {
		f.set(resultKey, "loss")
	} else {
		f.set(resultKey, "profit")
	}
}

// fillCitCredits - เครดิตภาษีที่มีหลักฐานในระบบ ตามคู่มือวิธีกรอกแบบของกรมสรรพากร (docs/kms/21-thai-tax-form-references.md §2–§3):
// ภ.ง.ด.50 ข้อ 3.(3) = ภาษีที่บริษัทถูกหัก ณ ที่จ่ายทั้งรอบบัญชี, ภ.ง.ด.51 ข้อ 5.(1) = เฉพาะ 6 เดือนแรก (from..to ของแบบ);
// ภ.ง.ด.50 ข้อ 3.(4) = ยอด "ชำระเพิ่มเติม" รายการที่ 2 ข้อ 6 ของ ภ.ง.ด.51 ที่บันทึกไว้ — ช่องอื่นของรายการเครดิตผู้ใช้กรอกเอง
func fillCitCredits(ctx context.Context, db *sql.DB, company, code string, year int, from, to time.Time, f *formFiller) ([]TaxFormNote, error) {
	var notes []TaxFormNote
	withheld, entries, err := generalledger.WithheldFromCompanyTotal(ctx, db, company, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	if entries > 0 {
		key := "less_wht"
		if code == "pnd51" {
			key = "r2_5_1_wht"
		}
		f.set(key, withheld.StringFixed(2))
		notes = append(notes, TaxFormNote{Key: "tax_form_note_cit_wht_credit", Count: entries, Amount: withheld.StringFixed(2)})
	}
	if code != "pnd50" {
		return notes, nil
	}
	paid, filings, err := pnd51PaidTax(ctx, db, company, year)
	if err != nil {
		return nil, err
	}
	if filings > 0 {
		f.set("less_pnd51_paid", paid.StringFixed(2))
		notes = append(notes, TaxFormNote{Key: "tax_form_note_cit_pnd51_paid", Count: filings, Amount: paid.StringFixed(2)})
	}
	return notes, nil
}

// pnd51PaidTax - ผลรวมยอด "ชำระเพิ่มเติม" (รายการที่ 2 ข้อ 6) ของ ภ.ง.ด.51 ทุกฉบับของปีเดียวกัน (ฉบับยื่นเพิ่มเติมหักยอดฉบับก่อนในข้อ 5.(3) แล้ว จึงรวมกันได้ตรง)
func pnd51PaidTax(ctx context.Context, db *sql.DB, company string, year int) (decimal.Decimal, int, error) {
	total, filings := decimal.Zero, 0
	if err := ensureTaxFilingSchema(ctx, db); err != nil {
		return total, filings, err
	}
	rows, err := db.QueryContext(ctx, `SELECT COALESCE(document->'values'->>'r2_6_balance','') FROM tax_filings
WHERE company_code=$1 AND form_code='pnd51' AND period_year=$2 AND document->'values'->>'r2_6_sign'='payable'`, company, year)
	if err != nil {
		return total, filings, fmt.Errorf("read pnd51 filings: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var balance string
		if err := rows.Scan(&balance); err != nil {
			return total, filings, fmt.Errorf("scan pnd51 filing: %w", err)
		}
		if strings.TrimSpace(balance) == "" {
			continue
		}
		amount, err := decimal.NewFromString(balance)
		if err != nil {
			return total, filings, fmt.Errorf("pnd51 balance %q: %w", balance, err)
		}
		total, filings = total.Add(amount), filings+1
	}
	return total, filings, rows.Err()
}

// profitAndLoss - รายได้ (เครดิต-เดบิต) และค่าใช้จ่าย (เดบิต-เครดิต) ของบัญชีหมวด 4/5 ในช่วง ไม่รวมรายการยกมา/ปิดบัญชี
func profitAndLoss(ctx context.Context, db *sql.DB, company, fiscal string, from, to time.Time) (decimal.Decimal, decimal.Decimal, error) {
	var revenue, expense decimal.Decimal
	err := db.QueryRowContext(ctx, `SELECT
COALESCE(SUM(CASE WHEN account_type='income' THEN credit-debit ELSE 0 END),0)::text,
COALESCE(SUM(CASE WHEN account_type='expense' THEN debit-credit ELSE 0 END),0)::text
FROM gl_lines WHERE company=$1 AND fiscal_year=$2 AND entry_date BETWEEN $3::date AND $4::date AND kind NOT IN ('opening','closing')`,
		company, fiscal, from.Format("2006-01-02"), to.Format("2006-01-02")).Scan(&revenue, &expense)
	if err != nil {
		return revenue, expense, fmt.Errorf("profit and loss: %w", err)
	}
	return revenue, expense, nil
}

// fillBalanceSheetTotals - ยอดรวมสินทรัพย์/หนี้สิน/ส่วนของผู้ถือหุ้น ณ วันสิ้นรอบ (กำไรที่ยังไม่ปิดบัญชีรวมในส่วนของผู้ถือหุ้น)
func fillBalanceSheetTotals(ctx context.Context, db *sql.DB, company string, fy fiscalPeriod, f *formFiller) error {
	var assets, liabilities, equity decimal.Decimal
	err := db.QueryRowContext(ctx, `SELECT
COALESCE(SUM(CASE WHEN account_type='asset' THEN debit-credit ELSE 0 END),0)::text,
COALESCE(SUM(CASE WHEN account_type='liability' THEN credit-debit ELSE 0 END),0)::text,
COALESCE(SUM(CASE WHEN account_type IN ('equity','income','expense') THEN credit-debit ELSE 0 END),0)::text
FROM gl_lines WHERE company=$1 AND fiscal_year=$2 AND entry_date<=$3::date`, company, fy.code, fy.end.Format("2006-01-02")).Scan(&assets, &liabilities, &equity)
	if err != nil {
		return fmt.Errorf("balance sheet totals: %w", err)
	}
	f.set("bs_total_assets", assets.StringFixed(2))
	f.set("bs_total_liabilities", liabilities.StringFixed(2))
	f.set("bs_total_equity", equity.StringFixed(2))
	f.set("bs_total_liabilities_equity", liabilities.Add(equity).StringFixed(2))
	return nil
}
