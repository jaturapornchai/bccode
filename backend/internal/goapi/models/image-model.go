package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ImageMetadata - ข้อมูล metadata ของรูปภาพที่เก็บใน MongoDB
// Note: R2Key เก็บไว้ใน DB แต่ไม่ส่งออกไป frontend (json:"-")
type ImageMetadata struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	HoldingCode  string             `bson:"holdingcode" json:"holdingcode"`
	FileName     string             `bson:"filename" json:"filename"`         // ชื่อไฟล์ใน R2 (hash + extension)
	OriginalName string             `bson:"originalname" json:"originalname"` // ชื่อไฟล์เดิม
	ContentType  string             `bson:"contenttype" json:"contenttype"`   // MIME type
	Size         int64              `bson:"size" json:"size"`                 // ขนาดไฟล์ (bytes)
	R2Key        string             `bson:"r2key" json:"-"`                   // key ใน R2 bucket (ไม่ส่งออก)
	Category     string             `bson:"category,omitempty" json:"category,omitempty"`
	Description  string             `bson:"description,omitempty" json:"description,omitempty"`
	Tags         []string           `bson:"tags,omitempty" json:"tags,omitempty"`
	UploadedBy   string             `bson:"uploadedby,omitempty" json:"uploadedby,omitempty"`
	CreatedAt    time.Time          `bson:"createdat" json:"createdat"`
	UpdatedAt    time.Time          `bson:"updatedat" json:"updatedat"`

	// Slip Verification (Thunder API)
	Verified     bool                   `bson:"verified" json:"verified"`                             // ผ่านการตรวจสอบหรือไม่
	VerifiedAt   *time.Time             `bson:"verifiedat,omitempty" json:"verifiedat,omitempty"`     // เวลาที่ตรวจสอบ
	VerifyStatus string                 `bson:"verifystatus,omitempty" json:"verifystatus,omitempty"` // success, error, not_found
	VerifyError  string                 `bson:"verifyerror,omitempty" json:"verifyerror,omitempty"`   // error message ถ้ามี
	SlipData     map[string]interface{} `bson:"slipdata,omitempty" json:"slipdata,omitempty"`         // ข้อมูล slip จาก Thunder API
	TransRef     string                 `bson:"transref,omitempty" json:"transref,omitempty"`         // Transaction Reference
	TransAmount  float64                `bson:"transamount,omitempty" json:"transamount,omitempty"`   // จำนวนเงิน
	TransDate    string                 `bson:"transdate,omitempty" json:"transdate,omitempty"`       // วันที่ทำรายการ
	SenderName   string                 `bson:"sendername,omitempty" json:"sendername,omitempty"`     // ชื่อผู้โอน
	SenderBank   string                 `bson:"senderbank,omitempty" json:"senderbank,omitempty"`     // ธนาคารผู้โอน
	ReceiverName string                 `bson:"receivername,omitempty" json:"receivername,omitempty"` // ชื่อผู้รับ
	ReceiverBank string                 `bson:"receiverbank,omitempty" json:"receiverbank,omitempty"` // ธนาคารผู้รับ
	IsDuplicate  bool                   `bson:"isduplicate,omitempty" json:"isduplicate,omitempty"`   // เป็น slip ซ้ำหรือไม่
	SlipType     string                 `bson:"sliptype,omitempty" json:"sliptype,omitempty"`         // ประเภท slip: bank หรือ truewallet
}

// ImageUploadRequest - request body สำหรับ upload
type ImageUploadRequest struct {
	HoldingCode string   `form:"holdingcode" json:"holdingcode"`
	Category    string   `form:"category" json:"category"`
	Description string   `form:"description" json:"description"`
	Tags        []string `form:"tags" json:"tags"`
	UploadedBy  string   `form:"uploadedby" json:"uploadedby"`
}

// ImageListRequest - request body สำหรับ list รูปภาพ
type ImageListRequest struct {
	HoldingCode string `json:"holdingcode"`
	Category    string `json:"category,omitempty"`
	Limit       int64  `json:"limit,omitempty"`
	Skip        int64  `json:"skip,omitempty"`
	IncludeData bool   `json:"includedata,omitempty"` // true = รวม base64 data ด้วย
}

// ImageDeleteRequest - request body สำหรับลบรูปภาพ
type ImageDeleteRequest struct {
	HoldingCode string `json:"holdingcode"`
	ImageID     string `json:"imageid"`
	FileName    string `json:"filename"` // สามารถใช้ filename แทน image_id ได้
}

// ImageGetRequest - request body สำหรับดึงรูปภาพ
type ImageGetRequest struct {
	HoldingCode string `json:"holdingcode"`
	FileName    string `json:"filename"`
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
	HoldingCode  string             `json:"holdingcode"`
	FileName     string             `json:"filename"`
	OriginalName string             `json:"originalname"`
	ContentType  string             `json:"contenttype"`
	Size         int64              `json:"size"`
	Category     string             `json:"category,omitempty"`
	Description  string             `json:"description,omitempty"`
	Tags         []string           `json:"tags,omitempty"`
	UploadedBy   string             `json:"uploadedby,omitempty"`
	CreatedAt    time.Time          `json:"createdat"`
	UpdatedAt    time.Time          `json:"updatedat"`
	URL          string             `json:"url,omitempty"`  // Private backend URL
	Data         string             `json:"data,omitempty"` // base64 data URL (deprecated, ใช้ url แทน)

	// Slip Verification fields (ถ้าเคยตรวจสอบแล้ว)
	Verified     *bool    `json:"verified"`               // true = ตรวจสอบแล้ว, null = ยังไม่ตรวจสอบ
	VerifyStatus *string  `json:"verifystatus"`           // "success", "duplicate", "not_found", "error"
	IsDuplicate  *bool    `json:"isduplicate"`            // true = สลิปซ้ำ
	TransRef     *string  `json:"transref,omitempty"`     // หมายเลขอ้างอิงการโอน
	TransAmount  *float64 `json:"transamount,omitempty"`  // จำนวนเงินที่โอน
	SenderName   *string  `json:"sendername,omitempty"`   // ชื่อผู้โอน
	ReceiverName *string  `json:"receivername,omitempty"` // ชื่อผู้รับ
	SlipType     *string  `json:"sliptype,omitempty"`     // bank หรือ truewallet
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
	HoldingCode    string `json:"holdingcode"`
	ImageID        string `json:"imageid"`        // หรือใช้ filename
	FileName       string `json:"filename"`       // สามารถใช้แทน image_id ได้
	CheckDuplicate bool   `json:"checkduplicate"` // ตรวจสอบ slip ซ้ำ
	Type           string `json:"type"`           // "bank" (default) หรือ "truewallet"
}

// ImageVerifyResponse - response สำหรับ verify slip
type ImageVerifyResponse struct {
	Status  string         `json:"status"`
	Code    int            `json:"code"`
	Message string         `json:"message,omitempty"`
	Data    *ImageMetadata `json:"data,omitempty"`
}
