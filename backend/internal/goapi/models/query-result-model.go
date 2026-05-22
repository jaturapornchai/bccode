package models

// Query Result Models (แปลงจาก Python models.py)

// ===== Request Models =====

// ResultFromQueryRequest - Request model สำหรับ /resultfromquery
type ResultFromQueryRequest struct {
	ShopID string `json:"shopid" binding:"required"` // Shop ID
	Query string `json:"query" binding:"required"`  // SQL SELECT query to execute
	GUID string `json:"guid,omitempty"`            // Optional GUID, will generate if not provided
}

// ResultGetRequest - Request model สำหรับ /resultget
type ResultGetRequest struct {
	ShopID string `json:"shopid" binding:"required"` // Shop ID
	GUID string `json:"guid" binding:"required"`   // GUID of the query result
	Limit int    `json:"limit,omitempty"`           // Number of rows to return (max 1000)
	Offset int    `json:"offset,omitempty"`          // Number of rows to skip
}

// PDFConfig - PDF configuration
type PDFConfig struct {
	Title string `json:"title,omitempty"`       // PDF title (default: "รายงาน")
	Orientation string `json:"orientation,omitempty"` // Page orientation (L=Landscape, P=Portrait)
	PageSize string `json:"page_size,omitempty"`   // Page size (A4, Letter, etc.)
}

// ResultToPDFRequest - Request model สำหรับ /resulttopdf
type ResultToPDFRequest struct {
	ShopID string            `json:"shopid" binding:"required"` // Shop ID
	GUID string            `json:"guid" binding:"required"`   // GUID of the query result
	PDFConfig PDFConfig         `json:"pdf_config,omitempty"`      // PDF configuration
	ColumnOrder []string          `json:"column_order,omitempty"`    // Order of columns in PDF
	ColumnNames map[string]string `json:"column_names,omitempty"`    // Thai names for columns
}

// ===== Response Models =====

// ResultFromQueryResponse - Response model สำหรับ /resultfromquery
type ResultFromQueryResponse struct {
	Status string `json:"status"`  // success or error
	GUID string `json:"guid"`    // GUID of the saved result
	Count int    `json:"count"`   // Number of rows saved
	Message string `json:"message"` // Success/error message
}

// PaginationInfo - Pagination information
type PaginationInfo struct {
	Total int `json:"total"`  // Total number of rows
	Limit int `json:"limit"`  // Rows per page
	Offset int `json:"offset"` // Current offset
	Count int `json:"count"`  // Number of rows in current page
}

// ResultGetResponse - Response model สำหรับ /resultget
type ResultGetResponse struct {
	Status string                   `json:"status"`               // success or error
	Data []map[string]interface{} `json:"data"`                 // Query result data
	Pagination PaginationInfo           `json:"pagination"`           // Pagination information
	Message string                   `json:"message,omitempty"`    // Optional message
}

// ErrorResponse - Error response model
type ErrorResponse struct {
	Status string `json:"status"`           // Always 'error'
	Message string `json:"message"`          // Error message
	Detail string `json:"detail,omitempty"` // Detailed error information
}

// NewErrorResponse creates a new error response
func NewErrorResponse(message string, detail string) ErrorResponse {
	return ErrorResponse{
		Status:  "error",
		Message: message,
		Detail:  detail,
	}
}

// NewSuccessResponse creates a success response for ResultFromQuery
func NewResultFromQueryResponse(guid string, count int, message string) ResultFromQueryResponse {
	return ResultFromQueryResponse{
		Status:  "success",
		GUID:    guid,
		Count:   count,
		Message: message,
	}
}
