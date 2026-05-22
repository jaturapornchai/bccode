package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AttachmentMetadata - ข้อมูลไฟล์แนบเอกสาร (PO, Sale, Purchase, etc.)
type AttachmentMetadata struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ShopID       string             `bson:"shopid" json:"shopid"`
	ScreenType   string             `bson:"screen_type" json:"screen_type"`                     // purchaseorder, sale, purchase, etc.
	DocNo        string             `bson:"docno" json:"docno"`                                 // เลขที่เอกสาร
	GuidFixed    string             `bson:"guid_fixed" json:"guid_fixed"`                       // GUID ของเอกสาร
	FileName     string             `bson:"file_name" json:"file_name"`                         // ชื่อไฟล์ใน R2 (timestamp_hash + extension)
	OriginalName string             `bson:"original_name" json:"original_name"`                 // ชื่อไฟล์เดิม
	ContentType  string             `bson:"content_type" json:"content_type"`                   // MIME type
	FileType     string             `bson:"file_type" json:"file_type"`                         // pdf, xlsx, jpg, png
	Size         int64              `bson:"size" json:"size"`                                   // ขนาดไฟล์ (bytes)
	R2Key        string             `bson:"r2_key" json:"-"`                                    // key ใน R2 bucket (ไม่ส่งออก)
	Description  string             `bson:"description,omitempty" json:"description,omitempty"` // คำอธิบาย
	UploadedBy   string             `bson:"uploaded_by" json:"uploaded_by"`                     // user_code
	UploadedName string             `bson:"uploaded_name" json:"uploaded_name"`                 // user_name
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

// AttachmentUploadRequest - request สำหรับ upload attachment
// Form data: file, shopid, screen_type, docno, guidfixed, description, uploaded_by, uploaded_name
type AttachmentUploadRequest struct {
	File         string `form:"file"` // multipart file
	ShopID       string `form:"shopid"`
	ScreenType   string `form:"screen_type"`
	DocNo        string `form:"docno"`
	GuidFixed    string `form:"guid_fixed"`
	Description  string `form:"description"`
	UploadedBy   string `form:"uploaded_by"`
	UploadedName string `form:"uploaded_name"`
}

// AttachmentListRequest - request สำหรับ list attachments
type AttachmentListRequest struct {
	ShopID     string `json:"shopid"`
	ScreenType string `json:"screen_type,omitempty"`
	DocNo      string `json:"docno,omitempty"`
	GuidFixed  string `json:"guid_fixed,omitempty"`
	Limit      int64  `json:"limit,omitempty"`
	Skip       int64  `json:"skip,omitempty"`
}

// AttachmentDeleteRequest - request สำหรับ delete attachment
type AttachmentDeleteRequest struct {
	ShopID       string `json:"shopid"`
	AttachmentID string `json:"attachment_id,omitempty"`
	FileName     string `json:"file_name,omitempty"`
}

// AttachmentResponse - response structure
type AttachmentResponse struct {
	Status  string              `json:"status"`
	Code    int                 `json:"code"`
	Message string              `json:"message,omitempty"`
	Data    *AttachmentMetadata `json:"data,omitempty"`
}

// AttachmentListDataItem - attachment item with private backend URL
type AttachmentListDataItem struct {
	ID           primitive.ObjectID `json:"id"`
	ShopID       string             `json:"shopid"`
	ScreenType   string             `json:"screen_type"`
	DocNo        string             `json:"docno"`
	GuidFixed    string             `json:"guid_fixed"`
	FileName     string             `json:"file_name"`
	OriginalName string             `json:"original_name"`
	ContentType  string             `json:"content_type"`
	FileType     string             `json:"file_type"`
	Size         int64              `json:"size"`
	Description  string             `json:"description,omitempty"`
	UploadedBy   string             `json:"uploaded_by"`
	UploadedName string             `json:"uploaded_name"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	URL          string             `json:"url,omitempty"` // Private backend URL
}

// AttachmentListDataResponse - response สำหรับ list attachments
type AttachmentListDataResponse struct {
	Status string                   `json:"status"`
	Code   int                      `json:"code"`
	Count  int                      `json:"count"`
	Total  int64                    `json:"total"`
	Data   []AttachmentListDataItem `json:"data"`
}
