package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AttachmentMetadata - ข้อมูลไฟล์แนบเอกสาร (PO, Sale, Purchase, etc.)
type AttachmentMetadata struct {
	ID           primitive.ObjectID `bson:"id,omitempty" json:"id"`
	HoldingCode  string             `bson:"holdingcode" json:"holdingcode"`
	ScreenType   string             `bson:"screentype" json:"screentype"`                       // purchaseorder, sale, purchase, etc.
	DocNo        string             `bson:"docno" json:"docno"`                                 // เลขที่เอกสาร
	GuidFixed    string             `bson:"guidfixed" json:"guidfixed"`                         // GUID ของเอกสาร
	FileName     string             `bson:"filename" json:"filename"`                           // ชื่อไฟล์ใน R2 (timestamp_hash + extension)
	OriginalName string             `bson:"originalname" json:"originalname"`                   // ชื่อไฟล์เดิม
	ContentType  string             `bson:"contenttype" json:"contenttype"`                     // MIME type
	FileType     string             `bson:"filetype" json:"filetype"`                           // pdf, xlsx, jpg, png
	Size         int64              `bson:"size" json:"size"`                                   // ขนาดไฟล์ (bytes)
	R2Key        string             `bson:"r2key" json:"-"`                                     // key ใน R2 bucket (ไม่ส่งออก)
	Description  string             `bson:"description,omitempty" json:"description,omitempty"` // คำอธิบาย
	UploadedBy   string             `bson:"uploadedby" json:"uploadedby"`                       // usercode
	UploadedName string             `bson:"uploadedname" json:"uploadedname"`                   // user_name
	CreatedAt    time.Time          `bson:"createdat" json:"createdat"`
	UpdatedAt    time.Time          `bson:"updatedat" json:"updatedat"`
}

// AttachmentUploadRequest - request สำหรับ upload attachment
// Form data: file, holdingcode, screentype, docno, guidfixed, description, uploadedby, uploadedname
type AttachmentUploadRequest struct {
	File         string `form:"file"` // multipart file
	HoldingCode  string `form:"holdingcode"`
	ScreenType   string `form:"screentype"`
	DocNo        string `form:"docno"`
	GuidFixed    string `form:"guidfixed"`
	Description  string `form:"description"`
	UploadedBy   string `form:"uploadedby"`
	UploadedName string `form:"uploadedname"`
}

// AttachmentListRequest - request สำหรับ list attachments
type AttachmentListRequest struct {
	HoldingCode string `json:"holdingcode"`
	ScreenType  string `json:"screentype,omitempty"`
	DocNo       string `json:"docno,omitempty"`
	GuidFixed   string `json:"guidfixed,omitempty"`
	Limit       int64  `json:"limit,omitempty"`
	Skip        int64  `json:"skip,omitempty"`
}

// AttachmentDeleteRequest - request สำหรับ delete attachment
type AttachmentDeleteRequest struct {
	HoldingCode  string `json:"holdingcode"`
	AttachmentID string `json:"attachmentid,omitempty"`
	FileName     string `json:"filename,omitempty"`
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
	HoldingCode  string             `json:"holdingcode"`
	ScreenType   string             `json:"screentype"`
	DocNo        string             `json:"docno"`
	GuidFixed    string             `json:"guidfixed"`
	FileName     string             `json:"filename"`
	OriginalName string             `json:"originalname"`
	ContentType  string             `json:"contenttype"`
	FileType     string             `json:"filetype"`
	Size         int64              `json:"size"`
	Description  string             `json:"description,omitempty"`
	UploadedBy   string             `json:"uploadedby"`
	UploadedName string             `json:"uploadedname"`
	CreatedAt    time.Time          `json:"createdat"`
	UpdatedAt    time.Time          `json:"updatedat"`
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
