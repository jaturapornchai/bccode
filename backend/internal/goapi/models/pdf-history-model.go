package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PdfHistory - เก็บประวัติการพิมพ์ PDF
type PdfHistory struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	HoldingCode  string             `bson:"holdingcode" json:"holdingcode"`
	Collection   string             `bson:"collection" json:"collection"`     // collection ที่ดึงข้อมูล
	DocNo        string             `bson:"docno" json:"docno"`               // เลขที่เอกสาร
	DocDate      time.Time          `bson:"docdate" json:"docdate"`           // วันที่เอกสาร
	Title        string             `bson:"title" json:"title"`               // ชื่อเอกสาร
	FileName     string             `bson:"filename" json:"filename"`         // ชื่อไฟล์ PDF
	R2Key        string             `bson:"r2key" json:"-"`                   // path ใน R2 (ไม่ส่งออก frontend)
	FileSize     int64              `bson:"filesize" json:"filesize"`         // ขนาดไฟล์ (bytes)
	PageSize     string             `bson:"pagesize" json:"pagesize"`         // A4, A3, etc.
	Orientation  string             `bson:"orientation" json:"orientation"`   // P, L
	Language     string             `bson:"language" json:"language"`         // th, en, etc.
	ThemeName    string             `bson:"themename" json:"themename"`       // theme ที่ใช้
	TemplateID   string             `bson:"templateid" json:"templateid"`     // template ที่ใช้
	PrintedBy    string             `bson:"printedby" json:"printedby"`       // ผู้พิมพ์ (optional)
	PrintedAt    time.Time          `bson:"printedat" json:"printedat"`       // วันเวลาที่พิมพ์
	ReprintCount int                `bson:"reprintcount" json:"reprintcount"` // จำนวนครั้งที่ reprint
	CreatedAt    time.Time          `bson:"createdat" json:"createdat"`

	// ข้อมูลเอกสาร
	CustomerName string  `bson:"customername" json:"customername"` // ชื่อลูกหนี้/ลูกค้า (AR)
	VendorName   string  `bson:"vendorname" json:"vendorname"`     // ชื่อเจ้าหนี้/ผู้ขาย (AP)
	TotalAmount  float64 `bson:"totalamount" json:"totalamount"`   // ยอดรวม
}

// PdfHistoryListRequest - request สำหรับดึงรายการประวัติ PDF
type PdfHistoryListRequest struct {
	HoldingCode string `json:"holdingcode"`
	Collection  string `json:"collection"` // filter by collection (optional)
	DocNo       string `json:"docno"`      // filter by docno (optional)
	Limit       int64  `json:"limit"`
	Skip        int64  `json:"skip"`
}

// PdfHistoryListResponse - response สำหรับรายการประวัติ PDF
type PdfHistoryListResponse struct {
	Status string               `json:"status"`
	Code   int                  `json:"code"`
	Count  int                  `json:"count"`
	Total  int64                `json:"total"`
	Data   []PdfHistoryListItem `json:"data"`
}

// PdfHistoryListItem - item ในรายการประวัติ PDF พร้อม URL
type PdfHistoryListItem struct {
	ID           primitive.ObjectID `json:"id"`
	HoldingCode  string             `json:"holdingcode"`
	Collection   string             `json:"collection"`
	DocNo        string             `json:"docno"`
	Title        string             `json:"title"`
	FileName     string             `json:"filename"`
	FileSize     int64              `json:"filesize"`
	PageSize     string             `json:"pagesize"`
	Orientation  string             `json:"orientation"`
	Language     string             `json:"language"`
	ThemeName    string             `json:"themename"`
	TemplateID   string             `json:"templateid"`
	PrintedBy    string             `json:"printedby"`
	PrintedAt    time.Time          `json:"printedat"`
	ReprintCount int                `json:"reprintcount"`
	URL          string             `json:"url,omitempty"` // Presigned URL (หมดอายุได้)
	CreatedAt    time.Time          `json:"createdat"`
}
