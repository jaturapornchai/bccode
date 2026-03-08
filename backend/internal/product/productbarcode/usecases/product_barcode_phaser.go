package usecases

import (
	"encoding/json"
	"smlcloudplatform/internal/product/productbarcode/models"
)

type IProductBarcodePhaser interface {
	PhaseProductBarcodeDoc(doc *models.ProductBarcodeDoc) (*models.ProductBarcodePg, error)
}

type ProductBarcodePhaser struct{}

func (ProductBarcodePhaser) PhaseProductBarcodeDoc(doc *models.ProductBarcodeDoc) (*models.ProductBarcodePg, error) {

	j, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}

	var productBarcode models.ProductBarcode
	err = json.Unmarshal(j, &productBarcode)
	if err != nil {
		return nil, err
	}

	var productBarcodePG models.ProductBarcodePg
	err = json.Unmarshal(j, &productBarcodePG)
	if err != nil {
		return nil, err
	}

	productBarcodePG.UnitCode = doc.ItemUnitCode

	isNormalProductAndHasRefBarcode := productBarcode.ItemType == 0 && productBarcode.RefBarcodes != nil && len(*productBarcode.RefBarcodes) > 0

	if isNormalProductAndHasRefBarcode {
		productBarcodePG.MainBarcodeRef = (*productBarcode.RefBarcodes)[0].Barcode
		productBarcodePG.StandValue = (*productBarcode.RefBarcodes)[0].StandValue
		productBarcodePG.DivideValue = (*productBarcode.RefBarcodes)[0].DivideValue
	} else {
		productBarcodePG.MainBarcodeRef = productBarcode.Barcode
		productBarcodePG.StandValue = 1
		productBarcodePG.DivideValue = 1
	}

	return &productBarcodePG, nil
}
