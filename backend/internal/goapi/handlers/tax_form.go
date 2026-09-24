package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lib/pq"

	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/rdform"
	msmodels "smlcloudplatform/pkg/microservice/models"
)

// แบบยื่นภาษีของกรมสรรพากร (ภ.ง.ด./ภ.พ./ภ.ธ.): จอกรอกแบบสร้างจากสเปกของ internal/rdform,
// backend เติมค่าจากบัญชีแยกประเภท (prefill) คำนวณยอดรวมตามสูตรที่พิมพ์บนแบบ (compute) บันทึกฉบับร่าง
// ในตาราง tax_filings ของกลุ่มกิจการ และพิมพ์ลงแบบฟอร์มทางการ (pdf) — ทุกช่องแก้ได้เสมอ รวมฐานภาษี

// taxFormMeta - งวดของแบบ + ผู้ยื่น (individual = บุคคลธรรมดา/ห้างสามัญ ไม่ใช่บริษัทที่เลือก)
type taxFormMeta struct {
	Period     string `json:"period"` // month | year
	Source     string `json:"source"` // wht | vat | cit | manual — แหล่งที่ระบบเติมค่าให้
	Individual bool   `json:"individual,omitempty"`
}

// taxFormOrder - ลำดับบนจอ: ภาษีหัก ณ ที่จ่าย → ภาษีมูลค่าเพิ่ม/ธุรกิจเฉพาะ → ภาษีเงินได้นิติบุคคล → บุคคลธรรมดา
// ไม่มี ภ.ง.ด.1/1ก (ระบบเงินเดือนอยู่นอกขอบเขตตามกฎโปรเจกต์)
var taxFormOrder = []string{"pnd3", "pnd53", "pnd2", "pnd2a", "pp30", "pp36", "pbt40", "pnd51", "pnd50", "pnd94", "pnd93"}

var taxForms = map[string]taxFormMeta{
	"pnd3":  {Period: "month", Source: "wht"},
	"pnd53": {Period: "month", Source: "wht"},
	"pnd2":  {Period: "month", Source: "wht"},
	"pnd2a": {Period: "year", Source: "wht"},
	"pp30":  {Period: "month", Source: "vat"},
	"pp36":  {Period: "month", Source: "manual"},
	"pbt40": {Period: "month", Source: "manual"},
	"pnd51": {Period: "year", Source: "cit"},
	"pnd50": {Period: "year", Source: "cit"},
	"pnd94": {Period: "year", Source: "manual", Individual: true},
	"pnd93": {Period: "year", Source: "manual", Individual: true},
}

type taxFormRequest struct {
	HoldingCode  string          `json:"holdingcode"`
	BusinessCode string          `json:"businesscode"`
	ID           int64           `json:"id,omitempty"`
	Version      int             `json:"version,omitempty"`
	Code         string          `json:"code"`
	Year         int             `json:"year"`  // ค.ศ. (ปีภาษี/ปีของวันสิ้นรอบบัญชีสำหรับแบบรายปี)
	Month        int             `json:"month"` // 1-12 สำหรับแบบรายเดือน, 0 สำหรับแบบรายปี
	Document     rdform.Document `json:"document"`
}

// TaxFiling - ฉบับที่บันทึกไว้
type TaxFiling struct {
	ID        int64            `json:"id"`
	Code      string           `json:"code"`
	Year      int              `json:"year"`
	Month     int              `json:"month"`
	FilingSeq int              `json:"filingseq"` // 0 = ยื่นปกติ, n = ยื่นเพิ่มเติมครั้งที่ n
	Version   int              `json:"version"`
	UpdatedBy string           `json:"updatedby"`
	UpdatedAt string           `json:"updatedat"`
	Document  *rdform.Document `json:"document,omitempty"`
}

// taxFormCall - บริบทที่ทุก endpoint ใช้ร่วมกัน (ตรวจสิทธิ์บริษัท + ภาษาของข้อความผิดพลาด)
type taxFormCall struct {
	c        echo.Context
	lang     string
	req      taxFormRequest
	holding  string
	company  string
	username string
}

