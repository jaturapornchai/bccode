package gentranspdf

import (
	"github.com/jung-kurt/gofpdf"
)

// GenerateDocumentPDF - สร้าง PDF ตาม collection ที่ระบุ
func GenerateDocumentPDF(document map[string]interface{}, payload GenPDFPayload) (*gofpdf.Fpdf, error) {
	// Route to specific PDF generator based on collection
	switch payload.Collection {
	case "transactionPurchaseOrder":
		return GeneratePurchaseOrderPDF(document, payload)
	case "transactionSaleInvoice":
		return GenerateSaleInvoicePDF(document, payload)
	default:
		return GenerateDefaultPDF(document, payload)
	}
}
