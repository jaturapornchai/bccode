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
	ProductCodes     []string `json:"productcodes" bson:"productcodes"`         // รายการ Code ของสินค้า
	GroupCodes       []string `json:"groupcodes" bson:"groupcodes"`             // รายการ GroupCode
	GroupSubOneCodes []string `json:"groupsubonecodes" bson:"groupsubonecodes"` // รายการ GroupsuboneCode
	GroupSubTwoCodes []string `json:"groupsubtwocodes" bson:"groupsubtwocodes"` // รายการ GroupsubtwoCode
	BrandCodes       []string `json:"brandcodes" bson:"brandcodes"`             // รายการ BrandCode
	DesignCodes      []string `json:"designcodes" bson:"designcodes"`           // รายการ DesignCode
	ModelCodes       []string `json:"modelcodes" bson:"modelcodes"`             // รายการ ModelCode
	PatternCodes     []string `json:"patterncodes" bson:"patterncodes"`         // รายการ PatternCode
	GradeCodes       []string `json:"gradecodes" bson:"gradecodes"`             // รายการ GradeCode
	CategoryCodes    []string `json:"categorycodes" bson:"categorycodes"`       // รายการ CategoryCode
	ClassCodes       []string `json:"classcodes" bson:"classcodes"`             // รายการ ClassCode
	MinimumAmount    float64  `json:"minimumamount" bson:"minimumamount"`       // มูลค่าขั้นต่ำของสินค้าที่เข้าเงื่อนไข (0 = ไม่จำกัด)
}

