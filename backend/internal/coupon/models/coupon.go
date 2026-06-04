package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const couponCollectionName = "coupons"

type CouponType int8

const (
	CouponTypeValueDiscount   CouponType = 0 // ลดตามมูลค่า
	CouponTypePercentDiscount CouponType = 1 // ลดตามเปอร์เซ็นต์
	CouponTypeCashVoucher     CouponType = 2 // คูปองแทนเงินสด
)

type CouponStatus int8

const (
	CouponStatusActive   CouponStatus = 0 // ปกติ
	CouponStatusCanceled CouponStatus = 1 // ยกเลิก
)

// เงื่อนไขสินค้าที่ใช้กับคูปอง
type CouponProductCondition struct {
	ProductCodes     []string `json:"product_codes" bson:"product_codes"`           // รายการ Code ของสินค้า
	GroupCodes       []string `json:"group_codes" bson:"group_codes"`               // รายการ GroupCode
	GroupSubOneCodes []string `json:"group_subone_codes" bson:"group_subone_codes"` // รายการ GroupsuboneCode
	GroupSubTwoCodes []string `json:"group_subtwo_codes" bson:"group_subtwo_codes"` // รายการ GroupsubtwoCode
	BrandCodes       []string `json:"brand_codes" bson:"brand_codes"`               // รายการ BrandCode
	DesignCodes      []string `json:"design_codes" bson:"design_codes"`             // รายการ DesignCode
	ModelCodes       []string `json:"model_codes" bson:"model_codes"`               // รายการ ModelCode
	PatternCodes     []string `json:"pattern_codes" bson:"pattern_codes"`           // รายการ PatternCode
	GradeCodes       []string `json:"grade_codes" bson:"grade_codes"`               // รายการ GradeCode
	CategoryCodes    []string `json:"category_codes" bson:"category_codes"`         // รายการ CategoryCode
	ClassCodes       []string `json:"class_codes" bson:"class_codes"`               // รายการ ClassCode
	MinimumAmount    float64  `json:"minimum_amount" bson:"minimum_amount"`         // มูลค่าขั้นต่ำของสินค้าที่เข้าเงื่อนไข (0 = ไม่จำกัด)
}

type Coupon struct {
	models.PartitionIdentity `bson:"inline"`
	CouponCode               string          `json:"coupon_code" bson:"coupon_code"` // หมายเลขคูปอง
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	CouponValue              float64         `json:"couponvalue" bson:"couponvalue"`                           // มูลค่าคูปอง
	IssuedDate               time.Time       `json:"issued_date" bson:"issued_date"`                           // วันที่ออกคูปอง
	ExpiryDate               time.Time       `json:"expiry_date" bson:"expiry_date"`                           // วันที่หมดอายุคูปอง
	CouponType               CouponType      `json:"coupon_type" bson:"coupon_type"`                           // ประเภทคูปอง (ลดตามมูลค่า/ลดตามเปอร์เซ็นต์)
	CustomerCodes            []string        `json:"customer_codes" bson:"customer_codes"`                     // รหัสลูกค้าที่สามารถใช้ได้ ([] = ทุกคน)
	Remark                   string          `json:"remark" bson:"remark"`                                     // หมายเหตุ
	Status                   CouponStatus    `json:"status" bson:"status"`                                     // สถานะ ยกเลิก (ปกติ/ยกเลิก)
	IsOneTimeUse             bool            `json:"isonetimeuse" bson:"isonetimeuse"`                         // ใช้ครั้งเดียว
	MaxUsageCount            int             `json:"maxusagecount" bson:"maxusagecount"`                       // จำนวนครั้งสูงสุดรวม (ใช้เมื่อ CustomerCodes = [])
	MaxUsageCountPerCustomer int             `json:"maxusagecountpercustomer" bson:"maxusagecountpercustomer"` // จำนวนครั้งสูงสุดต่อลูกค้า (ใช้เมื่อ CustomerCodes != [])
	IgnoreBranchCode         []string        `json:"ignore_branch_code" bson:"ignore_branch_code"`             // รายการสาขาที่ไม่ต้องการใช้คูปอง
	// เงื่อนไขสินค้าที่ใช้กับคูปอง
	ProductCondition *CouponProductCondition `json:"product_condition,omitempty" bson:"product_condition,omitempty"` // เงื่อนไขสินค้าที่ใช้กับคูปอง
}

