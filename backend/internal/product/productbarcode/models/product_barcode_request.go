package models

import "time"

type ProductBarcodeBranchRequest struct {
	Branch ProductBarcodeBranch `json:"branch"`
	Products []string             `json:"products"`
}

type ProductBarcodeBusinessTypeRequest struct {
	BusinessType ProductBarcodeBusinessType `json:"businesstype"`
	Products []string                   `json:"products"`
}

type BOMVersionRequest struct {
	GuidFixed string       `json:"guid_fixed"`
	StartDate time.Time    `json:"start_date"`
	EndDate   *time.Time   `json:"end_date"`
	BOM       []BOMRequest `json:"bom"`
}

type ProductBarcodeRequest struct {
	ProductBarcodeBase
	RefBarcodes []BarcodeRequest             `json:"refbarcodes"`
	BOM []BOMRequest                 `json:"bom"`
	BOMs []BOMVersionRequest          `json:"boms"`
	IgnoreBranches []ProductBarcodeBranch       `json:"ignorebranches"`
	BusinessTypes []ProductBarcodeBusinessType `json:"businesstypes"`
}

func (p ProductBarcodeRequest) ToProductBarcode() ProductBarcode {
	return ProductBarcode{
		ProductBarcodeBase: p.ProductBarcodeBase,
	}
}

type BarcodeRequest struct {
	Barcode string  `json:"barcode" bson:"barcode" validate:"required,min=1"`
	Condition bool    `json:"condition" bson:"condition"`
	DivideValue float64 `json:"dividevalue" bson:"dividevalue"`
	StandValue float64 `json:"standvalue" bson:"standvalue"`
	Qty float64 `json:"qty" bson:"qty"`
}

type BOMRequest struct {
	Barcode string  `json:"barcode" bson:"barcode" validate:"required,min=1"`
	Condition bool    `json:"condition" bson:"condition"`
	DivideValue float64 `json:"dividevalue" bson:"dividevalue"`
	StandValue float64 `json:"standvalue" bson:"standvalue"`
	Qty float64 `json:"qty" bson:"qty"`
}

type RefBarcodeImportRequest struct {
	Barcode string  `json:"barcode" validate:"required"`
	StandValue float64 `json:"standvalue" validate:"required"`
	DivideValue float64 `json:"dividevalue" validate:"required"`
	BarcodeRef string  `json:"barcoderef" validate:"required"`
}

type RefBarcodeImportResponse struct {
	Success bool                    `json:"success"`
	TotalItems int                     `json:"total_items"`
	Updated int                     `json:"updated"`
	Failed int                     `json:"failed"`
	Errors []RefBarcodeImportError `json:"errors,omitempty"`
}

type RefBarcodeImportError struct {
	Row int    `json:"row"`
	Barcode string `json:"barcode"`
	Error string `json:"error"`
}
