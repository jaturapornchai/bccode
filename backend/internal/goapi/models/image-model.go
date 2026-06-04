package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ImageMetadata - ข้อมูล metadata ของรูปภาพที่เก็บใน MongoDB
// Note: R2Key เก็บไว้ใน DB แต่ไม่ส่งออกไป frontend (json:"-")
type ImageMetadata struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	HoldingCode  string             `bson:"holding_code" json:"holding_code"`
	FileName     string             `bson:"file_name" json:"file_name"`         // ชื่อไฟล์ใน R2 (hash + extension)
	OriginalName string             `bson:"original_name" json:"original_name"` // ชื่อไฟล์เดิม
	ContentType  string             `bson:"content_type" json:"content_type"`   // MIME type
	Size         int64              `bson:"size" json:"size"`                   // ขนาดไฟล์ (bytes)
	R2Key        string             `bson:"r2_key" json:"-"`                    // key ใน R2 bucket (ไม่ส่งออก)
	Category     string             `bson:"category,omitempty" json:"category,omitempty"`
	Description  string             `bson:"description,omitempty" json:"description,omitempty"`
	Tags         []string           `bson:"tags,omitempty" json:"tags,omitempty"`
	UploadedBy   string             `bson:"uploaded_by,omitempty" json:"uploaded_by,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`

	// Slip Verification (Thunder API)
	Verified     bool                   `bson:"verified" json:"verified"`                               // ผ่านการตรวจสอบหรือไม่
	VerifiedAt   *time.Time             `bson:"verified_at,omitempty" json:"verified_at,omitempty"`     // เวลาที่ตรวจสอบ
	VerifyStatus string                 `bson:"verify_status,omitempty" json:"verify_status,omitempty"` // success, error, not_found
	VerifyError  string                 `bson:"verify_error,omitempty" json:"verify_error,omitempty"`   // error message ถ้ามี
	SlipData     map[string]interface{} `bson:"slip_data,omitempty" json:"slip_data,omitempty"`         // ข้อมูล slip จาก Thunder API
	TransRef     string                 `bson:"trans_ref,omitempty" json:"trans_ref,omitempty"`         // Transaction Reference
	TransAmount  float64                `bson:"trans_amount,omitempty" json:"trans_amount,omitempty"`   // จำนวนเงิน
	TransDate    string                 `bson:"trans_date,omitempty" json:"trans_date,omitempty"`       // วันที่ทำรายการ
	SenderName   string                 `bson:"sender_name,omitempty" json:"sender_name,omitempty"`     // ชื่อผู้โอน
	SenderBank   string                 `bson:"sender_bank,omitempty" json:"sender_bank,omitempty"`     // ธนาคารผู้โอน
	ReceiverName string                 `bson:"receiver_name,omitempty" json:"receiver_name,omitempty"` // ชื่อผู้รับ
	ReceiverBank string                 `bson:"receiver_bank,omitempty" json:"receiver_bank,omitempty"` // ธนาคารผู้รับ
	IsDuplicate  bool                   `bson:"is_duplicate,omitempty" json:"is_duplicate,omitempty"`   // เป็น slip ซ้ำหรือไม่
	SlipType     string                 `bson:"slip_type,omitempty" json:"slip_type,omitempty"`         // ประเภท slip: bank หรือ truewallet
}

// ImageUploadRequest - request body สำหรับ upload
type ImageUploadRequest struct {
	HoldingCode string   `form:"holding_code" json:"holding_code"`
	Category    string   `form:"category" json:"category"`
	Description string   `form:"description" json:"description"`
	Tags        []string `form:"tags" json:"tags"`
	UploadedBy  string   `form:"uploaded_by" json:"uploaded_by"`
}

// ImageListRequest - request body สำหรับ list รูปภาพ
type ImageListRequest struct {
	HoldingCode string `json:"holding_code"`
	Category    string `json:"category,omitempty"`
	Limit       int64  `json:"limit,omitempty"`
	Skip        int64  `json:"skip,omitempty"`
	IncludeData bool   `json:"include_data,omitempty"` // true = รวม base64 data ด้วย
}

// ImageDeleteRequest - request body สำหรับลบรูปภาพ
type ImageDeleteRequest struct {
	HoldingCode string `json:"holding_code"`
	ImageID     string `json:"image_id"`
	FileName    string `json:"file_name"` // สามารถใช้ filename แทน image_id ได้
}