type CouponInfo struct {
	models.DocIdentity `bson:"inline"`
	Coupon             `bson:"inline"`
}

func (CouponInfo) CollectionName() string {
	return couponCollectionName
}

type CouponData struct {
	models.HoldingCodeentity `bson:"inline"`
	CouponInfo               `bson:"inline"`
}

type CouponDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CouponData         `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (CouponDoc) CollectionName() string {
	return couponCollectionName
}

type CouponItemGuid struct {
	CouponCode string `json:"coupon_code" bson:"coupon_code" gorm:"coupon_code"`
}

func (CouponItemGuid) CollectionName() string {
	return couponCollectionName
}

type CouponActivity struct {
	CouponData          `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (CouponActivity) CollectionName() string {
	return couponCollectionName
}

type CouponDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (CouponDeleteActivity) CollectionName() string {
	return couponCollectionName
}

type CouponInfoResponse struct {
	Success bool       `json:"success"`
	Data    CouponInfo `json:"data,omitempty"`
}

type CouponPageResponse struct {
	Success    bool                          `json:"success"`
	Data       []CouponInfo                  `json:"data,omitempty"`
	Pagination models.PaginationDataResponse `json:"pagination,omitempty"`
}

// Coupon Reservation Models
const couponReservationCollectionName = "coupon_reservations"

type ReservationStatus int8

const (
	ReservationStatusActive   ReservationStatus = 0 // จองอยู่
	ReservationStatusUsed     ReservationStatus = 1 // ใช้แล้ว
	ReservationStatusExpired  ReservationStatus = 2 // หมดอายุ
	ReservationStatusCanceled ReservationStatus = 3 // ยกเลิก
)

type CouponReservation struct {
	models.Identity `bson:"inline"`
	CouponID        string            `json:"coupon_id" bson:"coupon_id"`
	CustomerID      string            `json:"customer_id" bson:"customer_id"`
	TransactionID   string            `json:"transaction_id" bson:"transaction_id"`
	ReservedAt      time.Time         `json:"reserved_at" bson:"reserved_at"`
	ExpiresAt       time.Time         `json:"expires_at" bson:"expires_at"`
	Status          ReservationStatus `json:"status" bson:"status"`
	UsedAt          *time.Time        `json:"used_at,omitempty" bson:"used_at,omitempty"`
	CanceledAt      *time.Time        `json:"canceled_at,omitempty" bson:"canceled_at,omitempty"`
}

func (CouponReservation) CollectionName() string {
	return couponReservationCollectionName
}

type CouponReservationData struct {
	models.HoldingCodeentity `bson:"inline"`
	CouponReservation        `bson:"inline"`
}

type CouponReservationDoc struct {
	ID                    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CouponReservationData `bson:"inline"`
	models.ActivityDoc    `bson:"inline"`
}

func (CouponReservationDoc) CollectionName() string {
	return couponReservationCollectionName
}

// Coupon Usage History Models
const couponUsageHistoryCollectionName = "coupon_usage_history"

