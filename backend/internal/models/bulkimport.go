package models

type BulkImport struct {
	Created []string `json:"created"`
	Updated []string `json:"updated"`
	UpdateFailed []string `json:"update_failed"`
	PayloadDuplicate []string `json:"payload_duplicate"`
}

type BulkResponse struct {
	Success bool `json:"success"`
	BulkImport
}

type BulkPreviewData struct {
	WillCreate []string `json:"will_create"`       // CouponCodes ที่จะสร้างใหม่
	WillUpdate []string `json:"will_update"`       // CouponCodes ที่จะอัพเดท
	PayloadDuplicate []string `json:"payload_duplicate"` // CouponCodes ที่ซ้ำใน payload
	ValidationErrors []string `json:"validation_errors"` // ข้อผิดพลาดในการ validate
}

type BulkPreviewResponse struct {
	Success bool            `json:"success"`
	PreviewData BulkPreviewData `json:"preview_data"`
	CanProceed bool            `json:"can_proceed"`   // สามารถดำเนินการนำเข้าได้หรือไม่
	ErrorCount int             `json:"error_count"`   // จำนวนข้อผิดพลาด
	TotalRecords int             `json:"total_records"` // จำนวนรายการทั้งหมด
}
