package models

import (
	"math"
	"time"
)

type ProductImportHeader struct {
	LanguangeCode string `json:"languangecode" validate:"required,min=2,max=3"`
}

type ProductImportRaw struct {
	Barcode       string  `json:"barcode" ch:"barcode"`
	Code          string  `json:"code" ch:"code"`
	Name          string  `json:"name" ch:"name"`
	UnitCode      string  `json:"unitcode" ch:"unitcode"`
	Price         float64 `json:"price" ch:"price"`
	PriceMember   float64 `json:"pricemember" ch:"pricemember"`
	PriceDelivery float64 `json:"pricedelivery" ch:"pricedelivery"`

	PriceOne   float64 `json:"priceone" ch:"priceone"`
	PriceTwo   float64 `json:"pricetwo" ch:"pricetwo"`
	PriceThree float64 `json:"pricethree" ch:"pricethree"`
	PriceFour  float64 `json:"pricefour" ch:"pricefour"`
	PriceFive  float64 `json:"pricefive" ch:"pricefive"`
	PriceSix   float64 `json:"pricesix" ch:"pricesix"`
	PriceSeven float64 `json:"priseseven" ch:"priseseven"`
	PriceEight float64 `json:"priceeight" ch:"priceeight"`
	PriceNine  float64 `json:"pricenine" ch:"pricenine"`

	IsDuplicate    bool `json:"isduplicate" ch:"isduplicate"`
	IsExist        bool `json:"isexist" ch:"isexist"`
	IsUnitNotExist bool `json:"isunitnotexist" ch:"isunitnotexist"`
	// เพิ่มฟิลด์ใหม่
	GroupCode       string  `json:"groupcode" ch:"groupcode"`
	GroupsuboneCode string  `json:"groupsubonecode" ch:"groupsubonecode"`
	GroupsubtwoCode string  `json:"groupsubtwocode" ch:"groupsubtwocode"`
	BrandCode       string  `json:"brandcode" ch:"brandcode"`
	DesignCode      string  `json:"designcode" ch:"designcode"`
	ModelCode       string  `json:"modelcode" ch:"modelcode"`
	PatternCode     string  `json:"patterncode" ch:"patterncode"`
	GradeCode       string  `json:"gradecode" ch:"gradecode"`
	CategoryCode    string  `json:"categorycode" ch:"categorycode"`
	ClassCode       string  `json:"classcode" ch:"classcode"`
	BarcodeRef      string  `json:"barcoderef"  ch:"barcoderef"`
	StandValue      float64 `json:"standvalue"  ch:"standvalue"`
	DivideValue     float64 `json:"dividevalue"  ch:"dividevalue"`
	IsSumPoint      bool    `json:"issumpoint"  ch:"issumpoint"`
}

type ProductImport struct {
	TaskID    string  `json:"taskid" ch:"taskid"`
	RowNumber float64 `json:"rownumber" ch:"rownumber"`
	ProductImportRaw
}

type ProductImportInfo struct {
	GUIDFixed   string `json:"guidfixed" ch:"guidfixed"`
	HoldingCode string `json:"holdingcode" ch:"holdingcode"`
	ProductImport
}

type ProductImportDoc struct {
	ProductImportInfo
	CreatedAt time.Time `json:"createdat" ch:"createdat"`
	CreatedBy string    `json:"createdby" ch:"createdby"`
}

func (ProductImportDoc) TableName() string {
	return "productbarcodeimport"
}

type TaskStatus int8

const (
	TaskStatusPending TaskStatus = iota
	TaskStatusProcessing
	TaskStatusDone
	TaskStatusError
	TaskStatusSaveSucceded
	TaskStatusSaveFailed
	TaskStatusNotFound
)

type PaginationData struct {
	Total     int64 `json:"total"`
	Page      int64 `json:"page"`
	PerPage   int64 `json:"perpage"`
	Prev      int64 `json:"prev"`
	Next      int64 `json:"next"`
	TotalPage int64 `json:"totalpage"`
}

func (p *PaginationData) Build() {
	totalPage := math.Ceil(float64(p.Total) / float64(p.PerPage))
	p.TotalPage = int64(totalPage)

	if p.Page == 0 {
		p.Page = 1
	}

	if p.Page > 1 {
		p.Prev = p.Page - 1
	}

	if p.Page < p.TotalPage {
		p.Next = p.Page + 1
	}
}

// Enhanced models สำหรับ compare และ insert/update functionality
type CompareResult struct {
	TaskID          string                  `json:"taskid"`
	TotalRecords    int                     `json:"totalrecords"`
	NewRecords      int                     `json:"newrecords"`
	ExistingRecords int                     `json:"existingrecords"`
	UpdatedRecords  int                     `json:"updatedrecords"`
	ConflictRecords int                     `json:"conflictrecords"`
	Items           []ProductBarcodeCompare `json:"items"`
	Summary         *CompareSummary         `json:"summary"`
	CreatedAt       time.Time               `json:"createdat"`
}