type CouponUsageHistory struct {
	models.Identity   `bson:"inline"`
	CouponID          string     `json:"coupon_id" bson:"coupon_id"`
	CouponCode        string     `json:"coupon_code" bson:"coupon_code"`
	CustomerID        string     `json:"customer_id" bson:"customer_id"`
	CustomerCode      string     `json:"customer_code" bson:"customer_code"`
	CustomerName      string     `json:"customer_name" bson:"customer_name"`
	SaleInvoiceID     string     `json:"sale_invoice_id" bson:"sale_invoice_id"`
	SaleInvoiceNumber string     `json:"sale_invoice_number" bson:"sale_invoice_number"`
	TransactionID     string     `json:"transaction_id" bson:"transaction_id"`
	ReservationID     string     `json:"reservation_id" bson:"reservation_id"`
	UsedAmount        float64    `json:"used_amount" bson:"used_amount"`
	DiscountAmount    float64    `json:"discount_amount" bson:"discount_amount"`
	CashVoucherAmount float64    `json:"cash_voucher_amount" bson:"cash_voucher_amount"`
	OrderAmount       float64    `json:"order_amount" bson:"order_amount"`
	CouponType        CouponType `json:"coupon_type" bson:"coupon_type"`
	CouponValue       float64    `json:"coupon_value" bson:"coupon_value"`
	UsedAt            time.Time  `json:"used_at" bson:"used_at"`
	UsedBy            string     `json:"used_by" bson:"used_by"` // ผู้ทำรายการ
	Remark            string     `json:"remark" bson:"remark"`
}

func (CouponUsageHistory) CollectionName() string {
	return couponUsageHistoryCollectionName
}

type CouponUsageHistoryData struct {
	models.HoldingCodeentity `bson:"inline"`
	CouponUsageHistory       `bson:"inline"`
}

type CouponUsageHistoryDoc struct {
	ID                     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CouponUsageHistoryData `bson:"inline"`
	models.ActivityDoc     `bson:"inline"`
}

func (CouponUsageHistoryDoc) CollectionName() string {
	return couponUsageHistoryCollectionName
}

// Request/Response Models
type ReserveCouponRequest struct {
	CustomerID    string `json:"customer_id"`
	TransactionID string `json:"transaction_id" validate:"required"`
}

type CancelReserveCouponRequest struct {
	CustomerID    string `json:"customer_id"`
	ReservationID string `json:"reservation_id" validate:"required"`
}

type UseCouponRequest struct {
	CustomerID        string  `json:"customer_id"`
	CustomerCode      string  `json:"customer_code,omitempty"`
	CustomerName      string  `json:"customer_name,omitempty"`
	ReservationID     string  `json:"reservation_id" validate:"required"`
	TransactionID     string  `json:"transaction_id" validate:"required"`
	SaleInvoiceID     string  `json:"sale_invoice_id" validate:"required"`
	SaleInvoiceNumber string  `json:"sale_invoice_number" validate:"required"`
	UseAmount         float64 `json:"use_amount" validate:"required,min=0"`
	OrderAmount       float64 `json:"order_amount" validate:"required,min=0"`
	Remark            string  `json:"remark,omitempty"`
}

type CouponAvailabilityResponse struct {
	Available      bool   `json:"available"`
	UsageCount     int    `json:"usage_count"`     // จำนวนครั้งที่ใช้ไปแล้ว
	MaxUsageCount  int    `json:"max_usage_count"` // จำนวนครั้งสูงสุดที่ใช้ได้
	RemainingUsage int    `json:"remaining_usage"` // จำนวนครั้งที่ใช้ได้อีก
	Status         int8   `json:"status"`
	IsExpired      bool   `json:"is_expired"`
	Message        string `json:"message,omitempty"`
}

type CouponReservationResponse struct {
	ReservationID string    `json:"reservation_id"`
	CouponID      string    `json:"coupon_id"`
	CouponCode    string    `json:"coupon_code"`
	TransactionID string    `json:"transaction_id"`
	ExpiresAt     time.Time `json:"expires_at"`
	Reserved      bool      `json:"reserved"`
	Message       string    `json:"message,omitempty"`
}