func (t *taxFormCall) fail(status int, key string, field ...string) error {
	body := map[string]any{"success": false, "code": key, "message": language.Text(key, t.lang)}
	if len(field) > 0 {
		body["field"] = field[0]
	}
	return t.c.JSON(status, body)
}

func (t *taxFormCall) failValue(err error) error {
	var fe *rdform.FieldError
	if errors.As(err, &fe) {
		body := map[string]any{"success": false, "code": "tax_form_value_invalid", "field": fe.Key, "message": language.Text("tax_form_value_invalid", t.lang)}
		if fe.Row > 0 {
			body["row"] = fe.Row
		}
		return t.c.JSON(http.StatusBadRequest, body)
	}
	return t.fail(http.StatusBadRequest, "tax_form_value_invalid")
}

// begin - อ่าน payload + ตรวจสิทธิ์บริษัท; needCode = ต้องระบุแบบที่ระบบรองรับ
func beginTaxForm(c echo.Context, needCode bool) (*taxFormCall, error) {
	t := &taxFormCall{c: c, lang: taxRequestLanguage(c)}
	if err := c.Bind(&t.req); err != nil {
		return nil, t.fail(http.StatusBadRequest, "tax_form_payload_invalid")
	}
	holding, company, scopeErr := authenticatedCompanyContext(c, t.req.HoldingCode, t.req.BusinessCode)
	if scopeErr != nil {
		return nil, taxScopeFail(c, scopeErr)
	}
	t.holding, t.company = holding, company
	if info, ok := c.Get("UserInfo").(msmodels.UserInfo); ok {
		t.username = info.Username
	}
	t.req.Code = strings.TrimSpace(t.req.Code)
	if _, ok := taxForms[t.req.Code]; needCode && !ok {
		return nil, t.fail(http.StatusBadRequest, "tax_form_unknown", "code")
	}
	return t, nil
}

// validPeriod - แบบรายเดือนต้องมีเดือน, แบบรายปี month = 0
func (t *taxFormCall) validPeriod() bool {
	if t.req.Year < 2000 || t.req.Year > 2200 {
		return false
	}
	if taxForms[t.req.Code].Period == "month" {
		return t.req.Month >= 1 && t.req.Month <= 12
	}
	return t.req.Month == 0
}

func (t *taxFormCall) db(ctx context.Context) (*sql.DB, error) {
	db, err := mypg.PgSqlFastConnect(t.holding) // pool กลางของกลุ่มกิจการ ห้าม Close
	if err != nil {
		return nil, err
	}
	return db, ensureTaxFilingSchema(ctx, db)
}

// TaxFormCatalogHandler - POST /api/report/tax/form/catalog: รายการแบบที่ระบบกรอก/พิมพ์ได้
func TaxFormCatalogHandler(c echo.Context) error {
	t, err := beginTaxForm(c, false)
	if t == nil {
		return err
	}
	type entry struct {
		Code          string `json:"code"`
		Title         string `json:"title"`
		HasAttachment bool   `json:"hasattachment"`
		// RdFile - แบบนี้สร้างไฟล์ยื่นภาษีด้วยสื่อ (Format กลาง) ได้: จอแสดงปุ่มสร้างไฟล์ตามค่านี้ ไม่ฝังรายการแบบไว้ที่ frontend
		RdFile bool `json:"rdfile,omitempty"`
		taxFormMeta
	}
	out := make([]entry, 0, len(taxFormOrder))
	for _, code := range taxFormOrder {
		spec, err := rdform.Lookup(code)
		if err != nil {
			logger.Error("TaxFormCatalog: %v", err)
			return t.fail(http.StatusInternalServerError, "tax_form_failed")
		}
		_, rdFile := rdFileKinds[code] // ภ.ง.ด.53/3/2 — รายการเดียวกับที่ TaxFormRdFileHandler รองรับ
		out = append(out, entry{Code: code, Title: spec.Title, HasAttachment: rdform.Attachment(code) != nil, RdFile: rdFile, taxFormMeta: taxForms[code]})
	}
	return c.JSON(http.StatusOK, map[string]any{"success": true, "data": out})
}

