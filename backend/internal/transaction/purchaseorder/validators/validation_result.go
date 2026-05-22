// Package validators — PO Validation Result Structures
//
// ============================================================================
// RESPONSE FORMAT / รูปแบบ JSON Response:
//
// Frontend/AI Agent ส่ง lang parameter มาเพื่อเลือกภาษา:
//   - lang=en (default) → ข้อความภาษาอังกฤษ
//   - lang=th → ข้อความภาษาไทย
//
// Error Response Example (HTTP 400):
//   {
//     "success": false,
//     "error_code": "PO_VALIDATION_FAILED",
//     "message": "Purchase order validation failed. Please review and correct the errors.",
//     "errors": [
//       {
//         "field": "custcode",
//         "code": "REQUIRED",
//         "message": "Supplier code is required. A purchase order must always specify a vendor."
//       }
//     ],
//     "error_count": 1
//   }
//
// FOR AI AGENTS:
//   - Use "code" field for programmatic error handling (e.g. "REQUIRED", "MUST_BE_POSITIVE")
//   - Use "field" for identifying which JSON field has the error
//   - Use "message" for human-readable description
//   - Send "lang" query parameter or header to choose language
// ============================================================================
package validators

import "fmt"

// BilingualMessage — ข้อความภายใน (internal) ที่เก็บทั้งสองภาษา
// Internal bilingual storage — NOT sent directly in JSON response.
// Use Get(lang) to pick one language for the response.
type BilingualMessage struct {
	TH string // ข้อความภาษาไทย
	EN string // English message
}

// Get — เลือกข้อความตามภาษาที่ต้องการ (default: English)
// Returns the message in the requested language. Falls back to English.
func (m BilingualMessage) Get(lang string) string {
	if lang == "th" {
		return m.TH
	}
	return m.EN
}

// ValidationError — ข้อผิดพลาดของแต่ละ field (internal, เก็บทั้งสองภาษา)
// Internal error representation with both languages stored.
type ValidationError struct {
	Field   string           // ชื่อ field เช่น "custcode", "details[2].qty"
	Code    string           // รหัสข้อผิดพลาด เช่น "REQUIRED", "MUST_BE_POSITIVE"
	Message BilingualMessage // ข้อความอธิบายเหตุผล (เก็บทั้ง TH/EN)
}

// ValidationErrorItem — รายการ error ใน JSON response (ภาษาเดียว)
// Single-language error item sent in the JSON response to the client.
//
// Example:
//
//	{"field": "custcode", "code": "REQUIRED", "message": "Supplier code is required..."}
type ValidationErrorItem struct {
	Field string `json:"field"`   // Field path, e.g. "custcode", "details[0].qty"
	Code string `json:"code"`    // Machine-readable error code, e.g. "REQUIRED"
	Message string `json:"message"` // Human-readable message in the selected language
}

// ValidationResult — ผลลัพธ์การตรวจสอบ รวบรวม errors ทั้งหมด
// Collects all validation errors. Empty Errors = valid.
type ValidationResult struct {
	Errors []ValidationError
}

// NewValidationResult — สร้าง ValidationResult ใหม่
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Errors: []ValidationError{},
	}
}

// IsValid — ตรวจสอบว่าผ่านการตรวจสอบหรือไม่ (true = ไม่มี error)
func (r *ValidationResult) IsValid() bool {
	return len(r.Errors) == 0
}

// AddError — เพิ่ม error พร้อมข้อความสองภาษา (internal)
func (r *ValidationResult) AddError(field, code, thMsg, enMsg string) {
	r.Errors = append(r.Errors, ValidationError{
		Field: field,
		Code:  code,
		Message: BilingualMessage{
			TH: thMsg,
			EN: enMsg,
		},
	})
}

// AddErrorf — เพิ่ม error พร้อม format string สองภาษา (internal)
// NOTE: args ถูกใช้กับทั้ง thFormat และ enFormat — ทั้งสอง format ต้องมี placeholders จำนวนเท่ากัน
func (r *ValidationResult) AddErrorf(field, code, thFormat, enFormat string, args ...interface{}) {
	// แบ่ง args ครึ่งหนึ่ง — ครึ่งแรกสำหรับ TH, ครึ่งหลังสำหรับ EN
	half := len(args) / 2
	thArgs := args[:half]
	enArgs := args[half:]

	r.Errors = append(r.Errors, ValidationError{
		Field: field,
		Code:  code,
		Message: BilingualMessage{
			TH: fmt.Sprintf(thFormat, thArgs...),
			EN: fmt.Sprintf(enFormat, enArgs...),
		},
	})
}