// ข้อมูลสินค้าสำหรับการตรวจสอบคูปอง
type CouponCheckItem struct {
	Barcode   string  `json:"barcode" validate:"required"` // บาร์โค้ดสินค้า
	Qty       float64 `json:"qty" validate:"min=1"`        // จำนวน
	Price     float64 `json:"price" validate:"min=0"`      // ราคาต่อหน่วย
	SumAmount float64 `json:"sum_amount" validate:"min=0"` // ยอดรวม (Qty * Price)
}

// ข้อมูลการตรวจสอบคูปองแบบขั้นสูง
type CouponAvailabilityCheckRequest struct {
	CustomerID string            `json:"customer_id,omitempty"`                // รหัสลูกค้า (optional)
	BranchCode string            `json:"branch_code" validate:"required"`      // รหัสสาขา
	Items      []CouponCheckItem `json:"items" validate:"required,min=1,dive"` // รายการสินค้า
}

// ผลการตรวจสอบสินค้าแต่ละรายการ
type CouponItemCheckResult struct {
	Barcode         string  `json:"barcode"`
	Qty             float64 `json:"qty"`
	Price           float64 `json:"price"`
	SumAmount       float64 `json:"sum_amount"`
	IsEligible      bool    `json:"is_eligible"`                // สินค้านี้เข้าเงื่อนไขคูปองหรือไม่
	DiscountAmount  float64 `json:"discount_amount"`            // ส่วนลดของสินค้านี้
	MatchedCategory string  `json:"matched_category,omitempty"` // หมวดหมู่ที่ตรงเงื่อนไข
	Message         string  `json:"message,omitempty"`
}

// ผลการตรวจสอบคูปองแบบขั้นสูง
type CouponAvailabilityAdvancedResponse struct {
	Available         bool                    `json:"available"`
	UsageCount        int                     `json:"usage_count"`
	MaxUsageCount     int                     `json:"max_usage_count"`
	RemainingUsage    int                     `json:"remaining_usage"`
	Status            int8                    `json:"status"`
	IsExpired         bool                    `json:"is_expired"`
	BranchAllowed     bool                    `json:"branch_allowed"`      // สาขานี้สามารถใช้คูปองได้หรือไม่
	TotalAmount       float64                 `json:"total_amount"`        // ยอดรวมทั้งหมด
	EligibleAmount    float64                 `json:"eligible_amount"`     // ยอดรวมที่เข้าเงื่อนไข
	TotalDiscount     float64                 `json:"total_discount"`      // ส่วนลดรวม
	EligibleItemCount int                     `json:"eligible_item_count"` // จำนวนสินค้าที่เข้าเงื่อนไข
	ItemResults       []CouponItemCheckResult `json:"item_results"`        // ผลการตรวจสอบแต่ละสินค้า
	Message           string                  `json:"message,omitempty"`
}

type UseCouponResponse struct {
	Used           bool    `json:"used"`
	DiscountAmount float64 `json:"discount_amount"`
	TransactionID  string  `json:"transaction_id"`
	UsageCount     int     `json:"usage_count"`     // จำนวนครั้งที่ใช้ไปแล้วหลังการใช้งานนี้
	RemainingUsage int     `json:"remaining_usage"` // จำนวนครั้งที่ใช้ได้อีก
	Message        string  `json:"message,omitempty"`
}

// Calculate Coupon Request และ Response Models
type CalculateCouponRequest struct {
	OrderAmount float64                      `json:"order_amount" validate:"required,gt=0"` // ยอดรวมการสั่งซื้อ
	Coupons     []CalculateCouponItemRequest `json:"coupons" validate:"required,min=1"`     // รายการคูปองที่ต้องการใช้
	CustomerID  string                       `json:"customer_id,omitempty"`                 // รหัสลูกค้า (ถ้ามี)
	BranchCode  string                       `json:"branch_code" validate:"required"`       // รหัสสาขา (สำหรับตรวจสอบ IgnoreBranchCode)
	Items       []CouponCheckItem            `json:"items" validate:"required,min=1,dive"`  // รายการสินค้า (สำหรับตรวจสอบเงื่อนไขสินค้า)
}