// TaxFormSchemaHandler - POST /api/report/tax/form/schema: ช่องของแบบสำหรับจอกรอก
func TaxFormSchemaHandler(c echo.Context) error {
	t, err := beginTaxForm(c, true)
	if t == nil {
		return err
	}
	schema, err := rdform.SchemaFor(t.req.Code)
	if err != nil {
		logger.Error("TaxFormSchema: %v", err)
		return t.fail(http.StatusInternalServerError, "tax_form_failed")
	}
	return c.JSON(http.StatusOK, map[string]any{"success": true, "data": schema, "meta": taxForms[t.req.Code]})
}

// TaxFormPrefillHandler - POST /api/report/tax/form/prefill: ค่าเริ่มต้นของงวดจากทะเบียนบริษัท ฉบับก่อน และบัญชีแยกประเภท
func TaxFormPrefillHandler(c echo.Context) error {
	t, err := beginTaxForm(c, true)
	if t == nil {
		return err
	}
	if !t.validPeriod() {
		return t.fail(http.StatusBadRequest, "tax_form_period_invalid", "period")
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 30*time.Second)
	defer cancel()
	db, err := t.db(ctx)
	if err != nil {
		logger.Error("TaxFormPrefill: db: %v", err)
		return t.fail(http.StatusInternalServerError, "tax_form_failed")
	}
	doc, notes, err := prefillTaxForm(ctx, db, t.holding, t.company, t.req.Code, t.req.Year, t.req.Month, time.Now())
	if err != nil {
		logger.Error("TaxFormPrefill %s: %v", t.req.Code, err)
		return t.fail(http.StatusInternalServerError, "tax_form_failed")
	}
	return c.JSON(http.StatusOK, map[string]any{"success": true, "data": doc, "notes": notes})
}

// TaxFormComputeHandler - POST /api/report/tax/form/compute: คำนวณยอดรวม/ยอดสุทธิตามสูตรที่พิมพ์บนแบบใหม่
// จากค่าที่ผู้ใช้แก้ (ฐานภาษีและยอดรายการเป็นของผู้ใช้เสมอ ระบบคำนวณเฉพาะบรรทัดรวม)
func TaxFormComputeHandler(c echo.Context) error {
	t, err := beginTaxForm(c, true)
	if t == nil {
		return err
	}
	doc, err := prepareTaxDocument(t.req.Code, t.req.Document)
	if err != nil {
		return t.failValue(err)
	}
	res := map[string]any{"success": true, "data": doc}
	if taxForms[t.req.Code].Source == "wht" {
		// หมายเหตุที่ตรวจจากเอกสารได้ ตรวจใหม่ทุกครั้ง: frontend แทนหมายเหตุ key ใน rechecked ด้วยชุดใหม่ (แก้แล้วหายเอง)
		res["notes"] = append(missingTaxIDNotes(doc.Rows), missingIncomeTypeNotes(t.req.Code, doc.Rows)...)
		res["rechecked"] = []string{"tax_form_note_missing_taxid", "tax_form_note_missing_income_type"}
	}
	return c.JSON(http.StatusOK, res)
}