type ProductBarcodeCompare struct {
	ImportData   ProductImportRaw        `json:"importdata"`
	ExistingData *ProductBarcodeExisting `json:"existingdata,omitempty"`
	Status       string                  `json:"status"` // "NEW", "EXISTING", "UPDATE", "CONFLICT"
	Changes      []FieldChange           `json:"changes,omitempty"`
	Conflicts    []FieldConflict         `json:"conflicts,omitempty"`
	CanUpdate    bool                    `json:"canupdate"`
	UpdateReason string                  `json:"updatereason,omitempty"`
}

type ProductBarcodeExisting struct {
	Barcode       string  `json:"barcode"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	UnitCode      string  `json:"unitcode"`
	Price         float64 `json:"price"`
	PriceMember   float64 `json:"pricemember"`
	PriceDelivery float64 `json:"pricedelivery"`

	// Extended price fields
	PriceOne   float64 `json:"priceone"`
	PriceTwo   float64 `json:"pricetwo"`
	PriceThree float64 `json:"pricethree"`
	PriceFour  float64 `json:"pricefour"`
	PriceFive  float64 `json:"pricefive"`
	PriceSix   float64 `json:"pricesix"`
	PriceSeven float64 `json:"priceseven"`
	PriceEight float64 `json:"priceeight"`
	PriceNine  float64 `json:"pricenine"`

	// Master data codes
	GroupCode       string `json:"groupcode"`
	GroupsuboneCode string `json:"groupsubonecode"`
	GroupsubtwoCode string `json:"groupsubtwocode"`
	BrandCode       string `json:"brandcode"`
	DesignCode      string `json:"designcode"`
	ModelCode       string `json:"modelcode"`
	PatternCode     string `json:"patterncode"`
	GradeCode       string `json:"gradecode"`
	CategoryCode    string `json:"categorycode"`
	ClassCode       string `json:"classcode"`
	IsSumPoint      bool   `json:"issumpoint"`

	// Reference barcode information
	BarcodeRef  string  `json:"barcoderef"`
	StandValue  float64 `json:"standvalue"`
	DivideValue float64 `json:"dividevalue"`

	LastModified time.Time `json:"lastmodified"`
	ModifiedBy   string    `json:"modifiedby"`
}

type FieldChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"oldvalue"`
	NewValue interface{} `json:"newvalue"`
	Reason   string      `json:"reason,omitempty"`
}

type FieldConflict struct {
	Field        string      `json:"field"`
	ImportValue  interface{} `json:"importvalue"`
	CurrentValue interface{} `json:"currentvalue"`
	Severity     string      `json:"severity"` // "LOW", "MEDIUM", "HIGH"
	Suggestion   string      `json:"suggestion,omitempty"`
}

type CompareSummary struct {
	ImportMode        string    `json:"importmode"`
	TotalChanges      int       `json:"totalchanges"`
	PriceChanges      int       `json:"pricechanges"`
	NameChanges       int       `json:"namechanges"`
	UnitChanges       int       `json:"unitchanges"`
	MasterDataChanges int       `json:"masterdatachanges"`
	RefBarcodeChanges int       `json:"refbarcodechanges"`
	TotalConflicts    int       `json:"totalconflicts"`
	HighConflicts     int       `json:"highconflicts"`
	MediumConflicts   int       `json:"mediumconflicts"`
	LowConflicts      int       `json:"lowconflicts"`
	EstimatedTime     string    `json:"estimatedtime"`
	CreatedAt         time.Time `json:"createdat"`
}

type PreviewResult struct {
	TaskID          string                  `json:"taskid"`
	ImportMode      string                  `json:"importmode"`
	WillInsert      []ProductImportRaw      `json:"willinsert"`
	WillUpdate      []ProductBarcodeCompare `json:"willupdate"`
	WillSkip        []ProductBarcodeCompare `json:"willskip"`
	Warnings        []PreviewWarning        `json:"warnings"`
	EstimatedTime   string                  `json:"estimatedtime"`
	RequiredActions []string                `json:"requiredactions"`
	CreatedAt       time.Time               `json:"createdat"`
}

type PreviewWarning struct {
	Type       string `json:"type"`     // "CONFLICT", "DUPLICATE", "VALIDATION"
	Severity   string `json:"severity"` // "LOW", "MEDIUM", "HIGH"
	Message    string `json:"message"`
	Field      string `json:"field,omitempty"`
	Barcode    string `json:"barcode,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

type ImportSummary struct {
	TaskID        string        `json:"taskid"`
	ImportMode    string        `json:"importmode"`
	StartTime     time.Time     `json:"starttime"`
	EndTime       time.Time     `json:"endtime"`
	Duration      string        `json:"duration"`
	TotalRecords  int           `json:"totalrecords"`
	SuccessInsert int           `json:"successinsert"`
	SuccessUpdate int           `json:"successupdate"`
	Failed        int           `json:"failed"`
	Skipped       int           `json:"skipped"`
	Errors        []ImportError `json:"errors,omitempty"`
	CreatedBy     string        `json:"createdby"`
	Status        string        `json:"status"` // "COMPLETED", "PARTIAL", "FAILED"
}

type ImportError struct {
	Barcode string `json:"barcode"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Type    string `json:"type"` // "VALIDATION", "DUPLICATE", "SYSTEM"
	Field   string `json:"field,omitempty"`
}
