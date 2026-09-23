package models

import (
	"errors"
	"time"

	common "smlcloudplatform/internal/models"
)

var (
	ErrBranchCompanyGuidRequired = errors.New("companyguid is required")
	ErrBranchCompanyNotFound     = errors.New("company not found")
)

// BranchOrgDoc is one row of the central branches table. In PostgreSQL the branch
// identity is (holding_code, company_code, code): guidfixed/branchuid carry the branch
// code and companyuid/companyguid carry the company code.
type BranchOrgDoc struct {
	HoldingCode  string       `json:"holdingcode"`
	HoldingUID   string       `json:"holdinguid"`
	GuidFixed    string       `json:"guidfixed"`
	CompanyGuid  string       `json:"companyguid"`
	CompanyUID   string       `json:"companyuid"`
	BranchUID    string       `json:"branchuid"`
	BusinessCode string       `json:"businesscode"`
	Code         string       `json:"code"`
	BranchCode   string       `json:"branchcode"`
	Names        common.JSONB `json:"names"`
	BranchSettings
	IsActive  bool      `json:"isactive"`
	CreatedAt time.Time `json:"createdat"`
	UpdatedAt time.Time `json:"updatedat"`
	CreatedBy string    `json:"createdby,omitempty"`
	UpdatedBy string    `json:"updatedby,omitempty"`
}

// BranchSettings is the value-only branch configuration stored in branches.settings.
type BranchSettings struct {
	BusinessTypes []string `json:"businesstypes"`
	LogoURI       string   `json:"logouri"`
	// Locale/date-time settings are explicit per branch. Timezone is the IANA
	// source of truth; label and offset are display-only derived values.
	Timezone       string `json:"timezone"`
	TimezoneLabel  string `json:"timezonelabel"`
	TimezoneOffset string `json:"timezoneoffset"`
	DateFormat     string `json:"dateformat"`
	YearType       string `json:"yeartype"`
	BaseCurrency   string `json:"basecurrency"`
	Language       string `json:"language"`
	// Tax / registration fields (ภ.พ.20). BranchType is one of
	// "head" | "permanent" | "temporary" (สำนักงานใหญ่ / สาขาถาวร / สาขาชั่วคราว).
	BranchType            string `json:"branchtype"`
	IsVatRegistered       bool   `json:"isvatregistered"`
	CompanyRegistrationNo string `json:"companyregistrationno"`
	Email                 string `json:"email"`
	ManagerName           string `json:"managername"`
	// Legal/document address per language (used when printing documents). Each
	// entry is one language; the Address text itself may be multi-line (\n).
	Addresses []BranchAddress `json:"addresses"`
	// Structured geo-address codes, language-independent — display names come from
	// the Thailand address dataset lookup, not stored here.
	CountryCode     string `json:"countrycode"`
	ProvinceCode    string `json:"provincecode"`
	DistrictCode    string `json:"districtcode"`
	SubDistrictCode string `json:"subdistrictcode"`
	ZipCode         string `json:"zipcode"`
	// Document / accounting config — value-only. The running-number generator and
	// e-Tax submission engines are separate subsystems that read these settings.
	FiscalStartMonth int8              `json:"fiscalstartmonth"`
	DocumentFormats  []BranchDocFormat `json:"documentformats"`
	ETaxEnabled      bool              `json:"etaxenabled"`
}

// BranchDocFormat = หนึ่งรูปแบบเลขที่เอกสาร โดยแต่ละประเภทเอกสาร (DocType) มีได้
// หลายรูปแบบ (multi-format ต่อประเภท). config-only เท่านั้น — running-number generator
// ยังไม่ใช้ค่าเหล่านี้. DocType = code ของประเภทเอกสาร เช่น SI, PO, SO, TF
// (ตรงกับ MODULE_NAME ของ transaction module). YearMode: none|be2|be4|ce2|ce4 ;
// ResetMode: none/never|yearly|monthly|daily.
type BranchDocFormat struct {
	DocType     string `json:"doctype"`
	Name        string `json:"name"`
	Prefix      string `json:"prefix"`
	UseBranch   bool   `json:"usebranch"`
	YearMode    string `json:"yearmode"`
	UseMonth    bool   `json:"usemonth"`
	UseDay      bool   `json:"useday"`
	Separator   bool   `json:"separator"`
	RunLength   int    `json:"runlength"`
	ResetMode   string `json:"resetmode"`
	StartNumber int    `json:"startnumber"`
	Enabled     bool   `json:"enabled"`
	IsDefault   bool   `json:"isdefault"`
}

// BranchAddress = ที่อยู่สาขาแยกตามภาษา ใช้สำหรับออกเอกสาร. Code = language code
// (เช่น th, en) ; Address = ข้อความที่อยู่ ป้อนได้หลายบรรทัด (เก็บ \n ตามที่ผู้ใช้ป้อน).
type BranchAddress struct {
	Code    string `json:"code"`
	Address string `json:"address"`
}