// TaxFormPDFHandler - POST /api/report/tax/form/pdf → application/pdf บนแบบฟอร์มทางการ
func TaxFormPDFHandler(c echo.Context) error {
	t, err := beginTaxForm(c, true)
	if t == nil {
		return err
	}
	// พิมพ์ยอดรวมที่คำนวณใหม่เสมอ — ยอดรวมจาก browser แก้ได้ จึงเชื่อได้แค่ยอดรายการ
	doc, err := prepareTaxDocument(t.req.Code, t.req.Document)
	if err != nil {
		return t.failValue(err)
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 30*time.Second)
	defer cancel()
	if !taxForms[t.req.Code].Individual {
		// ชื่อ/เลขผู้เสียภาษีของผู้ยื่นยึดทะเบียนบริษัทเสมอเมื่อมีค่า (กันพิมพ์ในนามบริษัทอื่น) เหมือน 50 ทวิ
		company, err := whtCompanyHeader(ctx, t.holding, t.company)
		if err != nil {
			logger.Error("TaxFormPDF: company header: %v", err)
			return t.fail(http.StatusInternalServerError, "tax_form_failed")
		}
		applyCompanyHeader(doc.Values, company)
	}
	pdf, err := rdform.Render(t.req.Code, doc)
	if errors.Is(err, rdform.ErrInvalidValue) {
		return t.failValue(err)
	}
	if err != nil {
		logger.Error("TaxFormPDF %s: %v", t.req.Code, err)
		return t.fail(http.StatusInternalServerError, "tax_form_failed")
	}
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s-%04d%02d.pdf"`, t.req.Code, t.req.Year, t.req.Month))
	c.Response().Header().Set("Cache-Control", "no-store")
	return c.Blob(http.StatusOK, "application/pdf", pdf)
}

func applyCompanyHeader(values map[string]string, company CompanyHeader) {
	if name := strings.TrimSpace(company.Name); name != "" {
		values["name"] = name
	}
	if taxID := digitsOnly(company.TaxID); len(taxID) == 13 {
		values["tax_id"] = taxID
	}
}

// prepareTaxDocument - เอกสารที่จะบันทึก/พิมพ์/ส่งกลับ: ตัดช่องว่าง → ตรวจค่า → คำนวณบรรทัดรวมใหม่ด้วยสูตรเดียวกับ /compute
// (ยอดรวมที่ส่งมาถูกแทนด้วยผลคำนวณเสมอ ยอดรวมที่ถูกแก้จึงไม่ถูกบันทึกหรือพิมพ์)
func prepareTaxDocument(code string, in rdform.Document) (rdform.Document, error) {
	doc := normalizeDocument(in)
	if err := rdform.Validate(code, doc); err != nil {
		return doc, err
	}
	err := computeTaxForm(code, &doc)
	return doc, err
}

// normalizeDocument - map ว่างแทน nil และตัดช่องว่างหัวท้าย (ผู้ใช้ก๊อปจาก Excel มักติดช่องว่าง)
func normalizeDocument(doc rdform.Document) rdform.Document {
	trim := func(m map[string]string) map[string]string {
		out := make(map[string]string, len(m))
		for k, v := range m {
			out[k] = strings.TrimSpace(v)
		}
		return out
	}
	out := rdform.Document{Values: trim(doc.Values)}
	for _, r := range doc.Rows {
		out.Rows = append(out.Rows, trim(r))
	}
	for _, s := range doc.Sheets {
		out.Sheets = append(out.Sheets, trim(s))
	}
	return out
}

// filingSeq - ยื่นปกติ = 0, ยื่นเพิ่มเติม = ครั้งที่ (ต้องเป็นเลข 1-99)
func filingSeq(values map[string]string) (int, bool) {
	if values["filing_type"] != "additional" {
		return 0, true
	}
	n, err := strconv.Atoi(strings.TrimSpace(values["additional_no"]))
	return n, err == nil && n >= 1 && n <= 99
}