type CalculateCouponItemRequest struct {
	CouponCode string  `json:"coupon_code" validate:"required"`       // รหัสคูปอง
	UseAmount  float64 `json:"use_amount,omitempty" validate:"gte=0"` // จำนวนที่ต้องการใช้ (สำหรับคูปองแบบบางส่วน)
}

type CalculateCouponResponse struct {
	Success          bool                          `json:"success"`
	TotalDiscount    float64                       `json:"total_discount"`     // ส่วนลดรวมทั้งหมด
	TotalCashVoucher float64                       `json:"total_cash_voucher"` // มูลค่าแทนเงินสดรวม
	FinalAmount      float64                       `json:"final_amount"`       // ยอดสุทธิหลังหักส่วนลด
	CouponResults    []CalculateCouponItemResponse `json:"coupon_results"`     // ผลการคำนวนแต่ละคูปอง
	Errors           []CouponCalculationError      `json:"errors,omitempty"`   // ข้อผิดพลาด (ถ้ามี)
}

type CalculateCouponItemResponse struct {
	CouponCode        string                  `json:"coupon_code"`            // รหัสคูปอง
	CouponType        CouponType              `json:"coupon_type"`            // ประเภทคูปอง
	CouponTypeName    string                  `json:"coupon_type_name"`       // ชื่อประเภทคูปอง
	DiscountAmount    float64                 `json:"discount_amount"`        // จำนวนส่วนลด
	CashVoucherAmount float64                 `json:"cash_voucher_amount"`    // จำนวนแทนเงินสด
	UsedAmount        float64                 `json:"used_amount"`            // จำนวนที่ใช้จริง
	UsageCount        int                     `json:"usage_count"`            // จำนวนครั้งที่ใช้ไปแล้ว
	RemainingUsage    int                     `json:"remaining_usage"`        // จำนวนครั้งที่ใช้ได้อีก
	Applied           bool                    `json:"applied"`                // ใช้งานได้หรือไม่
	BranchAllowed     bool                    `json:"branch_allowed"`         // สาขานี้สามารถใช้คูปองได้หรือไม่
	EligibleAmount    float64                 `json:"eligible_amount"`        // ยอดรวมของสินค้าที่เข้าเงื่อนไข
	EligibleItemCount int                     `json:"eligible_item_count"`    // จำนวนสินค้าที่เข้าเงื่อนไข
	MinimumAmount     float64                 `json:"minimum_amount"`         // ยอดขั้นต่ำที่ต้องมีสำหรับคูปองนี้
	ItemResults       []CouponItemCheckResult `json:"item_results,omitempty"` // ผลการตรวจสอบแต่ละสินค้า
	Message           string                  `json:"message"`                // ข้อความเพิ่มเติม
}

type CouponCalculationError struct {
	CouponCode string `json:"coupon_code"`
	Error      string `json:"error"`
	Code       string `json:"code"`
}

// Reservation Lookup Models
type ReservationLookupRequest struct {
	TransactionID    string `json:"transaction_id" validate:"required"`
	IncludeExpired   bool   `json:"include_expired,omitempty"`
	IncludeCancelled bool   `json:"include_cancelled,omitempty"`
	IncludeUsed      bool   `json:"include_used,omitempty"`
}

type ReservationStatusResponse struct {
	TransactionID     string            `json:"transaction_id"`
	Reservations      []ReservationInfo `json:"reservations"`
	ReservationStatus string            `json:"reservation_status"` // active, expired, used, cancelled
	Count             int               `json:"count"`
}

