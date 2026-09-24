package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/rdfile"
	"smlcloudplatform/internal/rdform"
)

// ไฟล์ยื่นภาษีหัก ณ ที่จ่ายด้วยสื่อ (Format กลาง V2.0 ของกรมสรรพากร สำหรับโปรแกรม SWC-UI) จาก "ฉบับที่บันทึกแล้ว" ใน tax_filings
// — ไม่ใช้เอกสารจาก browser: ตัวเลขต้องตรงกับฉบับที่ผู้ใช้ตรวจและพิมพ์ PDF (รวมฐานภาษีที่แก้แล้ว) และคำนวณยอดรวมใหม่ทุกครั้ง
// เนื้อไฟล์มีเลขประจำตัวประชาชนของผู้มีเงินได้ → ห้าม log เนื้อไฟล์

// rdFileKinds - แบบที่ทำไฟล์ได้ตอนนี้ (ภ.ง.ด.2ก/3ก และ ภ.พ.30 ยื่นรวม เลื่อนไปก่อน)
var rdFileKinds = map[string]rdfile.Kind{"pnd53": rdfile.PND53, "pnd3": rdfile.PND3, "pnd2": rdfile.PND2}

// rdFileSections - ตัวเลือก "นำส่งภาษีตาม" ของหน้าแบบ → ตำแหน่งธงมาตรา H10-H12
var rdFileSections = map[rdfile.Kind]map[string]int{
	rdfile.PND53: {"3tres": 0, "65jattawa": 1, "69tawi": 2},
	rdfile.PND3:  {"3tres": 0, "48tawi": 1, "50_3_4_5": 2},
}

// rdFilePND2IncomeTypes - ประเภทเงินได้ของใบแนบ ภ.ง.ด.2 → INC_TYPE_PND
// ([F2]: 1=มาตรา 40(3), 2=40(4)(ก), 3=40(4)(ข), 4=40(4)(ช), 5=40(4) อื่นๆ)
var rdFilePND2IncomeTypes = map[string]string{"royalty": "1", "interest": "2", "dividend": "3", "share_transfer": "4", "other_404": "5"}

// rdFileOptions - ค่าที่ไม่มีในแบบ ผู้ใช้กรอกในแผงสร้างไฟล์
type rdFileOptions struct {
	DeptName     string `json:"dept_name"`
	SubmissionNo string `json:"submission_no"` // "ครั้งที่ส่ง" 00-99 ท้ายชื่อไฟล์ ค่าเริ่มต้น 00
	LTO          string `json:"lto"`           // "" | "0" | "1"
	BranchType   string `json:"branch_type"`   // "" | "V" | "S"
}

const (
	maxRdFileIssues       = 100 // รายการผิดที่ส่งกลับ (total บอกจำนวนจริง)
	maxRdFileRequestBytes = 64 << 10
)

var errRdFileUnsupported = errors.New("tax form has no RD media file format")

// taxRdFileDB - จุดสลับให้เทสระดับ handler ใช้ PostgreSQL ทดสอบแทน pool ของกลุ่มกิจการ
var taxRdFileDB = func(ctx context.Context, t *taxFormCall) (*sql.DB, error) { return t.db(ctx) }

