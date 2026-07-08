package models

import (
	"time"

	common "smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BranchOrgDoc struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	HoldingCode string             `json:"holdingcode" bson:"holdingcode"`
	GuidFixed   string             `json:"guidfixed" bson:"guidfixed"`
	CompanyGuid string             `json:"companyguid" bson:"companyguid"`
	Code        string             `json:"code" bson:"code"`
	Names       common.JSONB       `json:"names" bson:"names"`
	LogoURI     string             `json:"logouri" bson:"logouri"`
	// Default locale/date-time settings carried on the branch so workspace
	// headers and date rendering have a value before the user edits the branch.
	// These mirror BranchData fields; kept here so the minimal head-office doc
	// created at company-create time already has sensible defaults.
	Timezone       string `json:"timezone" bson:"timezone"`
	TimezoneLabel  string `json:"timezonelabel" bson:"timezonelabel"`
	TimezoneOffset string `json:"timezoneoffset" bson:"timezoneoffset"`
	DateFormat     string `json:"dateformat" bson:"dateformat"`
	YearType       string `json:"yeartype" bson:"yeartype"`
	BaseCurrency   string `json:"basecurrency" bson:"basecurrency"`
	Language       string `json:"language" bson:"language"`
	// Tax / registration fields (ภ.พ.20). BranchType is one of
	// "head" | "permanent" | "temporary" (สำนักงานใหญ่ / สาขาถาวร / สาขาชั่วคราว).
	// IsVatRegistered + CompanyRegistrationNo mirror the flat fields on the full Branch struct.
	BranchType            string `json:"branchtype" bson:"branchtype"`
	IsVatRegistered       bool   `json:"isvatregistered" bson:"isvatregistered"`
	CompanyRegistrationNo string `json:"companyregistrationno" bson:"companyregistrationno"`
	Email                 string `json:"email" bson:"email"`
	ManagerName           string `json:"managername" bson:"managername"`
	// Legal/document address per language (used when printing documents). Each
	// entry is one language; the Address text itself may be multi-line (\n).
	Addresses []BranchAddress `json:"addresses" bson:"addresses"`
	// Structured geo-address (province/district/subdistrict/zipcode codes),
	// language-independent — display names come from the Thailand address
	// dataset lookup, not stored here. Separate from Addresses above.
	CountryCode     string `json:"countrycode" bson:"countrycode"`
	ProvinceCode    string `json:"provincecode" bson:"provincecode"`
	DistrictCode    string `json:"districtcode" bson:"districtcode"`
	SubDistrictCode string `json:"subdistrictcode" bson:"subdistrictcode"`
	ZipCode         string `json:"zipcode" bson:"zipcode"`
	// Document / accounting config — value-only. The running-number generator and
	// e-Tax submission engines are separate subsystems and are NOT implemented here;
	// these just store the per-branch settings they will read.
	FiscalStartMonth int8              `json:"fiscalstartmonth" bson:"fiscalstartmonth"`
	DocumentFormats  []BranchDocFormat `json:"documentformats" bson:"documentformats"`
	ETaxEnabled      bool              `json:"etaxenabled" bson:"etaxenabled"`
	IsActive         bool              `json:"isactive" bson:"isactive"`
	CreatedAt        time.Time         `json:"createdat" bson:"createdat"`
	UpdatedAt        time.Time         `json:"updatedat" bson:"updatedat"`
	DeletedAt        *time.Time        `json:"deletedat,omitempty" bson:"deletedat,omitempty"`
	CreatedBy        string            `json:"createdby,omitempty" bson:"createdby,omitempty"`
	UpdatedBy        string            `json:"updatedby,omitempty" bson:"updatedby,omitempty"`
	DeletedBy        string            `json:"deletedby,omitempty" bson:"deletedby,omitempty"`
}

func (BranchOrgDoc) CollectionName() string {
	return branchCollectionName
}

// BranchDocFormat = หนึ่งรูปแบบเลขที่เอกสาร โดยแต่ละประเภทเอกสาร (DocType) มีได้
// หลายรูปแบบ (multi-format ต่อประเภท). config-only เท่านั้น — running-number generator
// ยังไม่ใช้ค่าเหล่านี้. DocType = code ของประเภทเอกสาร เช่น SI, PO, SO, TF
// (ตรงกับ MODULE_NAME ของ transaction module). YearMode: none|be2|be4|ce2|ce4 ;
// ResetMode: none/never|yearly|monthly|daily.
type BranchDocFormat struct {
	DocType     string `json:"doctype" bson:"doctype"`
	Name        string `json:"name" bson:"name"`
	Prefix      string `json:"prefix" bson:"prefix"`
	UseBranch   bool   `json:"usebranch" bson:"usebranch"`
	YearMode    string `json:"yearmode" bson:"yearmode"`
	UseMonth    bool   `json:"usemonth" bson:"usemonth"`
	UseDay      bool   `json:"useday" bson:"useday"`
	Separator   bool   `json:"separator" bson:"separator"`
	RunLength   int    `json:"runlength" bson:"runlength"`
	ResetMode   string `json:"resetmode" bson:"resetmode"`
	StartNumber int    `json:"startnumber" bson:"startnumber"`
	Enabled     bool   `json:"enabled" bson:"enabled"`
	IsDefault   bool   `json:"isdefault" bson:"isdefault"`
}

// BranchAddress = ที่อยู่สาขาแยกตามภาษา ใช้สำหรับออกเอกสาร. Code = language code
// (เช่น th, en) ; Address = ข้อความที่อยู่ ป้อนได้หลายบรรทัด (เก็บ \n ตามที่ผู้ใช้ป้อน).
type BranchAddress struct {
	Code    string `json:"code" bson:"code"`
	Address string `json:"address" bson:"address"`
}