type ReservationInfo struct {
	ID         string            `json:"id"`
	CouponID   string            `json:"coupon_id"`
	CouponCode string            `json:"coupon_code"`
	CustomerID string            `json:"customer_id"`
	Status     ReservationStatus `json:"status"`
	StatusName string            `json:"status_name"`
	ReservedAt time.Time         `json:"reserved_at"`
	ExpiresAt  time.Time         `json:"expires_at"`
	UsedAt     *time.Time        `json:"used_at,omitempty"`
	CanceledAt *time.Time        `json:"canceled_at,omitempty"`
	IsExpired  bool              `json:"is_expired"`
}

type ReservationLookupResponse struct {
	TransactionID string            `json:"transaction_id"`
	Found         bool              `json:"found"`
	Reservations  []ReservationInfo `json:"reservations"`
	Summary       struct {
		TotalReservations int `json:"total_reservations"`
		ActiveCount       int `json:"active_count"`
		UsedCount         int `json:"used_count"`
		ExpiredCount      int `json:"expired_count"`
		CancelledCount    int `json:"cancelled_count"`
	} `json:"summary"`
}

// Helper functions
func (c *Coupon) IsExpired() bool {
	// Use UTC time for consistent comparison with ExpiryDate stored in UTC
	return time.Now().UTC().After(c.ExpiryDate.UTC())
}

func (c *Coupon) IsActive() bool {
	return c.Status == CouponStatusActive && !c.IsExpired()
}

func (r *CouponReservation) IsExpired() bool {
	// Use UTC time for consistent comparison with ExpiresAt stored in UTC
	return time.Now().UTC().After(r.ExpiresAt.UTC())
}

func (r *CouponReservation) IsActive() bool {
	return r.Status == ReservationStatusActive && !r.IsExpired()
}

// สร้างการจองใหม่ (15 นาที)
func NewCouponReservation(couponID, customerID, transactionID string) *CouponReservation {
	now := time.Now().UTC() // Use UTC time consistently
	return &CouponReservation{
		CouponID:      couponID,
		CustomerID:    customerID,
		TransactionID: transactionID,
		ReservedAt:    now,
		ExpiresAt:     now.Add(15 * time.Minute), // หมดอายุใน 15 นาที
		Status:        ReservationStatusActive,
	}
}

// ตรวจสอบว่าคูปองสามารถใช้งานได้หรือไม่
func (c *Coupon) CanUse() (bool, string) {
	if !c.IsActive() {
		if c.IsExpired() {
			return false, "คูปองหมดอายุแล้ว"
		}
		if c.Status == CouponStatusCanceled {
			return false, "คูปองถูกยกเลิกแล้ว"
		}
		return false, "คูปองไม่สามารถใช้งานได้"
	}

	// ไม่เช็ค RemainingValue อีกต่อไป เช็คเฉพาะจำนวนครั้งการใช้งาน
	// การเช็คจำนวนครั้งจะทำใน service layer

	return true, ""
}

// ตรวจสอบว่าลูกค้าสามารถใช้คูปองนี้ได้หรือไม่
func (c *Coupon) IsCustomerEligible(customerID string) bool {
	// ถ้าไม่กำหนดลูกค้า (CustomerCodes ว่าง) = ทุกคนใช้ได้
	if len(c.CustomerCodes) == 0 {
		return true
	}

	// ตรวจสอบว่าลูกค้าอยู่ในรายการหรือไม่
	for _, code := range c.CustomerCodes {
		if code == customerID {
			return true
		}
	}
	return false
}

// ตรวจสอบว่าใช้ MaxUsageCount หรือ MaxUsageCountPerCustomer
func (c *Coupon) IsGlobalUsageMode() bool {
	return len(c.CustomerCodes) == 0
}

// ได้จำนวนครั้งสูงสุดที่ใช้ได้ตามโหมดการใช้งาน
func (c *Coupon) GetMaxUsageLimit() int {
	if c.IsOneTimeUse {
		return 1
	}

	if c.IsGlobalUsageMode() {
		// โหมดรวม: ใช้ MaxUsageCount
		return c.MaxUsageCount
	} else {
		// โหมดต่อลูกค้า: ใช้ MaxUsageCountPerCustomer
		return c.MaxUsageCountPerCustomer
	}
}