// TaxFormRdFileHandler - POST /api/report/tax/form/rdfile {id, version, rdfile:{dept_name, submission_no, lto, branch_type}}
// → 200 text/plain (ไฟล์ .txt แนบ) | 400 {code: tax_rdfile_invalid, issues} | 404 | 409
func TaxFormRdFileHandler(c echo.Context) error {
	raw, err := io.ReadAll(io.LimitReader(c.Request().Body, maxRdFileRequestBytes+1))
	if err != nil || len(raw) > maxRdFileRequestBytes {
		return c.JSON(http.StatusBadRequest, map[string]any{"success": false, "code": "tax_form_payload_invalid",
			"message": language.Text("tax_form_payload_invalid", taxRequestLanguage(c))})
	}
	c.Request().Body = io.NopCloser(bytes.NewReader(raw)) // beginTaxForm อ่าน holding/company/id/version จาก body เดิม
	t, err := beginTaxForm(c, false)
	if t == nil {
		return err
	}
	var body struct {
		RdFile rdFileOptions `json:"rdfile"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return t.fail(http.StatusBadRequest, "tax_form_payload_invalid")
	}
	if t.req.ID <= 0 {
		return t.fail(http.StatusNotFound, "tax_form_not_found") // ไฟล์สร้างจากฉบับที่บันทึกแล้วเท่านั้น
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 30*time.Second)
	defer cancel()
	db, err := taxRdFileDB(ctx, t)
	if err != nil {
		logger.Error("TaxFormRdFile: db: %v", err)
		return t.fail(http.StatusInternalServerError, "tax_form_failed")
	}
	file, err := buildTaxRdFile(ctx, db, t.holding, t.company, t.req.ID, t.req.Version, body.RdFile.normalized())
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return t.fail(http.StatusNotFound, "tax_form_not_found")
	case errors.Is(err, errTaxFilingConflict):
		return t.fail(http.StatusConflict, "tax_form_version_conflict")
	case errors.Is(err, errRdFileUnsupported):
		return t.fail(http.StatusBadRequest, "tax_rdfile_unsupported_form", "code")
	case errors.Is(err, rdform.ErrInvalidValue):
		return t.failValue(err)
	case err != nil:
		logger.Error("TaxFormRdFile %d: %v", t.req.ID, err)
		return t.fail(http.StatusInternalServerError, "tax_form_failed")
	}
	if len(file.issues) > 0 {
		return t.rdFileIssues(file.issues)
	}
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.name))
	c.Response().Header().Set("Cache-Control", "no-store")
	return c.Blob(http.StatusOK, "text/plain; charset=utf-8", file.content)
}

func (o rdFileOptions) normalized() rdFileOptions {
	o.DeptName, o.LTO, o.BranchType = strings.TrimSpace(o.DeptName), strings.TrimSpace(o.LTO), strings.TrimSpace(o.BranchType)
	if o.SubmissionNo = strings.TrimSpace(o.SubmissionNo); o.SubmissionNo == "" {
		o.SubmissionNo = "00"
	}
	return o
}

// rdFileIssues - 400 พร้อมทุกจุดที่ต้องแก้ (ไม่เกิน maxRdFileIssues) ข้อความตามภาษาผู้ใช้ แทน {char}/{max} ด้วยค่าจริง
func (t *taxFormCall) rdFileIssues(issues []rdfile.Issue) error {
	type issueBody struct {
		Key     string            `json:"key"`
		Field   string            `json:"field,omitempty"`
		Row     int               `json:"row,omitempty"`
		Args    map[string]string `json:"args,omitempty"`
		Message string            `json:"message"`
	}
	out := make([]issueBody, 0, min(len(issues), maxRdFileIssues))
	for _, i := range issues[:min(len(issues), maxRdFileIssues)] {
		msg := language.Text(i.Key, t.lang)
		for k, v := range i.Args {
			msg = strings.ReplaceAll(msg, "{"+k+"}", v)
		}
		out = append(out, issueBody{Key: i.Key, Field: i.Field, Row: i.Row, Args: i.Args, Message: msg})
	}
	return t.c.JSON(http.StatusBadRequest, map[string]any{"success": false, "code": "tax_rdfile_invalid",
		"message": language.Text("tax_rdfile_invalid", t.lang), "total": len(issues), "issues": out})
}

type taxRdFile struct {
	name    string
	content []byte
	issues  []rdfile.Issue
}

// buildTaxRdFile - ฉบับที่บันทึก (id + version ต้องตรง) → คำนวณยอดรวมใหม่ → หัวแบบตามทะเบียนบริษัทแบบเดียวกับ PDF → ตรวจ → ไฟล์
func buildTaxRdFile(ctx context.Context, db *sql.DB, holding, company string, id int64, version int, opts rdFileOptions) (taxRdFile, error) {
	filing, err := loadTaxFiling(ctx, db, company, id)
	if err != nil {
		return taxRdFile{}, err
	}
	if filing.Version != version {
		return taxRdFile{}, errTaxFilingConflict
	}
	kind, ok := rdFileKinds[filing.Code]
	if !ok {
		return taxRdFile{}, errRdFileUnsupported
	}
	doc, err := prepareTaxDocument(filing.Code, *filing.Document)
	if err != nil {
		return taxRdFile{}, err
	}
	registered, err := whtCompanyHeader(ctx, holding, company)
	if err != nil {
		return taxRdFile{}, fmt.Errorf("company header: %w", err)
	}
	applyCompanyHeader(doc.Values, registered)
	h, ds := rdFileContent(kind, filing, doc, opts)
	if issues := rdfile.Check(kind, h, ds); len(issues) > 0 {
		return taxRdFile{issues: issues}, nil
	}
	return taxRdFile{name: rdfile.FileName(kind, h), content: rdfile.Encode(rdfile.Lines(kind, h, ds))}, nil
}

// rdFileContent - เอกสารที่คำนวณแล้ว → บรรทัด H/D ของไฟล์ (แปลงรูปแบบอย่างเดียว ค่าที่ผิดคงไว้ให้ rdfile.Check รายงาน)
func rdFileContent(kind rdfile.Kind, filing TaxFiling, doc rdform.Document, opts rdFileOptions) (rdfile.Header, []rdfile.Detail) {
	v := doc.Values
	nid, branch := digitsOnly(v["tax_id"]), rdFileBranch(v["branch_no"])
	h := rdfile.Header{SenderID: rdfile.SenderIDMedia, SenderNID: nid, SenderBranch: branch, SenderRole: rdfile.SenderRoleSelf,
		NID: nid, BranchNo: branch, DeptName: opts.DeptName, LTO: opts.LTO, TaxMonth: fmt.Sprintf("%02d", filing.Month),
		TaxYear: strconv.Itoa(filing.Year + 543), BranchType: opts.BranchType, FormType: fmt.Sprintf("%02d", filing.FilingSeq),
		SurAmt: rdFileDecimal(v["surcharge"]), UserID: v["media_ref_no"], FormFlag: rdfile.FormFlagMedia, SubmissionNo: opts.SubmissionNo}
	if h.DeptName == "" && branch == "000000" {
		h.DeptName = rdfile.HeadOfficeDept
	}
	if sections, ok := rdFileSections[kind]; ok {
		h.Sections = [3]string{"0", "0", "0"}
		if i, ok := sections[v["tax_section"]]; ok {
			h.Sections[i] = "1"
		}
	}
	order := rdFileRowOrder(kind, doc.Rows)
	ds := make([]rdfile.Detail, 0, len(order))
	for seq, i := range order {
		ds = append(ds, rdFileDetail(kind, doc.Rows[i], i+1, seq+1, branch))
	}
	h.NilReturn = len(doc.Rows) == 0 && rdFileZero(v["total_income"]) && rdFileZero(v["total_tax"]) && rdFileZero(v["surcharge"])
	h, ds = rdfile.Trim(h, ds)
	return rdfile.Summarize(kind, h, ds), ds
}

// rdFileZero - ช่องยอดเงินว่างหรือเป็นศูนย์: ฉบับที่ไม่มีใบแนบและยอดทุกช่องเป็นศูนย์ = ยื่นหัวแบบอย่างเดียว (TOT_NUM 0)
// ไม่มีใบแนบแต่ผู้ใช้กรอกยอดเอง (ยื่นทางสื่อจากระบบอื่น) ยังคง tax_rdfile_no_rows — ไฟล์ต้องมีบรรทัด D ให้ตรงกับยอดหัวแบบ
func rdFileZero(raw string) bool {
	s := strings.ReplaceAll(strings.TrimSpace(raw), ",", "")
	if s == "" {
		return true
	}
	d, err := decimal.NewFromString(s)
	return err == nil && d.IsZero()
}

// rdFileRowOrder - ลำดับบรรทัด D = ลำดับที่บนใบแนบที่พิมพ์: ใบแนบ ภ.ง.ด.2 แยกแผ่นตามประเภทเงินได้ (ประเภทที่พบก่อนขึ้นก่อน)
// แบบเดียวกับตัวกรอก PDF (rdform) จึงให้ SEQ_NO ตรงกับกระดาษ
func rdFileRowOrder(kind rdfile.Kind, rows []map[string]string) []int {
	order := make([]int, 0, len(rows))
	if kind != rdfile.PND2 {
		for i := range rows {
			order = append(order, i)
		}
		return order
	}
	var groups []string
	members := map[string][]int{}
	for i, r := range rows {
		g := strings.TrimSpace(r["income_type"])
		if _, ok := members[g]; !ok {
			groups = append(groups, g)
		}
		members[g] = append(members[g], i)
	}
	for _, g := range groups {
		order = append(order, members[g]...)
	}
	return order
}

// rdFileDetail - แถวใบแนบ → บรรทัด D; สาขาในบรรทัด D คือสาขาของผู้จ่ายเงินได้ ([F53] D3 "ระบุสาขาของผู้จ่ายเงินได้")
func rdFileDetail(kind rdfile.Kind, row map[string]string, rowNo, seq int, branch string) rdfile.Detail {
	d := rdfile.Detail{Row: rowNo, Seq: strconv.Itoa(seq), BranchNo: branch, NID: digitsOnly(row["tax_id"]), TIN: rdfile.NoTIN,
		Title: row["title"], FirstName: rdFileFirstName(row["name"], row["title"])}
	switch kind {
	case rdfile.PND2:
		// ACC_NO บังคับเฉพาะเงินปันผล (รหัส 3) ใน rdfile.Check ตาม [F2] D6 — ดอกเบี้ย (รหัส 2) ส่งตามที่กรอก
		d.AccNo, d.LastName = digitsOnly(row["bank_account_no"]), row["surname"]
		d.Items[0] = rdFileItem(row, "l1_", rdFilePND2IncomeTypes[row["income_type"]])
	default:
		for k := range d.Items {
			p := "l" + strconv.Itoa(k+1) + "_"
			d.Items[k] = rdFileItem(row, p, row[p+"income_type"])
		}
		if kind == rdfile.PND3 {
			d.LastName = row["surname"]
			d.Address.Amphur, d.Address.Province, d.Address.PostalCode = row["addr_district"], row["addr_province"], row["addr_postcode"]
		}
	}
	return d
}

// rdFileItem - รายการที่ k มีอยู่เมื่อช่องใดช่องหนึ่งของ lk_* มีค่า; ว่างทั้งหมด = รายการว่าง (00000000|0.00|0.00|0.00||)
func rdFileItem(row map[string]string, p, incomeType string) rdfile.Item {
	date, rate, amount, tax, condition := row[p+"date"], row[p+"rate"], row[p+"amount"], row[p+"tax"], row[p+"condition"]
	if date == "" && rate == "" && amount == "" && tax == "" && incomeType == "" && condition == "" {
		return rdfile.Item{}
	}
	return rdfile.Item{PaidDate: rdFileDate(date), TaxRate: rdFileDecimal(rate), PaidAmt: rdFileDecimal(amount),
		TaxAmt: rdFileDecimal(tax), IncType: incomeType, PayCon: condition}
}

// rdFileFirstName - FNAME: ชื่อในใบแนบตัดคำนำหน้าที่ขึ้นต้นตรงกับช่องคำนำหน้าออก 1 ครั้ง
// ("บริษัท สยาม จำกัด" + "บริษัท" → "สยาม จำกัด"); ไม่ขึ้นต้นด้วยคำนำหน้า = ใช้ชื่อเต็ม
// ช่องชื่อของใบแนบต้องมีคำนำหน้าตามแบบ ("ให้ระบุให้ชัดเจนว่าเป็น นาย นาง นางสาว หรือยศ") คำนำหน้าไทยจึงติดกับชื่อได้ ("นายสมชาย");
// ที่เหลือขึ้นต้นด้วยสระ/วรรณยุกต์ที่ขึ้นต้นคำไม่ได้ ("นายิกา" − "นาย" = "ิกา") = ตัดกลางพยางค์ ไม่ใช่คำนำหน้า → ใช้ชื่อเต็ม
func rdFileFirstName(name, title string) string {
	name, title = strings.TrimSpace(name), strings.TrimSpace(title)
	if title == "" || title == "-" || !strings.HasPrefix(name, title) {
		return name
	}
	rest := strings.TrimSpace(strings.TrimPrefix(name, title))
	if r, _ := utf8.DecodeRuneInString(rest); (r >= 0x0E30 && r <= 0x0E3A) || (r >= 0x0E45 && r <= 0x0E4E) {
		return name
	}
	return rest
}

// rdFileBranch - สาขา 5 หลักของหน้าแบบ → 6 หลักในไฟล์ (เติม 0 ข้างหน้า; 000000 = สำนักงานใหญ่) ว่าง/เกิน 6 หลักคงไว้ให้ Check แจ้ง
func rdFileBranch(raw string) string {
	branch := digitsOnly(raw)
	if branch == "" || len(branch) > 6 {
		return branch
	}
	return strings.Repeat("0", 6-len(branch)) + branch
}

var thaiSlashDate = regexp.MustCompile(`^([0-9]{1,2})/([0-9]{1,2})/([0-9]{4})$`)

// rdFileDate - "DD/MM/YYYY" (พ.ศ. ตามที่ prefill เขียน — thaiDate) → "DDMMYYYY"; รูปแบบอื่นคงไว้ให้ Check แจ้ง
func rdFileDate(raw string) string {
	m := thaiSlashDate.FindStringSubmatch(strings.TrimSpace(raw))
	if m == nil {
		return strings.TrimSpace(raw)
	}
	day, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	return fmt.Sprintf("%02d%02d%s", day, month, m[3])
}

// rdFileDecimal - ยอดเงิน/อัตรา → ทศนิยม 2 ตำแหน่ง (ตัดคอมมา); ทศนิยมเกิน 2 ตำแหน่งที่มีค่าจริงไม่ปัดเงียบ — คืนค่าเดิมให้ Check บล็อก
func rdFileDecimal(raw string) string {
	s := strings.ReplaceAll(strings.TrimSpace(raw), ",", "")
	if s == "" {
		return ""
	}
	d, err := decimal.NewFromString(s)
	if err != nil || !d.Equal(d.Round(2)) {
		return raw
	}
	return d.StringFixed(2)
}