// ============================================================================
// JSON RESPONSE STRUCTURES / โครงสร้าง JSON Response
// ============================================================================

// ValidationErrorResponse — JSON response สำหรับ validation error (ภาษาเดียว)
// Sent to the client as HTTP 400 response.
//
// Fields:
//   - success: always false for validation errors
//   - error_code: machine-readable code (e.g. "PO_VALIDATION_FAILED")
//   - message: human-readable summary in the selected language
//   - errors: array of ValidationErrorItem with details per field
//   - error_count: total number of errors found
type ValidationErrorResponse struct {
	Success bool                  `json:"success"`
	ErrorCode string                `json:"error_code"`
	Message string                `json:"message"`
	Errors []ValidationErrorItem `json:"errors"`
	ErrorCount int                   `json:"error_count"`
}

// ToErrorResponse — แปลง ValidationResult เป็น JSON response ตามภาษาที่เลือก
//
// Parameters:
//   - lang: "th" = ภาษาไทย, อื่นๆ = English (default)
//
// Usage:
//
//	result := validators.ValidatePurchaseOrder(&po)
//	if !result.IsValid() {
//	    ctx.Response(http.StatusBadRequest, result.ToErrorResponse(lang))
//	}
func (r *ValidationResult) ToErrorResponse(lang string) ValidationErrorResponse {
	summary := BilingualMessage{
		TH: "ข้อมูลใบสั่งซื้อไม่ถูกต้อง กรุณาตรวจสอบและแก้ไข",
		EN: "Purchase order validation failed. Please review and correct the errors.",
	}

	errors := make([]ValidationErrorItem, len(r.Errors))
	for i, e := range r.Errors {
		errors[i] = ValidationErrorItem{
			Field:   e.Field,
			Code:    e.Code,
			Message: e.Message.Get(lang),
		}
	}

	return ValidationErrorResponse{
		Success:    false,
		ErrorCode:  "PO_VALIDATION_FAILED",
		Message:    summary.Get(lang),
		Errors:     errors,
		ErrorCount: len(r.Errors),
	}
}

// ToUpdateErrorResponse — แปลง ValidationResult เป็น JSON response สำหรับ update
func (r *ValidationResult) ToUpdateErrorResponse(lang string) ValidationErrorResponse {
	resp := r.ToErrorResponse(lang)
	resp.ErrorCode = "PO_UPDATE_VALIDATION_FAILED"

	updateSummary := BilingualMessage{
		TH: "ไม่สามารถแก้ไขใบสั่งซื้อได้ กรุณาตรวจสอบและแก้ไข",
		EN: "Cannot update purchase order. Please review and correct the errors.",
	}
	resp.Message = updateSummary.Get(lang)

	return resp
}

// POSuccessResponse — JSON response สำหรับ PO สร้าง/แก้ไขสำเร็จ
// Sent to the client as HTTP 201 (create) or 200 (update) response.
//
// Example:
//
//	{"success": true, "id": "abc123...", "docno": "PO20260208xxxxx", "message": "Purchase order saved successfully"}
type POSuccessResponse struct {
	Success bool   `json:"success"`
	ID string `json:"id,omitempty"`    // GUID ของเอกสาร
	DocNo string `json:"docno,omitempty"` // เลขที่เอกสาร
	Message string `json:"message"`         // ข้อความแจ้งผลสำเร็จ
}

// NewCreateSuccessResponse — สร้าง success response สำหรับ Create PO
func NewCreateSuccessResponse(id, docNo, lang string) POSuccessResponse {
	msg := BilingualMessage{
		TH: "บันทึกใบสั่งซื้อสำเร็จ",
		EN: "Purchase order saved successfully.",
	}
	return POSuccessResponse{
		Success: true,
		ID:      id,
		DocNo:   docNo,
		Message: msg.Get(lang),
	}
}

// NewUpdateSuccessResponse — สร้าง success response สำหรับ Update PO
func NewUpdateSuccessResponse(id, lang string) POSuccessResponse {
	msg := BilingualMessage{
		TH: "แก้ไขใบสั่งซื้อสำเร็จ",
		EN: "Purchase order updated successfully.",
	}
	return POSuccessResponse{
		Success: true,
		ID:      id,
		Message: msg.Get(lang),
	}
}