// Helper method สำหรับแปลง CouponType เป็น string
func (ct CouponType) String() string {
	switch ct {
	case CouponTypeValueDiscount:
		return "ลดตามมูลค่า"
	case CouponTypePercentDiscount:
		return "ลดตามเปอร์เซ็นต์"
	case CouponTypeCashVoucher:
		return "คูปองแทนเงินสด"
	default:
		return "ไม่ทราบประเภท"
	}
}

// Helper method สำหรับตรวจสอบว่าเป็นคูปองแทนเงินสด
func (ct CouponType) IsCashVoucher() bool {
	return ct == CouponTypeCashVoucher
}

// Helper method สำหรับตรวจสอบความถูกต้องของ CouponType
func (ct CouponType) IsValid() bool {
	return ct >= CouponTypeValueDiscount && ct <= CouponTypeCashVoucher
}

// Helper method สำหรับตรวจสอบว่าคูปองมีเงื่อนไขสินค้าหรือไม่
func (c *Coupon) HasProductCondition() bool {
	if c.ProductCondition == nil {
		return false
	}

	// ตรวจสอบว่ามีการกำหนดเงื่อนไขสินค้าใด ๆ หรือไม่
	return len(c.ProductCondition.ProductCodes) > 0 ||
		len(c.ProductCondition.GroupCodes) > 0 ||
		len(c.ProductCondition.GroupSubOneCodes) > 0 ||
		len(c.ProductCondition.GroupSubTwoCodes) > 0 ||
		len(c.ProductCondition.BrandCodes) > 0 ||
		len(c.ProductCondition.DesignCodes) > 0 ||
		len(c.ProductCondition.ModelCodes) > 0 ||
		len(c.ProductCondition.PatternCodes) > 0 ||
		len(c.ProductCondition.GradeCodes) > 0 ||
		len(c.ProductCondition.CategoryCodes) > 0 ||
		len(c.ProductCondition.ClassCodes) > 0 ||
		c.ProductCondition.MinimumAmount > 0
}

// Coupon Usage History API Models
type CouponUsageHistoryRequest struct {
	CouponID          string     `json:"coupon_id,omitempty"`
	CouponCode        string     `json:"coupon_code,omitempty"`
	CustomerID        string     `json:"customer_id,omitempty"`
	SaleInvoiceID     string     `json:"sale_invoice_id,omitempty"`
	SaleInvoiceNumber string     `json:"sale_invoice_number,omitempty"`
	TransactionID     string     `json:"transaction_id,omitempty"`
	StartDate         *time.Time `json:"start_date,omitempty"`
	EndDate           *time.Time `json:"end_date,omitempty"`
	Page              int        `json:"page,omitempty"`
	PageSize          int        `json:"page_size,omitempty"`
}

type CouponUsageHistoryResponse struct {
	UsageHistory []CouponUsageHistoryItem `json:"usage_history"`
	Pagination   struct {
		Page      int `json:"page"`
		PageSize  int `json:"page_size"`
		Total     int `json:"total"`
		TotalPage int `json:"total_page"`
	} `json:"pagination"`
	Summary struct {
		TotalUsed        int     `json:"total_used"`
		TotalDiscount    float64 `json:"total_discount"`
		TotalCashVoucher float64 `json:"total_cash_voucher"`
		TotalOrderAmount float64 `json:"total_order_amount"`
	} `json:"summary"`
}

