package gentranspdf

import (
	"github.com/jung-kurt/gofpdf"
)

// GenerateSaleInvoicePDF - สร้าง PDF ใบแจ้งหนี้/ใบกำกับภาษี
func GenerateSaleInvoicePDF(document map[string]interface{}, payload GenPDFPayload) (*gofpdf.Fpdf, error) {
	// Set default title for Sale Invoice
	if payload.Title == "เอกสาร" || payload.Title == "" {
		payload.Title = "ใบแจ้งหนี้/ใบกำกับภาษี"
	}
	return GenerateBasePDF(document, payload)
}