// TaxFormSaveHandler - POST /api/report/tax/form/save: บันทึกฉบับร่าง (ใหม่ หรือแก้ตาม id + version)
func TaxFormSaveHandler(c echo.Context) error {
	t, err := beginTaxForm(c, true)
	if t == nil {
		return err
	}
	if !t.validPeriod() {
		return t.fail(http.StatusBadRequest, "tax_form_period_invalid", "period")
	}
	// เก็บยอดรวมที่คำนวณใหม่เสมอ (ไม่เชื่อยอดรวมจาก browser)
	doc, err := prepareTaxDocument(t.req.Code, t.req.Document)
	if err != nil {
		return t.failValue(err)
	}
	seq, ok := filingSeq(doc.Values)
	if !ok {
		return t.fail(http.StatusBadRequest, "tax_form_value_invalid", "additional_no")
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 15*time.Second)
	defer cancel()
	db, err := t.db(ctx)
	if err != nil {
		logger.Error("TaxFormSave: db: %v", err)
		return t.fail(http.StatusInternalServerError, "tax_form_failed")
	}
	filing := TaxFiling{ID: t.req.ID, Code: t.req.Code, Year: t.req.Year, Month: t.req.Month, FilingSeq: seq, Version: t.req.Version, Document: &doc}
	switch err := saveTaxFiling(ctx, db, t.company, t.username, &filing); {
	case errors.Is(err, errTaxFilingConflict):
		return t.fail(http.StatusConflict, "tax_form_version_conflict")
	case errors.Is(err, errTaxFilingDuplicate):
		return t.fail(http.StatusConflict, "tax_form_duplicate")
	case err != nil:
		logger.Error("TaxFormSave: %v", err)
		return t.fail(http.StatusInternalServerError, "tax_form_failed")
	}
	filing.Document = nil
	return c.JSON(http.StatusOK, map[string]any{"success": true, "data": filing})
}

// TaxFormListHandler - POST /api/report/tax/form/list: ฉบับที่บันทึกไว้ของบริษัท (กรองแบบ/ปีได้)
func TaxFormListHandler(c echo.Context) error {
	t, err := beginTaxForm(c, false)
	if t == nil {
		return err
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 15*time.Second)
	defer cancel()
	db, err := t.db(ctx)
	if err != nil {
		logger.Error("TaxFormList: db: %v", err)
		return t.fail(http.StatusInternalServerError, "tax_form_failed")
	}
	rows, err := listTaxFilings(ctx, db, t.company, t.req.Code, t.req.Year)
	if err != nil {
		logger.Error("TaxFormList: %v", err)
		return t.fail(http.StatusInternalServerError, "tax_form_failed")
	}
	return c.JSON(http.StatusOK, map[string]any{"success": true, "data": rows})
}

// TaxFormLoadHandler - POST /api/report/tax/form/load {id}
func TaxFormLoadHandler(c echo.Context) error {
	t, err := beginTaxForm(c, false)
	if t == nil {
		return err
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 15*time.Second)
	defer cancel()
	db, err := t.db(ctx)
	if err != nil {
		logger.Error("TaxFormLoad: db: %v", err)
		return t.fail(http.StatusInternalServerError, "tax_form_failed")
	}
	filing, err := loadTaxFiling(ctx, db, t.company, t.req.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return t.fail(http.StatusNotFound, "tax_form_not_found")
	}
	if err != nil {
		logger.Error("TaxFormLoad: %v", err)
		return t.fail(http.StatusInternalServerError, "tax_form_failed")
	}
	return c.JSON(http.StatusOK, map[string]any{"success": true, "data": filing})
}

// TaxFormDeleteHandler - POST /api/report/tax/form/delete {id, version}: ลบฉบับร่าง (ประวัติทุกรุ่นยังอยู่ใน tax_filing_history)
func TaxFormDeleteHandler(c echo.Context) error {
	t, err := beginTaxForm(c, false)
	if t == nil {
		return err
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 15*time.Second)
	defer cancel()
	db, err := t.db(ctx)
	if err != nil {
		logger.Error("TaxFormDelete: db: %v", err)
		return t.fail(http.StatusInternalServerError, "tax_form_failed")
	}
	switch err := deleteTaxFiling(ctx, db, t.company, t.req.ID, t.req.Version); {
	case errors.Is(err, errTaxFilingNotFound):
		return t.fail(http.StatusNotFound, "tax_form_not_found")
	case errors.Is(err, errTaxFilingConflict):
		return t.fail(http.StatusConflict, "tax_form_version_conflict")
	case err != nil:
		logger.Error("TaxFormDelete: %v", err)
		return t.fail(http.StatusInternalServerError, "tax_form_failed")
	}
	return c.JSON(http.StatusOK, map[string]any{"success": true})
}