// ImageGetRequest - request body สำหรับดึงรูปภาพ
type ImageGetRequest struct {
	HoldingCode string `json:"holding_code"`
	FileName    string `json:"file_name"`
}

// ImageResponse - response เมื่ออัปโหลดสำเร็จ
type ImageResponse struct {
	Status  string         `json:"status"`
	Code    int            `json:"code"`
	Message string         `json:"message,omitempty"`
	Data    *ImageMetadata `json:"data,omitempty"`
}

// ImageDataResponse - response พร้อม base64 data
type ImageDataResponse struct {
	Status  string         `json:"status"`
	Code    int            `json:"code"`
	Message string         `json:"message,omitempty"`
	Data    *ImageMetadata `json:"data,omitempty"`
	Base64  string         `json:"base64,omitempty"` // base64 encoded image data
}

// ImageListResponse - response สำหรับ list รูปภาพ
type ImageListResponse struct {
	Status string          `json:"status"`
	Code   int             `json:"code"`
	Count  int             `json:"count"`
	Data   []ImageMetadata `json:"data"`
}

// ImageListDataItem - item ใน list พร้อม URL หรือ base64
// ใช้ fields แยกแทน embedding เพื่อให้ json marshal ทำงานถูกต้อง
type ImageListDataItem struct {
	ID           primitive.ObjectID `json:"id"`
	HoldingCode  string             `json:"holding_code"`
	FileName     string             `json:"file_name"`
	OriginalName string             `json:"original_name"`
	ContentType  string             `json:"content_type"`
	Size         int64              `json:"size"`
	Category     string             `json:"category,omitempty"`
	Description  string             `json:"description,omitempty"`
	Tags         []string           `json:"tags,omitempty"`
	UploadedBy   string             `json:"uploaded_by,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	URL          string             `json:"url,omitempty"`  // Private backend URL
	Data         string             `json:"data,omitempty"` // base64 data URL (deprecated, ใช้ url แทน)

	// Slip Verification fields (ถ้าเคยตรวจสอบแล้ว)
	Verified     *bool    `json:"verified"`                // true = ตรวจสอบแล้ว, null = ยังไม่ตรวจสอบ
	VerifyStatus *string  `json:"verify_status"`           // "success", "duplicate", "not_found", "error"
	IsDuplicate  *bool    `json:"is_duplicate"`            // true = สลิปซ้ำ
	TransRef     *string  `json:"trans_ref,omitempty"`     // หมายเลขอ้างอิงการโอน
	TransAmount  *float64 `json:"trans_amount,omitempty"`  // จำนวนเงินที่โอน
	SenderName   *string  `json:"sender_name,omitempty"`   // ชื่อผู้โอน
	ReceiverName *string  `json:"receiver_name,omitempty"` // ชื่อผู้รับ
	SlipType     *string  `json:"slip_type,omitempty"`     // bank หรือ truewallet
}

// ImageListDataResponse - response สำหรับ list รูปภาพพร้อม base64
type ImageListDataResponse struct {
	Status string              `json:"status"`
	Code   int                 `json:"code"`
	Count  int                 `json:"count"`
	Total  int64               `json:"total"` // จำนวนทั้งหมด (ไม่รวม limit)
	Data   []ImageListDataItem `json:"data"`
}

// ImageListResponseWithTotal - response สำหรับ list รูปภาพพร้อม total
type ImageListResponseWithTotal struct {
	Status string          `json:"status"`
	Code   int             `json:"code"`
	Count  int             `json:"count"`
	Total  int64           `json:"total"` // จำนวนทั้งหมด (ไม่รวม limit)
	Data   []ImageMetadata `json:"data"`
}

// ImageVerifyRequest - request body สำหรับ verify slip
type ImageVerifyRequest struct {
	HoldingCode    string `json:"holding_code"`
	ImageID        string `json:"image_id"`        // หรือใช้ filename
	FileName       string `json:"file_name"`       // สามารถใช้แทน image_id ได้
	CheckDuplicate bool   `json:"check_duplicate"` // ตรวจสอบ slip ซ้ำ
	Type           string `json:"type"`            // "bank" (default) หรือ "truewallet"
}

// ImageVerifyResponse - response สำหรับ verify slip
type ImageVerifyResponse struct {
	Status  string         `json:"status"`
	Code    int            `json:"code"`
	Message string         `json:"message,omitempty"`
	Data    *ImageMetadata `json:"data,omitempty"`
}
