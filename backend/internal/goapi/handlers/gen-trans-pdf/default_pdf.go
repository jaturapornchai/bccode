package gentranspdf

import (
	"github.com/jung-kurt/gofpdf"
)

// GenerateDefaultPDF - สร้าง PDF แบบทั่วไป
func GenerateDefaultPDF(document map[string]interface{}, payload GenPDFPayload) (*gofpdf.Fpdf, error) {
	return GenerateBasePDF(document, payload)
}
