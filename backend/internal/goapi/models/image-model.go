package models

import "time"

// ImageMetadata - ข้อมูลรูปที่อัปโหลดสำเร็จ (คืนให้ผู้เรียกเท่านั้น ไม่บันทึกลงฐานข้อมูล)
// ไฟล์จริงอยู่ใน S3/MinIO: ต้นฉบับที่ ObjectKey และ thumbnail WebP ที่ ThumbnailKey
type ImageMetadata struct {
	ID           string    `json:"id"` // = object key ของไฟล์ต้นฉบับ
	HoldingCode  string    `json:"holdingcode"`
	FileName     string    `json:"filename"`     // ชื่อไฟล์ใน bucket (timestamp + hash + extension)
	OriginalName string    `json:"originalname"` // ชื่อไฟล์เดิม
	ContentType  string    `json:"contenttype"`  // MIME type
	Size         int64     `json:"size"`         // ขนาดไฟล์ (bytes)
	ObjectKey    string    `json:"objectkey"`    // key ของไฟล์ต้นฉบับใน bucket
	URL          string    `json:"url"`          // URI ผ่าน proxy (/s3/file/<key>)
	ThumbnailKey string    `json:"thumbnailkey"` // key ของ thumbnail WebP ใน bucket
	ThumbnailURL string    `json:"thumbnailurl"` // URI ของ thumbnail ผ่าน proxy
	Category     string    `json:"category,omitempty"`
	Description  string    `json:"description,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
	UploadedBy   string    `json:"uploadedby,omitempty"`
	CreatedAt    time.Time `json:"createdat"`
	UpdatedAt    time.Time `json:"updatedat"`
}

// ImageResponse - response เมื่ออัปโหลดสำเร็จ
type ImageResponse struct {
	Status  string         `json:"status"`
	Code    int            `json:"code"`
	Message string         `json:"message,omitempty"`
	Data    *ImageMetadata `json:"data,omitempty"`
}