// ---- storage ----

var (
	errTaxFilingConflict  = errors.New("tax filing changed by someone else")
	errTaxFilingDuplicate = errors.New("tax filing for this period already exists")
	errTaxFilingNotFound  = errors.New("tax filing not found")
)

// deleteTaxFiling - ลบตาม id + version ของบริษัท; ไม่มีฉบับนี้ (หรือเป็นของบริษัทอื่น) = not found ไม่ใช่ "มีผู้อื่นแก้"
func deleteTaxFiling(ctx context.Context, db *sql.DB, company string, id int64, version int) error {
	res, err := db.ExecContext(ctx, `DELETE FROM tax_filings WHERE id=$1 AND company_code=$2 AND version=$3`, id, company, version)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil || n > 0 {
		return err
	}
	var exists bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM tax_filings WHERE id=$1 AND company_code=$2)`, id, company).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return errTaxFilingNotFound
	}
	return errTaxFilingConflict
}

// ensureTaxFilingSchema - ตารางของกลุ่มกิจการ (โค้ดใหม่คือ migration ตามกฎ "เดินหน้าอย่างเดียว")
// ยอดเงินในเอกสารเป็นสตริงทศนิยมใน JSONB ไม่ใช่ตัวเลข float
func ensureTaxFilingSchema(ctx context.Context, db *sql.DB) error {
	gate, _ := taxFilingSchemaGates.LoadOrStore(db, &taxFilingSchemaGate{})
	g := gate.(*taxFilingSchemaGate)
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.done {
		return nil
	}
	if err := migrateTaxFilingSchema(ctx, db); err != nil {
		return err // ไม่จำว่าเสร็จ — request ถัดไปลองใหม่
	}
	g.done = true
	return nil
}

// taxFilingSchemaGates - ตรวจโครงตาราง tax_filings ครั้งเดียวต่อ pool ของกลุ่มกิจการต่อ process
// (เดิมทุก request รัน CREATE IF NOT EXISTS + ค้น information_schema); key = *sql.DB จาก PgSqlFastConnect
// pool ที่สร้างใหม่ได้ pointer ใหม่ จึงถูกตรวจใหม่เอง
var taxFilingSchemaGates sync.Map

type taxFilingSchemaGate struct {
	mu   sync.Mutex
	done bool
}

// migrateTaxFilingSchema - สร้าง/ขยายตารางแบบยื่นภาษี (เรียกผ่าน ensureTaxFilingSchema เท่านั้น)
func migrateTaxFilingSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS tax_filings (
    id           BIGSERIAL PRIMARY KEY,
    company_code TEXT NOT NULL,
    form_code    VARCHAR(30) NOT NULL,
    period_year  INT NOT NULL CHECK (period_year BETWEEN 2000 AND 2200),
    period_month SMALLINT NOT NULL CHECK (period_month BETWEEN 0 AND 12),
    filing_seq   SMALLINT NOT NULL DEFAULT 0 CHECK (filing_seq BETWEEN 0 AND 99),
    document     JSONB NOT NULL,
    version      INT NOT NULL DEFAULT 1,
    created_by   TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by   TEXT NOT NULL DEFAULT '',
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_tax_filings_period UNIQUE (company_code, form_code, period_year, period_month, filing_seq)
);
CREATE TABLE IF NOT EXISTS tax_filing_history (
    id        BIGSERIAL PRIMARY KEY,
    filing_id BIGINT NOT NULL,
    version   INT NOT NULL,
    document  JSONB NOT NULL,
    saved_by  TEXT NOT NULL DEFAULT '',
    saved_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_tax_filing_history_filing ON tax_filing_history(filing_id, version);
-- Tables made before these columns became TEXT (company codes and usernames/emails did not fit
-- VARCHAR(20)/(100)): widen once; later calls find nothing to change.
DO $$
DECLARE col record;
BEGIN
    FOR col IN SELECT table_name, column_name FROM information_schema.columns
        WHERE table_schema = current_schema() AND data_type <> 'text'
          AND ((table_name = 'tax_filings' AND column_name IN ('company_code', 'created_by', 'updated_by'))
            OR (table_name = 'tax_filing_history' AND column_name = 'saved_by'))
    LOOP
        EXECUTE format('ALTER TABLE %I ALTER COLUMN %I TYPE TEXT', col.table_name, col.column_name);
    END LOOP;
END $$;`)
	if err != nil {
		return fmt.Errorf("ensure tax_filings: %w", err)
	}
	return nil
}