type CouponUsageHistoryItem struct {
	ID                string     `json:"id"`
	CouponID          string     `json:"coupon_id"`
	CouponCode        string     `json:"coupon_code"`
	CouponType        CouponType `json:"coupon_type"`
	CouponTypeName    string     `json:"coupon_type_name"`
	CouponValue       float64    `json:"coupon_value"`
	CustomerID        string     `json:"customer_id"`
	CustomerCode      string     `json:"customer_code"`
	CustomerName      string     `json:"customer_name"`
	SaleInvoiceID     string     `json:"sale_invoice_id"`
	SaleInvoiceNumber string     `json:"sale_invoice_number"`
	TransactionID     string     `json:"transaction_id"`
	ReservationID     string     `json:"reservation_id"`
	UsedAmount        float64    `json:"used_amount"`
	DiscountAmount    float64    `json:"discount_amount"`
	CashVoucherAmount float64    `json:"cash_voucher_amount"`
	OrderAmount       float64    `json:"order_amount"`
	UsedAt            time.Time  `json:"used_at"`
	UsedBy            string     `json:"used_by"`
	Remark            string     `json:"remark"`
}

type CreateUsageHistoryRequest struct {
	CouponID          string     `json:"coupon_id" validate:"required"`
	CouponCode        string     `json:"coupon_code" validate:"required"`
	CustomerID        string     `json:"customer_id" validate:"required"`
	CustomerCode      string     `json:"customer_code,omitempty"`
	CustomerName      string     `json:"customer_name,omitempty"`
	SaleInvoiceID     string     `json:"sale_invoice_id" validate:"required"`
	SaleInvoiceNumber string     `json:"sale_invoice_number" validate:"required"`
	TransactionID     string     `json:"transaction_id" validate:"required"`
	ReservationID     string     `json:"reservation_id,omitempty"`
	UsedAmount        float64    `json:"used_amount" validate:"required,min=0"`
	DiscountAmount    float64    `json:"discount_amount" validate:"min=0"`
	CashVoucherAmount float64    `json:"cash_voucher_amount" validate:"min=0"`
	OrderAmount       float64    `json:"order_amount" validate:"required,min=0"`
	CouponType        CouponType `json:"coupon_type" validate:"required"`
	CouponValue       float64    `json:"coupon_value" validate:"required,min=0"`
	Remark            string     `json:"remark,omitempty"`
}

// Excel Import Related Models
type CouponExcelImportRequest struct {
	FileName          string `json:"file_name"`
	FileData          []byte `json:"file_data"`
	SheetName         string `json:"sheet_name,omitempty"`
	StartRow          int    `json:"start_row,omitempty"`
	ValidateOnly      bool   `json:"validate_only,omitempty"`
	OverwriteExisting bool   `json:"overwrite_existing,omitempty"`
}

type CouponExcelImportResponse struct {
	Success       bool                `json:"success"`
	BatchID       string              `json:"batch_id,omitempty"`
	ImportedCount int                 `json:"imported_count"`
	ErrorCount    int                 `json:"error_count"`
	TotalRows     int                 `json:"total_rows"`
	Message       string              `json:"message"`
	Errors        []CouponImportError `json:"errors,omitempty"`
}

type CouponImportError struct {
	Row        int    `json:"row"`
	CouponCode string `json:"coupon_code"`
	Field      string `json:"field"`
	Error      string `json:"error"`
}

type CouponBatchImportStatus struct {
	BatchID       string    `json:"batch_id"`
	Status        string    `json:"status"`   // pending, processing, completed, failed
	Progress      int       `json:"progress"` // 0-100
	TotalRows     int       `json:"total_rows"`
	ProcessedRows int       `json:"processed_rows"`
	SuccessCount  int       `json:"success_count"`
	ErrorCount    int       `json:"error_count"`
	Message       string    `json:"message"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CouponBatchImportResponse struct {
	BatchID       string    `json:"batch_id"`
	Status        string    `json:"status"`   // pending, processing, completed, failed
	Progress      int       `json:"progress"` // 0-100
	TotalRows     int       `json:"total_rows"`
	ProcessedRows int       `json:"processed_rows"`
	SuccessCount  int       `json:"success_count"`
	ErrorCount    int       `json:"error_count"`
	Message       string    `json:"message"`
	UpdatedAt     time.Time `json:"updated_at"`
}
