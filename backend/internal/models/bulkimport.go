package models

type BulkImport struct {
	Created          []string `json:"created"`
	Updated          []string `json:"updated"`
	UpdateFailed     []string `json:"updatefailed"`
	PayloadDuplicate []string `json:"payloadduplicate"`
}

type BulkResponse struct {
	Success bool `json:"success"`
	BulkImport
}

type BulkPreviewData struct {
	WillCreate       []string `json:"willcreate"`       // CouponCodes ที่จะสร้างใหม่
	WillUpdate       []string `json:"willupdate"`       // CouponCodes ที่จะอัพเดท
	PayloadDuplicate []string `json:"payloadduplicate"` // CouponCodes ที่ซ้ำใน payload
	ValidationErrors []string `json:"validationerrors"` // ข้อผิดพลาดในการ validate
}

type BulkPreviewResponse struct {
	Success      bool            `json:"success"`
	PreviewData  BulkPreviewData `json:"previewdata"`
	CanProceed   bool            `json:"canproceed"`   // สามารถดำเนินการนำเข้าได้หรือไม่
	ErrorCount   int             `json:"errorcount"`   // จำนวนข้อผิดพลาด
	TotalRecords int             `json:"totalrecords"` // จำนวนรายการทั้งหมด
}
