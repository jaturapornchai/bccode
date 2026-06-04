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
	GroupCode       string  `json:"group_code" ch:"group_code"`
	GroupsuboneCode string  `json:"groupsubonecode" ch:"groupsubonecode"`
	GroupsubtwoCode string  `json:"groupsubtwocode" ch:"groupsubtwocode"`
	BrandCode       string  `json:"brand_code" ch:"brand_code"`
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
	GUIDFixed   string `json:"guid_fixed" ch:"guid_fixed"`
	HoldingCode string `json:"holding_code" ch:"holding_code"`
	ProductImport
}

type ProductImportDoc struct {
	ProductImportInfo
	CreatedAt time.Time `json:"created_at" ch:"created_at"`
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
	PerPage   int64 `json:"per_page"`
	Prev      int64 `json:"prev"`
	Next      int64 `json:"next"`
	TotalPage int64 `json:"total_page"`
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
	TaskID          string                  `json:"task_id"`
	TotalRecords    int                     `json:"total_records"`
	NewRecords      int                     `json:"new_records"`
	ExistingRecords int                     `json:"existing_records"`
	UpdatedRecords  int                     `json:"updated_records"`
	ConflictRecords int                     `json:"conflict_records"`
	Items           []ProductBarcodeCompare `json:"items"`
	Summary         *CompareSummary         `json:"summary"`
	CreatedAt       time.Time               `json:"created_at"`
}

type ProductBarcodeCompare struct {
	ImportData   ProductImportRaw        `json:"import_data"`
	ExistingData *ProductBarcodeExisting `json:"existing_data,omitempty"`
	Status       string                  `json:"status"` // "NEW", "EXISTING", "UPDATE", "CONFLICT"
	Changes      []FieldChange           `json:"changes,omitempty"`
	Conflicts    []FieldConflict         `json:"conflicts,omitempty"`
	CanUpdate    bool                    `json:"can_update"`
	UpdateReason string                  `json:"update_reason,omitempty"`
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
	GroupCode       string `json:"group_code"`
	GroupsuboneCode string `json:"groupsubonecode"`
	GroupsubtwoCode string `json:"groupsubtwocode"`
	BrandCode       string `json:"brand_code"`
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

	LastModified time.Time `json:"last_modified"`
	ModifiedBy   string    `json:"modified_by"`
}

type FieldChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
	Reason   string      `json:"reason,omitempty"`
}

type FieldConflict struct {
	Field        string      `json:"field"`
	ImportValue  interface{} `json:"import_value"`
	CurrentValue interface{} `json:"current_value"`
	Severity     string      `json:"severity"` // "LOW", "MEDIUM", "HIGH"
	Suggestion   string      `json:"suggestion,omitempty"`
}

type CompareSummary struct {
	ImportMode        string    `json:"import_mode"`
	TotalChanges      int       `json:"total_changes"`
	PriceChanges      int       `json:"price_changes"`
	NameChanges       int       `json:"name_changes"`
	UnitChanges       int       `json:"unit_changes"`
	MasterDataChanges int       `json:"master_data_changes"`
	RefBarcodeChanges int       `json:"ref_barcode_changes"`
	TotalConflicts    int       `json:"total_conflicts"`
	HighConflicts     int       `json:"high_conflicts"`
	MediumConflicts   int       `json:"medium_conflicts"`
	LowConflicts      int       `json:"low_conflicts"`
	EstimatedTime     string    `json:"estimated_time"`
	CreatedAt         time.Time `json:"created_at"`
}

type PreviewResult struct {
	TaskID          string                  `json:"task_id"`
	ImportMode      string                  `json:"import_mode"`
	WillInsert      []ProductImportRaw      `json:"will_insert"`
	WillUpdate      []ProductBarcodeCompare `json:"will_update"`
	WillSkip        []ProductBarcodeCompare `json:"will_skip"`
	Warnings        []PreviewWarning        `json:"warnings"`
	EstimatedTime   string                  `json:"estimated_time"`
	RequiredActions []string                `json:"required_actions"`
	CreatedAt       time.Time               `json:"created_at"`
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
	TaskID        string        `json:"task_id"`
	ImportMode    string        `json:"import_mode"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	Duration      string        `json:"duration"`
	TotalRecords  int           `json:"total_records"`
	SuccessInsert int           `json:"success_insert"`
	SuccessUpdate int           `json:"success_update"`
	Failed        int           `json:"failed"`
	Skipped       int           `json:"skipped"`
	Errors        []ImportError `json:"errors,omitempty"`
	CreatedBy     string        `json:"created_by"`
	Status        string        `json:"status"` // "COMPLETED", "PARTIAL", "FAILED"
}

type ImportError struct {
	Barcode string `json:"barcode"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Type    string `json:"type"` // "VALIDATION", "DUPLICATE", "SYSTEM"
	Field   string `json:"field,omitempty"`
}