type Coupon struct {
	models.PartitionIdentity `bson:"inline"`
	CouponCode               string          `json:"couponcode" bson:"couponcode"` // หมายเลขคูปอง
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	CouponValue              float64         `json:"couponvalue" bson:"couponvalue"`                           // มูลค่าคูปอง
	IssuedDate               time.Time       `json:"issueddate" bson:"issueddate"`                             // วันที่ออกคูปอง
	ExpiryDate               time.Time       `json:"expirydate" bson:"expirydate"`                             // วันที่หมดอายุคูปอง
	CouponType               CouponType      `json:"coupontype" bson:"coupontype"`                             // ประเภทคูปอง (ลดตามมูลค่า/ลดตามเปอร์เซ็นต์)
	CustomerCodes            []string        `json:"customercodes" bson:"customercodes"`                       // รหัสลูกค้าที่สามารถใช้ได้ ([] = ทุกคน)
	Remark                   string          `json:"remark" bson:"remark"`                                     // หมายเหตุ
	Status                   CouponStatus    `json:"status" bson:"status"`                                     // สถานะ ยกเลิก (ปกติ/ยกเลิก)
	IsOneTimeUse             bool            `json:"isonetimeuse" bson:"isonetimeuse"`                         // ใช้ครั้งเดียว
	MaxUsageCount            int             `json:"maxusagecount" bson:"maxusagecount"`                       // จำนวนครั้งสูงสุดรวม (ใช้เมื่อ CustomerCodes = [])
	MaxUsageCountPerCustomer int             `json:"maxusagecountpercustomer" bson:"maxusagecountpercustomer"` // จำนวนครั้งสูงสุดต่อลูกค้า (ใช้เมื่อ CustomerCodes != [])
	IgnoreBranchCode         []string        `json:"ignorebranchcode" bson:"ignorebranchcode"`                 // รายการสาขาที่ไม่ต้องการใช้คูปอง
	// เงื่อนไขสินค้าที่ใช้กับคูปอง
	ProductCondition *CouponProductCondition `json:"productcondition,omitempty" bson:"productcondition,omitempty"` // เงื่อนไขสินค้าที่ใช้กับคูปอง
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
	ID                 primitive.ObjectID `json:"id" bson:"id,omitempty"`
	CouponData         `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (CouponDoc) CollectionName() string {
	return couponCollectionName
}

type CouponItemGuid struct {
	CouponCode string `json:"couponcode" bson:"couponcode" gorm:"coupon_code"`
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
	CouponID        string            `json:"couponid" bson:"couponid"`
	CustomerID      string            `json:"customerid" bson:"customerid"`
	TransactionID   string            `json:"transactionid" bson:"transactionid"`
	ReservedAt      time.Time         `json:"reservedat" bson:"reservedat"`
	ExpiresAt       time.Time         `json:"expiresat" bson:"expiresat"`
	Status          ReservationStatus `json:"status" bson:"status"`
	UsedAt          *time.Time        `json:"usedat,omitempty" bson:"usedat,omitempty"`
	CanceledAt      *time.Time        `json:"canceledat,omitempty" bson:"canceledat,omitempty"`
}

func (CouponReservation) CollectionName() string {
	return couponReservationCollectionName
}

type CouponReservationData struct {
	models.HoldingCodeentity `bson:"inline"`
	CouponReservation        `bson:"inline"`
}

type CouponReservationDoc struct {
	ID                    primitive.ObjectID `json:"id" bson:"id,omitempty"`
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
	CouponID          string     `json:"couponid" bson:"couponid"`
	CouponCode        string     `json:"couponcode" bson:"couponcode"`
	CustomerID        string     `json:"customerid" bson:"customerid"`
	CustomerCode      string     `json:"customercode" bson:"customercode"`
	CustomerName      string     `json:"customername" bson:"customername"`
	SaleInvoiceID     string     `json:"saleinvoiceid" bson:"saleinvoiceid"`
	SaleInvoiceNumber string     `json:"saleinvoicenumber" bson:"saleinvoicenumber"`
	TransactionID     string     `json:"transactionid" bson:"transactionid"`
	ReservationID     string     `json:"reservationid" bson:"reservationid"`
	UsedAmount        float64    `json:"usedamount" bson:"usedamount"`
	DiscountAmount    float64    `json:"discountamount" bson:"discountamount"`
	CashVoucherAmount float64    `json:"cashvoucheramount" bson:"cashvoucheramount"`
	OrderAmount       float64    `json:"orderamount" bson:"orderamount"`
	CouponType        CouponType `json:"coupontype" bson:"coupontype"`
	CouponValue       float64    `json:"couponvalue" bson:"couponvalue"`
	UsedAt            time.Time  `json:"usedat" bson:"usedat"`
	UsedBy            string     `json:"usedby" bson:"usedby"` // ผู้ทำรายการ
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
	ID                     primitive.ObjectID `json:"id" bson:"id,omitempty"`
	CouponUsageHistoryData `bson:"inline"`
	models.ActivityDoc     `bson:"inline"`
}

func (CouponUsageHistoryDoc) CollectionName() string {
	return couponUsageHistoryCollectionName
}

// Request/Response Models
type ReserveCouponRequest struct {
	CustomerID    string `json:"customerid"`
	TransactionID string `json:"transactionid" validate:"required"`
}

type CancelReserveCouponRequest struct {
	CustomerID    string `json:"customerid"`
	ReservationID string `json:"reservationid" validate:"required"`
}

type UseCouponRequest struct {
	CustomerID        string  `json:"customerid"`
	CustomerCode      string  `json:"customercode,omitempty"`
	CustomerName      string  `json:"customername,omitempty"`
	ReservationID     string  `json:"reservationid" validate:"required"`
	TransactionID     string  `json:"transactionid" validate:"required"`
	SaleInvoiceID     string  `json:"saleinvoiceid" validate:"required"`
	SaleInvoiceNumber string  `json:"saleinvoicenumber" validate:"required"`
	UseAmount         float64 `json:"useamount" validate:"required,min=0"`
	OrderAmount       float64 `json:"orderamount" validate:"required,min=0"`
	Remark            string  `json:"remark,omitempty"`
}

type CouponAvailabilityResponse struct {
	Available      bool   `json:"available"`
	UsageCount     int    `json:"usagecount"`     // จำนวนครั้งที่ใช้ไปแล้ว
	MaxUsageCount  int    `json:"maxusagecount"`  // จำนวนครั้งสูงสุดที่ใช้ได้
	RemainingUsage int    `json:"remainingusage"` // จำนวนครั้งที่ใช้ได้อีก
	Status         int8   `json:"status"`
	IsExpired      bool   `json:"isexpired"`
	Message        string `json:"message,omitempty"`
}

type CouponReservationResponse struct {
	ReservationID string    `json:"reservationid"`
	CouponID      string    `json:"couponid"`
	CouponCode    string    `json:"couponcode"`
	TransactionID string    `json:"transactionid"`
	ExpiresAt     time.Time `json:"expiresat"`
	Reserved      bool      `json:"reserved"`
	Message       string    `json:"message,omitempty"`
}

// ข้อมูลสินค้าสำหรับการตรวจสอบคูปอง
type CouponCheckItem struct {
	Barcode   string  `json:"barcode" validate:"required"` // บาร์โค้ดสินค้า
	Qty       float64 `json:"qty" validate:"min=1"`        // จำนวน
	Price     float64 `json:"price" validate:"min=0"`      // ราคาต่อหน่วย
	SumAmount float64 `json:"sumamount" validate:"min=0"`  // ยอดรวม (Qty * Price)
}

// ข้อมูลการตรวจสอบคูปองแบบขั้นสูง
type CouponAvailabilityCheckRequest struct {
	CustomerID string            `json:"customerid,omitempty"`                 // รหัสลูกค้า (optional)
	BranchCode string            `json:"branchcode" validate:"required"`       // รหัสสาขา
	Items      []CouponCheckItem `json:"items" validate:"required,min=1,dive"` // รายการสินค้า
}

// ผลการตรวจสอบสินค้าแต่ละรายการ
type CouponItemCheckResult struct {
	Barcode         string  `json:"barcode"`
	Qty             float64 `json:"qty"`
	Price           float64 `json:"price"`
	SumAmount       float64 `json:"sumamount"`
	IsEligible      bool    `json:"iseligible"`                // สินค้านี้เข้าเงื่อนไขคูปองหรือไม่
	DiscountAmount  float64 `json:"discountamount"`            // ส่วนลดของสินค้านี้
	MatchedCategory string  `json:"matchedcategory,omitempty"` // หมวดหมู่ที่ตรงเงื่อนไข
	Message         string  `json:"message,omitempty"`
}

// ผลการตรวจสอบคูปองแบบขั้นสูง
type CouponAvailabilityAdvancedResponse struct {
	Available         bool                    `json:"available"`
	UsageCount        int                     `json:"usagecount"`
	MaxUsageCount     int                     `json:"maxusagecount"`
	RemainingUsage    int                     `json:"remainingusage"`
	Status            int8                    `json:"status"`
	IsExpired         bool                    `json:"isexpired"`
	BranchAllowed     bool                    `json:"branchallowed"`     // สาขานี้สามารถใช้คูปองได้หรือไม่
	TotalAmount       float64                 `json:"totalamount"`       // ยอดรวมทั้งหมด
	EligibleAmount    float64                 `json:"eligibleamount"`    // ยอดรวมที่เข้าเงื่อนไข
	TotalDiscount     float64                 `json:"totaldiscount"`     // ส่วนลดรวม
	EligibleItemCount int                     `json:"eligibleitemcount"` // จำนวนสินค้าที่เข้าเงื่อนไข
	ItemResults       []CouponItemCheckResult `json:"itemresults"`       // ผลการตรวจสอบแต่ละสินค้า
	Message           string                  `json:"message,omitempty"`
}

type UseCouponResponse struct {
	Used           bool    `json:"used"`
	DiscountAmount float64 `json:"discountamount"`
	TransactionID  string  `json:"transactionid"`
	UsageCount     int     `json:"usagecount"`     // จำนวนครั้งที่ใช้ไปแล้วหลังการใช้งานนี้
	RemainingUsage int     `json:"remainingusage"` // จำนวนครั้งที่ใช้ได้อีก
	Message        string  `json:"message,omitempty"`
}

// Calculate Coupon Request และ Response Models
type CalculateCouponRequest struct {
	OrderAmount float64                      `json:"orderamount" validate:"required,gt=0"` // ยอดรวมการสั่งซื้อ
	Coupons     []CalculateCouponItemRequest `json:"coupons" validate:"required,min=1"`    // รายการคูปองที่ต้องการใช้
	CustomerID  string                       `json:"customerid,omitempty"`                 // รหัสลูกค้า (ถ้ามี)
	BranchCode  string                       `json:"branchcode" validate:"required"`       // รหัสสาขา (สำหรับตรวจสอบ IgnoreBranchCode)
	Items       []CouponCheckItem            `json:"items" validate:"required,min=1,dive"` // รายการสินค้า (สำหรับตรวจสอบเงื่อนไขสินค้า)
}

type CalculateCouponItemRequest struct {
	CouponCode string  `json:"couponcode" validate:"required"`       // รหัสคูปอง
	UseAmount  float64 `json:"useamount,omitempty" validate:"gte=0"` // จำนวนที่ต้องการใช้ (สำหรับคูปองแบบบางส่วน)
}

type CalculateCouponResponse struct {
	Success          bool                          `json:"success"`
	TotalDiscount    float64                       `json:"totaldiscount"`    // ส่วนลดรวมทั้งหมด
	TotalCashVoucher float64                       `json:"totalcashvoucher"` // มูลค่าแทนเงินสดรวม
	FinalAmount      float64                       `json:"finalamount"`      // ยอดสุทธิหลังหักส่วนลด
	CouponResults    []CalculateCouponItemResponse `json:"couponresults"`    // ผลการคำนวนแต่ละคูปอง
	Errors           []CouponCalculationError      `json:"errors,omitempty"` // ข้อผิดพลาด (ถ้ามี)
}

type CalculateCouponItemResponse struct {
	CouponCode        string                  `json:"couponcode"`            // รหัสคูปอง
	CouponType        CouponType              `json:"coupontype"`            // ประเภทคูปอง
	CouponTypeName    string                  `json:"coupontypename"`        // ชื่อประเภทคูปอง
	DiscountAmount    float64                 `json:"discountamount"`        // จำนวนส่วนลด
	CashVoucherAmount float64                 `json:"cashvoucheramount"`     // จำนวนแทนเงินสด
	UsedAmount        float64                 `json:"usedamount"`            // จำนวนที่ใช้จริง
	UsageCount        int                     `json:"usagecount"`            // จำนวนครั้งที่ใช้ไปแล้ว
	RemainingUsage    int                     `json:"remainingusage"`        // จำนวนครั้งที่ใช้ได้อีก
	Applied           bool                    `json:"applied"`               // ใช้งานได้หรือไม่
	BranchAllowed     bool                    `json:"branchallowed"`         // สาขานี้สามารถใช้คูปองได้หรือไม่
	EligibleAmount    float64                 `json:"eligibleamount"`        // ยอดรวมของสินค้าที่เข้าเงื่อนไข
	EligibleItemCount int                     `json:"eligibleitemcount"`     // จำนวนสินค้าที่เข้าเงื่อนไข
	MinimumAmount     float64                 `json:"minimumamount"`         // ยอดขั้นต่ำที่ต้องมีสำหรับคูปองนี้
	ItemResults       []CouponItemCheckResult `json:"itemresults,omitempty"` // ผลการตรวจสอบแต่ละสินค้า
	Message           string                  `json:"message"`               // ข้อความเพิ่มเติม
}

type CouponCalculationError struct {
	CouponCode string `json:"couponcode"`
	Error      string `json:"error"`
	Code       string `json:"code"`
}

// Reservation Lookup Models
type ReservationLookupRequest struct {
	TransactionID    string `json:"transactionid" validate:"required"`
	IncludeExpired   bool   `json:"includeexpired,omitempty"`
	IncludeCancelled bool   `json:"includecancelled,omitempty"`
	IncludeUsed      bool   `json:"includeused,omitempty"`
}

type ReservationStatusResponse struct {
	TransactionID     string            `json:"transactionid"`
	Reservations      []ReservationInfo `json:"reservations"`
	ReservationStatus string            `json:"reservationstatus"` // active, expired, used, cancelled
	Count             int               `json:"count"`
}

type ReservationInfo struct {
	ID         string            `json:"id"`
	CouponID   string            `json:"couponid"`
	CouponCode string            `json:"couponcode"`
	CustomerID string            `json:"customerid"`
	Status     ReservationStatus `json:"status"`
	StatusName string            `json:"statusname"`
	ReservedAt time.Time         `json:"reservedat"`
	ExpiresAt  time.Time         `json:"expiresat"`
	UsedAt     *time.Time        `json:"usedat,omitempty"`
	CanceledAt *time.Time        `json:"canceledat,omitempty"`
	IsExpired  bool              `json:"isexpired"`
}

type ReservationLookupResponse struct {
	TransactionID string            `json:"transactionid"`
	Found         bool              `json:"found"`
	Reservations  []ReservationInfo `json:"reservations"`
	Summary       struct {
		TotalReservations int `json:"totalreservations"`
		ActiveCount       int `json:"activecount"`
		UsedCount         int `json:"usedcount"`
		ExpiredCount      int `json:"expiredcount"`
		CancelledCount    int `json:"cancelledcount"`
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
	CouponID          string     `json:"couponid,omitempty"`
	CouponCode        string     `json:"couponcode,omitempty"`
	CustomerID        string     `json:"customerid,omitempty"`
	SaleInvoiceID     string     `json:"saleinvoiceid,omitempty"`
	SaleInvoiceNumber string     `json:"saleinvoicenumber,omitempty"`
	TransactionID     string     `json:"transactionid,omitempty"`
	StartDate         *time.Time `json:"startdate,omitempty"`
	EndDate           *time.Time `json:"enddate,omitempty"`
	Page              int        `json:"page,omitempty"`
	PageSize          int        `json:"pagesize,omitempty"`
}

type CouponUsageHistoryResponse struct {
	UsageHistory []CouponUsageHistoryItem `json:"usagehistory"`
	Pagination   struct {
		Page      int `json:"page"`
		PageSize  int `json:"pagesize"`
		Total     int `json:"total"`
		TotalPage int `json:"totalpage"`
	} `json:"pagination"`
	Summary struct {
		TotalUsed        int     `json:"totalused"`
		TotalDiscount    float64 `json:"totaldiscount"`
		TotalCashVoucher float64 `json:"totalcashvoucher"`
		TotalOrderAmount float64 `json:"totalorderamount"`
	} `json:"summary"`
}

type CouponUsageHistoryItem struct {
	ID                string     `json:"id"`
	CouponID          string     `json:"couponid"`
	CouponCode        string     `json:"couponcode"`
	CouponType        CouponType `json:"coupontype"`
	CouponTypeName    string     `json:"coupontypename"`
	CouponValue       float64    `json:"couponvalue"`
	CustomerID        string     `json:"customerid"`
	CustomerCode      string     `json:"customercode"`
	CustomerName      string     `json:"customername"`
	SaleInvoiceID     string     `json:"saleinvoiceid"`
	SaleInvoiceNumber string     `json:"saleinvoicenumber"`
	TransactionID     string     `json:"transactionid"`
	ReservationID     string     `json:"reservationid"`
	UsedAmount        float64    `json:"usedamount"`
	DiscountAmount    float64    `json:"discountamount"`
	CashVoucherAmount float64    `json:"cashvoucheramount"`
	OrderAmount       float64    `json:"orderamount"`
	UsedAt            time.Time  `json:"usedat"`
	UsedBy            string     `json:"usedby"`
	Remark            string     `json:"remark"`
}

type CreateUsageHistoryRequest struct {
	CouponID          string     `json:"couponid" validate:"required"`
	CouponCode        string     `json:"couponcode" validate:"required"`
	CustomerID        string     `json:"customerid" validate:"required"`
	CustomerCode      string     `json:"customercode,omitempty"`
	CustomerName      string     `json:"customername,omitempty"`
	SaleInvoiceID     string     `json:"saleinvoiceid" validate:"required"`
	SaleInvoiceNumber string     `json:"saleinvoicenumber" validate:"required"`
	TransactionID     string     `json:"transactionid" validate:"required"`
	ReservationID     string     `json:"reservationid,omitempty"`
	UsedAmount        float64    `json:"usedamount" validate:"required,min=0"`
	DiscountAmount    float64    `json:"discountamount" validate:"min=0"`
	CashVoucherAmount float64    `json:"cashvoucheramount" validate:"min=0"`
	OrderAmount       float64    `json:"orderamount" validate:"required,min=0"`
	CouponType        CouponType `json:"coupontype" validate:"required"`
	CouponValue       float64    `json:"couponvalue" validate:"required,min=0"`
	Remark            string     `json:"remark,omitempty"`
}

// Excel Import Related Models
type CouponExcelImportRequest struct {
	FileName          string `json:"filename"`
	FileData          []byte `json:"filedata"`
	SheetName         string `json:"sheetname,omitempty"`
	StartRow          int    `json:"startrow,omitempty"`
	ValidateOnly      bool   `json:"validateonly,omitempty"`
	OverwriteExisting bool   `json:"overwriteexisting,omitempty"`
}

type CouponExcelImportResponse struct {
	Success       bool                `json:"success"`
	BatchID       string              `json:"batchid,omitempty"`
	ImportedCount int                 `json:"importedcount"`
	ErrorCount    int                 `json:"errorcount"`
	TotalRows     int                 `json:"totalrows"`
	Message       string              `json:"message"`
	Errors        []CouponImportError `json:"errors,omitempty"`
}

type CouponImportError struct {
	Row        int    `json:"row"`
	CouponCode string `json:"couponcode"`
	Field      string `json:"field"`
	Error      string `json:"error"`
}

type CouponBatchImportStatus struct {
	BatchID       string    `json:"batchid"`
	Status        string    `json:"status"`   // pending, processing, completed, failed
	Progress      int       `json:"progress"` // 0-100
	TotalRows     int       `json:"totalrows"`
	ProcessedRows int       `json:"processedrows"`
	SuccessCount  int       `json:"successcount"`
	ErrorCount    int       `json:"errorcount"`
	Message       string    `json:"message"`
	UpdatedAt     time.Time `json:"updatedat"`
}

type CouponBatchImportResponse struct {
	BatchID       string    `json:"batchid"`
	Status        string    `json:"status"`   // pending, processing, completed, failed
	Progress      int       `json:"progress"` // 0-100
	TotalRows     int       `json:"totalrows"`
	ProcessedRows int       `json:"processedrows"`
	SuccessCount  int       `json:"successcount"`
	ErrorCount    int       `json:"errorcount"`
	Message       string    `json:"message"`
	UpdatedAt     time.Time `json:"updatedat"`
}
