package validators

import "fmt"

type BilingualMessage struct {
	TH string
	EN string
}

func (m BilingualMessage) Get(lang string) string {
	if lang == "th" {
		return m.TH
	}
	return m.EN
}

type ValidationError struct {
	Field   string
	Code    string
	Message BilingualMessage
}

type ValidationErrorItem struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ValidationResult struct {
	Errors []ValidationError
}

func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Errors: []ValidationError{},
	}
}

func (r *ValidationResult) IsValid() bool {
	return len(r.Errors) == 0
}

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

func (r *ValidationResult) AddErrorf(field, code, thFormat, enFormat string, args ...interface{}) {
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

type ValidationErrorResponse struct {
	Success    bool                  `json:"success"`
	ErrorCode  string                `json:"errorcode"`
	Message    string                `json:"message"`
	Errors     []ValidationErrorItem `json:"errors"`
	ErrorCount int                   `json:"errorcount"`
}

func (r *ValidationResult) ToErrorResponse(lang string) ValidationErrorResponse {
	summary := BilingualMessage{
		TH: "ข้อมูลใบสืบราคาไม่ถูกต้อง กรุณาตรวจสอบและแก้ไข",
		EN: "Request for quotation validation failed. Please review and correct the errors.",
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
		ErrorCode:  "RFQ_VALIDATION_FAILED",
		Message:    summary.Get(lang),
		Errors:     errors,
		ErrorCount: len(r.Errors),
	}
}

func (r *ValidationResult) ToUpdateErrorResponse(lang string) ValidationErrorResponse {
	resp := r.ToErrorResponse(lang)
	resp.ErrorCode = "RFQ_UPDATE_VALIDATION_FAILED"
	updateSummary := BilingualMessage{
		TH: "ไม่สามารถแก้ไขใบสืบราคาได้ กรุณาตรวจสอบและแก้ไข",
		EN: "Cannot update request for quotation. Please review and correct the errors.",
	}
	resp.Message = updateSummary.Get(lang)
	return resp
}

type RFQSuccessResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	DocNo   string `json:"docno,omitempty"`
	Message string `json:"message"`
}

func NewCreateSuccessResponse(id, docNo, lang string) RFQSuccessResponse {
	msg := BilingualMessage{
		TH: "บันทึกใบสืบราคาสำเร็จ",
		EN: "Request for quotation saved successfully.",
	}
	return RFQSuccessResponse{
		Success: true,
		ID:      id,
		DocNo:   docNo,
		Message: msg.Get(lang),
	}
}

func NewUpdateSuccessResponse(id, lang string) RFQSuccessResponse {
	msg := BilingualMessage{
		TH: "แก้ไขใบสืบราคาสำเร็จ",
		EN: "Request for quotation updated successfully.",
	}
	return RFQSuccessResponse{
		Success: true,
		ID:      id,
		Message: msg.Get(lang),
	}
}