func saveTaxFiling(ctx context.Context, db *sql.DB, company, user string, f *TaxFiling) error {
	raw, err := json.Marshal(f.Document)
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if f.ID == 0 {
		err = tx.QueryRowContext(ctx, `INSERT INTO tax_filings (company_code, form_code, period_year, period_month, filing_seq, document, created_by, updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$7) RETURNING id, version`, company, f.Code, f.Year, f.Month, f.FilingSeq, raw, user).Scan(&f.ID, &f.Version)
	} else {
		err = tx.QueryRowContext(ctx, `UPDATE tax_filings SET form_code=$3, period_year=$4, period_month=$5, filing_seq=$6, document=$7,
version=version+1, updated_by=$8, updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND company_code=$2 AND version=$9 RETURNING version`,
			f.ID, company, f.Code, f.Year, f.Month, f.FilingSeq, raw, user, f.Version).Scan(&f.Version)
		if errors.Is(err, sql.ErrNoRows) {
			return errTaxFilingConflict
		}
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return errTaxFilingDuplicate
	}
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO tax_filing_history (filing_id, version, document, saved_by) VALUES ($1,$2,$3,$4)`, f.ID, f.Version, raw, user); err != nil {
		return err
	}
	if err = tx.QueryRowContext(ctx, `SELECT updated_by, to_char(updated_at AT TIME ZONE 'Asia/Bangkok','YYYY-MM-DD"T"HH24:MI:SS') FROM tax_filings WHERE id=$1`, f.ID).Scan(&f.UpdatedBy, &f.UpdatedAt); err != nil {
		return err
	}
	return tx.Commit()
}

const taxFilingColumns = `id, form_code, period_year, period_month, filing_seq, version, updated_by, to_char(updated_at AT TIME ZONE 'Asia/Bangkok','YYYY-MM-DD"T"HH24:MI:SS')`

func scanTaxFiling(scan func(...any) error, f *TaxFiling) error {
	return scan(&f.ID, &f.Code, &f.Year, &f.Month, &f.FilingSeq, &f.Version, &f.UpdatedBy, &f.UpdatedAt)
}

func listTaxFilings(ctx context.Context, db *sql.DB, company, code string, year int) ([]TaxFiling, error) {
	rows, err := db.QueryContext(ctx, `SELECT `+taxFilingColumns+` FROM tax_filings
WHERE company_code=$1 AND ($2='' OR form_code=$2) AND ($3=0 OR period_year=$3)
ORDER BY period_year DESC, period_month DESC, form_code, filing_seq LIMIT 500`, company, code, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []TaxFiling{}
	for rows.Next() {
		var f TaxFiling
		if err := scanTaxFiling(rows.Scan, &f); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func loadTaxFiling(ctx context.Context, db *sql.DB, company string, id int64) (TaxFiling, error) {
	var f TaxFiling
	var raw []byte
	row := db.QueryRowContext(ctx, `SELECT `+taxFilingColumns+`, document FROM tax_filings WHERE id=$1 AND company_code=$2`, id, company)
	if err := row.Scan(&f.ID, &f.Code, &f.Year, &f.Month, &f.FilingSeq, &f.Version, &f.UpdatedBy, &f.UpdatedAt, &raw); err != nil {
		return f, err
	}
	f.Document = &rdform.Document{}
	return f, json.Unmarshal(raw, f.Document)
}
